package ui

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
)

type WAFState int

const (
	WAFStateMenu WAFState = iota
	WAFStateWebACLs
	WAFStateIPSets
)

type wafItem struct {
	title       string
	description string
	id          string
	arn         string
}

func (i wafItem) Title() string       { return i.title }
func (i wafItem) Description() string { return i.description }
func (i wafItem) FilterValue() string { return i.title }

type WAFModel struct {
	list          list.Model
	styles        Styles
	state         WAFState
	scope         types.Scope
	profile       string
	region        string
	width         int
	height        int
	err           error
	cache         *cache.Cache
	cacheKeys     *cache.KeyBuilder
}

type wafItemDelegate struct {
	list.DefaultDelegate
	styles Styles
}

var wafWebACLColumns = []Column{
	{Title: "Name", Width: 0.4},
	{Title: "ID", Width: 0.3},
	{Title: "Description", Width: 0.3},
}

func (d wafItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(wafItem)
	if !ok {
		return
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, wafWebACLColumns)
	isSelected := index == m.Index()

	values := []string{
		i.title,
		i.id,
		i.description,
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d wafItemDelegate) Height() int { return 1 }

func NewWAFModel(profile string, styles Styles, appCache *cache.Cache, region string) WAFModel {
	d := wafItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "WAFv2"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(styles.Primary).PaddingLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(2)

	return WAFModel{
		list:      l,
		styles:    styles,
		state:     WAFStateMenu,
		profile:   profile,
		region:    region,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type WAFWebACLsMsg []aws.WebACLInfo
type WAFIPSetsMsg []aws.IPSetInfo
type WAFErrorMsg error
type WAFMenuMsg []list.Item

func (m WAFModel) Init() tea.Cmd {
	return m.showMenu()
}

func (m WAFModel) showMenu() tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			wafItem{title: "CloudFront Web ACLs", description: "Global scope (us-east-1)"},
			wafItem{title: "Regional Web ACLs", description: fmt.Sprintf("Regional scope (%s)", m.region)},
			wafItem{title: "CloudFront IP Sets", description: "Global scope (us-east-1)"},
			wafItem{title: "Regional IP Sets", description: fmt.Sprintf("Regional scope (%s)", m.region)},
		}
		return WAFMenuMsg(items)
	}
}

func (m WAFModel) fetchWebACLs() tea.Cmd {
	return func() tea.Msg {
		region := m.region
		if m.scope == types.ScopeCloudfront {
			region = "us-east-1"
		}

		cacheKey := m.cacheKeys.WAFResources("webacls", string(m.scope))
		if cached, ok := m.cache.Get(cacheKey); ok {
			if acls, ok := cached.([]aws.WebACLInfo); ok {
				return WAFWebACLsMsg(acls)
			}
		}

		client, err := aws.NewWAFClient(context.Background(), m.profile, region)
		if err != nil {
			return WAFErrorMsg(err)
		}
		acls, err := client.ListWebACLs(context.Background(), m.scope)
		if err != nil {
			return WAFErrorMsg(err)
		}

		m.cache.Set(cacheKey, acls, cache.TTLWAFResources)
		return WAFWebACLsMsg(acls)
	}
}

func (m WAFModel) fetchIPSets() tea.Cmd {
	return func() tea.Msg {
		region := m.region
		if m.scope == types.ScopeCloudfront {
			region = "us-east-1"
		}

		cacheKey := m.cacheKeys.WAFResources("ipsets", string(m.scope))
		if cached, ok := m.cache.Get(cacheKey); ok {
			if ipSets, ok := cached.([]aws.IPSetInfo); ok {
				return WAFIPSetsMsg(ipSets)
			}
		}

		client, err := aws.NewWAFClient(context.Background(), m.profile, region)
		if err != nil {
			return WAFErrorMsg(err)
		}
		ipSets, err := client.ListIPSets(context.Background(), m.scope)
		if err != nil {
			return WAFErrorMsg(err)
		}

		m.cache.Set(cacheKey, ipSets, cache.TTLWAFResources)
		return WAFIPSetsMsg(ipSets)
	}
}

func (m WAFModel) Update(msg tea.Msg) (WAFModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case WAFMenuMsg:
		m.list.SetItems(msg)
		m.state = WAFStateMenu

	case WAFWebACLsMsg:
		items := make([]list.Item, len(msg))
		for i, acl := range msg {
			items[i] = wafItem{
				title:       acl.Name,
				description: acl.Description,
				id:          acl.ID,
				arn:         acl.ARN,
			}
		}
		m.list.SetItems(items)
		m.state = WAFStateWebACLs

	case WAFIPSetsMsg:
		items := make([]list.Item, len(msg))
		for i, ipSet := range msg {
			items[i] = wafItem{
				title:       ipSet.Name,
				description: ipSet.Description,
				id:          ipSet.ID,
				arn:         ipSet.ARN,
			}
		}
		m.list.SetItems(items)
		m.state = WAFStateIPSets

	case WAFErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			if m.state == WAFStateWebACLs {
				m.cache.Delete(m.cacheKeys.WAFResources("webacls", string(m.scope)))
				return m, m.fetchWebACLs()
			} else if m.state == WAFStateIPSets {
				m.cache.Delete(m.cacheKeys.WAFResources("ipsets", string(m.scope)))
				return m, m.fetchIPSets()
			}
		case "enter":
			if m.state == WAFStateMenu {
				selected := m.list.SelectedItem().(wafItem).title
				switch selected {
				case "CloudFront Web ACLs":
					m.scope = types.ScopeCloudfront
					return m, m.fetchWebACLs()
				case "Regional Web ACLs":
					m.scope = types.ScopeRegional
					return m, m.fetchWebACLs()
				case "CloudFront IP Sets":
					m.scope = types.ScopeCloudfront
					return m, m.fetchIPSets()
				case "Regional IP Sets":
					m.scope = types.ScopeRegional
					return m, m.fetchIPSets()
				}
			}
		case "esc", "backspace":
			if m.state != WAFStateMenu {
				return m, m.showMenu()
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m WAFModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state == WAFStateMenu {
		return m.list.View()
	}

	_, header := RenderTableHelpers(m.list, m.styles, wafWebACLColumns)
	return header + "\n" + m.list.View()
}

func (m *WAFModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
