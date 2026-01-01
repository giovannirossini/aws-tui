package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

type SMState int

const (
	SMStateSecrets SMState = iota
	SMStateValue
)

type smItem struct {
	title       string
	description string
	arn         string
	values      []string
}

func (i smItem) Title() string       { return i.title }
func (i smItem) Description() string { return i.description }
func (i smItem) FilterValue() string { return i.title + " " + i.description + " " + i.arn }

type SMModel struct {
	client         *aws.SecretsManagerClient
	list           list.Model
	styles         Styles
	state          SMState
	width          int
	height         int
	profile        string
	err            error
	cache          *cache.Cache
	cacheKeys      *cache.KeyBuilder
	selectedValue  string
	selectedSecret string
}

type smItemDelegate struct {
	list.DefaultDelegate
	styles Styles
	state  SMState
}

var smSecretColumns = []Column{
	{Title: "Secret Name", Width: 0.4},
	{Title: "Last Changed", Width: 0.2},
	{Title: "Description", Width: 0.4},
}

func (d smItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(smItem)
	if !ok {
		return
	}

	var columns []Column
	switch d.state {
	case SMStateSecrets:
		columns = smSecretColumns
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, columns)
	isSelected := index == m.Index()

	RenderTableRow(w, m, d.styles, colStyles, i.values, isSelected)
}

func (d smItemDelegate) Height() int {
	return 1
}

func NewSMModel(profile string, styles Styles, appCache *cache.Cache) SMModel {
	d := smItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
		state:           SMStateSecrets,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	d.Styles.SelectedDesc = styles.ListSelectedDesc

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Secrets Manager"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return SMModel{
		list:      l,
		styles:    styles,
		state:     SMStateSecrets,
		profile:   profile,
		cache:     appCache,
		cacheKeys: cache.NewKeyBuilder(profile),
	}
}

type SMSecretsMsg []aws.SecretInfo
type SMSecretValueMsg string
type SMErrorMsg error

func (m SMModel) Init() tea.Cmd {
	return m.fetchSecrets()
}

func (m SMModel) fetchSecrets() tea.Cmd {
	return func() tea.Msg {
		if cached, ok := m.cache.Get(m.cacheKeys.SMResources("secrets")); ok {
			if secrets, ok := cached.([]aws.SecretInfo); ok {
				return SMSecretsMsg(secrets)
			}
		}

		client, err := aws.NewSecretsManagerClient(context.Background(), m.profile)
		if err != nil {
			return SMErrorMsg(err)
		}
		secrets, err := client.ListSecrets(context.Background())
		if err != nil {
			return SMErrorMsg(err)
		}
		m.cache.Set(m.cacheKeys.SMResources("secrets"), secrets, cache.TTLSMResources)
		return SMSecretsMsg(secrets)
	}
}

func (m SMModel) fetchSecretValue(secretID string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewSecretsManagerClient(context.Background(), m.profile)
		if err != nil {
			return SMErrorMsg(err)
		}
		value, err := client.GetSecretValue(context.Background(), secretID)
		if err != nil {
			return SMErrorMsg(err)
		}
		return SMSecretValueMsg(value)
	}
}

func (m SMModel) Update(msg tea.Msg) (SMModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(GetInnerListSize(msg.Width, msg.Height))

	case SMSecretsMsg:
		items := make([]list.Item, len(msg))
		for i, v := range msg {
			lastChanged := "-"
			if v.LastChanged != nil {
				lastChanged = v.LastChanged.Format("2006-01-02 15:04")
			}
			items[i] = smItem{
				title:       v.Name,
				description: v.Description,
				arn:         v.ARN,
				values:      []string{v.Name, lastChanged, v.Description},
			}
		}
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.state = SMStateSecrets

	case SMSecretValueMsg:
		m.selectedValue = string(msg)
		m.state = SMStateValue

	case SMErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		switch msg.String() {
		case "r":
			if m.state == SMStateSecrets {
				m.cache.Delete(m.cacheKeys.SMResources("secrets"))
				return m, m.fetchSecrets()
			} else if m.state == SMStateValue {
				return m, m.fetchSecretValue(m.selectedSecret)
			}
		case "enter":
			if m.state == SMStateSecrets {
				if item, ok := m.list.SelectedItem().(smItem); ok {
					m.selectedSecret = item.title
					return m, m.fetchSecretValue(m.selectedSecret)
				}
			}
		case "backspace", "esc":
			if m.state == SMStateValue {
				m.state = SMStateSecrets
				return m, nil
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SMModel) highlightSecret(content string) string {
	lexer := lexers.Analyse(content)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// If it's JSON, force JSON lexer for better results
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(content), &jsonObj); err == nil {
		lexer = lexers.Get("json")
		if pretty, err := json.MarshalIndent(jsonObj, "", "  "); err == nil {
			content = string(pretty)
		}
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

func (m SMModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	if m.state == SMStateValue {
		displayValue := m.highlightSecret(m.selectedValue)

		return lipgloss.NewStyle().
			Width(m.width - InnerContentWidthOffset).
			Padding(1, 2).
			Render(displayValue)
	}

	var columns []Column
	switch m.state {
	case SMStateSecrets:
		columns = smSecretColumns
	}
	_, header := RenderTableHelpers(m.list, m.styles, columns)
	return header + "\n" + m.list.View()
}

func (m *SMModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(GetInnerListSize(width, height))
}
