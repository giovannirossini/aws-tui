package ui

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

type billingItem struct {
	service string
	amount  string
	unit    string
}

func (i billingItem) Title() string       { return i.service }
func (i billingItem) Description() string { return fmt.Sprintf("%s %s", i.amount, i.unit) }
func (i billingItem) FilterValue() string { return i.service }

type BillingModel struct {
	list      list.Model
	styles    Styles
	profile   string
	width     int
	height    int
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type billingItemDelegate struct {
	list.DefaultDelegate
	styles Styles
}

var billingColumns = []Column{
	{Title: "Service", Width: 0.7},
	{Title: "Cost (This Month)", Width: 0.3},
}

func (d billingItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(billingItem)
	if !ok {
		return
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, billingColumns)
	isSelected := index == m.Index()

	values := []string{
		i.service,
		fmt.Sprintf("%s %s", i.amount, i.unit),
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d billingItemDelegate) Height() int { return 1 }

func NewBillingModel(profile string, styles Styles, appCache *cache.Cache) BillingModel {
	d := billingItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Billing & Costs"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(styles.Primary).PaddingLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(2)

	return BillingModel{
		list:      l,
		styles:    styles,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type BillingMsg []aws.CostInfo
type BillingErrorMsg error

func (m BillingModel) Init() tea.Cmd {
	return m.fetchCosts()
}

func (m BillingModel) fetchCosts() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.BillingResources()); ok {
			if costs, ok := cached.([]aws.CostInfo); ok {
				return BillingMsg(costs)
			}
		}

		client, err := aws.NewBillingClient(context.Background(), m.profile)
		if err != nil {
			return BillingErrorMsg(err)
		}
		costs, err := client.GetMonthlyCosts(context.Background())
		if err != nil {
			return BillingErrorMsg(err)
		}

		sort.Slice(costs, func(i, j int) bool {
			valI, _ := strconv.ParseFloat(costs[i].Amount, 64)
			valJ, _ := strconv.ParseFloat(costs[j].Amount, 64)
			return valI > valJ
		})

		m.cache.Set(m.cacheKeys.BillingResources(), costs, cache.TTLBillingResources)
		return BillingMsg(costs)
	}
}

func (m BillingModel) Update(msg tea.Msg) (BillingModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case BillingMsg:
		items := make([]list.Item, len(msg))
		for i, c := range msg {
			amount := c.Amount
			if val, err := strconv.ParseFloat(c.Amount, 64); err == nil {
				amount = fmt.Sprintf("%.2f", val)
			}
			items[i] = billingItem{
				service: c.Service,
				amount:  amount,
				unit:    c.Unit,
			}
		}
		m.list.SetItems(items)

	case BillingErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.cache.Delete(m.cacheKeys.BillingResources())
			return m, m.fetchCosts()
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m BillingModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	_, header := RenderTableHelpers(m.list, m.styles, billingColumns)
	return header + "\n" + m.list.View()
}

func (m *BillingModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
