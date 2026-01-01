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

type TransferState int

const (
	TransferStateServers TransferState = iota
	TransferStateUsers
)

type transferItem struct {
	title       string
	description string
	serverId    string
	isUser      bool
}

func (i transferItem) Title() string       { return i.title }
func (i transferItem) Description() string { return i.description }
func (i transferItem) FilterValue() string { return i.title }

type TransferModel struct {
	client        *aws.TransferClient
	list          list.Model
	styles        Styles
	state         TransferState
	currentServer string
	width         int
	height        int
	profile       string
	err           error
	cache         *cache.Cache
	cacheKeys     *cache.KeyBuilder
}

type transferItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  TransferState
}

var transferServerColumns = []Column{
	{Title: "Server ID", Width: 0.25},
	{Title: "State", Width: 0.15},
	{Title: "Endpoint", Width: 0.2},
	{Title: "Identity Provider", Width: 0.25},
	{Title: "Users", Width: 0.15},
}

var transferUserColumns = []Column{
	{Title: "User Name", Width: 0.25},
	{Title: "Role", Width: 0.35},
	{Title: "Home Directory", Width: 0.25},
	{Title: "SSH Keys", Width: 0.15},
}

func (d transferItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(transferItem)
	if !ok {
		return
	}

	var columns []Column
	if d.state == TransferStateServers {
		columns = transferServerColumns
	} else {
		columns = transferUserColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	var values []string
	if d.state == TransferStateServers {
		parts := strings.Split(i.description, " | ")
		state := ""
		endpoint := ""
		idp := ""
		users := ""
		if len(parts) >= 4 {
			state = strings.TrimPrefix(parts[0], "State: ")
			endpoint = strings.TrimPrefix(parts[1], "Endpoint: ")
			idp = strings.TrimPrefix(parts[2], "IDP: ")
			users = strings.TrimPrefix(parts[3], "Users: ")
		}
		values = []string{
			"🖥️ " + i.title,
			state,
			endpoint,
			idp,
			users,
		}
	} else {
		if i.title == ".." {
			values = []string{"..", "", "", ""}
		} else {
			parts := strings.Split(i.description, " | ")
			role := ""
			home := ""
			keys := ""
			if len(parts) >= 3 {
				role = strings.TrimPrefix(parts[0], "Role: ")
				home = strings.TrimPrefix(parts[1], "Home: ")
				keys = strings.TrimPrefix(parts[2], "SSH Keys: ")
			}
			values = []string{
				"👤 " + i.title,
				role,
				home,
				keys,
			}
		}
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d transferItemDelegate) Height() int { return 1 }

func NewTransferModel(profile string, styles Styles, appCache *cache.Cache) TransferModel {
	d := transferItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           TransferStateServers,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "AWS Transfer Servers"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(styles.Primary).PaddingLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(2)

	return TransferModel{
		list:      l,
		styles:    styles,
		state:     TransferStateServers,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type TransferServersMsg []aws.TransferServerInfo
type TransferUsersMsg []aws.TransferUserInfo
type TransferErrorMsg error

func (m TransferModel) Init() tea.Cmd {
	return m.fetchServers()
}

func (m TransferModel) fetchServers() tea.Cmd {
	return func() tea.Msg {
		cacheKey := m.cacheKeys.TransferResources("servers")
		if cached, ok := m.cache.Get(cacheKey); ok {
			if servers, ok := cached.([]aws.TransferServerInfo); ok {
				return TransferServersMsg(servers)
			}
		}

		client, err := aws.NewTransferClient(context.Background(), m.profile)
		if err != nil {
			return TransferErrorMsg(err)
		}
		servers, err := client.ListServers(context.Background())
		if err != nil {
			return TransferErrorMsg(err)
		}

		m.cache.Set(cacheKey, servers, cache.TTLTransferResources)
		return TransferServersMsg(servers)
	}
}

func (m TransferModel) fetchUsers() tea.Cmd {
	return func() tea.Msg {
		cacheKey := m.cacheKeys.TransferUsers(m.currentServer)
		if cached, ok := m.cache.Get(cacheKey); ok {
			if users, ok := cached.([]aws.TransferUserInfo); ok {
				return TransferUsersMsg(users)
			}
		}

		client, err := aws.NewTransferClient(context.Background(), m.profile)
		if err != nil {
			return TransferErrorMsg(err)
		}
		users, err := client.ListUsers(context.Background(), m.currentServer)
		if err != nil {
			return TransferErrorMsg(err)
		}

		m.cache.Set(cacheKey, users, cache.TTLTransferResources)
		return TransferUsersMsg(users)
	}
}

func (m TransferModel) Update(msg tea.Msg) (TransferModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case TransferServersMsg:
		items := make([]list.Item, len(msg))
		for i, s := range msg {
			items[i] = transferItem{
				title:       s.ServerId,
				description: fmt.Sprintf("State: %s | Endpoint: %s | IDP: %s | Users: %d", s.State, s.EndpointType, s.IdentityProviderType, s.UserCount),
				serverId:    s.ServerId,
			}
		}
		m.list.SetItems(items)
		m.state = TransferStateServers
		m.list.Title = "AWS Transfer Servers"

		d := transferItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           TransferStateServers,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case TransferUsersMsg:
		items := make([]list.Item, 0)
		items = append(items, transferItem{title: "..", description: "Back"})

		for _, u := range msg {
			items = append(items, transferItem{
				title:       u.UserName,
				description: fmt.Sprintf("Role: %s | Home: %s | SSH Keys: %d", u.Role, u.HomeDirectory, u.SshPublicKeyCount),
				isUser:      true,
			})
		}
		m.list.SetItems(items)
		m.state = TransferStateUsers
		m.list.Title = fmt.Sprintf("Transfer Users: %s", m.currentServer)

		d := transferItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           TransferStateUsers,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case TransferErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			if m.state == TransferStateServers {
				m.cache.Delete(m.cacheKeys.TransferResources("servers"))
				return m, m.fetchServers()
			} else {
				m.cache.Delete(m.cacheKeys.TransferUsers(m.currentServer))
				return m, m.fetchUsers()
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(transferItem); ok {
				if item.title == ".." {
					m.state = TransferStateServers
					return m, m.fetchServers()
				}
				if !item.isUser {
					m.currentServer = item.serverId
					m.state = TransferStateUsers
					return m, m.fetchUsers()
				}
			}
		case "backspace", "esc":
			if m.state == TransferStateUsers {
				m.state = TransferStateServers
				return m, m.fetchServers()
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m TransferModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	return m.renderHeader() + "\n" + m.list.View()
}

func (m TransferModel) renderHeader() string {
	var columns []Column
	if m.state == TransferStateServers {
		columns = transferServerColumns
	} else {
		columns = transferUserColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header
}

func (m *TransferModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
