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

type ECRState int

const (
	ECRStateRepositories ECRState = iota
	ECRStateImages
)

type ecrItem struct {
	title       string
	description string
	isRepo      bool
	repository  string
}

func (i ecrItem) Title() string       { return i.title }
func (i ecrItem) Description() string { return i.description }
func (i ecrItem) FilterValue() string { return i.title }

type ECRModel struct {
	client            *aws.ECRClient
	list              list.Model
	styles            Styles
	state             ECRState
	currentRepository string
	width             int
	height            int
	profile           string
	err               error
	cache             *cache.Cache
	cacheKeys         *cache.KeyBuilder
}

type ecrItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  ECRState
}

var ecrRepoColumns = []Column{
	{Title: "Repository Name", Width: 0.3},
	{Title: "Created At", Width: 0.2},
	{Title: "URI", Width: 0.5},
}

var ecrImageColumns = []Column{
	{Title: "Tags", Width: 0.4},
	{Title: "Pushed At", Width: 0.2},
	{Title: "Size", Width: 0.1},
	{Title: "Digest", Width: 0.3},
}

func (d ecrItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ecrItem)
	if !ok {
		return
	}

	var columns []Column
	if d.state == ECRStateRepositories {
		columns = ecrRepoColumns
	} else {
		columns = ecrImageColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	var values []string
	if d.state == ECRStateRepositories {
		parts := strings.Split(i.description, " | URI: ")
		createdAt := ""
		uri := ""
		if len(parts) == 2 {
			createdAt = parts[0]
			uri = parts[1]
		}
		values = []string{
			"📦 " + i.title,
			createdAt,
			uri,
		}
	} else {
		if i.title == ".." {
			values = []string{"..", "", "", ""}
		} else {
			parts := strings.Split(i.description, " | ")
			pushedAt := ""
			size := ""
			digest := ""
			if len(parts) == 3 {
				pushedAt = strings.TrimPrefix(parts[0], "Pushed: ")
				size = strings.TrimPrefix(parts[1], "Size: ")
				digest = strings.TrimPrefix(parts[2], "Digest: ")
			}
			values = []string{
				"🖼  " + i.title,
				pushedAt,
				size,
				digest,
			}
		}
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d ecrItemDelegate) Height() int { return 1 }

func NewECRModel(profile string, styles Styles, appCache *cache.Cache) ECRModel {
	d := ecrItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           ECRStateRepositories,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "ECR Repositories"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(styles.Primary).PaddingLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(2)

	return ECRModel{
		list:      l,
		styles:    styles,
		state:     ECRStateRepositories,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type ECRReposMsg []aws.RepositoryInfo
type ECRImagesMsg []aws.ImageInfo
type ECRErrorMsg error

func (m ECRModel) Init() tea.Cmd {
	return m.fetchRepositories()
}

func (m ECRModel) fetchRepositories() tea.Cmd {
	return func() tea.Msg {
		cacheKey := m.cacheKeys.ECRResources("repositories")
		if cached, ok := m.cache.Get(cacheKey); ok {
			if repos, ok := cached.([]aws.RepositoryInfo); ok {
				return ECRReposMsg(repos)
			}
		}

		client, err := aws.NewECRClient(context.Background(), m.profile)
		if err != nil {
			return ECRErrorMsg(err)
		}
		repos, err := client.ListRepositories(context.Background())
		if err != nil {
			return ECRErrorMsg(err)
		}

		m.cache.Set(cacheKey, repos, cache.TTLECRResources)
		return ECRReposMsg(repos)
	}
}

func (m ECRModel) fetchImages() tea.Cmd {
	return func() tea.Msg {
		cacheKey := m.cacheKeys.ECRImages(m.currentRepository)
		if cached, ok := m.cache.Get(cacheKey); ok {
			if images, ok := cached.([]aws.ImageInfo); ok {
				return ECRImagesMsg(images)
			}
		}

		client, err := aws.NewECRClient(context.Background(), m.profile)
		if err != nil {
			return ECRErrorMsg(err)
		}
		images, err := client.ListImages(context.Background(), m.currentRepository)
		if err != nil {
			return ECRErrorMsg(err)
		}

		m.cache.Set(cacheKey, images, cache.TTLECRResources)
		return ECRImagesMsg(images)
	}
}

func (m ECRModel) Update(msg tea.Msg) (ECRModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case ECRReposMsg:
		items := make([]list.Item, len(msg))
		for i, r := range msg {
			items[i] = ecrItem{
				title:       r.Name,
				description: fmt.Sprintf("%s | URI: %s", r.CreatedAt.Format("2006-01-02 15:04"), r.RepositoryUri),
				isRepo:      true,
			}
		}
		m.list.SetItems(items)
		m.state = ECRStateRepositories
		m.list.Title = "ECR Repositories"

		d := ecrItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           ECRStateRepositories,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case ECRImagesMsg:
		items := make([]list.Item, 0)
		items = append(items, ecrItem{title: "..", description: "Back", isRepo: false})

		for _, img := range msg {
			tags := strings.Join(img.Tags, ", ")
			if tags == "" {
				tags = "<untagged>"
			}
			sizeMB := float64(img.Size) / 1024 / 1024
			items = append(items, ecrItem{
				title:       tags,
				description: fmt.Sprintf("Pushed: %s | Size: %.2f MB | Digest: %s", img.PushedAt.Format("2006-01-02 15:04"), sizeMB, img.Digest),
				isRepo:      false,
				repository:  m.currentRepository,
			})
		}
		m.list.SetItems(items)
		m.state = ECRStateImages
		m.list.Title = fmt.Sprintf("ECR Images: %s", m.currentRepository)

		d := ecrItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           ECRStateImages,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case ECRErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			if m.state == ECRStateRepositories {
				m.cache.Delete(m.cacheKeys.ECRResources("repositories"))
				return m, m.fetchRepositories()
			} else {
				m.cache.Delete(m.cacheKeys.ECRImages(m.currentRepository))
				return m, m.fetchImages()
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(ecrItem); ok {
				if item.title == ".." {
					m.state = ECRStateRepositories
					return m, m.fetchRepositories()
				}
				if item.isRepo {
					m.currentRepository = item.title
					m.state = ECRStateImages
					return m, m.fetchImages()
				}
			}
		case "backspace", "esc":
			if m.state == ECRStateImages {
				m.state = ECRStateRepositories
				return m, m.fetchRepositories()
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ECRModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	return m.renderHeader() + "\n" + m.list.View()
}

func (m ECRModel) renderHeader() string {
	var columns []Column
	if m.state == ECRStateRepositories {
		columns = ecrRepoColumns
	} else {
		columns = ecrImageColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header
}

func (m *ECRModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
