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

type DMSState int

const (
	DMSStateMenu DMSState = iota
	DMSStateTasks
	DMSStateEndpoints
	DMSStateInstances
	DMSStateActions
)

type dmsItem struct {
	title       string
	description string
	id          string
	arn         string
	values      []string
}

func (i dmsItem) Title() string       { return i.title }
func (i dmsItem) Description() string { return i.description }
func (i dmsItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type DMSModel struct {
	client       *aws.DMSClient
	list         list.Model
	actionList   list.Model
	delegate     dmsItemDelegate
	styles       Styles
	state        DMSState
	width        int
	height       int
	profile      string
	err          error
	cache        *cache.Cache
	cacheKeys    *cache.KeyBuilder
	selectedTask string
}

type dmsItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  DMSState
}

var dmsMenuColumns = []Column{
	{Title: "Resource Type", Width: 1.0},
}

var dmsTaskColumns = []Column{
	{Title: "Task Identifier", Width: 0.3},
	{Title: "Status", Width: 0.15},
	{Title: "Type", Width: 0.15},
	{Title: "Progress", Width: 0.1},
	{Title: "Instance", Width: 0.3},
}

var dmsEndpointColumns = []Column{
	{Title: "Endpoint Identifier", Width: 0.3},
	{Title: "Type", Width: 0.1},
	{Title: "Engine", Width: 0.15},
	{Title: "Server", Width: 0.3},
	{Title: "Status", Width: 0.15},
}

var dmsInstanceColumns = []Column{
	{Title: "Instance Identifier", Width: 0.3},
	{Title: "Class", Width: 0.2},
	{Title: "Status", Width: 0.15},
	{Title: "Engine", Width: 0.1},
	{Title: "AZ", Width: 0.1},
	{Title: "Public", Width: 0.15},
}

func (d dmsItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(dmsItem)
	if !ok {
		return
	}

	var columns []Column
	switch d.state {
	case DMSStateMenu:
		columns = dmsMenuColumns
	case DMSStateTasks:
		columns = dmsTaskColumns
	case DMSStateEndpoints:
		columns = dmsEndpointColumns
	case DMSStateInstances:
		columns = dmsInstanceColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d dmsItemDelegate) Height() int {
	return 1
}

func NewDMSModel(profile string, styles Styles, appCache *cache.Cache) DMSModel {
	d := dmsItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           DMSStateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "DMS"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	m := DMSModel{
		list:      l,
		delegate:  d,
		styles:    styles,
		state:     DMSStateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
	m.loadMenu()
	return m
}

func (m *DMSModel) loadMenu() {
	items := []list.Item{
		dmsItem{title: "Replication Tasks", id: "tasks", values: []string{"Replication Tasks"}},
		dmsItem{title: "Endpoints", id: "endpoints", values: []string{"Endpoints"}},
		dmsItem{title: "Replication Instances", id: "instances", values: []string{"Replication Instances"}},
	}
	m.list.SetItems(items)
	m.list.ResetSelected()
	m.state = DMSStateMenu
}

func (m *DMSModel) loadActionMenu() {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc

	m.actionList = list.New([]list.Item{
		dmsItem{title: "Start", description: "Start the replication task"},
		dmsItem{title: "Stop", description: "Stop the replication task"},
		dmsItem{title: "Resume", description: "Resume the replication task"},
		dmsItem{title: "Reload", description: "Reload the target tables"},
	}, d, 30, 10)
	m.actionList.Title = "Task Actions"
	m.actionList.SetShowStatusBar(false)
	m.actionList.SetShowHelp(false)
	m.actionList.SetShowTitle(true)
}

type DMSTasksMsg []aws.ReplicationTaskInfo
type DMSEndpointsMsg []aws.DMSEndpointInfo
type DMSInstancesMsg []aws.ReplicationInstanceInfo
type DMSSuccessMsg string
type DMSErrorMsg error

func (m DMSModel) Init() tea.Cmd {
	return nil
}

func (m DMSModel) fetchTasks() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.DMSResources("tasks")); ok {
			if tasks, ok := cached.([]aws.ReplicationTaskInfo); ok {
				return DMSTasksMsg(tasks)
			}
		}

		client, err := aws.NewDMSClient(context.Background(), m.profile)
		if err != nil {
			return DMSErrorMsg(err)
		}
		tasks, err := client.ListReplicationTasks(context.Background())
		if err != nil {
			return DMSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.DMSResources("tasks"), tasks, cache.TTLDMSResources)
		return DMSTasksMsg(tasks)
	}
}

func (m DMSModel) fetchEndpoints() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.DMSResources("endpoints")); ok {
			if endpoints, ok := cached.([]aws.DMSEndpointInfo); ok {
				return DMSEndpointsMsg(endpoints)
			}
		}

		client, err := aws.NewDMSClient(context.Background(), m.profile)
		if err != nil {
			return DMSErrorMsg(err)
		}
		endpoints, err := client.ListEndpoints(context.Background())
		if err != nil {
			return DMSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.DMSResources("endpoints"), endpoints, cache.TTLDMSResources)
		return DMSEndpointsMsg(endpoints)
	}
}

func (m DMSModel) fetchInstances() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.DMSResources("instances")); ok {
			if instances, ok := cached.([]aws.ReplicationInstanceInfo); ok {
				return DMSInstancesMsg(instances)
			}
		}

		client, err := aws.NewDMSClient(context.Background(), m.profile)
		if err != nil {
			return DMSErrorMsg(err)
		}
		instances, err := client.ListReplicationInstances(context.Background())
		if err != nil {
			return DMSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.DMSResources("instances"), instances, cache.TTLDMSResources)
		return DMSInstancesMsg(instances)
	}
}

func (m DMSModel) runAction(action string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewDMSClient(context.Background(), m.profile)
		if err != nil {
			return DMSErrorMsg(err)
		}

		var cmdErr error
		switch action {
		case "Start":
			cmdErr = client.StartReplicationTask(context.Background(), m.selectedTask, "start-replication")
		case "Resume":
			cmdErr = client.StartReplicationTask(context.Background(), m.selectedTask, "resume-processing")
		case "Reload":
			cmdErr = client.StartReplicationTask(context.Background(), m.selectedTask, "reload-target")
		case "Stop":
			cmdErr = client.StopReplicationTask(context.Background(), m.selectedTask)
		}

		if cmdErr != nil {
			return DMSErrorMsg(cmdErr)
		}
		return DMSSuccessMsg(fmt.Sprintf("Task %s successful", action))
	}
}

func (m DMSModel) Update(msg tea.Msg) (DMSModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case DMSTasksMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			status := v.Status
			if status == "running" {
				status = lipgloss.NewStyle().Foreground(m.styles.Success.GetForeground()).Render(status)
			} else if strings.Contains(status, "failed") || strings.Contains(status, "error") {
				status = lipgloss.NewStyle().Foreground(m.styles.Error.GetForeground()).Render(status)
			}

			// Shorten ARNs for display
			instance := v.Instance
			if idx := strings.LastIndex(instance, ":"); idx != -1 {
				instance = instance[idx+1:]
			}

			items[i] = dmsItem{
				title:       v.ID,
				description: v.Status,
				id:          v.ID,
				arn:         v.ARN,
				values: []string{
					v.ID,
					status,
					v.Type,
					fmt.Sprintf("%d%%", v.FullLoadProgress),
					instance,
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = DMSStateTasks

	case DMSEndpointsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			status := v.Status
			if status == "active" {
				status = lipgloss.NewStyle().Foreground(m.styles.Success.GetForeground()).Render(status)
			}

			items[i] = dmsItem{
				title:       v.ID,
				description: v.Type,
				id:          v.ID,
				values: []string{
					v.ID,
					v.Type,
					v.Engine,
					v.Server,
					status,
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = DMSStateEndpoints

	case DMSInstancesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			status := v.Status
			if status == "available" {
				status = lipgloss.NewStyle().Foreground(m.styles.Success.GetForeground()).Render(status)
			}

			public := "No"
			if v.PubliclyAccessible {
				public = "Yes"
			}

			items[i] = dmsItem{
				title:       v.ID,
				description: v.Class,
				id:          v.ID,
				values: []string{
					v.ID,
					v.Class,
					status,
					v.EngineVersion,
					v.AZ,
					public,
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = DMSStateInstances

	case DMSSuccessMsg:
		m.err = nil
		m.state = DMSStateTasks
		// Refresh tasks after action
		return m, m.fetchTasks()

	case DMSErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		if m.state == DMSStateActions {
			switch msg.String() {
			case "esc", "q":
				m.state = DMSStateTasks
				return m, nil
			case "enter":
				item := m.actionList.SelectedItem().(dmsItem)
				return m, m.runAction(item.title)
			}
			m.actionList, cmd = m.actionList.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "o":
			if m.state == DMSStateTasks {
				if item, ok := m.list.SelectedItem().(dmsItem); ok {
					m.selectedTask = item.arn
					m.state = DMSStateActions
					m.loadActionMenu()
					return m, nil
				}
			}
		case "r":
			switch m.state {
			case DMSStateTasks:
				m.cache.Delete(m.cacheKeys.DMSResources("tasks"))
				return m, m.fetchTasks()
			case DMSStateEndpoints:
				m.cache.Delete(m.cacheKeys.DMSResources("endpoints"))
				return m, m.fetchEndpoints()
			case DMSStateInstances:
				m.cache.Delete(m.cacheKeys.DMSResources("instances"))
				return m, m.fetchInstances()
			}
		case "enter":
			if m.state == DMSStateMenu {
				item := m.list.SelectedItem().(dmsItem)
				switch item.id {
				case "tasks":
					return m, m.fetchTasks()
				case "endpoints":
					return m, m.fetchEndpoints()
				case "instances":
					return m, m.fetchInstances()
				}
			}
		case "backspace", "esc":
			if m.state != DMSStateMenu {
				m.loadMenu()
				return m, nil
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	// Update delegate state to match model state
	m.delegate.state = m.state
	m.list.SetDelegate(m.delegate)
	return m, cmd
}

func (m DMSModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state == DMSStateActions {
		popup := m.styles.Popup.Width(38).Render(
			m.actionList.View(),
		)
		w, h := GetMainContainerSize(m.width, m.height)
		return lipgloss.Place(w, h-AppInternalFooterHeight-2, lipgloss.Center, lipgloss.Center, popup)
	}

	var columns []Column
	switch m.state {
	case DMSStateMenu:
		columns = dmsMenuColumns
	case DMSStateTasks:
		columns = dmsTaskColumns
	case DMSStateEndpoints:
		columns = dmsEndpointColumns
	case DMSStateInstances:
		columns = dmsInstanceColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header + "\n" + m.list.View()
}

func (m *DMSModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
