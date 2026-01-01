package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type profileItem string

func (p profileItem) FilterValue() string { return string(p) }
func (p profileItem) Title() string       { return string(p) }
func (p profileItem) Description() string { return "" }

type ProfileSelector struct {
	list     list.Model
	active   bool
	profiles []string
	selected string
	styles   Styles
}

func NewProfileSelector(profiles []string, initial string, styles Styles) ProfileSelector {
	items := make([]list.Item, len(profiles))
	for i, p := range profiles {
		items[i] = profileItem(p)
	}

	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetHeight(1)
	d.SetSpacing(0)
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	// Calculate height: title (2) + filter (2) + items + pagination (1)
	h := len(profiles) + 5
	if h > 15 {
		h = 15
	}

	l := list.New(items, d, 34, h)
	l.Title = "Select AWS Profile"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(true)
	l.SetFilteringEnabled(true)
	if len(profiles) <= 10 {
		l.SetShowPagination(false)
	}
	l.Styles.Title = styles.AppTitle.Copy().
		Background(styles.DarkGray).
		Foreground(styles.Primary).
		Margin(0, 0, 1, 0).
		Width(34).
		Align(lipgloss.Center)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(styles.Primary)
	l.KeyMap.Quit.SetEnabled(false)

	return ProfileSelector{
		list:     l,
		active:   false,
		profiles: profiles,
		selected: initial,
		styles:   styles,
	}
}

func (m ProfileSelector) Update(msg tea.Msg) (ProfileSelector, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(profileItem); ok {
				m.selected = string(i)
				m.active = false
				return m, func() tea.Msg { return ProfileSelectedMsg(m.selected) }
			}
		case "esc":
			m.active = false
			m.list.ResetSelected()
			m.list.FilterInput.Blur()
		}
	}

	return m, cmd
}

func (m ProfileSelector) View() string {
	if !m.active {
		return ""
	}
	return m.list.View()
}

type ProfileSelectedMsg string

func (m *ProfileSelector) SetSize(width, height int) {
	h := len(m.profiles) + 5
	if h > 15 {
		h = 15
	}
	if h > height-10 {
		h = height - 10
	}
	if h < 5 {
		h = 5
	}
	m.list.SetSize(34, h)
}
