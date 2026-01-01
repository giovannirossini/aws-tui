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

type EFSState int

const (
	EFSStateFileSystems EFSState = iota
	EFSStateMountTargets
)

type efsItem struct {
	title        string
	description  string
	fileSystemId string
	isMount      bool
}

func (i efsItem) Title() string       { return i.title }
func (i efsItem) Description() string { return i.description }
func (i efsItem) FilterValue() string { return i.title }

type EFSModel struct {
	client            *aws.EFSClient
	list              list.Model
	styles            Styles
	state             EFSState
	currentFileSystem string
	width             int
	height            int
	profile           string
	err               error
	cache             *cache.Cache
	cacheKeys         *cache.KeyBuilder
}

type efsItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  EFSState
}

var efsFileSystemColumns = []Column{
	{Title: "File System ID", Width: 0.25},
	{Title: "Name", Width: 0.25},
	{Title: "State", Width: 0.15},
	{Title: "Size", Width: 0.15},
	{Title: "Mount Targets", Width: 0.2},
}

var efsMountTargetColumns = []Column{
	{Title: "Mount Target ID", Width: 0.25},
	{Title: "Subnet ID", Width: 0.25},
	{Title: "IP Address", Width: 0.25},
	{Title: "State", Width: 0.25},
}

func (d efsItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(efsItem)
	if !ok {
		return
	}

	var columns []Column
	if d.state == EFSStateFileSystems {
		columns = efsFileSystemColumns
	} else {
		columns = efsMountTargetColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	var values []string
	if d.state == EFSStateFileSystems {
		parts := strings.Split(i.description, " | ")
		name := ""
		state := ""
		size := ""
		targets := ""
		if len(parts) >= 4 {
			name = strings.TrimPrefix(parts[0], "Name: ")
			state = strings.TrimPrefix(parts[1], "State: ")
			size = strings.TrimPrefix(parts[2], "Size: ")
			targets = strings.TrimPrefix(parts[3], "Targets: ")
		}
		values = []string{
			"📁 " + i.title,
			name,
			state,
			size,
			targets,
		}
	} else {
		if i.title == ".." {
			values = []string{"..", "", "", ""}
		} else {
			parts := strings.Split(i.description, " | ")
			subnetId := ""
			ip := ""
			state := ""
			if len(parts) >= 3 {
				subnetId = strings.TrimPrefix(parts[0], "Subnet: ")
				ip = strings.TrimPrefix(parts[1], "IP: ")
				state = strings.TrimPrefix(parts[2], "State: ")
			}
			values = []string{
				"📍 " + i.title,
				subnetId,
				ip,
				state,
			}
		}
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d efsItemDelegate) Height() int { return 1 }

func NewEFSModel(profile string, styles Styles, appCache *cache.Cache) EFSModel {
	d := efsItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           EFSStateFileSystems,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "EFS File Systems"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(styles.Primary).PaddingLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(2)

	return EFSModel{
		list:      l,
		styles:    styles,
		state:     EFSStateFileSystems,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type EFSFileSystemsMsg []aws.FileSystemInfo
type EFSMountTargetsMsg []aws.MountTargetInfo
type EFSErrorMsg error

func (m EFSModel) Init() tea.Cmd {
	return m.fetchFileSystems()
}

func (m EFSModel) fetchFileSystems() tea.Cmd {
	return func() tea.Msg {
		cacheKey := m.cacheKeys.EFSResources("file-systems")
		if cached, ok := m.cache.Get(cacheKey); ok {
			if fss, ok := cached.([]aws.FileSystemInfo); ok {
				return EFSFileSystemsMsg(fss)
			}
		}

		client, err := aws.NewEFSClient(context.Background(), m.profile)
		if err != nil {
			return EFSErrorMsg(err)
		}
		fss, err := client.ListFileSystems(context.Background())
		if err != nil {
			return EFSErrorMsg(err)
		}

		m.cache.Set(cacheKey, fss, cache.TTLEFSResources)
		return EFSFileSystemsMsg(fss)
	}
}

func (m EFSModel) fetchMountTargets() tea.Cmd {
	return func() tea.Msg {
		cacheKey := m.cacheKeys.EFSMountTargets(m.currentFileSystem)
		if cached, ok := m.cache.Get(cacheKey); ok {
			if targets, ok := cached.([]aws.MountTargetInfo); ok {
				return EFSMountTargetsMsg(targets)
			}
		}

		client, err := aws.NewEFSClient(context.Background(), m.profile)
		if err != nil {
			return EFSErrorMsg(err)
		}
		targets, err := client.ListMountTargets(context.Background(), m.currentFileSystem)
		if err != nil {
			return EFSErrorMsg(err)
		}

		m.cache.Set(cacheKey, targets, cache.TTLEFSResources)
		return EFSMountTargetsMsg(targets)
	}
}

func (m EFSModel) Update(msg tea.Msg) (EFSModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case EFSFileSystemsMsg:
		items := make([]list.Item, len(msg))
		for i, fs := range msg {
			sizeMB := float64(fs.SizeInBytes) / 1024 / 1024
			items[i] = efsItem{
				title:        fs.FileSystemId,
				description:  fmt.Sprintf("Name: %s | State: %s | Size: %.2f MB | Targets: %d", fs.Name, fs.LifeCycleState, sizeMB, fs.NumberOfMountTargets),
				fileSystemId: fs.FileSystemId,
			}
		}
		m.list.SetItems(items)
		m.state = EFSStateFileSystems
		m.list.Title = "EFS File Systems"

		d := efsItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           EFSStateFileSystems,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case EFSMountTargetsMsg:
		items := make([]list.Item, 0)
		items = append(items, efsItem{title: "..", description: "Back"})

		for _, mt := range msg {
			items = append(items, efsItem{
				title:       mt.MountTargetId,
				description: fmt.Sprintf("Subnet: %s | IP: %s | State: %s", mt.SubnetId, mt.IpAddress, mt.LifeCycleState),
				isMount:     true,
			})
		}
		m.list.SetItems(items)
		m.state = EFSStateMountTargets
		m.list.Title = fmt.Sprintf("EFS Mount Targets: %s", m.currentFileSystem)

		d := efsItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           EFSStateMountTargets,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case EFSErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			if m.state == EFSStateFileSystems {
				m.cache.Delete(m.cacheKeys.EFSResources("file-systems"))
				return m, m.fetchFileSystems()
			} else {
				m.cache.Delete(m.cacheKeys.EFSMountTargets(m.currentFileSystem))
				return m, m.fetchMountTargets()
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(efsItem); ok {
				if item.title == ".." {
					m.state = EFSStateFileSystems
					return m, m.fetchFileSystems()
				}
				if !item.isMount {
					m.currentFileSystem = item.fileSystemId
					m.state = EFSStateMountTargets
					return m, m.fetchMountTargets()
				}
			}
		case "backspace", "esc":
			if m.state == EFSStateMountTargets {
				m.state = EFSStateFileSystems
				return m, m.fetchFileSystems()
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m EFSModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	return m.renderHeader() + "\n" + m.list.View()
}

func (m EFSModel) renderHeader() string {
	var columns []Column
	if m.state == EFSStateFileSystems {
		columns = efsFileSystemColumns
	} else {
		columns = efsMountTargetColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header
}

func (m *EFSModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
