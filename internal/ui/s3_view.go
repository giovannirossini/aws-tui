package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

type S3State int

const (
	S3StateBuckets S3State = iota
	S3StateObjects
	S3StateInput
	S3StateConfirmDelete
)

type S3Action int

const (
	S3ActionNone S3Action = iota
	S3ActionCreateBucket
	S3ActionCreateFolder
	S3ActionDeleteBucket
	S3ActionDeleteObject
	S3ActionUploadFile
	S3ActionEditFile
)

type s3Item struct {
	title       string
	description string
	isBucket    bool
	isFolder    bool
	key         string
}

func (i s3Item) Title() string       { return i.title }
func (i s3Item) Description() string { return i.description }
func (i s3Item) FilterValue() string { return i.title }

type S3Model struct {
	client        *aws.S3Client
	list          list.Model
	input         textinput.Model
	styles        Styles
	state         S3State
	action        S3Action
	currentBucket string
	currentPrefix string
	selectedItem  s3Item
	width         int
	height        int
	profile       string
	err           error
	cache         *cache.Cache
	cacheKeys     *cache.KeyBuilder
}

type s3ItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  S3State
}

var s3BucketColumns = []Column{
	{Title: "Bucket Name", Width: 0.7},
	{Title: "Created", Width: 0.3},
}

var s3ObjectColumns = []Column{
	{Title: "Name", Width: 0.6},
	{Title: "Size", Width: 0.15},
	{Title: "Last Modified", Width: 0.25},
}

func (d s3ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(s3Item)
	if !ok {
		return
	}

	var columns []Column
	if d.state == S3StateBuckets {
		columns = s3BucketColumns
	} else {
		columns = s3ObjectColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	var values []string
	if d.state == S3StateBuckets {
		values = []string{
			"🪣 " + i.title,
			i.description,
		}
	} else {
		icon := "📄 "
		if i.isFolder {
			icon = "📁 "
		}

		size := ""
		modified := ""
		if !i.isFolder && i.key != "back" {
			parts := strings.Split(i.description, ", Modified: ")
			if len(parts) == 2 {
				size = strings.TrimPrefix(parts[0], "Size: ")
				modified = parts[1]
			}
		} else if i.isFolder && i.key != "back" {
			size = "Folder"
		}

		values = []string{
			icon + i.title,
			size,
			modified,
		}
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d s3ItemDelegate) Height() int { return 1 }

func NewS3Model(profile string, styles Styles, appCache *cache.Cache) S3Model {
	d := s3ItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           S3StateBuckets,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "S3 Buckets"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false) // Hide the internal list title
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)
	l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(styles.Primary).PaddingLeft(2)
	l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(styles.Muted).PaddingLeft(2)

	ti := textinput.New()
	ti.Placeholder = "Enter name..."
	ti.Focus()

	return S3Model{
		list:      l,
		input:     ti,
		styles:    styles,
		state:     S3StateBuckets,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type S3BucketsMsg []aws.BucketInfo
type S3ObjectsMsg []aws.ObjectInfo
type S3ErrorMsg error
type S3SuccessMsg string

func (m S3Model) Init() tea.Cmd {
	return m.fetchBuckets()
}

func (m S3Model) fetchBuckets() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		if cached, ok := m.cache.Get(m.cacheKeys.S3Buckets()); ok {
			if buckets, ok := cached.([]aws.BucketInfo); ok {
				return S3BucketsMsg(buckets)
			}
		}

		client, err := aws.NewS3Client(context.Background(), m.profile)
		if err != nil {
			return S3ErrorMsg(err)
		}
		buckets, err := client.ListBuckets(context.Background())
		if err != nil {
			return S3ErrorMsg(err)
		}

		// Cache the result
		m.cache.Set(m.cacheKeys.S3Buckets(), buckets, cache.TTLS3Buckets)

		return S3BucketsMsg(buckets)
	}
}

func (m S3Model) fetchObjects() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		cacheKey := m.cacheKeys.S3Objects(m.currentBucket, m.currentPrefix)
		if cached, ok := m.cache.Get(cacheKey); ok {
			if objects, ok := cached.([]aws.ObjectInfo); ok {
				return S3ObjectsMsg(objects)
			}
		}

		client, err := aws.NewS3Client(context.Background(), m.profile)
		if err != nil {
			return S3ErrorMsg(err)
		}
		objects, err := client.ListObjects(context.Background(), m.currentBucket, m.currentPrefix, "/")
		if err != nil {
			return S3ErrorMsg(err)
		}

		// Use shorter TTL for objects as they change more frequently
		ttl := cache.TTLShortS3Objects
		if len(objects) > 100 {
			// Longer TTL for large buckets
			ttl = cache.TTLLongS3Objects
		}
		m.cache.Set(cacheKey, objects, ttl)

		return S3ObjectsMsg(objects)
	}
}

func (m S3Model) createBucket(name string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewS3Client(context.Background(), m.profile)
		if err != nil {
			return S3ErrorMsg(err)
		}
		err = client.CreateBucket(context.Background(), name, "")
		if err != nil {
			return S3ErrorMsg(err)
		}
		
		// Invalidate buckets cache
		m.cache.Delete(m.cacheKeys.S3Buckets())
		
		return S3SuccessMsg("Bucket created")
	}
}

func (m S3Model) deleteBucket(name string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewS3Client(context.Background(), m.profile)
		if err != nil {
			return S3ErrorMsg(err)
		}
		err = client.DeleteBucket(context.Background(), name)
		if err != nil {
			return S3ErrorMsg(err)
		}
		
		// Invalidate caches
		m.cache.Delete(m.cacheKeys.S3Buckets())
		m.cache.DeletePrefix(m.cacheKeys.S3BucketPrefix(name))
		
		return S3SuccessMsg("Bucket deleted")
	}
}

func (m S3Model) createFolder(name string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewS3Client(context.Background(), m.profile)
		if err != nil {
			return S3ErrorMsg(err)
		}
		err = client.CreateFolder(context.Background(), m.currentBucket, m.currentPrefix+name)
		if err != nil {
			return S3ErrorMsg(err)
		}
		
		// Invalidate current prefix cache
		m.cache.Delete(m.cacheKeys.S3Objects(m.currentBucket, m.currentPrefix))
		
		return S3SuccessMsg("Folder created")
	}
}

func (m S3Model) deleteObject(key string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewS3Client(context.Background(), m.profile)
		if err != nil {
			return S3ErrorMsg(err)
		}
		err = client.DeleteObject(context.Background(), m.currentBucket, key)
		if err != nil {
			return S3ErrorMsg(err)
		}
		
		// Invalidate current prefix cache
		m.cache.Delete(m.cacheKeys.S3Objects(m.currentBucket, m.currentPrefix))
		
		return S3SuccessMsg("Object deleted")
	}
}

func (m S3Model) uploadFile(localPath string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewS3Client(context.Background(), m.profile)
		if err != nil {
			return S3ErrorMsg(err)
		}
		key := m.currentPrefix + strings.Split(localPath, "/")[len(strings.Split(localPath, "/"))-1]
		err = client.UploadFile(context.Background(), m.currentBucket, key, localPath)
		if err != nil {
			return S3ErrorMsg(err)
		}
		
		// Invalidate current prefix cache
		m.cache.Delete(m.cacheKeys.S3Objects(m.currentBucket, m.currentPrefix))
		
		return S3SuccessMsg("File uploaded")
	}
}

func (m S3Model) Update(msg tea.Msg) (S3Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case S3BucketsMsg:
		items := make([]list.Item, len(msg))
		for i, b := range msg {
			items[i] = s3Item{
				title:       b.Name,
				description: fmt.Sprintf("%s", b.CreationDate.Format("2006-01-02")),
				isBucket:    true,
			}
		}
		m.list.SetItems(items)
		m.state = S3StateBuckets
		
		// Update delegate state for tabular rendering
		d := s3ItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           S3StateBuckets,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case S3ObjectsMsg:
		items := make([]list.Item, 0)

		// Add "back" item if not at root
		if m.currentPrefix != "" {
			items = append(items, s3Item{title: "..", description: "Back", isFolder: true, key: "back"})
		}

		for _, o := range msg {
			desc := fmt.Sprintf("Size: %d bytes, Modified: %s", o.Size, o.LastModified.Format("2006-01-02 15:04"))
			if o.IsFolder {
				desc = "Folder"
			}
			items = append(items, s3Item{
				title:       o.Key,
				description: desc,
				isFolder:    o.IsFolder,
				key:         o.Key,
			})
		}
		m.list.SetItems(items)
		m.state = S3StateObjects
		
		// Update delegate state for tabular rendering
		d := s3ItemDelegate{
			DefaultDelegate: list.NewDefaultDelegate(),
			styles:          m.styles,
			state:           S3StateObjects,
		}
		d.Styles.SelectedTitle = m.styles.ListSelectedTitle
		d.Styles.SelectedDesc = m.styles.ListSelectedDesc
		m.list.SetDelegate(d)

	case S3SuccessMsg:
		m.err = nil
		m.state = S3StateBuckets
		if m.currentBucket != "" {
			m.state = S3StateObjects
		}
		if m.state == S3StateBuckets || m.action == S3ActionCreateBucket || m.action == S3ActionDeleteBucket {
			return m, m.fetchBuckets()
		}
		return m, m.fetchObjects()

	case S3ErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		if m.state == S3StateInput {
			switch msg.String() {
			case "enter":
				name := m.input.Value()
				if name == "" {
					m.state = S3StateBuckets
					if m.currentBucket != "" {
						m.state = S3StateObjects
					}
					return m, nil
				}
				var actionCmd tea.Cmd
				if m.action == S3ActionCreateBucket {
					actionCmd = m.createBucket(name)
				} else if m.action == S3ActionCreateFolder {
					actionCmd = m.createFolder(name)
				} else if m.action == S3ActionUploadFile {
					actionCmd = m.uploadFile(name)
				}
				m.input.Reset()
				return m, actionCmd
			case "esc":
				m.state = S3StateBuckets
				if m.currentBucket != "" {
					m.state = S3StateObjects
				}
				return m, nil
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		if m.state == S3StateConfirmDelete {
			switch msg.String() {
			case "y", "Y":
				var actionCmd tea.Cmd
				if m.action == S3ActionDeleteBucket {
					actionCmd = m.deleteBucket(m.selectedItem.title)
				} else if m.action == S3ActionDeleteObject {
					actionCmd = m.deleteObject(m.selectedItem.key)
				}
				return m, actionCmd
			default:
				m.state = S3StateBuckets
				if m.currentBucket != "" {
					m.state = S3StateObjects
				}
				return m, nil
			}
		}

		switch msg.String() {
		case "r": // Manual refresh for S3
			if m.state == S3StateBuckets {
				m.cache.Delete(m.cacheKeys.S3Buckets())
				return m, m.fetchBuckets()
			} else if m.state == S3StateObjects {
				m.cache.Delete(m.cacheKeys.S3Objects(m.currentBucket, m.currentPrefix))
				return m, m.fetchObjects()
			}
		case "e":
			// Handled in main model to use tea.ExecProcess
			return m, nil
		case "n":
			if m.state == S3StateBuckets {
				m.state = S3StateInput
				m.action = S3ActionCreateBucket
				m.input.Placeholder = "Bucket name"
				m.input.Focus()
				return m, nil
			} else if m.state == S3StateObjects {
				m.state = S3StateInput
				m.action = S3ActionCreateFolder
				m.input.Placeholder = "Folder name"
				m.input.Focus()
				return m, nil
			}
		case "u":
			if m.state == S3StateObjects {
				m.state = S3StateInput
				m.action = S3ActionUploadFile
				m.input.Placeholder = "Local file path"
				m.input.Focus()
				return m, nil
			}
		case "d":
			if item, ok := m.list.SelectedItem().(s3Item); ok {
				if item.key == "back" {
					return m, nil
				}
				m.selectedItem = item
				m.state = S3StateConfirmDelete
				if item.isBucket {
					m.action = S3ActionDeleteBucket
				} else {
					m.action = S3ActionDeleteObject
				}
				return m, nil
			}
		case "enter":
			if item, ok := m.list.SelectedItem().(s3Item); ok {
				if item.isBucket {
					m.currentBucket = item.title
					m.currentPrefix = ""
					m.state = S3StateObjects
					return m, m.fetchObjects()
				}
				if item.isFolder {
					if item.key == "back" {
						parts := strings.Split(strings.TrimSuffix(m.currentPrefix, "/"), "/")
						if len(parts) <= 1 {
							m.currentPrefix = ""
						} else {
							m.currentPrefix = strings.Join(parts[:len(parts)-1], "/") + "/"
						}
					} else {
						m.currentPrefix = item.key
					}
					return m, m.fetchObjects()
				}
			}
		case "backspace", "esc":
			if m.state == S3StateObjects {
				if m.currentPrefix != "" {
					parts := strings.Split(strings.TrimSuffix(m.currentPrefix, "/"), "/")
					if len(parts) <= 1 {
						m.currentPrefix = ""
					} else {
						m.currentPrefix = strings.Join(parts[:len(parts)-1], "/") + "/"
					}
					return m, m.fetchObjects()
				} else {
					m.state = S3StateBuckets
					m.currentBucket = ""
					return m, m.fetchBuckets()
				}
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m S3Model) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	switch m.state {
	case S3StateInput:
		header := m.renderHeader()
		return RenderOverlay(header+"\n"+m.list.View(), m.styles.Popup.Width(40).Render(fmt.Sprintf(
			" %s\n\n %s\n\n %s",
			lipgloss.NewStyle().Foreground(m.styles.Primary).Render(m.input.Placeholder),
			m.input.View(),
			m.styles.StatusMuted.Render("(esc to cancel)"),
		)), m.width, m.height)
	case S3StateConfirmDelete:
		header := m.renderHeader()
		return RenderOverlay(header+"\n"+m.list.View(), m.styles.Popup.Width(40).BorderForeground(ErrorColor).Render(fmt.Sprintf(
			" %s\n\n %s %s\n\n %s",
			m.styles.Error.Bold(true).Render("⚠ Confirm Deletion"),
			"Are you sure you want to delete",
			lipgloss.NewStyle().Foreground(m.styles.Primary).Bold(true).Render(m.selectedItem.title),
			m.styles.StatusMuted.Render("(y/n)"),
		)), m.width, m.height)
	default:
		return m.renderHeader() + "\n" + m.list.View()
	}
}

func (m S3Model) renderHeader() string {
	var columns []Column
	if m.state == S3StateBuckets {
		columns = s3BucketColumns
	} else {
		columns = s3ObjectColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header
}

var lastTmpPath string
var lastTmpModTime time.Time

func (m S3Model) getEditCommand(key string) *exec.Cmd {
	client, err := aws.NewS3Client(context.Background(), m.profile)
	if err != nil {
		return nil
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "aws-tui-edit-*")
	if err != nil {
		return nil
	}
	lastTmpPath = tmpFile.Name()
	tmpFile.Close()

	// Download
	err = client.DownloadFile(context.Background(), m.currentBucket, key, lastTmpPath)
	if err != nil {
		return nil
	}

	// Get initial mod time
	if info, err := os.Stat(lastTmpPath); err == nil {
		lastTmpModTime = info.ModTime()
	}

	// Open editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	return exec.Command(editor, lastTmpPath)
}

func (m S3Model) uploadEditedFile(key string) tea.Msg {
	if lastTmpPath == "" {
		return nil
	}
	defer os.Remove(lastTmpPath)

	// Check if file was modified
	if info, err := os.Stat(lastTmpPath); err == nil {
		if info.ModTime().Equal(lastTmpModTime) {
			return S3SuccessMsg("No changes made, skipping upload")
		}
	}

	client, err := aws.NewS3Client(context.Background(), m.profile)
	if err != nil {
		return S3ErrorMsg(err)
	}

	err = client.UploadFile(context.Background(), m.currentBucket, key, lastTmpPath)
	if err != nil {
		return S3ErrorMsg(err)
	}

	return S3SuccessMsg("File edited successfully")
}

func (m *S3Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
