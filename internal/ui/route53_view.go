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

type Route53State int

const (
	Route53StateZones Route53State = iota
	Route53StateRecords
)

type route53Item struct {
	title       string
	description string
	id          string
	values      []string
}

func (i route53Item) Title() string       { return i.title }
func (i route53Item) Description() string { return i.description }
func (i route53Item) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type Route53Model struct {
	client           *aws.Route53Client
	list             list.Model
	delegate         route53ItemDelegate
	styles           Styles
	state            Route53State
	width            int
	height           int
	profile          string
	err              error
	cache            *cache.Cache
	cacheKeys        *cache.KeyBuilder
	selectedZone     string
	selectedZoneName string
}

type route53ItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  Route53State
}

var hostedZoneColumns = []Column{
	{Title: "Domain Name", Width: 0.4},
	{Title: "Type", Width: 0.1},
	{Title: "ID", Width: 0.2},
	{Title: "Records", Width: 0.1},
	{Title: "Comment", Width: 0.2},
}

var route53RecordColumns = []Column{
	{Title: "Record Name", Width: 0.4},
	{Title: "Type", Width: 0.1},
	{Title: "Value", Width: 0.45},
	{Title: "TTL", Width: 0.05},
}

func (d route53ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(route53Item)
	if !ok {
		return
	}

	var columns []Column
	switch d.state {
	case Route53StateZones:
		columns = hostedZoneColumns
	case Route53StateRecords:
		columns = route53RecordColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d route53ItemDelegate) Height() int {
	return 1
}

func NewRoute53Model(profile string, styles Styles, appCache *cache.Cache) Route53Model {
	d := route53ItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           Route53StateZones,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Route 53"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return Route53Model{
		list:      l,
		delegate:  d,
		styles:    styles,
		state:     Route53StateZones,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type HostedZonesMsg []aws.HostedZoneInfo
type RecordSetsMsg []aws.ResourceRecordSetInfo
type Route53ErrorMsg error

func (m Route53Model) Init() tea.Cmd {
	return m.fetchHostedZones()
}

func (m Route53Model) fetchHostedZones() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.Route53Resources("hosted-zones")); ok {
			if zones, ok := cached.([]aws.HostedZoneInfo); ok {
				return HostedZonesMsg(zones)
			}
		}

		client, err := aws.NewRoute53Client(context.Background(), m.profile)
		if err != nil {
			return Route53ErrorMsg(err)
		}
		zones, err := client.ListHostedZones(context.Background())
		if err != nil {
			return Route53ErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.Route53Resources("hosted-zones"), zones, cache.TTLRoute53Resources)
		return HostedZonesMsg(zones)
	}
}

func (m Route53Model) fetchRecordSets(zoneID string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewRoute53Client(context.Background(), m.profile)
		if err != nil {
			return Route53ErrorMsg(err)
		}
		records, err := client.ListResourceRecordSets(context.Background(), zoneID)
		if err != nil {
			return Route53ErrorMsg(err)
		}
		return RecordSetsMsg(records)
	}
}

func (m Route53Model) Update(msg tea.Msg) (Route53Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case HostedZonesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			zoneType := "Public"
			if v.IsPrivate {
				zoneType = "Private"
			}
			items[i] = route53Item{
				title:       v.Name,
				description: v.ID,
				id:          v.ID,
				values: []string{
					strings.TrimSpace(v.Name),
					strings.TrimSpace(zoneType),
					strings.TrimSpace(v.ID),
					strings.TrimSpace(fmt.Sprintf("%d", v.RecordCount)),
					strings.TrimSpace(v.Comment),
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.delegate.state = Route53StateZones
		m.list.SetDelegate(m.delegate)
		m.state = Route53StateZones

	case RecordSetsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			recordType := v.Type
			val := strings.Join(v.Values, ", ")
			if v.Alias != "" {
				aliasTag := lipgloss.NewStyle().
					Foreground(m.styles.Muted).
					Render("(ALIAS)")
				recordType = fmt.Sprintf("%s %s", v.Type, aliasTag)
				val = v.Alias
			}
			ttl := fmt.Sprintf("%d", v.TTL)
			if v.TTL == 0 && v.Alias != "" {
				ttl = "-"
			}
			items[i] = route53Item{
				title:       v.Name,
				description: v.Type,
				id:          v.Name,
				values: []string{
					strings.TrimSpace(v.Name),
					strings.TrimSpace(recordType),
					strings.TrimSpace(val),
					strings.TrimSpace(ttl),
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.delegate.state = Route53StateRecords
		m.list.SetDelegate(m.delegate)
		m.state = Route53StateRecords

	case Route53ErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			if m.state == Route53StateZones {
				m.cache.Delete(m.cacheKeys.Route53Resources("hosted-zones"))
				return m, m.fetchHostedZones()
			} else if m.state == Route53StateRecords {
				return m, m.fetchRecordSets(m.selectedZone)
			}
		case "enter":
			if m.state == Route53StateZones {
				if item, ok := m.list.SelectedItem().(route53Item); ok {
					m.selectedZone = item.id
					m.selectedZoneName = item.title
					return m, m.fetchRecordSets(m.selectedZone)
				}
			}
		case "backspace", "esc":
			if m.state == Route53StateRecords {
				return m, m.fetchHostedZones()
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Route53Model) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	var columns []Column
	switch m.state {
	case Route53StateZones:
		columns = hostedZoneColumns
	case Route53StateRecords:
		columns = route53RecordColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header + "\n" + m.list.View()
}

func (m *Route53Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
