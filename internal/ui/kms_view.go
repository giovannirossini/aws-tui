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

type kmsItem struct {
	title       string
	description string
	id          string
	values      []string
}

func (i kmsItem) Title() string       { return i.title }
func (i kmsItem) Description() string { return i.description }
func (i kmsItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type KMSModel struct {
	client    *aws.KMSClient
	list      list.Model
	styles    Styles
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type kmsItemDelegate struct {
	list.DefaultDelegate
	styles Styles
}

var kmsColumns = []Column{
	{Title: "Alias / Key ID", Width: 0.4},
	{Title: "Status", Width: 0.15},
	{Title: "Manager", Width: 0.15},
	{Title: "Description", Width: 0.3},
}

func (d kmsItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(kmsItem)
	if !ok {
		return
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, kmsColumns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d kmsItemDelegate) Height() int {
	return 1
}

func NewKMSModel(profile string, styles Styles, appCache *cache.Cache) KMSModel {
	d := kmsItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "KMS Keys"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return KMSModel{
		list:      l,
		styles:    styles,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type KMSKeysMsg []aws.KMSKeyInfo
type KMSErrorMsg error

func (m KMSModel) Init() tea.Cmd {
	return m.fetchKeys()
}

func (m KMSModel) fetchKeys() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.KMSResources("keys")); ok {
			if keys, ok := cached.([]aws.KMSKeyInfo); ok {
				return KMSKeysMsg(keys)
			}
		}

		client, err := aws.NewKMSClient(context.Background(), m.profile)
		if err != nil {
			return KMSErrorMsg(err)
		}
		keys, err := client.ListKeys(context.Background())
		if err != nil {
			return KMSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.KMSResources("keys"), keys, cache.TTLKMSResources)
		return KMSKeysMsg(keys)
	}
}

func (m KMSModel) Update(msg tea.Msg) (KMSModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case KMSKeysMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			name := v.Alias
			if name == "" {
				name = v.ID
			} else {
				name = strings.TrimPrefix(name, "alias/")
			}

			status := v.State
			switch status {
			case "Enabled":
				status = lipgloss.NewStyle().Foreground(m.styles.Success.GetForeground()).Render(status)
			case "Disabled":
				status = lipgloss.NewStyle().Foreground(m.styles.Error.GetForeground()).Render(status)
			case "PendingDeletion":
				status = lipgloss.NewStyle().Foreground(m.styles.Warning.GetForeground()).Render(status)
			}

			manager := v.Manager
			if manager == "AWS" {
				manager = lipgloss.NewStyle().Foreground(m.styles.Muted).Render(manager)
			} else {
				manager = lipgloss.NewStyle().Foreground(m.styles.Accent).Render(manager)
			}

			items[i] = kmsItem{
				title:       name,
				description: v.ARN,
				id:          v.ID,
				values: []string{
					name,
					status,
					manager,
					v.Description,
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()

	case KMSErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.cache.Delete(m.cacheKeys.KMSResources("keys"))
			return m, m.fetchKeys()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m KMSModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	_, header := RenderTableHelpers(m.list, m.styles, kmsColumns)
	return header + "\n" + m.list.View()
}

func (m *KMSModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
