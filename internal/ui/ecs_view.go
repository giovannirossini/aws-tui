package ui

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

type ECSState int

const (
	ECSStateMenu ECSState = iota
	ECSStateClusters
	ECSStateServices
	ECSStateTasks
	ECSStateEvents
	ECSStateTaskDefFamilies
	ECSStateTaskDefRevisions
	ECSStateTaskDefJSON
	ECSStateSubMenu
	ECSStateTaskActions
	ECSStateServiceActions
)

type ecsItem struct {
	title       string
	description string
	id          string
	arn         string
	taskDef     string
	values      []string
}

func (i ecsItem) Title() string       { return i.title }
func (i ecsItem) Description() string { return i.description }
func (i ecsItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type ECSModel struct {
	client                 *aws.ECSClient
	list                   list.Model
	actionList             list.Model
	serviceActionList      list.Model
	viewport               viewport.Model
	delegate               ecsItemDelegate
	styles                 Styles
	state                  ECSState
	width                  int
	height                 int
	profile                string
	err                    error
	cache                  *cache.Cache
	cacheKeys              *cache.KeyBuilder
	selectedCluster        string
	selectedService        string
	selectedTask           string
	selectedServiceTaskDef string
	selectedTaskDefFamily  string
	selectedTaskDefJSON    string
	allTaskDefs            []aws.TaskDefinitionInfo
}

type ecsItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  ECSState
}

var ecsMenuColumns = []Column{
	{Title: "Resource Category", Width: 1.0},
}

var ecsClusterColumns = []Column{
	{Title: "Cluster Name", Width: 0.3},
	{Title: "Status", Width: 0.1},
	{Title: "Services", Width: 0.15},
	{Title: "Running Tasks", Width: 0.15},
	{Title: "Pending Tasks", Width: 0.15},
}

var ecsServiceColumns = []Column{
	{Title: "Service Name", Width: 0.3},
	{Title: "Status", Width: 0.1},
	{Title: "Running/Desired", Width: 0.2},
	{Title: "Launch Type", Width: 0.15},
}

var ecsTaskColumns = []Column{
	{Title: "Task ID", Width: 0.2},
	{Title: "Status", Width: 0.15},
	{Title: "Desired", Width: 0.2},
	{Title: "CPU", Width: 0.1},
	{Title: "Memory", Width: 0.1},
	{Title: "Created", Width: 0.25},
}

var ecsEventColumns = []Column{
	{Title: "Time", Width: 0.15},
	{Title: "Message", Width: 0.85},
}

var ecsTaskDefFamilyColumns = []Column{
	{Title: "Family", Width: 1.0},
}

var ecsTaskDefRevisionColumns = []Column{
	{Title: "Revision", Width: 0.2},
	{Title: "Status", Width: 0.3},
	{Title: "ARN", Width: 0.5},
}

var ecsTaskDefColumns = []Column{
	{Title: "Family", Width: 0.5},
	{Title: "Revision", Width: 0.2},
	{Title: "Status", Width: 0.3},
}

func (d ecsItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ecsItem)
	if !ok {
		return
	}

	var columns []Column
	switch d.state {
	case ECSStateMenu:
		columns = ecsMenuColumns
	case ECSStateClusters:
		columns = ecsClusterColumns
	case ECSStateServices:
		columns = ecsServiceColumns
	case ECSStateTasks:
		columns = ecsTaskColumns
	case ECSStateEvents:
		columns = ecsEventColumns
	case ECSStateTaskDefFamilies:
		columns = ecsTaskDefFamilyColumns
	case ECSStateTaskDefRevisions:
		columns = ecsTaskDefRevisionColumns
	case ECSStateSubMenu:
		columns = ecsMenuColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d ecsItemDelegate) Height() int {
	return 1
}

func NewECSModel(profile string, styles Styles, appCache *cache.Cache) ECSModel {
	d := ecsItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           ECSStateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "ECS"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	m := ECSModel{
		list:      l,
		delegate:  d,
		viewport:  viewport.New(0, 0),
		styles:    styles,
		state:     ECSStateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
	m.loadMenu()
	return m
}

func (m *ECSModel) loadMenu() {
	items := []list.Item{
		ecsItem{title: "Clusters", id: "clusters", values: []string{"Clusters"}},
		ecsItem{title: "Task Definitions", id: "task-def-families", values: []string{"Task Definitions"}},
	}
	m.list.SetItems(items)
	m.list.ResetSelected()
	m.state = ECSStateMenu
}

func (m *ECSModel) loadServiceSubMenu(serviceName string) {
	items := []list.Item{
		ecsItem{title: "Tasks", id: "tasks", values: []string{"Tasks"}},
		ecsItem{title: "Logs", id: "logs", values: []string{"Logs"}},
		ecsItem{title: "Events", id: "events", values: []string{"Events"}},
	}
	m.list.SetItems(items)
	m.list.ResetSelected()
	// Use menu columns for this submenu
}

func (m *ECSModel) loadActionMenu() {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc

	m.actionList = list.New([]list.Item{
		ecsItem{title: "Restart", id: "restart-task", description: "Stop the task (ECS will restart it if in a service)"},
	}, d, 30, 10)
	m.actionList.Title = "Task Actions"
	m.actionList.SetShowStatusBar(false)
	m.actionList.SetShowHelp(false)
	m.actionList.SetShowTitle(true)
}

func (m *ECSModel) loadServiceActionMenu() {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc

	m.serviceActionList = list.New([]list.Item{
		ecsItem{title: "Restart", id: "restart-service", description: "Force a new deployment"},
		ecsItem{title: "Stop", id: "stop-service", description: "Set desired tasks to 0"},
	}, d, 30, 10)
	m.serviceActionList.Title = "Service Actions"
	m.serviceActionList.SetShowStatusBar(false)
	m.serviceActionList.SetShowHelp(false)
	m.serviceActionList.SetShowTitle(true)
}

type ECSClustersMsg []aws.ECSClusterInfo
type ECSServicesMsg []aws.ServiceInfo
type ECSTasksMsg []aws.ECSTaskInfo
type ECSEventsMsg []aws.ECSEventInfo
type ECSTaskDefsMsg []aws.TaskDefinitionInfo
type ECSTaskDefFamiliesMsg []string
type ECSTaskDefJSONMsg string
type ECSLogGroupMsg string
type ECSSuccessMsg string
type ECSErrorMsg error

func (m ECSModel) Init() tea.Cmd {
	return nil
}

func (m ECSModel) fetchClusters() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.ECSResources("clusters")); ok {
			if clusters, ok := cached.([]aws.ECSClusterInfo); ok {
				return ECSClustersMsg(clusters)
			}
		}

		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		clusters, err := client.ListClusters(context.Background())
		if err != nil {
			return ECSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.ECSResources("clusters"), clusters, cache.TTLECSResources)
		return ECSClustersMsg(clusters)
	}
}

func (m ECSModel) fetchServices(cluster string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		services, err := client.ListServices(context.Background(), cluster)
		if err != nil {
			return ECSErrorMsg(err)
		}
		return ECSServicesMsg(services)
	}
}

func (m ECSModel) fetchTasks(cluster, service string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		var svc *string
		if service != "" {
			svc = &service
		}
		tasks, err := client.ListTasks(context.Background(), cluster, svc)
		if err != nil {
			return ECSErrorMsg(err)
		}
		return ECSTasksMsg(tasks)
	}
}

func (m ECSModel) fetchEvents(cluster, service string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		events, err := client.GetServiceEvents(context.Background(), cluster, service)
		if err != nil {
			return ECSErrorMsg(err)
		}
		return ECSEventsMsg(events)
	}
}

func (m ECSModel) fetchAllTaskDefs() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.ECSResources("all-task-defs")); ok {
			if defs, ok := cached.([]aws.TaskDefinitionInfo); ok {
				return ECSTaskDefsMsg(defs)
			}
		}

		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		defs, err := client.ListAllTaskDefinitions(context.Background())
		if err != nil {
			return ECSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.ECSResources("all-task-defs"), defs, cache.TTLECSResources)
		return ECSTaskDefsMsg(defs)
	}
}

func (m ECSModel) fetchTaskDefFamilies() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.ECSResources("task-def-families")); ok {
			if families, ok := cached.([]string); ok {
				return ECSTaskDefFamiliesMsg(families)
			}
		}

		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		families, err := client.ListTaskDefinitionFamilies(context.Background())
		if err != nil {
			return ECSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.ECSResources("task-def-families"), families, cache.TTLECSResources)
		return ECSTaskDefFamiliesMsg(families)
	}
}

func (m ECSModel) fetchTaskDefRevisions(family string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		revisions, err := client.ListTaskDefinitionRevisions(context.Background(), family)
		if err != nil {
			return ECSErrorMsg(err)
		}
		return ECSTaskDefsMsg(revisions)
	}
}

func (m ECSModel) fetchTaskDefJSON(arn string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		json, err := client.GetTaskDefinitionJSON(context.Background(), arn)
		if err != nil {
			return ECSErrorMsg(err)
		}
		return ECSTaskDefJSONMsg(json)
	}
}

func (m ECSModel) fetchLogGroup(taskDefArn string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		group, err := client.GetLogGroupForTaskDefinition(context.Background(), taskDefArn)
		if err != nil {
			return ECSErrorMsg(err)
		}
		return ECSLogGroupMsg(group)
	}
}

func (m ECSModel) restartTask() tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		err = client.StopTask(context.Background(), m.selectedCluster, m.selectedTask)
		if err != nil {
			return ECSErrorMsg(err)
		}
		return ECSSuccessMsg("Task stopped successfully")
	}
}

func (m ECSModel) stopServiceAction() tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		err = client.StopService(context.Background(), m.selectedCluster, m.selectedService)
		if err != nil {
			return ECSErrorMsg(err)
		}
		return ECSSuccessMsg("Service stopped successfully")
	}
}

func (m ECSModel) restartServiceAction() tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewECSClient(context.Background(), m.profile)
		if err != nil {
			return ECSErrorMsg(err)
		}
		err = client.RestartService(context.Background(), m.selectedCluster, m.selectedService)
		if err != nil {
			return ECSErrorMsg(err)
		}
		return ECSSuccessMsg("Service restart initiated successfully")
	}
}

func (m ECSModel) Update(msg tea.Msg) (ECSModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case ECSClustersMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			status := v.Status
			if status == "ACTIVE" {
				status = lipgloss.NewStyle().Foreground(m.styles.Success.GetForeground()).Render(status)
			}
			items[i] = ecsItem{
				title:       v.Name,
				description: v.ARN,
				id:          v.Name,
				arn:         v.ARN,
				values: []string{
					v.Name,
					status,
					fmt.Sprintf("%d", v.ActiveServices),
					fmt.Sprintf("%d", v.RunningTasks),
					fmt.Sprintf("%d", v.PendingTasks),
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = ECSStateClusters

	case ECSServicesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			status := v.Status
			if status == "ACTIVE" {
				status = lipgloss.NewStyle().Foreground(m.styles.Success.GetForeground()).Render(status)
			}
			items[i] = ecsItem{
				title:       v.Name,
				description: v.ARN,
				id:          v.Name,
				arn:         v.ARN,
				taskDef:     v.TaskDefinition,
				values: []string{
					v.Name,
					status,
					fmt.Sprintf("%d/%d", v.RunningTasks, v.DesiredTasks),
					v.LaunchType,
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = ECSStateServices

	case ECSTasksMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			status := v.LastStatus
			if status == "RUNNING" {
				status = lipgloss.NewStyle().Foreground(m.styles.Success.GetForeground()).Render(status)
			} else if status == "STOPPED" {
				status = lipgloss.NewStyle().Foreground(m.styles.Error.GetForeground()).Render(status)
			}
			items[i] = ecsItem{
				title:       v.ID,
				description: v.ARN,
				id:          v.ID,
				arn:         v.ARN,
				taskDef:     v.TaskDefinition,
				values: []string{
					v.ID,
					status,
					v.DesiredStatus,
					v.CPU,
					v.Memory,
					v.CreatedAt,
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = ECSStateTasks

	case ECSEventsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = ecsItem{
				title:       v.Message,
				description: v.CreatedAt,
				id:          v.ID,
				values: []string{
					v.CreatedAt,
					v.Message,
				},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = ECSStateEvents

	case ECSTaskDefsMsg:
		m.allTaskDefs = msg
		// Sort revisions globally by revision number descending
		sort.Slice(m.allTaskDefs, func(i, j int) bool {
			return m.allTaskDefs[i].Revision > m.allTaskDefs[j].Revision
		})

		// If we are in Menu state, it means we just clicked "Task Definitions"
		if m.state == ECSStateMenu || m.state == ECSStateTaskDefFamilies {
			familiesMap := make(map[string]bool)
			var families []string
			for _, td := range msg {
				if !familiesMap[td.Family] {
					familiesMap[td.Family] = true
					families = append(families, td.Family)
				}
			}
			items := make([]list.Item, len(families))
			for i, f := range families {
				items[i] = ecsItem{
					title:  f,
					id:     f,
					values: []string{f},
				}
			}
			m.list.SetItems(items)
			m.list.ResetSelected()
			m.state = ECSStateTaskDefFamilies
		} else if m.state == ECSStateTaskDefRevisions {
			// Filtering revisions for the selected family
			var items []list.Item
			for _, v := range m.allTaskDefs {
				if v.Family == m.selectedTaskDefFamily {
					items = append(items, ecsItem{
						title:       fmt.Sprintf("Revision %d", v.Revision),
						description: v.ARN,
						id:          v.ARN,
						arn:         v.ARN,
						values: []string{
							fmt.Sprintf("%d", v.Revision),
							v.Status,
							v.ARN,
						},
					})
				}
			}
			m.list.SetItems(items)
			// Don't reset selected if we are refreshing?
			// But here we usually want to show the list
			m.list.ResetSelected()
		}

	case ECSTaskDefFamiliesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = ecsItem{
				title:  v,
				id:     v,
				values: []string{v},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = ECSStateTaskDefFamilies

	case ECSTaskDefJSONMsg:
		m.selectedTaskDefJSON = string(msg)
		m.state = ECSStateTaskDefJSON
		m.viewport.SetContent(m.highlightTaskDef(string(msg)))
		m.viewport.YOffset = 0

	case ECSSuccessMsg:
		m.err = nil
		if m.state == ECSStateServiceActions {
			m.state = ECSStateServices
			return m, m.fetchServices(m.selectedCluster)
		}
		m.state = ECSStateTasks
		return m, m.fetchTasks(m.selectedCluster, m.selectedService)

	case ECSErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		if m.state == ECSStateTaskDefJSON {
			switch msg.String() {
			case "esc", "backspace", "q":
				m.state = ECSStateTaskDefRevisions
			default:
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}

		if m.state == ECSStateTaskActions {
			switch msg.String() {
			case "esc", "q":
				m.state = ECSStateTasks
			case "enter":
				return m, m.restartTask()
			default:
				m.actionList, cmd = m.actionList.Update(msg)
				return m, cmd
			}
		}

		if m.state == ECSStateServiceActions {
			switch msg.String() {
			case "esc", "q":
				m.state = ECSStateServices
			case "enter":
				if item, ok := m.serviceActionList.SelectedItem().(ecsItem); ok {
					if item.id == "stop-service" {
						return m, m.stopServiceAction()
					} else if item.id == "restart-service" {
						return m, m.restartServiceAction()
					}
				}
			default:
				m.serviceActionList, cmd = m.serviceActionList.Update(msg)
				return m, cmd
			}
		}

		switch msg.String() {
		case "o":
			if m.state == ECSStateTasks {
				if item, ok := m.list.SelectedItem().(ecsItem); ok {
					m.selectedTask = item.arn
					m.state = ECSStateTaskActions
					m.loadActionMenu()
					return m, nil
				}
			}
			if m.state == ECSStateServices {
				if item, ok := m.list.SelectedItem().(ecsItem); ok {
					m.selectedService = item.id
					m.state = ECSStateServiceActions
					m.loadServiceActionMenu()
					return m, nil
				}
			}
		case "r":
			switch m.state {
			case ECSStateClusters:
				m.cache.Delete(m.cacheKeys.ECSResources("clusters"))
				return m, m.fetchClusters()
			case ECSStateServices:
				return m, m.fetchServices(m.selectedCluster)
			case ECSStateTasks:
				return m, m.fetchTasks(m.selectedCluster, m.selectedService)
			case ECSStateEvents:
				return m, m.fetchEvents(m.selectedCluster, m.selectedService)
			case ECSStateTaskDefFamilies:
				m.cache.Delete(m.cacheKeys.ECSResources("all-task-defs"))
				return m, m.fetchAllTaskDefs()
			case ECSStateTaskDefRevisions:
				return m, func() tea.Msg { return ECSTaskDefsMsg(m.allTaskDefs) }
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(ecsItem); ok {
				switch m.state {
				case ECSStateMenu:
					if item.id == "clusters" {
						return m, m.fetchClusters()
					} else {
						return m, m.fetchAllTaskDefs()
					}
				case ECSStateClusters:
					m.selectedCluster = item.id
					return m, m.fetchServices(m.selectedCluster)
				case ECSStateServices:
					m.selectedService = item.id
					m.selectedServiceTaskDef = item.taskDef
					m.loadServiceSubMenu(item.id)
					m.state = ECSStateSubMenu
					return m, nil
				case ECSStateTaskDefFamilies:
					m.selectedTaskDefFamily = item.id
					// We already have the data in m.allTaskDefs
					m.state = ECSStateTaskDefRevisions
					return m, func() tea.Msg { return ECSTaskDefsMsg(m.allTaskDefs) }
				case ECSStateTaskDefRevisions:
					return m, m.fetchTaskDefJSON(item.arn)
				case ECSStateTasks:
					// Options handled by 'o'
				case ECSStateEvents:
					// Just list
				case ECSStateTaskDefJSON:
					// Already viewing
				default:
					// Submenu handling
					if m.selectedService != "" {
						switch item.id {
						case "tasks":
							return m, m.fetchTasks(m.selectedCluster, m.selectedService)
						case "logs":
							return m, m.fetchLogGroup(m.selectedServiceTaskDef)
						case "events":
							return m, m.fetchEvents(m.selectedCluster, m.selectedService)
						}
					}
				}
			}
		case "backspace", "esc":
			switch m.state {
			case ECSStateClusters, ECSStateTaskDefFamilies:
				m.loadMenu()
			case ECSStateServices:
				return m, m.fetchClusters()
			case ECSStateTasks, ECSStateEvents:
				m.loadServiceSubMenu(m.selectedService)
				m.state = ECSStateSubMenu
			case ECSStateTaskDefRevisions:
				m.state = ECSStateTaskDefFamilies
				return m, func() tea.Msg { return ECSTaskDefsMsg(m.allTaskDefs) }
			case ECSStateTaskDefJSON:
				m.state = ECSStateTaskDefRevisions
				return m, nil
			default:
				if m.selectedService != "" {
					m.selectedService = ""
					return m, m.fetchServices(m.selectedCluster)
				}
				if m.selectedCluster != "" {
					m.selectedCluster = ""
					return m, m.fetchClusters()
				}
				m.loadMenu()
			}
			return m, nil
		}
	}

	m.list, cmd = m.list.Update(msg)
	// Update delegate state
	m.delegate.state = m.state
	m.list.SetDelegate(m.delegate)
	return m, cmd
}

func (m ECSModel) highlightTaskDef(content string) string {
	lexer := lexers.Get("json")
	if lexer == nil {
		lexer = lexers.Fallback
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

func (m ECSModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state == ECSStateTaskDefJSON {
		return lipgloss.NewStyle().
			Padding(1, 2).
			Render(m.viewport.View())
	}

	if m.state == ECSStateTaskActions {
		popup := m.styles.Popup.Width(38).Render(
			m.actionList.View(),
		)
		w, h := GetMainContainerSize(m.width, m.height)
		return lipgloss.Place(w, h-AppInternalFooterHeight-2, lipgloss.Center, lipgloss.Center, popup)
	}

	if m.state == ECSStateServiceActions {
		popup := m.styles.Popup.Width(38).Render(
			m.serviceActionList.View(),
		)
		w, h := GetMainContainerSize(m.width, m.height)
		return lipgloss.Place(w, h-AppInternalFooterHeight-2, lipgloss.Center, lipgloss.Center, popup)
	}

	var columns []Column
	switch m.state {
	case ECSStateMenu:
		columns = ecsMenuColumns
	case ECSStateClusters:
		columns = ecsClusterColumns
	case ECSStateServices:
		columns = ecsServiceColumns
	case ECSStateTasks:
		columns = ecsTaskColumns
	case ECSStateEvents:
		columns = ecsEventColumns
	case ECSStateTaskDefFamilies:
		columns = ecsTaskDefFamilyColumns
	case ECSStateTaskDefRevisions:
		columns = ecsTaskDefRevisionColumns
	default:
		// Submenu or unknown state
		columns = ecsMenuColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header + "\n" + m.list.View()
}

func (m *ECSModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
	m.viewport.Width = width - InnerContentWidthOffset
	m.viewport.Height = height - AppInternalFooterHeight - 4
}
