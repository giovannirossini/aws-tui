package ui

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

type acmItem struct {
	title       string
	description string
	id          string
	values      []string
}

func (i acmItem) Title() string       { return i.title }
func (i acmItem) Description() string { return i.description }
func (i acmItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type ACMModel struct {
	client    *aws.ACMClient
	list      list.Model
	delegate  acmItemDelegate
	styles    Styles
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type acmItemDelegate struct {
	list.DefaultDelegate
	styles Styles
}

var acmColumns = []Column{
	{Title: "Domain Name", Width: 0.4},
	{Title: "Status", Width: 0.15},
	{Title: "Type", Width: 0.15},
	{Title: "Expiration", Width: 0.3},
}

func (d acmItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(acmItem)
	if !ok {
		return
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, acmColumns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d acmItemDelegate) Height() int {
	return 1
}

func NewACMModel(profile string, styles Styles, appCache *cache.Cache) ACMModel {
	d := acmItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "ACM Certificates"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return ACMModel{
		list:      l,
		delegate:  d,
		styles:    styles,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type CertificatesMsg []aws.CertificateInfo
type ACMErrorMsg error

func (m ACMModel) Init() tea.Cmd {
	return m.fetchCertificates()
}

func (m ACMModel) fetchCertificates() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.ACMResources("certificates")); ok {
			if certs, ok := cached.([]aws.CertificateInfo); ok {
				return CertificatesMsg(certs)
			}
		}

		client, err := aws.NewACMClient(context.Background(), m.profile)
		if err != nil {
			return ACMErrorMsg(err)
		}
		certs, err := client.ListCertificates(context.Background())
		if err != nil {
			return ACMErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.ACMResources("certificates"), certs, cache.TTLACMResources)
		return CertificatesMsg(certs)
	}
}

func (m ACMModel) Update(msg tea.Msg) (ACMModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case CertificatesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			expires := "N/A"
			if v.ExpiresAt != nil {
				expires = v.ExpiresAt.Format("2006-01-02 15:04")
				if v.ExpiresAt.Before(time.Now()) {
					expires = lipgloss.NewStyle().Foreground(m.styles.Error.GetForeground()).Render(expires)
				} else if v.ExpiresAt.Before(time.Now().AddDate(0, 1, 0)) {
					expires = lipgloss.NewStyle().Foreground(m.styles.Warning.GetForeground()).Render(expires)
				}
			}

			status := v.Status
			switch status {
			case "ISSUED":
				status = lipgloss.NewStyle().Foreground(m.styles.Success.GetForeground()).Render(status)
			case "PENDING_VALIDATION":
				status = lipgloss.NewStyle().Foreground(m.styles.Warning.GetForeground()).Render(status)
			case "EXPIRED", "FAILED":
				status = lipgloss.NewStyle().Foreground(m.styles.Error.GetForeground()).Render(status)
			}

			items[i] = acmItem{
				title:       v.DomainName,
				description: v.ARN,
				id:          v.ARN,
				values: []string{
					v.DomainName,
					status,
					v.Type,
					expires,
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()

	case ACMErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.cache.Delete(m.cacheKeys.ACMResources("certificates"))
			return m, m.fetchCertificates()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ACMModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	_, header := RenderTableHelpers(m.list, m.styles, acmColumns)
	return header + "\n" + m.list.View()
}

func (m *ACMModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
