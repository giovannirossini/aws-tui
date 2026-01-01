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

type EC2State int

const (
	EC2StateMenu EC2State = iota
	EC2StateInstances
	EC2StateSecurityGroups
	EC2StateVolumes
	EC2StateTargetGroups
)

type ec2Item struct {
	title       string
	description string
	id          string
	category    string
	values      []string
}

func (i ec2Item) Title() string       { return i.title }
func (i ec2Item) Description() string { return i.description }
func (i ec2Item) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type EC2Model struct {
	client    *aws.EC2ResourcesClient
	list      list.Model
	styles    Styles
	state     EC2State
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type ec2ItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  EC2State
}

var instanceColumns = []Column{
	{Title: "Name", Width: 0.25},
	{Title: "Instance ID", Width: 0.2},
	{Title: "Type", Width: 0.15},
	{Title: "State", Width: 0.1},
	{Title: "Public IP", Width: 0.15},
	{Title: "AZ", Width: 0.15},
}

var sgColumns = []Column{
	{Title: "Name", Width: 0.25},
	{Title: "Group ID", Width: 0.2},
	{Title: "Description", Width: 0.35},
	{Title: "VPC ID", Width: 0.2},
}

var volumeColumns = []Column{
	{Title: "Name", Width: 0.25},
	{Title: "Volume ID", Width: 0.2},
	{Title: "Size (GB)", Width: 0.1},
	{Title: "Type", Width: 0.1},
	{Title: "State", Width: 0.1},
	{Title: "Instance ID", Width: 0.25},
}

var tgColumns = []Column{
	{Title: "Name", Width: 0.25},
	{Title: "Protocol", Width: 0.1},
	{Title: "Port", Width: 0.1},
	{Title: "Type", Width: 0.15},
	{Title: "VPC ID", Width: 0.4},
}

func (d ec2ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ec2Item)
	if !ok {
		return
	}

	if d.state == EC2StateMenu {
		d.DefaultDelegate.Render(w, m, index, listItem)
		return
	}

	var columns []Column
	switch d.state {
	case EC2StateInstances:
		columns = instanceColumns
	case EC2StateSecurityGroups:
		columns = sgColumns
	case EC2StateVolumes:
		columns = volumeColumns
	case EC2StateTargetGroups:
		columns = tgColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d ec2ItemDelegate) Height() int {
	if d.state == EC2StateMenu {
		return 2
	}
	return 1
}

func NewEC2Model(profile string, styles Styles, appCache *cache.Cache) EC2Model {
	d := ec2ItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           EC2StateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "EC2 Resources"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return EC2Model{
		list:      l,
		styles:    styles,
		state:     EC2StateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type InstancesMsg []aws.InstanceInfo
type SecurityGroupsMsg []aws.SecurityGroupInfo
type VolumesMsg []aws.VolumeInfo
type TargetGroupsMsg []aws.TargetGroupInfo
type EC2ErrorMsg error
type EC2MenuMsg []list.Item

func (m EC2Model) Init() tea.Cmd {
	return m.showMenu()
}

func (m EC2Model) showMenu() tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			ec2Item{title: "Instances", description: "EC2 Virtual Servers", category: "menu"},
			ec2Item{title: "Security Groups", description: "Network Firewall Rules", category: "menu"},
			ec2Item{title: "Volumes", description: "Elastic Block Store Volumes", category: "menu"},
			ec2Item{title: "Target Groups", description: "Load Balancer Target Groups", category: "menu"},
		}
		return EC2MenuMsg(items)
	}
}

func (m EC2Model) fetchInstances() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.EC2Resources("instances")); ok {
			if instances, ok := cached.([]aws.InstanceInfo); ok {
				return InstancesMsg(instances)
			}
		}

		client, err := aws.NewEC2ResourcesClient(context.Background(), m.profile)
		if err != nil {
			return EC2ErrorMsg(err)
		}
		instances, err := client.ListInstances(context.Background())
		if err != nil {
			return EC2ErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.EC2Resources("instances"), instances, cache.TTLEC2Resources)
		return InstancesMsg(instances)
	}
}

func (m EC2Model) fetchSecurityGroups() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.EC2Resources("security-groups")); ok {
			if sgs, ok := cached.([]aws.SecurityGroupInfo); ok {
				return SecurityGroupsMsg(sgs)
			}
		}

		client, err := aws.NewEC2ResourcesClient(context.Background(), m.profile)
		if err != nil {
			return EC2ErrorMsg(err)
		}
		sgs, err := client.ListSecurityGroups(context.Background())
		if err != nil {
			return EC2ErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.EC2Resources("security-groups"), sgs, cache.TTLEC2Resources)
		return SecurityGroupsMsg(sgs)
	}
}

func (m EC2Model) fetchVolumes() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.EC2Resources("volumes")); ok {
			if volumes, ok := cached.([]aws.VolumeInfo); ok {
				return VolumesMsg(volumes)
			}
		}

		client, err := aws.NewEC2ResourcesClient(context.Background(), m.profile)
		if err != nil {
			return EC2ErrorMsg(err)
		}
		volumes, err := client.ListVolumes(context.Background())
		if err != nil {
			return EC2ErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.EC2Resources("volumes"), volumes, cache.TTLEC2Resources)
		return VolumesMsg(volumes)
	}
}

func (m EC2Model) fetchTargetGroups() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.EC2Resources("target-groups")); ok {
			if tgs, ok := cached.([]aws.TargetGroupInfo); ok {
				return TargetGroupsMsg(tgs)
			}
		}

		client, err := aws.NewEC2ResourcesClient(context.Background(), m.profile)
		if err != nil {
			return EC2ErrorMsg(err)
		}
		tgs, err := client.ListTargetGroups(context.Background())
		if err != nil {
			return EC2ErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.EC2Resources("target-groups"), tgs, cache.TTLEC2Resources)
		return TargetGroupsMsg(tgs)
	}
}

func (m EC2Model) Update(msg tea.Msg) (EC2Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case EC2MenuMsg:
		m.list.SetItems(msg)
		m.list.ResetSelected()
		m.state = EC2StateMenu
		m.updateDelegate()

	case InstancesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = ec2Item{
				title:       v.Name,
				description: v.ID,
				id:          v.ID,
				category:    "instance",
				values:      []string{v.Name, v.ID, v.Type, v.State, v.PublicIP, v.AvailabilityZone},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = EC2StateInstances
		m.updateDelegate()

	case SecurityGroupsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = ec2Item{
				title:       v.Name,
				description: v.ID,
				id:          v.ID,
				category:    "sg",
				values:      []string{v.Name, v.ID, v.Description, v.VpcID},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = EC2StateSecurityGroups
		m.updateDelegate()

	case VolumesMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = ec2Item{
				title:       v.Name,
				description: v.ID,
				id:          v.ID,
				category:    "volume",
				values:      []string{v.Name, v.ID, fmt.Sprintf("%d", v.Size), v.Type, v.State, v.InstanceID},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = EC2StateVolumes
		m.updateDelegate()

	case TargetGroupsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = ec2Item{
				title:       v.Name,
				description: v.ARN,
				id:          v.ARN,
				category:    "tg",
				values:      []string{v.Name, v.Protocol, fmt.Sprintf("%d", v.Port), v.TargetType, v.VpcID},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = EC2StateTargetGroups
		m.updateDelegate()

	case EC2ErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			switch m.state {
			case EC2StateInstances:
				m.cache.Delete(m.cacheKeys.EC2Resources("instances"))
				return m, m.fetchInstances()
			case EC2StateSecurityGroups:
				m.cache.Delete(m.cacheKeys.EC2Resources("security-groups"))
				return m, m.fetchSecurityGroups()
			case EC2StateVolumes:
				m.cache.Delete(m.cacheKeys.EC2Resources("volumes"))
				return m, m.fetchVolumes()
			case EC2StateTargetGroups:
				m.cache.Delete(m.cacheKeys.EC2Resources("target-groups"))
				return m, m.fetchTargetGroups()
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(ec2Item); ok {
				if m.state == EC2StateMenu {
					switch item.title {
					case "Instances":
						return m, m.fetchInstances()
					case "Security Groups":
						return m, m.fetchSecurityGroups()
					case "Volumes":
						return m, m.fetchVolumes()
					case "Target Groups":
						return m, m.fetchTargetGroups()
					}
				}
			}
		case "backspace", "esc":
			if m.state != EC2StateMenu {
				return m, m.showMenu()
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *EC2Model) updateDelegate() {
	d := ec2ItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          m.styles,
		state:           m.state,
	}
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc
	m.list.SetDelegate(d)
}

func (m EC2Model) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state != EC2StateMenu {
		var columns []Column
		switch m.state {
		case EC2StateInstances:
			columns = instanceColumns
		case EC2StateSecurityGroups:
			columns = sgColumns
		case EC2StateVolumes:
			columns = volumeColumns
		case EC2StateTargetGroups:
			columns = tgColumns
		}
		_, header := RenderTableHelpers(m.list, m.styles, columns)
		return header + "\n" + m.list.View()
	}

	return m.list.View()
}

func (m *EC2Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
