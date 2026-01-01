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

type RDSState int

const (
	RDSStateMenu RDSState = iota
	RDSStateInstances
	RDSStateClusters
	RDSStateSnapshots
	RDSStateSubnetGroups
)

type rdsItem struct {
	title       string
	description string
	id          string
	category    string
	values      []string
}

func (i rdsItem) Title() string       { return i.title }
func (i rdsItem) Description() string { return i.description }
func (i rdsItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type RDSModel struct {
	client    *aws.RDSClient
	list      list.Model
	styles    Styles
	state     RDSState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type rdsItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  RDSState
}

var rdsInstanceColumns = []Column{
	{Title: "Instance ID", Width: 0.25},
	{Title: "Engine", Width: 0.15},
	{Title: "Status", Width: 0.15},
	{Title: "Class", Width: 0.15},
	{Title: "Endpoint", Width: 0.3},
}

var rdsClusterColumns = []Column{
	{Title: "Cluster ID", Width: 0.3},
	{Title: "Engine", Width: 0.2},
	{Title: "Status", Width: 0.2},
	{Title: "VPC ID", Width: 0.3},
}

var rdsSnapshotColumns = []Column{
	{Title: "Snapshot ID", Width: 0.3},
	{Title: "Instance ID", Width: 0.2},
	{Title: "Status", Width: 0.15},
	{Title: "Type", Width: 0.1},
	{Title: "Created At", Width: 0.25},
}

var rdsSubnetColumns = []Column{
	{Title: "Group Name", Width: 0.3},
	{Title: "Description", Width: 0.3},
	{Title: "VPC ID", Width: 0.2},
	{Title: "Status", Width: 0.2},
}

func (d rdsItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(rdsItem)
	if !ok {
		return
	}

	if d.state == RDSStateMenu {
		d.DefaultDelegate.Render(w, m, index, listItem)
		return
	}

	var columns []Column
	switch d.state {
	case RDSStateInstances:
		columns = rdsInstanceColumns
	case RDSStateClusters:
		columns = rdsClusterColumns
	case RDSStateSnapshots:
		columns = rdsSnapshotColumns
	case RDSStateSubnetGroups:
		columns = rdsSubnetColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d rdsItemDelegate) Height() int {
	if d.state == RDSStateMenu {
		return 2
	}
	return 1
}

func NewRDSModel(profile string, styles Styles, appCache *cache.Cache) RDSModel {
	d := rdsItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           RDSStateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "RDS Resources"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return RDSModel{
		list:      l,
		styles:    styles,
		state:     RDSStateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type RDSInstancesMsg []aws.RDSInstanceInfo
type RDSClustersMsg []aws.RDSClusterInfo
type RDSSnapshotsMsg []aws.RDSSnapshotInfo
type RDSSubnetGroupsMsg []aws.RDSSubnetGroupInfo
type RDSErrorMsg error
type RDSMenuMsg []list.Item

func (m RDSModel) Init() tea.Cmd {
	return m.showMenu()
}

func (m RDSModel) showMenu() tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			rdsItem{title: "Databases", description: "RDS DB Instances", category: "menu"},
			rdsItem{title: "Clusters", description: "RDS DB Clusters", category: "menu"},
			rdsItem{title: "Snapshots", description: "DB Snapshots", category: "menu"},
			rdsItem{title: "Subnet Groups", description: "DB Subnet Groups", category: "menu"},
		}
		return RDSMenuMsg(items)
	}
}

func (m RDSModel) fetchInstances() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.RDSResources("instances")); ok {
			if instances, ok := cached.([]aws.RDSInstanceInfo); ok {
				return RDSInstancesMsg(instances)
			}
		}

		client, err := aws.NewRDSClient(context.Background(), m.profile)
		if err != nil {
			return RDSErrorMsg(err)
		}
		instances, err := client.ListInstances(context.Background())
		if err != nil {
			return RDSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.RDSResources("instances"), instances, cache.TTLRDSResources)
		return RDSInstancesMsg(instances)
	}
}

func (m RDSModel) fetchClusters() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.RDSResources("clusters")); ok {
			if clusters, ok := cached.([]aws.RDSClusterInfo); ok {
				return RDSClustersMsg(clusters)
			}
		}

		client, err := aws.NewRDSClient(context.Background(), m.profile)
		if err != nil {
			return RDSErrorMsg(err)
		}
		clusters, err := client.ListClusters(context.Background())
		if err != nil {
			return RDSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.RDSResources("clusters"), clusters, cache.TTLRDSResources)
		return RDSClustersMsg(clusters)
	}
}

func (m RDSModel) fetchSnapshots() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.RDSResources("snapshots")); ok {
			if snapshots, ok := cached.([]aws.RDSSnapshotInfo); ok {
				return RDSSnapshotsMsg(snapshots)
			}
		}

		client, err := aws.NewRDSClient(context.Background(), m.profile)
		if err != nil {
			return RDSErrorMsg(err)
		}
		snapshots, err := client.ListSnapshots(context.Background())
		if err != nil {
			return RDSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.RDSResources("snapshots"), snapshots, cache.TTLRDSResources)
		return RDSSnapshotsMsg(snapshots)
	}
}

func (m RDSModel) fetchSubnetGroups() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.RDSResources("subnet-groups")); ok {
			if groups, ok := cached.([]aws.RDSSubnetGroupInfo); ok {
				return RDSSubnetGroupsMsg(groups)
			}
		}

		client, err := aws.NewRDSClient(context.Background(), m.profile)
		if err != nil {
			return RDSErrorMsg(err)
		}
		groups, err := client.ListSubnetGroups(context.Background())
		if err != nil {
			return RDSErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.RDSResources("subnet-groups"), groups, cache.TTLRDSResources)
		return RDSSubnetGroupsMsg(groups)
	}
}

func (m RDSModel) Update(msg tea.Msg) (RDSModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case RDSMenuMsg:
		m.list.SetItems(msg)
		m.list.ResetSelected()
		m.state = RDSStateMenu
		m.updateDelegate()

	case RDSInstancesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = rdsItem{
				title:       v.ID,
				description: v.Engine,
				id:          v.ID,
				category:    "instance",
				values:      []string{v.ID, v.Engine, v.Status, v.Class, v.Endpoint},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = RDSStateInstances
		m.updateDelegate()

	case RDSClustersMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = rdsItem{
				title:       v.ID,
				description: v.Engine,
				id:          v.ID,
				category:    "cluster",
				values:      []string{v.ID, v.Engine, v.Status, v.VpcID},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = RDSStateClusters
		m.updateDelegate()

	case RDSSnapshotsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = rdsItem{
				title:       v.ID,
				description: v.InstanceID,
				id:          v.ID,
				category:    "snapshot",
				values:      []string{v.ID, v.InstanceID, v.Status, v.Type, v.CreateTime},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = RDSStateSnapshots
		m.updateDelegate()

	case RDSSubnetGroupsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = rdsItem{
				title:       v.Name,
				description: v.VpcID,
				id:          v.Name,
				category:    "subnet-group",
				values:      []string{v.Name, v.Description, v.VpcID, v.Status},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = RDSStateSubnetGroups
		m.updateDelegate()

	case RDSErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			switch m.state {
			case RDSStateInstances:
				m.cache.Delete(m.cacheKeys.RDSResources("instances"))
				return m, m.fetchInstances()
			case RDSStateClusters:
				m.cache.Delete(m.cacheKeys.RDSResources("clusters"))
				return m, m.fetchClusters()
			case RDSStateSnapshots:
				m.cache.Delete(m.cacheKeys.RDSResources("snapshots"))
				return m, m.fetchSnapshots()
			case RDSStateSubnetGroups:
				m.cache.Delete(m.cacheKeys.RDSResources("subnet-groups"))
				return m, m.fetchSubnetGroups()
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(rdsItem); ok {
				if m.state == RDSStateMenu {
					switch item.title {
					case "Databases":
						return m, m.fetchInstances()
					case "Clusters":
						return m, m.fetchClusters()
					case "Snapshots":
						return m, m.fetchSnapshots()
					case "Subnet Groups":
						return m, m.fetchSubnetGroups()
					}
				}
			}
		case "backspace", "esc":
			if m.state != RDSStateMenu {
				return m, m.showMenu()
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *RDSModel) updateDelegate() {
	d := rdsItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          m.styles,
		state:           m.state,
	}
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc
	m.list.SetDelegate(d)
}

func (m RDSModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state != RDSStateMenu {
		var columns []Column
		switch m.state {
		case RDSStateInstances:
			columns = rdsInstanceColumns
		case RDSStateClusters:
			columns = rdsClusterColumns
		case RDSStateSnapshots:
			columns = rdsSnapshotColumns
		case RDSStateSubnetGroups:
			columns = rdsSubnetColumns
		}
		_, header := RenderTableHelpers(m.list, m.styles, columns)
		return header + "\n" + m.list.View()
	}

	return m.list.View()
}

func (m *RDSModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
