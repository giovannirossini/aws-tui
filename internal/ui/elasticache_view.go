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

type ElastiCacheState int

const (
	ElastiCacheStateMenu ElastiCacheState = iota
	ElastiCacheStateReplicationGroups
	ElastiCacheStateCacheClusters
)

type elasticacheItem struct {
	title       string
	description string
	id          string
	category    string
	values      []string
}

func (i elasticacheItem) Title() string       { return i.title }
func (i elasticacheItem) Description() string { return i.description }
func (i elasticacheItem) FilterValue() string { return i.title + " " + i.description + " " + i.id }

type ElastiCacheModel struct {
	client    *aws.ElastiCacheClient
	list      list.Model
	styles    Styles
	state     ElastiCacheState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type elasticacheItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  ElastiCacheState
}

var replicationGroupColumns = []Column{
	{Title: "ID", Width: 0.3},
	{Title: "Status", Width: 0.15},
	{Title: "Engine", Width: 0.15},
	{Title: "Node Type", Width: 0.2},
	{Title: "Nodes", Width: 0.1},
	{Title: "Description", Width: 0.1},
}

var cacheClusterColumns = []Column{
	{Title: "ID", Width: 0.3},
	{Title: "Status", Width: 0.15},
	{Title: "Engine", Width: 0.15},
	{Title: "Version", Width: 0.1},
	{Title: "Node Type", Width: 0.2},
	{Title: "AZ", Width: 0.1},
}

func (d elasticacheItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(elasticacheItem)
	if !ok {
		return
	}

	if d.state == ElastiCacheStateMenu {
		d.DefaultDelegate.Render(w, m, index, listItem)
		return
	}

	var columns []Column
	switch d.state {
	case ElastiCacheStateReplicationGroups:
		columns = replicationGroupColumns
	case ElastiCacheStateCacheClusters:
		columns = cacheClusterColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d elasticacheItemDelegate) Height() int {
	if d.state == ElastiCacheStateMenu {
		return 2
	}
	return 1
}

func NewElastiCacheModel(profile string, styles Styles, appCache *cache.Cache) ElastiCacheModel {
	d := elasticacheItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           ElastiCacheStateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "ElastiCache Resources"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return ElastiCacheModel{
		list:      l,
		styles:    styles,
		state:     ElastiCacheStateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type ReplicationGroupsMsg []aws.ReplicationGroupInfo
type CacheClustersMsg []aws.CacheClusterInfo
type ElastiCacheErrorMsg error
type ElastiCacheMenuMsg []list.Item

func (m ElastiCacheModel) Init() tea.Cmd {
	return m.showMenu()
}

func (m ElastiCacheModel) showMenu() tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{
			elasticacheItem{title: "Replication Groups", description: "Redis Replication Groups", category: "menu"},
			elasticacheItem{title: "Cache Clusters", description: "Individual Cache Clusters", category: "menu"},
		}
		return ElastiCacheMenuMsg(items)
	}
}

func (m ElastiCacheModel) fetchReplicationGroups() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.ElastiCacheResources("replication-groups")); ok {
			if groups, ok := cached.([]aws.ReplicationGroupInfo); ok {
				return ReplicationGroupsMsg(groups)
			}
		}

		client, err := aws.NewElastiCacheClient(context.Background(), m.profile)
		if err != nil {
			return ElastiCacheErrorMsg(err)
		}
		groups, err := client.ListReplicationGroups(context.Background())
		if err != nil {
			return ElastiCacheErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.ElastiCacheResources("replication-groups"), groups, cache.TTLElastiCacheResources)
		return ReplicationGroupsMsg(groups)
	}
}

func (m ElastiCacheModel) fetchCacheClusters() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.ElastiCacheResources("cache-clusters")); ok {
			if clusters, ok := cached.([]aws.CacheClusterInfo); ok {
				return CacheClustersMsg(clusters)
			}
		}

		client, err := aws.NewElastiCacheClient(context.Background(), m.profile)
		if err != nil {
			return ElastiCacheErrorMsg(err)
		}
		clusters, err := client.ListCacheClusters(context.Background())
		if err != nil {
			return ElastiCacheErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.ElastiCacheResources("cache-clusters"), clusters, cache.TTLElastiCacheResources)
		return CacheClustersMsg(clusters)
	}
}

func (m ElastiCacheModel) Update(msg tea.Msg) (ElastiCacheModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case ElastiCacheMenuMsg:
		m.list.SetItems(msg)
		m.list.ResetSelected()
		m.state = ElastiCacheStateMenu
		m.updateDelegate()

	case ReplicationGroupsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = elasticacheItem{
				title:       v.ID,
				description: v.Description,
				id:          v.ID,
				category:    "replication-group",
				values:      []string{v.ID, v.Status, v.Engine, v.CacheNodeType, fmt.Sprintf("%d", v.Nodes), v.Description},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = ElastiCacheStateReplicationGroups
		m.updateDelegate()

	case CacheClustersMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			items[i] = elasticacheItem{
				title:       v.ID,
				description: v.Status,
				id:          v.ID,
				category:    "cache-cluster",
				values:      []string{v.ID, v.Status, v.Engine, v.EngineVersion, v.CacheNodeType, v.AZ},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = ElastiCacheStateCacheClusters
		m.updateDelegate()

	case ElastiCacheErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			switch m.state {
			case ElastiCacheStateReplicationGroups:
				m.cache.Delete(m.cacheKeys.ElastiCacheResources("replication-groups"))
				return m, m.fetchReplicationGroups()
			case ElastiCacheStateCacheClusters:
				m.cache.Delete(m.cacheKeys.ElastiCacheResources("cache-clusters"))
				return m, m.fetchCacheClusters()
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(elasticacheItem); ok {
				if m.state == ElastiCacheStateMenu {
					switch item.title {
					case "Replication Groups":
						return m, m.fetchReplicationGroups()
					case "Cache Clusters":
						return m, m.fetchCacheClusters()
					}
				}
			}
		case "backspace", "esc":
			if m.state != ElastiCacheStateMenu {
				return m, m.showMenu()
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *ElastiCacheModel) updateDelegate() {
	d := elasticacheItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          m.styles,
		state:           m.state,
	}
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc
	m.list.SetDelegate(d)
}

func (m ElastiCacheModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state != ElastiCacheStateMenu {
		var columns []Column
		switch m.state {
		case ElastiCacheStateReplicationGroups:
			columns = replicationGroupColumns
		case ElastiCacheStateCacheClusters:
			columns = cacheClusterColumns
		}
		_, header := RenderTableHelpers(m.list, m.styles, columns)
		return header + "\n" + m.list.View()
	}

	return m.list.View()
}

func (m *ElastiCacheModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
