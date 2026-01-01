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
)

type securityHubItem struct {
	finding aws.SecurityFinding
}

func (i securityHubItem) Title() string       { return i.finding.Title }
func (i securityHubItem) Description() string { return i.finding.Description }
func (i securityHubItem) FilterValue() string { return i.finding.Title + " " + i.finding.ResourceID }

type SecurityHubModel struct {
	list      list.Model
	styles    Styles
	profile   string
	width     int
	height    int
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type securityHubItemDelegate struct {
	list.DefaultDelegate
	styles Styles
}

var securityHubColumns = []Column{
	{Title: "Severity", Width: 0.1},
	{Title: "Title", Width: 0.4},
	{Title: "Resource", Width: 0.3},
	{Title: "Compliance", Width: 0.2},
}

func (d securityHubItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(securityHubItem)
	if !ok {
		return
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, securityHubColumns)
	isSelected := index == m.Index()

	severityStyle := lipgloss.NewStyle()
	switch i.finding.Severity {
	case "CRITICAL":
		severityStyle = severityStyle.Foreground(lipgloss.Color("#FF0000")).Bold(true)
	case "HIGH":
		severityStyle = severityStyle.Foreground(lipgloss.Color("#FF4500"))
	case "MEDIUM":
		severityStyle = severityStyle.Foreground(lipgloss.Color("#FFA500"))
	case "LOW":
		severityStyle = severityStyle.Foreground(lipgloss.Color("#FFFF00"))
	}

	values := []string{
		severityStyle.Render(i.finding.Severity),
		i.finding.Title,
		i.finding.ResourceID,
		i.finding.Compliance,
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d securityHubItemDelegate) Height() int { return 1 }

func NewSecurityHubModel(profile string, styles Styles, appCache *cache.Cache) SecurityHubModel {
	d := securityHubItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Security Hub Findings"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(styles.Primary).PaddingLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(2)

	return SecurityHubModel{
		list:      l,
		styles:    styles,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type SecurityHubMsg []aws.SecurityFinding
type SecurityHubErrorMsg error

func (m SecurityHubModel) Init() tea.Cmd {
	return m.fetchFindings()
}

func (m SecurityHubModel) fetchFindings() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.SecurityHubResources()); ok {
			if findings, ok := cached.([]aws.SecurityFinding); ok {
				return SecurityHubMsg(findings)
			}
		}

		client, err := aws.NewSecurityHubClient(context.Background(), m.profile)
		if err != nil {
			return SecurityHubErrorMsg(err)
		}
		findings, err := client.GetFindings(context.Background())
		if err != nil {
			return SecurityHubErrorMsg(err)
		}

		m.cache.Set(m.cacheKeys.SecurityHubResources(), findings, cache.TTLSecurityHubResources)
		return SecurityHubMsg(findings)
	}
}

func (m SecurityHubModel) Update(msg tea.Msg) (SecurityHubModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case SecurityHubMsg:
		items := make([]list.Item, len(msg))
		for i, f := range msg {
			items[i] = securityHubItem{finding: f}
		}
		m.list.SetItems(items)

	case SecurityHubErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.cache.Delete(m.cacheKeys.SecurityHubResources())
			return m, m.fetchFindings()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SecurityHubModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	_, header := RenderTableHelpers(m.list, m.styles, securityHubColumns)
	return header + "\n" + m.list.View()
}

func (m *SecurityHubModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
