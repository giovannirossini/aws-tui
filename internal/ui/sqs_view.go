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

type SQSState int

const (
	SQSStateQueues SQSState = iota
)

type sqsItem struct {
	title       string
	description string
	url         string
	values      []string
}

func (i sqsItem) Title() string       { return i.title }
func (i sqsItem) Description() string { return i.description }
func (i sqsItem) FilterValue() string { return i.title + " " + i.description + " " + i.url }

type SQSModel struct {
	client    *aws.SQSClient
	list      list.Model
	styles    Styles
	state     SQSState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type sqsItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  SQSState
}

var sqsQueueColumns = []Column{
	{Title: "Queue Name", Width: 0.5},
	{Title: "Type", Width: 0.1},
	{Title: "Available", Width: 0.1},
	{Title: "Delayed", Width: 0.1},
	{Title: "Not Visible", Width: 0.1},
	{Title: "Timeout (s)", Width: 0.1},
}

func (d sqsItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(sqsItem)
	if !ok {
		return
	}

	var columns []Column
	switch d.state {
	case SQSStateQueues:
		columns = sqsQueueColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d sqsItemDelegate) Height() int {
	return 1
}

func NewSQSModel(profile string, styles Styles, appCache *cache.Cache) SQSModel {
	d := sqsItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           SQSStateQueues,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "SQS Queues"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return SQSModel{
		list:      l,
		styles:    styles,
		state:     SQSStateQueues,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type SQSQueuesMsg []aws.QueueInfo
type SQSErrorMsg error

func (m SQSModel) Init() tea.Cmd {
	return m.fetchQueues()
}

func (m SQSModel) fetchQueues() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.SQSResources("queues")); ok {
			if queues, ok := cached.([]aws.QueueInfo); ok {
				return SQSQueuesMsg(queues)
			}
		}

		client, err := aws.NewSQSClient(context.Background(), m.profile)
		if err != nil {
			return SQSErrorMsg(err)
		}
		queues, err := client.ListQueues(context.Background())
		if err != nil {
			return SQSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.SQSResources("queues"), queues, cache.TTLSQSResources)
		return SQSQueuesMsg(queues)
	}
}

func (m SQSModel) Update(msg tea.Msg) (SQSModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case SQSQueuesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = sqsItem{
				title:       v.Name,
				description: v.URL,
				url:         v.URL,
				values:      []string{v.Name, v.Type, v.MessagesAvailable, v.MessagesDelayed, v.MessagesNotVisible, v.VisibilityTimeout},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = SQSStateQueues

	case SQSErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.cache.Delete(m.cacheKeys.SQSResources("queues"))
			return m, m.fetchQueues()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SQSModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	var columns []Column
	switch m.state {
	case SQSStateQueues:
		columns = sqsQueueColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header + "\n" + m.list.View()
}

func (m *SQSModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
