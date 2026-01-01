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

type BackupState int

const (
	BackupStateMenu BackupState = iota
	BackupStatePlans
	BackupStateJobs
)

type backupItem struct {
	title       string
	description string
	state       BackupState
}

func (i backupItem) Title() string       { return i.title }
func (i backupItem) Description() string { return i.description }
func (i backupItem) FilterValue() string { return i.title }

type BackupModel struct {
	client    *aws.BackupClient
	list      list.Model
	styles    Styles
	state     BackupState
	width     int
	height    int
	profile   string
	err       error
	cache     *cache.Cache
	cacheKeys *cache.KeyBuilder
}

type backupItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  BackupState
}

var backupPlanColumns = []Column{
	{Title: "Plan Name", Width: 0.4},
	{Title: "Plan ID", Width: 0.3},
	{Title: "Created At", Width: 0.3},
}

var backupJobColumns = []Column{
	{Title: "Job ID", Width: 0.2},
	{Title: "Resource Type", Width: 0.2},
	{Title: "State", Width: 0.15},
	{Title: "Size", Width: 0.15},
	{Title: "Created At", Width: 0.3},
}

func (d backupItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(backupItem)
	if !ok {
		return
	}

	var columns []Column
	switch d.state {
	case BackupStatePlans:
		columns = backupPlanColumns
	case BackupStateJobs:
		columns = backupJobColumns
	default:
		d.DefaultDelegate.Render(w, m, index, listItem)
		return
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	var values []string
	if d.state == BackupStatePlans {
		parts := strings.Split(i.description, " | ")
		planId := ""
		createdAt := ""
		if len(parts) >= 2 {
			planId = strings.TrimPrefix(parts[0], "ID: ")
			createdAt = strings.TrimPrefix(parts[1], "Created: ")
		}
		values = []string{
			"📋 " + i.title,
			planId,
			createdAt,
		}
	} else if d.state == BackupStateJobs {
		parts := strings.Split(i.description, " | ")
		resType := ""
		state := ""
		size := ""
		createdAt := ""
		if len(parts) >= 4 {
			resType = strings.TrimPrefix(parts[0], "Type: ")
			state = strings.TrimPrefix(parts[1], "State: ")
			size = strings.TrimPrefix(parts[2], "Size: ")
			createdAt = strings.TrimPrefix(parts[3], "Created: ")
		}
		values = []string{
			"⚙️ " + i.title,
			resType,
			state,
			size,
			createdAt,
		}
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d backupItemDelegate) Height() int {
	if d.state == BackupStateMenu {
		return 2
	}
	return 1
}

func NewBackupModel(profile string, styles Styles, appCache *cache.Cache) BackupModel {
	d := backupItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           BackupStateMenu,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "AWS Backup"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(styles.Primary).PaddingLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(2)

	m := BackupModel{
		list:      l,
		styles:    styles,
		state:     BackupStateMenu,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
	m.setMenu()
	return m
}

func (m *BackupModel) setMenu() {
	items := []list.Item{
		backupItem{title: "Backup Plans", description: "View configured backup plans", state: BackupStatePlans},
		backupItem{title: "Backup Jobs", description: "View recent backup jobs", state: BackupStateJobs},
	}
	m.list.SetItems(items)
	m.state = BackupStateMenu
	
	d := backupItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          m.styles,
		state:           BackupStateMenu,
	}
	d.Styles.SelectedTitle = m.styles.ListSelectedTitle
	d.Styles.SelectedDesc = m.styles.ListSelectedDesc
	m.list.SetDelegate(d)
}

type BackupPlansMsg []aws.BackupPlanInfo
type BackupJobsMsg []aws.BackupJobInfo
type BackupErrorMsg error

func (m BackupModel) Init() tea.Cmd {
	return nil
}

func (m BackupModel) fetchPlans() tea.Cmd {
	return func() tea.Msg {
		cacheKey := m.cacheKeys.BackupResources("plans")
		if cached, ok := m.cache.Get(cacheKey); ok {
			if plans, ok := cached.([]aws.BackupPlanInfo); ok {
				return BackupPlansMsg(plans)
			}
		}

		client, err := aws.NewBackupClient(context.Background(), m.profile)
		if err != nil {
			return BackupErrorMsg(err)
		}
		plans, err := client.ListBackupPlans(context.Background())
		if err != nil {
			return BackupErrorMsg(err)
		}

		m.cache.Set(cacheKey, plans, cache.TTLBackupResources)
		return BackupPlansMsg(plans)
	}
}

func (m BackupModel) fetchJobs() tea.Cmd {
	return func() tea.Msg {
		cacheKey := m.cacheKeys.BackupResources("jobs")
		if cached, ok := m.cache.Get(cacheKey); ok {
			if jobs, ok := cached.([]aws.BackupJobInfo); ok {
				return BackupJobsMsg(jobs)
			}
		}

		client, err := aws.NewBackupClient(context.Background(), m.profile)
		if err != nil {
			return BackupErrorMsg(err)
		}
		jobs, err := client.ListBackupJobs(context.Background())
		if err != nil {
			return BackupErrorMsg(err)
		}

		m.cache.Set(cacheKey, jobs, cache.TTLBackupResources)
		return BackupJobsMsg(jobs)
	}
}

func (m BackupModel) Update(msg tea.Msg) (BackupModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case BackupPlansMsg:
		items := make([]list.Item, len(msg))
		for i, p := range msg {
			items[i] = backupItem{
				title:       p.BackupPlanName,
				description: fmt.Sprintf("ID: %s | Created: %s", p.BackupPlanId, p.CreationDate.Format("2006-01-02")),
			}
		}
		m.list.SetItems(items)
		m.state = BackupStatePlans
		m.list.Title = "Backup Plans"

		d := backupItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           BackupStatePlans,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case BackupJobsMsg:
		items := make([]list.Item, len(msg))
		for i, j := range msg {
			sizeMB := float64(j.BackupSizeInBytes) / 1024 / 1024
			items[i] = backupItem{
				title:       j.BackupJobId,
				description: fmt.Sprintf("Type: %s | State: %s | Size: %.2f MB | Created: %s", j.ResourceType, j.State, sizeMB, j.CreationDate.Format("2006-01-02 15:04")),
			}
		}
		m.list.SetItems(items)
		m.state = BackupStateJobs
		m.list.Title = "Backup Jobs"

		d := backupItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           BackupStateJobs,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case BackupErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			if m.state == BackupStatePlans {
				m.cache.Delete(m.cacheKeys.BackupResources("plans"))
				return m, m.fetchPlans()
			} else if m.state == BackupStateJobs {
				m.cache.Delete(m.cacheKeys.BackupResources("jobs"))
				return m, m.fetchJobs()
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(backupItem); ok {
				if m.state == BackupStateMenu {
					if item.state == BackupStatePlans {
						return m, m.fetchPlans()
					} else if item.state == BackupStateJobs {
						return m, m.fetchJobs()
					}
				}
			}
		case "backspace", "esc":
			if m.state != BackupStateMenu {
				m.setMenu()
				return m, nil
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m BackupModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state == BackupStateMenu {
		return m.list.View()
	}

	return m.renderHeader() + "\n" + m.list.View()
}

func (m BackupModel) renderHeader() string {
	var columns []Column
	if m.state == BackupStatePlans {
		columns = backupPlanColumns
	} else if m.state == BackupStateJobs {
		columns = backupJobColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header
}

func (m *BackupModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
