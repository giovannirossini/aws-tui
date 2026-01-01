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

type APIGatewayState int

const (
	APIGatewayStateMenu APIGatewayState = iota
	APIGatewayStateRestAPIs
	APIGatewayStateHTTPAPIs
)

type apiGatewayItem struct {
	title       string
	description string
	values      []string
}

func (i apiGatewayItem) Title() string       { return i.title }
func (i apiGatewayItem) Description() string { return i.description }
func (i apiGatewayItem) FilterValue() string { return i.title + " " + i.description }

type APIGatewayModel struct {
	client    *aws.APIGatewayClient
	list      list.Model
	styles    Styles
	state     APIGatewayState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type apiGatewayItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  APIGatewayState
}

var apiGatewayRestAPIColumns = []Column{
	{Title: "Name", Width: 0.3},
	{Title: "ID", Width: 0.25},
	{Title: "Endpoint Type", Width: 0.15},
	{Title: "Version", Width: 0.1},
	{Title: "Created", Width: 0.2},
}

var apiGatewayHTTPAPIColumns = []Column{
	{Title: "Name", Width: 0.3},
	{Title: "API ID", Width: 0.25},
	{Title: "Protocol", Width: 0.15},
	{Title: "Endpoint", Width: 0.2},
	{Title: "Created", Width: 0.1},
}

func (d apiGatewayItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(apiGatewayItem)
	if !ok {
		return
	}

	if d.state == APIGatewayStateMenu {
		d.DefaultDelegate.Render(w, m, index, listItem)
		return
	}

	var columns []Column
	switch d.state {
	case APIGatewayStateRestAPIs:
		columns = apiGatewayRestAPIColumns
	case APIGatewayStateHTTPAPIs:
		columns = apiGatewayHTTPAPIColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d apiGatewayItemDelegate) Height() int {
	return 1
}

func NewAPIGatewayModel(profile string, styles Styles, appCache *cache.Cache) APIGatewayModel {
	d := apiGatewayItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           APIGatewayStateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "API Gateway"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	m := APIGatewayModel{
		list:      l,
		styles:    styles,
		state:     APIGatewayStateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
	m.updateDelegate()
	return m
}

func (m *APIGatewayModel) updateDelegate() {
	d := apiGatewayItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          m.styles,
		state:           m.state,
	}
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc
	m.list.SetDelegate(d)
}

type APIGatewayRestAPIsMsg []aws.RestAPIInfo
type APIGatewayHTTPAPIsMsg []aws.HTTPAPIInfo
type APIGatewayErrorMsg error
type APIGatewayMenuMsg []list.Item

func (m APIGatewayModel) Init() tea.Cmd {
	return m.showMenu()
}

func (m APIGatewayModel) showMenu() tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			apiGatewayItem{
				title:       "REST APIs",
				description: "View REST APIs",
				values:      []string{"REST APIs", "View REST APIs"},
			},
			apiGatewayItem{
				title:       "HTTP APIs",
				description: "View HTTP APIs",
				values:      []string{"HTTP APIs", "View HTTP APIs"},
			},
		}
		return APIGatewayMenuMsg(items)
	}
}

func (m APIGatewayModel) fetchRestAPIs() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.APIGatewayResources("rest-apis")); ok {
			if apis, ok := cached.([]aws.RestAPIInfo); ok {
				return APIGatewayRestAPIsMsg(apis)
			}
		}

		client, err := aws.NewAPIGatewayClient(context.Background(), m.profile)
		if err != nil {
			return APIGatewayErrorMsg(err)
		}
		apis, err := client.ListRestAPIs(context.Background())
		if err != nil {
			return APIGatewayErrorMsg(err)
		}

		m.cache.Set(m.cacheKeys.APIGatewayResources("rest-apis"), apis, cache.TTLAPIGatewayResources)
		return APIGatewayRestAPIsMsg(apis)
	}
}

func (m APIGatewayModel) fetchHTTPAPIs() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.APIGatewayResources("http-apis")); ok {
			if apis, ok := cached.([]aws.HTTPAPIInfo); ok {
				return APIGatewayHTTPAPIsMsg(apis)
			}
		}

		client, err := aws.NewAPIGatewayClient(context.Background(), m.profile)
		if err != nil {
			return APIGatewayErrorMsg(err)
		}
		apis, err := client.ListHTTPAPIs(context.Background())
		if err != nil {
			return APIGatewayErrorMsg(err)
		}

		m.cache.Set(m.cacheKeys.APIGatewayResources("http-apis"), apis, cache.TTLAPIGatewayResources)
		return APIGatewayHTTPAPIsMsg(apis)
	}
}

func (m APIGatewayModel) Update(msg tea.Msg) (APIGatewayModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case APIGatewayMenuMsg:
		m.list.SetItems([]list.Item(msg))
		m.list.ResetSelected()
		m.state = APIGatewayStateMenu
		m.updateDelegate()

	case APIGatewayRestAPIsMsg:
		items := make([]list.Item, len(msg))
		for i, api := range msg {
			items[i] = apiGatewayItem{
				title:       api.Name,
				description: api.Description,
				values: []string{
					api.Name,
					api.ID,
					api.EndpointType,
					api.Version,
					api.CreatedDate.Format("2006-01-02 15:04"),
				},
			}
		}
		m.list.SetItems(items)
		m.state = APIGatewayStateRestAPIs
		m.updateDelegate()

	case APIGatewayHTTPAPIsMsg:
		items := make([]list.Item, len(msg))
		for i, api := range msg {
			items[i] = apiGatewayItem{
				title:       api.Name,
				description: api.Description,
				values: []string{
					api.Name,
					api.APIID,
					api.ProtocolType,
					api.APIEndpoint,
					api.CreatedDate.Format("2006-01-02 15:04"),
				},
			}
		}
		m.list.SetItems(items)
		m.state = APIGatewayStateHTTPAPIs
		m.updateDelegate()

	case APIGatewayErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			switch m.state {
			case APIGatewayStateRestAPIs:
				m.cache.Delete(m.cacheKeys.APIGatewayResources("rest-apis"))
				return m, m.fetchRestAPIs()
			case APIGatewayStateHTTPAPIs:
				m.cache.Delete(m.cacheKeys.APIGatewayResources("http-apis"))
				return m, m.fetchHTTPAPIs()
			}
		case "backspace":
			if m.state != APIGatewayStateMenu {
				return m, m.showMenu()
			}
		case "enter":
			if m.state == APIGatewayStateMenu {
				selected := m.list.SelectedItem()
				if item, ok := selected.(apiGatewayItem); ok {
					switch item.title {
					case "REST APIs":
						return m, m.fetchRestAPIs()
					case "HTTP APIs":
						return m, m.fetchHTTPAPIs()
					}
				}
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m APIGatewayModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	var columns []Column
	switch m.state {
	case APIGatewayStateRestAPIs:
		columns = apiGatewayRestAPIColumns
	case APIGatewayStateHTTPAPIs:
		columns = apiGatewayHTTPAPIColumns
	}

	if len(columns) > 0 {
		_, header := RenderTableHelpers(m.list, m.styles, columns)
		return header + "\n" + m.list.View()
	}

	return m.list.View()
}

func (m *APIGatewayModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
