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

type VPCState int

const (
	VPCStateMenu VPCState = iota
	VPCStateVPCs
	VPCStateSubnets
	VPCStateNatGateways
	VPCStateRouteTables
	VPCStateVpnGateways
)

type vpcItem struct {
	title       string
	description string
	id          string
	category    string
	values      []string // Added for tabular rendering
}

func (i vpcItem) Title() string       { return i.title }
func (i vpcItem) Description() string { return i.description }
func (i vpcItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type VPCModel struct {
	client    *aws.EC2Client
	list      list.Model
	styles    Styles
	state     VPCState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
	vpcNames  map[string]string // ID -> Name lookup
}

type vpcItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  VPCState
}

var vpcColumns = []Column{
	{Title: "Name", Width: 0.3},
	{Title: "VPC ID", Width: 0.3},
	{Title: "CIDR Block", Width: 0.2},
	{Title: "State", Width: 0.1},
	{Title: "Default", Width: 0.1},
}

var subnetColumns = []Column{
	{Title: "Name", Width: 0.3},
	{Title: "Subnet ID", Width: 0.2},
	{Title: "VPC ID", Width: 0.3},
	{Title: "CIDR Block", Width: 0.1},
	{Title: "AZ", Width: 0.1},
}

var natColumns = []Column{
	{Title: "Name", Width: 0.25},
	{Title: "NAT ID", Width: 0.2},
	{Title: "VPC ID", Width: 0.25},
	{Title: "Public IP", Width: 0.15},
	{Title: "State", Width: 0.15},
}

var rtColumns = []Column{
	{Title: "Name", Width: 0.3},
	{Title: "RT ID", Width: 0.3},
	{Title: "VPC ID", Width: 0.4},
}

var vpnColumns = []Column{
	{Title: "Name", Width: 0.4},
	{Title: "VPN ID", Width: 0.3},
	{Title: "State", Width: 0.15},
	{Title: "Type", Width: 0.15},
}

func (d vpcItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(vpcItem)
	if !ok {
		return
	}

	if d.state == VPCStateMenu {
		d.DefaultDelegate.Render(w, m, index, listItem)
		return
	}

	var columns []Column
	switch d.state {
	case VPCStateVPCs:
		columns = vpcColumns
	case VPCStateSubnets:
		columns = subnetColumns
	case VPCStateNatGateways:
		columns = natColumns
	case VPCStateRouteTables:
		columns = rtColumns
	case VPCStateVpnGateways:
		columns = vpnColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d vpcItemDelegate) Height() int {
	if d.state == VPCStateMenu {
		return 2
	}
	return 1
}

func NewVPCModel(profile string, styles Styles, appCache *cache.Cache) VPCModel {
	d := vpcItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           VPCStateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "VPC Resources"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return VPCModel{
		list:      l,
		styles:    styles,
		state:     VPCStateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
		vpcNames:  make(map[string]string),
	}
}

func (m *VPCModel) getVPCDisplayName(id string) string {
	if name, ok := m.vpcNames[id]; ok && name != "" {
		return fmt.Sprintf("%s (%s)", id, name)
	}
	
	// Try to load from cache if map is empty
	if len(m.vpcNames) == 0 {
		if cached, ok := m.cache.Get(m.cacheKeys.VPCResources("vpcs")); ok {
			if vpcs, ok := cached.([]aws.VPCInfo); ok {
				for _, v := range vpcs {
					m.vpcNames[v.ID] = v.Name
				}
				if name, ok := m.vpcNames[id]; ok && name != "" {
					return fmt.Sprintf("%s (%s)", id, name)
				}
			}
		}
	}
	
	return id
}

type VPCsMsg []aws.VPCInfo
type SubnetsMsg []aws.SubnetInfo
type NatGatewaysMsg []aws.NatGatewayInfo
type RouteTablesMsg []aws.RouteTableInfo
type VpnGatewaysMsg []aws.VpnGatewayInfo
type VPCErrorMsg error

type VPCMenuMsg []list.Item

func (m VPCModel) Init() tea.Cmd {
	return m.showMenu()
}

func (m VPCModel) showMenu() tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			vpcItem{title: "VPCs", description: "Virtual Private Clouds", category: "menu", values: []string{"VPCs", "Virtual Private Clouds"}},
			vpcItem{title: "Subnets", description: "VPC Subnets", category: "menu", values: []string{"Subnets", "VPC Subnets"}},
			vpcItem{title: "NAT Gateways", description: "NAT Gateways", category: "menu", values: []string{"NAT Gateways", "NAT Gateways"}},
			vpcItem{title: "Route Tables", description: "Route Tables", category: "menu", values: []string{"Route Tables", "Route Tables"}},
			vpcItem{title: "VPN Gateways", description: "Virtual Private Gateways", category: "menu", values: []string{"VPN Gateways", "Virtual Private Gateways"}},
		}
		return VPCMenuMsg(items)
	}
}

func (m VPCModel) fetchVPCs() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.VPCResources("vpcs")); ok {
			if vpcs, ok := cached.([]aws.VPCInfo); ok {
				return VPCsMsg(vpcs)
			}
		}

		client, err := aws.NewEC2Client(context.Background(), m.profile)
		if err != nil {
			return VPCErrorMsg(err)
		}
		vpcs, err := client.ListVpcs(context.Background())
		if err != nil {
			return VPCErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.VPCResources("vpcs"), vpcs, cache.TTLVPCResources)
		return VPCsMsg(vpcs)
	}
}

func (m VPCModel) fetchSubnets() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.VPCResources("subnets")); ok {
			if subnets, ok := cached.([]aws.SubnetInfo); ok {
				return SubnetsMsg(subnets)
			}
		}

		client, err := aws.NewEC2Client(context.Background(), m.profile)
		if err != nil {
			return VPCErrorMsg(err)
		}
		subnets, err := client.ListSubnets(context.Background())
		if err != nil {
			return VPCErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.VPCResources("subnets"), subnets, cache.TTLVPCResources)
		return SubnetsMsg(subnets)
	}
}

func (m VPCModel) fetchNatGateways() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.VPCResources("nats")); ok {
			if nats, ok := cached.([]aws.NatGatewayInfo); ok {
				return NatGatewaysMsg(nats)
			}
		}

		client, err := aws.NewEC2Client(context.Background(), m.profile)
		if err != nil {
			return VPCErrorMsg(err)
		}
		nats, err := client.ListNatGateways(context.Background())
		if err != nil {
			return VPCErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.VPCResources("nats"), nats, cache.TTLVPCResources)
		return NatGatewaysMsg(nats)
	}
}

func (m VPCModel) fetchRouteTables() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.VPCResources("rts")); ok {
			if rts, ok := cached.([]aws.RouteTableInfo); ok {
				return RouteTablesMsg(rts)
			}
		}

		client, err := aws.NewEC2Client(context.Background(), m.profile)
		if err != nil {
			return VPCErrorMsg(err)
		}
		rts, err := client.ListRouteTables(context.Background())
		if err != nil {
			return VPCErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.VPCResources("rts"), rts, cache.TTLVPCResources)
		return RouteTablesMsg(rts)
	}
}

func (m VPCModel) fetchVpnGateways() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.VPCResources("vpns")); ok {
			if vpns, ok := cached.([]aws.VpnGatewayInfo); ok {
				return VpnGatewaysMsg(vpns)
			}
		}

		client, err := aws.NewEC2Client(context.Background(), m.profile)
		if err != nil {
			return VPCErrorMsg(err)
		}
		vpns, err := client.ListVpnGateways(context.Background())
		if err != nil {
			return VPCErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.VPCResources("vpns"), vpns, cache.TTLVPCResources)
		return VpnGatewaysMsg(vpns)
	}
}

func (m VPCModel) Update(msg tea.Msg) (VPCModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case VPCMenuMsg:
		m.list.SetItems(msg)
		m.state = VPCStateMenu
		m.updateDelegate()

	case VPCsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			m.vpcNames[v.ID] = v.Name
			def := "No"
			if v.IsDefault {
				def = "Yes"
			}
			items[i] = vpcItem{
				title:       v.Name,
				description: v.CidrBlock,
				id:          v.ID,
				category:    "vpc",
				values:      []string{v.Name, v.ID, v.CidrBlock, v.State, def},
			}
		}
		m.list.SetItems(items)
		m.state = VPCStateVPCs
		m.updateDelegate()

	case SubnetsMsg:
		items := make([]list.Item, len(msg))
		for i, s := range msg {
			vpcDisplay := m.getVPCDisplayName(s.VpcID)
			items[i] = vpcItem{
				title:       s.Name,
				description: s.VpcID,
				id:          s.ID,
				category:    "subnet",
				values:      []string{s.Name, s.ID, vpcDisplay, s.CidrBlock, s.AvailabilityZone},
			}
		}
		m.list.SetItems(items)
		m.state = VPCStateSubnets
		m.updateDelegate()

	case NatGatewaysMsg:
		items := make([]list.Item, len(msg))
		for i, n := range msg {
			vpcDisplay := m.getVPCDisplayName(n.VpcID)
			items[i] = vpcItem{
				title:       n.Name,
				description: n.VpcID,
				id:          n.ID,
				category:    "nat",
				values:      []string{n.Name, n.ID, vpcDisplay, n.PublicIP, n.State},
			}
		}
		m.list.SetItems(items)
		m.state = VPCStateNatGateways
		m.updateDelegate()

	case RouteTablesMsg:
		items := make([]list.Item, len(msg))
		for i, r := range msg {
			vpcDisplay := m.getVPCDisplayName(r.VpcID)
			items[i] = vpcItem{
				title:       r.Name,
				description: r.VpcID,
				id:          r.ID,
				category:    "rt",
				values:      []string{r.Name, r.ID, vpcDisplay},
			}
		}
		m.list.SetItems(items)
		m.state = VPCStateRouteTables
		m.updateDelegate()

	case VpnGatewaysMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = vpcItem{
				title:       v.Name,
				description: v.State,
				id:          v.ID,
				category:    "vpn",
				values:      []string{v.Name, v.ID, v.State, v.Type},
			}
		}
		m.list.SetItems(items)
		m.state = VPCStateVpnGateways
		m.updateDelegate()

	case VPCErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			if m.state == VPCStateVPCs {
				m.cache.Delete(m.cacheKeys.VPCResources("vpcs"))
				return m, m.fetchVPCs()
			} else if m.state == VPCStateSubnets {
				m.cache.Delete(m.cacheKeys.VPCResources("subnets"))
				return m, m.fetchSubnets()
			} else if m.state == VPCStateNatGateways {
				m.cache.Delete(m.cacheKeys.VPCResources("nats"))
				return m, m.fetchNatGateways()
			} else if m.state == VPCStateRouteTables {
				m.cache.Delete(m.cacheKeys.VPCResources("rts"))
				return m, m.fetchRouteTables()
			} else if m.state == VPCStateVpnGateways {
				m.cache.Delete(m.cacheKeys.VPCResources("vpns"))
				return m, m.fetchVpnGateways()
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(vpcItem); ok {
				if m.state == VPCStateMenu {
					switch item.title {
					case "VPCs":
						return m, m.fetchVPCs()
					case "Subnets":
						return m, m.fetchSubnets()
					case "NAT Gateways":
						return m, m.fetchNatGateways()
					case "Route Tables":
						return m, m.fetchRouteTables()
					case "VPN Gateways":
						return m, m.fetchVpnGateways()
					}
				}
			}
		case "backspace", "esc":
			if m.state != VPCStateMenu {
				return m, m.showMenu()
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}


func (m *VPCModel) updateDelegate() {
	d := vpcItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          m.styles,
		state:           m.state,
	}
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc
	m.list.SetDelegate(d)
}

func (m VPCModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	header := ""
	if m.state != VPCStateMenu {
		var columns []Column
		switch m.state {
		case VPCStateVPCs:
			columns = vpcColumns
		case VPCStateSubnets:
			columns = subnetColumns
		case VPCStateNatGateways:
			columns = natColumns
		case VPCStateRouteTables:
			columns = rtColumns
		case VPCStateVpnGateways:
			columns = vpnColumns
		}
		_, header = RenderTableHelpers(m.list, m.styles, columns)
		return header + "\n" + m.list.View()
	}

	return m.list.View()
}

func (m *VPCModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
