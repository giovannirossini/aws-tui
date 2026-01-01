package ui

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

type MSKState int

const (
	MSKStateClusters MSKState = iota
)

type mskItem struct {
	title       string
	description string
	id          string
	values      []string
}

func (i mskItem) Title() string       { return i.title }
func (i mskItem) Description() string { return i.description }
func (i mskItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type MSKModel struct {
	client    *aws.MSKClient
	list      list.Model
	styles    Styles
	state     MSKState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type mskItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  MSKState
}

var mskClusterColumns = []Column{
	{Title: "Name", Width: 0.3},
	{Title: "Status", Width: 0.15},
	{Title: "Kafka Version", Width: 0.15},
	{Title: "Nodes", Width: 0.1},
	{Title: "ARN", Width: 0.3},
}

func (d mskItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(mskItem)
	if !ok {
		return
	}

	var columns []Column
	switch d.state {
	case MSKStateClusters:
		columns = mskClusterColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d mskItemDelegate) Height() int {
	return 1
}

func NewMSKModel(profile string, styles Styles, appCache *cache.Cache) MSKModel {
	d := mskItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           MSKStateClusters,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "MSK Clusters"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return MSKModel{
		list:      l,
		styles:    styles,
		state:     MSKStateClusters,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type MSKClustersMsg []aws.ClusterInfo
type MSKErrorMsg error

func (m MSKModel) Init() tea.Cmd {
	return m.fetchClusters()
}

func (m MSKModel) fetchClusters() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.MSKResources("clusters")); ok {
			if clusters, ok := cached.([]aws.ClusterInfo); ok {
				return MSKClustersMsg(clusters)
			}
		}

		client, err := aws.NewMSKClient(context.Background(), m.profile)
		if err != nil {
			return MSKErrorMsg(err)
		}
		clusters, err := client.ListClustersV2(context.Background())
		if err != nil {
			return MSKErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.MSKResources("clusters"), clusters, cache.TTLMSKResources)
		return MSKClustersMsg(clusters)
	}
}

func (m MSKModel) Update(msg tea.Msg) (MSKModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case MSKClustersMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = mskItem{
				title:       v.Name,
				description: v.ARN,
				id:          v.ARN,
				values:      []string{v.Name, v.Status, v.EngineVersion, fmt.Sprintf("%d", v.Nodes), v.ARN},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = MSKStateClusters

	case MSKErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.cache.Delete(m.cacheKeys.MSKResources("clusters"))
			return m, m.fetchClusters()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m MSKModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	var columns []Column
	switch m.state {
	case MSKStateClusters:
		columns = mskClusterColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header + "\n" + m.list.View()
}

func (m *MSKModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
