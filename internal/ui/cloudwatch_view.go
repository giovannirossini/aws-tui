package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

type CWState int

const (
	CWStateMenu CWState = iota
	CWStateLogGroups
	CWStateLogStreams
	CWStateLogEvents
	CWStateLogDetail
)

type cwItem struct {
	title       string
	description string
	id          string
	category    string
	values      []string
}

func (i cwItem) Title() string       { return i.title }
func (i cwItem) Description() string { return i.description }
func (i cwItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type CWModel struct {
	client           *aws.CloudWatchClient
	list             list.Model
	styles           Styles
	state            CWState
	width            int
	height           int
	profile          string
	err              error
	cache            *cache.Cache
	cacheKeys        *cache.KeyBuilder
	selectedGroup    string
	selectedStream   string
	selectedMessage  string
}

type cwItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  CWState
}

var logGroupColumns = []Column{
	{Title: "Log Group Name", Width: 0.5},
	{Title: "Retention", Width: 0.1},
	{Title: "Stored (Bytes)", Width: 0.15},
	{Title: "Created", Width: 0.25},
}

var logStreamColumns = []Column{
	{Title: "Log Stream Name", Width: 0.5},
	{Title: "Last Event", Width: 0.25},
	{Title: "Created", Width: 0.25},
}

var logEventColumns = []Column{
	{Title: "Timestamp", Width: 0.25},
	{Title: "Message", Width: 0.75},
}

func (d cwItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(cwItem)
	if !ok {
		return
	}

	if d.state == CWStateMenu {
		d.DefaultDelegate.Render(w, m, index, listItem)
		return
	}

	var columns []Column
	switch d.state {
	case CWStateLogGroups:
		columns = logGroupColumns
	case CWStateLogStreams:
		columns = logStreamColumns
	case CWStateLogEvents:
		columns = logEventColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d cwItemDelegate) Height() int {
	if d.state == CWStateMenu {
		return 2
	}
	return 1
}

func NewCWModel(profile string, styles Styles, appCache *cache.Cache) CWModel {
	d := cwItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           CWStateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "CloudWatch Logs"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return CWModel{
		list:      l,
		styles:    styles,
		state:     CWStateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type CWLogGroupsMsg []aws.LogGroupInfo
type CWLogStreamsMsg []aws.LogStreamInfo
type CWLogEventsMsg []aws.LogEventInfo
type CWErrorMsg error
type CWMenuMsg []list.Item

func (m CWModel) Init() tea.Cmd {
	return m.showMenu()
}

func (m CWModel) showMenu() tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			cwItem{title: "Log Groups", description: "CloudWatch Log Groups", category: "menu"},
		}
		return CWMenuMsg(items)
	}
}

func (m CWModel) fetchLogGroups() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.CWResources("log-groups")); ok {
			if groups, ok := cached.([]aws.LogGroupInfo); ok {
				return CWLogGroupsMsg(groups)
			}
		}

		client, err := aws.NewCloudWatchClient(context.Background(), m.profile)
		if err != nil {
			return CWErrorMsg(err)
		}
		groups, err := client.ListLogGroups(context.Background())
		if err != nil {
			return CWErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.CWResources("log-groups"), groups, cache.TTLCWResources)
		return CWLogGroupsMsg(groups)
	}
}

func (m CWModel) fetchLogStreams(groupName string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewCloudWatchClient(context.Background(), m.profile)
		if err != nil {
			return CWErrorMsg(err)
		}
		streams, err := client.ListLogStreams(context.Background(), groupName)
		if err != nil {
			return CWErrorMsg(err)
		}
		return CWLogStreamsMsg(streams)
	}
}

func (m CWModel) fetchLogEvents(groupName, streamName string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewCloudWatchClient(context.Background(), m.profile)
		if err != nil {
			return CWErrorMsg(err)
		}
		events, err := client.GetLogEvents(context.Background(), groupName, streamName)
		if err != nil {
			return CWErrorMsg(err)
		}
		return CWLogEventsMsg(events)
	}
}

func (m CWModel) Update(msg tea.Msg) (CWModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case CWMenuMsg:
		m.list.SetItems(msg)
		m.list.ResetSelected()
		m.state = CWStateMenu
		m.updateDelegate()

	case CWLogGroupsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			retention := "Never"
			if v.RetentionDays > 0 {
				retention = fmt.Sprintf("%d days", v.RetentionDays)
			}
			items[i] = cwItem{
				title:       v.Name,
				description: v.Arn,
				id:          v.Name,
				category:    "log-group",
				values:      []string{v.Name, retention, fmt.Sprintf("%d", v.StoredBytes), v.CreationTime},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = CWStateLogGroups
		m.updateDelegate()

	case CWLogStreamsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = cwItem{
				title:       v.Name,
				description: v.Arn,
				id:          v.Name,
				category:    "log-stream",
				values:      []string{v.Name, v.LastEventTimestamp, v.CreationTime},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = CWStateLogStreams
		m.updateDelegate()

	case CWLogEventsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			ts := time.Unix(v.Timestamp/1000, 0).Format("15:04:05")
			items[i] = cwItem{
				title:       v.Message,
				description: ts,
				id:          fmt.Sprintf("%d", v.Timestamp),
				category:    "log-event",
				values:      []string{ts, v.Message},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = CWStateLogEvents
		m.updateDelegate()

	case CWErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			if m.state == CWStateLogGroups {
				m.cache.Delete(m.cacheKeys.CWResources("log-groups"))
				return m, m.fetchLogGroups()
			} else if m.state == CWStateLogStreams {
				return m, m.fetchLogStreams(m.selectedGroup)
			} else if m.state == CWStateLogEvents {
				return m, m.fetchLogEvents(m.selectedGroup, m.selectedStream)
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(cwItem); ok {
				if m.state == CWStateMenu {
					if item.title == "Log Groups" {
						return m, m.fetchLogGroups()
					}
				} else if m.state == CWStateLogGroups {
					m.selectedGroup = item.id
					return m, m.fetchLogStreams(m.selectedGroup)
				} else if m.state == CWStateLogStreams {
					m.selectedStream = item.id
					return m, m.fetchLogEvents(m.selectedGroup, m.selectedStream)
				} else if m.state == CWStateLogEvents {
					m.selectedMessage = item.title
					m.state = CWStateLogDetail
					return m, nil
				}
			}
		case "backspace", "esc":
			if m.state == CWStateLogGroups {
				return m, m.showMenu()
			} else if m.state == CWStateLogStreams {
				return m, m.fetchLogGroups()
			} else if m.state == CWStateLogEvents {
				return m, m.fetchLogStreams(m.selectedGroup)
			} else if m.state == CWStateLogDetail {
				m.state = CWStateLogEvents
				return m, nil
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *CWModel) updateDelegate() {
	d := cwItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          m.styles,
		state:           m.state,
	}
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc
	m.list.SetDelegate(d)
}

func (m CWModel) highlightLog(content string) string {
	lexer := lexers.Analyse(content)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// If it's JSON, force JSON lexer for better results
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(content), &jsonObj); err == nil {
		lexer = lexers.Get("json")
		if pretty, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
			content = string(pretty)
		}
	}

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}

	var sb strings.Builder
	err = formatter.Format(&sb, style, iterator)
	if err != nil {
		return content
	}

	// Add line numbers
	lines := strings.Split(sb.String(), "\n")
	var numberedLines []string
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}
		lineNumber := m.styles.StatusMuted.Render(fmt.Sprintf("%3d | ", i+1))
		numberedLines = append(numberedLines, lineNumber+line)
	}

	return strings.Join(numberedLines, "\n")
}

func (m CWModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state == CWStateLogDetail {
		displayMsg := m.highlightLog(m.selectedMessage)

		return lipgloss.NewStyle().
			Width(m.width - InnerContentWidthOffset).
			Padding(1, 2).
			Render(displayMsg)
	}

	if m.state != CWStateMenu {
		var columns []Column
		switch m.state {
		case CWStateLogGroups:
			columns = logGroupColumns
		case CWStateLogStreams:
			columns = logStreamColumns
		case CWStateLogEvents:
			columns = logEventColumns
		}
		_, header := RenderTableHelpers(m.list, m.styles, columns)
		return header + "\n" + m.list.View()
	}

	return m.list.View()
}

func (m *CWModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
