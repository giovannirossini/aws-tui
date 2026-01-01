package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

type DynamoDBState int

const (
	DynamoDBStateTables DynamoDBState = iota
)

type dynamoItem struct {
	title       string
	description string
}

func (i dynamoItem) Title() string       { return i.title }
func (i dynamoItem) Description() string { return i.description }
func (i dynamoItem) FilterValue() string { return i.title }

type DynamoDBModel struct {
	client    *aws.DynamoDBClient
	list      list.Model
	styles    Styles
	state     DynamoDBState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type dynamoItemDelegate struct {
	list.DefaultDelegate
	styles Styles
}

var dynamoTableColumns = []Column{
	{Title: "Table Name", Width: 0.3},
	{Title: "Status", Width: 0.15},
	{Title: "Items", Width: 0.15},
	{Title: "Size", Width: 0.15},
	{Title: "Partition Key", Width: 0.25},
}

func (d dynamoItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(dynamoItem)
	if !ok {
		return
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, dynamoTableColumns)
	isSelected := index == m.Index()

	parts := strings.Split(i.description, " | ")
	status := ""
	items := ""
	size := ""
	pk := ""
	if len(parts) >= 4 {
		status = strings.TrimPrefix(parts[0], "Status: ")
		items = strings.TrimPrefix(parts[1], "Items: ")
		size = strings.TrimPrefix(parts[2], "Size: ")
		pk = strings.TrimPrefix(parts[3], "PK: ")
	}

	values := []string{
		"📊 " + i.title,
		status,
		items,
		size,
		pk,
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d dynamoItemDelegate) Height() int { return 1 }

func NewDynamoDBModel(profile string, styles Styles, appCache *cache.Cache) DynamoDBModel {
	d := dynamoItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "DynamoDB Tables"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(styles.Primary).PaddingLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(2)

	return DynamoDBModel{
		list:      l,
		styles:    styles,
		state:     DynamoDBStateTables,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type DynamoTablesMsg []aws.DynamoTableInfo
type DynamoErrorMsg error

func (m DynamoDBModel) Init() tea.Cmd {
	return m.fetchTables()
}

func (m DynamoDBModel) fetchTables() tea.Cmd {
	return func() tea.Msg {
		cacheKey := m.cacheKeys.DynamoDBResources("tables")
		if cached, ok := m.cache.Get(cacheKey); ok {
			if tables, ok := cached.([]aws.DynamoTableInfo); ok {
				return DynamoTablesMsg(tables)
			}
		}

		client, err := aws.NewDynamoDBClient(context.Background(), m.profile)
		if err != nil {
			return DynamoErrorMsg(err)
		}
		tables, err := client.ListTables(context.Background())
		if err != nil {
			return DynamoErrorMsg(err)
		}

		m.cache.Set(cacheKey, tables, cache.TTLDynamoDBResources)
		return DynamoTablesMsg(tables)
	}
}

func (m DynamoDBModel) Update(msg tea.Msg) (DynamoDBModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case DynamoTablesMsg:
		items := make([]list.Item, len(msg))
		for i, t := range msg {
			sizeMB := float64(t.TableSize) / 1024 / 1024
			items[i] = dynamoItem{
				title:       t.Name,
				description: fmt.Sprintf("Status: %s | Items: %d | Size: %.2f MB | PK: %s", t.Status, t.ItemCount, sizeMB, t.PartitionKey),
			}
		}
		m.list.SetItems(items)
		m.state = DynamoDBStateTables
		m.list.Title = "DynamoDB Tables"

	case DynamoErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.cache.Delete(m.cacheKeys.DynamoDBResources("tables"))
			return m, m.fetchTables()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m DynamoDBModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	return m.renderHeader() + "\n" + m.list.View()
}

func (m DynamoDBModel) renderHeader() string {
	_, header := RenderTableHelpers(m.list, m.styles, dynamoTableColumns)
	return header
}

func (m *DynamoDBModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
