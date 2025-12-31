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

type LambdaState int

const (
	LambdaStateFunctions LambdaState = iota
)

type lambdaItem struct {
	title       string
	description string
	values      []string
}

func (i lambdaItem) Title() string       { return i.title }
func (i lambdaItem) Description() string { return i.description }
func (i lambdaItem) FilterValue() string { return i.title + " " + i.description }

type LambdaModel struct {
	client    *aws.LambdaClient
	list      list.Model
	styles    Styles
	state     LambdaState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type lambdaItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  LambdaState
}

var lambdaColumns = []Column{
	{Title: "Function Name", Width: 0.4},
	{Title: "Runtime", Width: 0.15},
	{Title: "Memory", Width: 0.1},
	{Title: "Timeout", Width: 0.1},
	{Title: "Last Modified", Width: 0.25},
}

func (d lambdaItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(lambdaItem)
	if !ok {
		return
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, lambdaColumns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d lambdaItemDelegate) Height() int { return 1 }

func NewLambdaModel(profile string, styles Styles, appCache *cache.Cache) LambdaModel {
	d := lambdaItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           LambdaStateFunctions,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Lambda Functions"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return LambdaModel{
		list:      l,
		styles:    styles,
		state:     LambdaStateFunctions,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type LambdaFunctionsMsg []aws.FunctionInfo
type LambdaErrorMsg error

func (m LambdaModel) Init() tea.Cmd {
	return m.fetchFunctions()
}

func (m LambdaModel) fetchFunctions() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.LambdaFunctions()); ok {
			if functions, ok := cached.([]aws.FunctionInfo); ok {
				return LambdaFunctionsMsg(functions)
			}
		}

		client, err := aws.NewLambdaClient(context.Background(), m.profile)
		if err != nil {
			return LambdaErrorMsg(err)
		}
		functions, err := client.ListFunctions(context.Background())
		if err != nil {
			return LambdaErrorMsg(err)
		}

		m.cache.Set(m.cacheKeys.LambdaFunctions(), functions, cache.TTLLambdaFunctions)
		return LambdaFunctionsMsg(functions)
	}
}

func (m LambdaModel) Update(msg tea.Msg) (LambdaModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case LambdaFunctionsMsg:
		items := make([]list.Item, len(msg))
		for i, f := range msg {
			items[i] = lambdaItem{
				title:       f.Name,
				description: f.Description,
				values: []string{
					f.Name,
					f.Runtime,
					fmt.Sprintf("%d MB", f.MemorySize),
					fmt.Sprintf("%d s", f.Timeout),
					f.LastModified.Format("2006-01-02 15:04"),
				},
			}
		}
		m.list.SetItems(items)
		m.state = LambdaStateFunctions

	case LambdaErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.cache.Delete(m.cacheKeys.LambdaFunctions())
			return m, m.fetchFunctions()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m LambdaModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	_, header := RenderTableHelpers(m.list, m.styles, lambdaColumns)
	return header + "\n" + m.list.View()
}

func (m *LambdaModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
