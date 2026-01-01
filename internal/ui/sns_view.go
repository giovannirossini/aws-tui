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

type SNSState int

const (
	SNSStateTopics SNSState = iota
)

type snsItem struct {
	title       string
	description string
	arn         string
	values      []string
}

func (i snsItem) Title() string       { return i.title }
func (i snsItem) Description() string { return i.description }
func (i snsItem) FilterValue() string { return i.title + " " + i.description + " " + i.arn }

type SNSModel struct {
	client    *aws.SNSClient
	list      list.Model
	styles    Styles
	state     SNSState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type snsItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  SNSState
}

var snsTopicColumns = []Column{
	{Title: "Topic Name", Width: 0.4},
	{Title: "Type", Width: 0.1},
	{Title: "Confirmed Subscriptions", Width: 0.25},
	{Title: "Pending Subscriptions", Width: 0.25},
}

func (d snsItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(snsItem)
	if !ok {
		return
	}

	var columns []Column
	switch d.state {
	case SNSStateTopics:
		columns = snsTopicColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d snsItemDelegate) Height() int {
	return 1
}

func NewSNSModel(profile string, styles Styles, appCache *cache.Cache) SNSModel {
	d := snsItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           SNSStateTopics,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "SNS Topics"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return SNSModel{
		list:      l,
		styles:    styles,
		state:     SNSStateTopics,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type SNSTopicsMsg []aws.TopicInfo
type SNSErrorMsg error

func (m SNSModel) Init() tea.Cmd {
	return m.fetchTopics()
}

func (m SNSModel) fetchTopics() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.SNSResources("topics")); ok {
			if topics, ok := cached.([]aws.TopicInfo); ok {
				return SNSTopicsMsg(topics)
			}
		}

		client, err := aws.NewSNSClient(context.Background(), m.profile)
		if err != nil {
			return SNSErrorMsg(err)
		}
		topics, err := client.ListTopics(context.Background())
		if err != nil {
			return SNSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.SNSResources("topics"), topics, cache.TTLSNSResources)
		return SNSTopicsMsg(topics)
	}
}

func (m SNSModel) Update(msg tea.Msg) (SNSModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case SNSTopicsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = snsItem{
				title:       v.Name,
				description: v.ARN,
				arn:         v.ARN,
				values:      []string{v.Name, v.Type, v.SubscriptionsConfirmed, v.SubscriptionsPending},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = SNSStateTopics

	case SNSErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.cache.Delete(m.cacheKeys.SNSResources("topics"))
			return m, m.fetchTopics()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SNSModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	var columns []Column
	switch m.state {
	case SNSStateTopics:
		columns = snsTopicColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header + "\n" + m.list.View()
}

func (m *SNSModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
