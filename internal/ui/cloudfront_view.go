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

type CFState int

const (
	CFStateMenu CFState = iota
	CFStateDistributions
	CFStateDistroSubMenu
	CFStateOrigins
	CFStateBehaviors
	CFStateInvalidations
	CFStatePolicies
	CFStateFunctions
)

type cfItem struct {
	title       string
	description string
	id          string
	category    string
	values      []string
}

func (i cfItem) Title() string       { return i.title }
func (i cfItem) Description() string { return i.description }
func (i cfItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type CFModel struct {
	client         *aws.CloudFrontClient
	list           list.Model
	styles         Styles
	state          CFState
	width          int
	height         int
	profile        string
	err            error
	cache          *cache.Cache
	cacheKeys      *cache.KeyBuilder
	selectedDistro string
}

type cfItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  CFState
}

var cfDistroColumns = []Column{
	{Title: "ID", Width: 0.35},
	{Title: "Status", Width: 0.15},
	{Title: "Domain Name", Width: 0.25},
	{Title: "Comment", Width: 0.15},
	{Title: "Enabled", Width: 0.1},
}

var cfOriginColumns = []Column{
	{Title: "ID", Width: 0.3},
	{Title: "Domain Name", Width: 0.4},
	{Title: "Path", Width: 0.3},
}

var cfBehaviorColumns = []Column{
	{Title: "Path Pattern", Width: 0.3},
	{Title: "Target Origin", Width: 0.3},
	{Title: "Viewer Protocol", Width: 0.4},
}

var cfInvalidationColumns = []Column{
	{Title: "ID", Width: 0.3},
	{Title: "Status", Width: 0.3},
	{Title: "Created At", Width: 0.4},
}

var cfPolicyColumns = []Column{
	{Title: "ID", Width: 0.3},
	{Title: "Name", Width: 0.5},
	{Title: "Type", Width: 0.2},
}

var cfFunctionColumns = []Column{
	{Title: "Name", Width: 0.4},
	{Title: "Status", Width: 0.3},
	{Title: "Runtime", Width: 0.3},
}

func (d cfItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(cfItem)
	if !ok {
		return
	}

	if d.state == CFStateMenu || d.state == CFStateDistroSubMenu {
		d.DefaultDelegate.Render(w, m, index, listItem)
		return
	}

	var columns []Column
	switch d.state {
	case CFStateDistributions:
		columns = cfDistroColumns
	case CFStateOrigins:
		columns = cfOriginColumns
	case CFStateBehaviors:
		columns = cfBehaviorColumns
	case CFStateInvalidations:
		columns = cfInvalidationColumns
	case CFStatePolicies:
		columns = cfPolicyColumns
	case CFStateFunctions:
		columns = cfFunctionColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d cfItemDelegate) Height() int {
	if d.state == CFStateMenu || d.state == CFStateDistroSubMenu {
		return 2
	}
	return 1
}

func NewCFModel(profile string, styles Styles, appCache *cache.Cache) CFModel {
	d := cfItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           CFStateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "CloudFront"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return CFModel{
		list:      l,
		styles:    styles,
		state:     CFStateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

func (m *CFModel) getDistroDisplayName(id, alias string) string {
	if alias != "" {
		return fmt.Sprintf("%s (%s)", id, alias)
	}
	return id
}

type CFDistributionsMsg []aws.CFDistributionInfo
type CFOriginsMsg []aws.CFOriginInfo
type CFBehaviorsMsg []aws.CFBehaviorInfo
type CFInvalidationsMsg []aws.CFInvalidationInfo
type CFPoliciesMsg []aws.CFPolicyInfo
type CFFunctionsMsg []aws.CFFunctionInfo
type CFErrorMsg error
type CFMenuMsg []list.Item

func (m CFModel) Init() tea.Cmd {
	return m.showMenu()
}

func (m CFModel) showMenu() tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			cfItem{title: "Distributions", description: "CloudFront Content Delivery Networks", category: "menu"},
			cfItem{title: "Policies", description: "Response Headers Policies", category: "menu"},
			cfItem{title: "Functions", description: "CloudFront Edge Functions", category: "menu"},
		}
		return CFMenuMsg(items)
	}
}

func (m CFModel) showDistroSubMenu(distroID string) tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			cfItem{title: "Origins", description: "Distribution Origins", id: distroID, category: "submenu"},
			cfItem{title: "Behaviors", description: "Distribution Cache Behaviors", id: distroID, category: "submenu"},
			cfItem{title: "Invalidations", description: "Cache Invalidation Requests", id: distroID, category: "submenu"},
		}
		return CFMenuMsg(items)
	}
}

func (m CFModel) fetchDistributions() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.CFResources("distributions")); ok {
			if distros, ok := cached.([]aws.CFDistributionInfo); ok {
				return CFDistributionsMsg(distros)
			}
		}

		client, err := aws.NewCloudFrontClient(context.Background(), m.profile)
		if err != nil {
			return CFErrorMsg(err)
		}
		distros, err := client.ListDistributions(context.Background())
		if err != nil {
			return CFErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.CFResources("distributions"), distros, cache.TTLCFResources)
		return CFDistributionsMsg(distros)
	}
}

func (m CFModel) fetchDistroDetails(distroID string, resourceType string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewCloudFrontClient(context.Background(), m.profile)
		if err != nil {
			return CFErrorMsg(err)
		}

		switch resourceType {
		case "origins", "behaviors":
			origins, behaviors, err := client.GetDistributionDetails(context.Background(), distroID)
			if err != nil {
				return CFErrorMsg(err)
			}
			if resourceType == "origins" {
				return CFOriginsMsg(origins)
			} else {
				return CFBehaviorsMsg(behaviors)
			}
		case "invalidations":
			invalidations, err := client.ListInvalidations(context.Background(), distroID)
			if err != nil {
				return CFErrorMsg(err)
			}
			return CFInvalidationsMsg(invalidations)
		}
		return nil
	}
}

func (m CFModel) fetchPolicies() tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewCloudFrontClient(context.Background(), m.profile)
		if err != nil {
			return CFErrorMsg(err)
		}
		policies, err := client.ListResponseHeadersPolicies(context.Background())
		if err != nil {
			return CFErrorMsg(err)
		}
		return CFPoliciesMsg(policies)
	}
}

func (m CFModel) fetchFunctions() tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewCloudFrontClient(context.Background(), m.profile)
		if err != nil {
			return CFErrorMsg(err)
		}
		fns, err := client.ListFunctions(context.Background())
		if err != nil {
			return CFErrorMsg(err)
		}
		return CFFunctionsMsg(fns)
	}
}

func (m CFModel) Update(msg tea.Msg) (CFModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case CFMenuMsg:
		m.list.SetItems(msg)
		m.list.ResetSelected()
		// Determine state based on items
		if len(msg) > 0 {
			if item, ok := msg[0].(cfItem); ok && item.category == "submenu" {
				m.state = CFStateDistroSubMenu
			} else {
				m.state = CFStateMenu
			}
		}
		m.updateDelegate()

	case CFDistributionsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			enabled := "No"
			if v.Enabled {
				enabled = "Yes"
			}
			displayName := m.getDistroDisplayName(v.ID, v.FirstAlias)
			items[i] = cfItem{
				title:       v.ID,
				description: v.Comment,
				id:          v.ID,
				category:    "distribution",
				values:      []string{displayName, v.Status, v.Domain, v.Comment, enabled},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = CFStateDistributions
		m.updateDelegate()
		return m, nil

	case CFOriginsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = cfItem{
				title:       v.ID,
				description: v.Domain,
				id:          v.ID,
				category:    "origin",
				values:      []string{v.ID, v.Domain, v.Path},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = CFStateOrigins
		m.updateDelegate()

	case CFBehaviorsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = cfItem{
				title:       v.PathPattern,
				description: v.TargetOriginID,
				id:          v.PathPattern,
				category:    "behavior",
				values:      []string{v.PathPattern, v.TargetOriginID, v.ViewerProtocol},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = CFStateBehaviors
		m.updateDelegate()

	case CFInvalidationsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = cfItem{
				title:       v.ID,
				description: v.Status,
				id:          v.ID,
				category:    "invalidation",
				values:      []string{v.ID, v.Status, v.CreateTime},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = CFStateInvalidations
		m.updateDelegate()

	case CFPoliciesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = cfItem{
				title:       v.Name,
				description: v.ID,
				id:          v.ID,
				category:    "policy",
				values:      []string{v.ID, v.Name, v.Type},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = CFStatePolicies
		m.updateDelegate()

	case CFFunctionsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = cfItem{
				title:       v.Name,
				description: v.Status,
				id:          v.Name,
				category:    "function",
				values:      []string{v.Name, v.Status, v.Runtime},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = CFStateFunctions
		m.updateDelegate()

	case CFErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			switch m.state {
			case CFStateDistributions:
				m.cache.Delete(m.cacheKeys.CFResources("distributions"))
				return m, m.fetchDistributions()
			case CFStateOrigins:
				return m, m.fetchDistroDetails(m.selectedDistro, "origins")
			case CFStateBehaviors:
				return m, m.fetchDistroDetails(m.selectedDistro, "behaviors")
			case CFStateInvalidations:
				return m, m.fetchDistroDetails(m.selectedDistro, "invalidations")
			case CFStatePolicies:
				return m, m.fetchPolicies()
			case CFStateFunctions:
				return m, m.fetchFunctions()
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(cfItem); ok {
				switch m.state {
				case CFStateMenu:
					switch item.title {
					case "Distributions":
						return m, m.fetchDistributions()
					case "Policies":
						return m, m.fetchPolicies()
					case "Functions":
						return m, m.fetchFunctions()
					}
				case CFStateDistributions:
					m.selectedDistro = item.id
					return m, m.showDistroSubMenu(item.id)
				case CFStateDistroSubMenu:
					switch item.title {
					case "Origins":
						return m, m.fetchDistroDetails(m.selectedDistro, "origins")
					case "Behaviors":
						return m, m.fetchDistroDetails(m.selectedDistro, "behaviors")
					case "Invalidations":
						return m, m.fetchDistroDetails(m.selectedDistro, "invalidations")
					}
				}
			}
		case "backspace", "esc":
			switch m.state {
			case CFStateDistributions, CFStatePolicies, CFStateFunctions:
				return m, m.showMenu()
			case CFStateDistroSubMenu:
				return m, m.fetchDistributions()
			case CFStateOrigins, CFStateBehaviors, CFStateInvalidations:
				return m, m.showDistroSubMenu(m.selectedDistro)
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *CFModel) updateDelegate() {
	d := cfItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          m.styles,
		state:           m.state,
	}
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc
	m.list.SetDelegate(d)
}

func (m CFModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state != CFStateMenu && m.state != CFStateDistroSubMenu {
		var columns []Column
		switch m.state {
		case CFStateDistributions:
			columns = cfDistroColumns
		case CFStateOrigins:
			columns = cfOriginColumns
		case CFStateBehaviors:
			columns = cfBehaviorColumns
		case CFStateInvalidations:
			columns = cfInvalidationColumns
		case CFStatePolicies:
			columns = cfPolicyColumns
		case CFStateFunctions:
			columns = cfFunctionColumns
		}
		_, header := RenderTableHelpers(m.list, m.styles, columns)
		return header + "\n" + m.list.View()
	}

	return m.list.View()
}

func (m *CFModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
