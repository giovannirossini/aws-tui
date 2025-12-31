package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/giovannirossini/aws-tui/internal/aws"
	"github.com/giovannirossini/aws-tui/internal/cache"
)

type IAMState int

const (
	IAMStateLoading IAMState = iota
	IAMStateUsers
	IAMStateActions
	IAMStateInput
	IAMStateConfirmDelete
	IAMStateConfirmConsoleToggle
)

type IAMAction int

const (
	IAMActionNone IAMAction = iota
	IAMActionCreateUser
	IAMActionDeleteUser
	IAMActionResetPassword
	IAMActionEnableConsole
	IAMActionDisableConsole
)

type iamActionItem struct {
	title string
	key   string
}

func (i iamActionItem) Title() string       { return i.title }
func (i iamActionItem) Description() string { return "" }
func (i iamActionItem) FilterValue() string { return i.title }

type iamItem struct {
	userName         string
	userID           string
	arn              string
	path             string
	createDate       string
	passwordLastUsed string
}

func (i iamItem) Title() string { return i.userName }
func (i iamItem) Description() string {
	desc := fmt.Sprintf("ID: %s | Path: %s\nARN: %s\nCreated: %s", i.userID, i.path, i.arn, i.createDate)
	if i.passwordLastUsed != "" {
		desc += " | Last Login: " + i.passwordLastUsed
	}
	return desc
}
func (i iamItem) FilterValue() string { return i.userName }

type iamItemDelegate struct {
	list.DefaultDelegate
	styles Styles
}

var iamColumns = []Column{
	{Title: "Username", Width: 0.18},
	{Title: "User ID", Width: 0.15},
	{Title: "Path", Width: 0.08},
	{Title: "Last Login", Width: 0.14},
	{Title: "Created", Width: 0.14},
	{Title: "Arn", Width: 0.31},
}

func (d iamItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(iamItem)
	if !ok {
		return
	}

	colStyles, _ := RenderTableHelpers(m, d.styles, iamColumns)
	isSelected := index == m.Index()

	lastLogin := i.passwordLastUsed
	if lastLogin == "" {
		lastLogin = "Never"
	}

	values := []string{
		"👤 " + i.userName,
		i.userID,
		i.path,
		lastLogin,
		i.createDate,
		i.arn,
	}

	RenderTableRow(w, m, d.styles, colStyles, values, isSelected)
}

func (d iamItemDelegate) Height() int { return 1 }

type actionDelegate struct {
	styles Styles
}

func (d actionDelegate) Height() int { return 1 }
func (d actionDelegate) Spacing() int { return 0 }
func (d actionDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d actionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(iamActionItem)
	if !ok {
		return
	}

	str := i.title

	if index == m.Index() {
		str = d.styles.SelectedMenuItem.Render("➜ " + str)
	} else {
		str = d.styles.MenuItem.Render("  " + str)
	}

	fmt.Fprint(w, str)
}

type IAMModel struct {
	list          list.Model
	actionList    list.Model
	input         textinput.Model
	styles        Styles
	state         IAMState
	action        IAMAction
	selectedUser  iamItem
	userDetail    *aws.IAMUserInfo
	userKeys      []aws.AccessKeyInfo
	width         int
	height        int
	profile       string
	err           error
	cache         *cache.Cache
	cacheKeys     *cache.KeyBuilder
}

func NewIAMModel(profile string, styles Styles, appCache *cache.Cache) IAMModel {
	d := iamItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		styles:          styles,
	}
	d.Styles.SelectedTitle = styles.ListSelectedTitle
	
	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "IAM Users"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false) // Hide the internal list title
	l.Styles.Title = styles.AppTitle.Copy().Background(styles.Squid)

	ad := actionDelegate{styles: styles}
	al := list.New([]list.Item{}, ad, 36, 0)
	al.SetShowStatusBar(false)
	al.SetShowHelp(false)
	al.SetShowTitle(false)
	al.SetShowPagination(false)
	al.KeyMap.Quit.SetEnabled(false) // Don't let the list handle quit, let the model do it

	ti := textinput.New()
	ti.Placeholder = "Username..."
	ti.Focus()

	return IAMModel{
		list:       l,
		actionList: al,
		input:      ti,
		styles:     styles,
		state:      IAMStateLoading,
		profile:    profile,
		cache:      appCache,
		cacheKeys:  cache.NewKeyBuilder(profile),
	}
}

type IAMUsersMsg []aws.IAMUserInfo
type IAMUserDetailsMsg struct {
	Info *aws.IAMUserInfo
	Keys []aws.AccessKeyInfo
}
type IAMErrorMsg error
type IAMSuccessMsg string

func (m IAMModel) Init() tea.Cmd {
	return m.fetchUsers()
}

func (m IAMModel) fetchUsers() tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		if cached, ok := m.cache.Get(m.cacheKeys.IAMUsers()); ok {
			if users, ok := cached.([]aws.IAMUserInfo); ok {
				return IAMUsersMsg(users)
			}
		}

		client, err := aws.NewIAMClient(context.Background(), m.profile)
		if err != nil {
			return IAMErrorMsg(err)
		}
		users, err := client.ListUsers(context.Background())
		if err != nil {
			return IAMErrorMsg(err)
		}

		// Cache the result
		m.cache.Set(m.cacheKeys.IAMUsers(), users, cache.TTLIAMUsers)

		return IAMUsersMsg(users)
	}
}

func (m IAMModel) fetchUserDetails(userName string) tea.Cmd {
	return func() tea.Msg {
		// Check cache first
		cacheKey := m.cacheKeys.IAMUserDetails(userName)
		if cached, ok := m.cache.Get(cacheKey); ok {
			if details, ok := cached.(IAMUserDetailsMsg); ok {
				return details
			}
		}

		client, err := aws.NewIAMClient(context.Background(), m.profile)
		if err != nil {
			return IAMErrorMsg(err)
		}
		info, keys, err := client.GetUserDetails(context.Background(), userName)
		if err != nil {
			return IAMErrorMsg(err)
		}

		result := IAMUserDetailsMsg{Info: info, Keys: keys}
		
		// Cache the result
		m.cache.Set(cacheKey, result, cache.TTLIAMUserDetails)

		return result
	}
}

func (m IAMModel) resetPassword(userName, password string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewIAMClient(context.Background(), m.profile)
		if err != nil {
			return IAMErrorMsg(err)
		}
		err = client.UpdateLoginProfile(context.Background(), userName, password)
		if err != nil {
			return IAMErrorMsg(err)
		}
		
		// Invalidate user details cache
		m.cache.Delete(m.cacheKeys.IAMUserDetails(userName))
		
		return IAMSuccessMsg("Password reset successfully")
	}
}

func (m IAMModel) toggleConsoleAccess(userName string, enable bool) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewIAMClient(context.Background(), m.profile)
		if err != nil {
			return IAMErrorMsg(err)
		}
		if enable {
			// Using a default temporary password for enabling
			err = client.CreateLoginProfile(context.Background(), userName, "TempPass123!")
		} else {
			err = client.DeleteLoginProfile(context.Background(), userName)
		}
		if err != nil {
			return IAMErrorMsg(err)
		}
		
		// Invalidate user details cache
		m.cache.Delete(m.cacheKeys.IAMUserDetails(userName))
		
		msg := "Console access disabled"
		if enable {
			msg = "Console access enabled (TempPass123!)"
		}
		return IAMSuccessMsg(msg)
	}
}

func (m IAMModel) createUser(name string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewIAMClient(context.Background(), m.profile)
		if err != nil {
			return IAMErrorMsg(err)
		}
		err = client.CreateUser(context.Background(), name)
		if err != nil {
			return IAMErrorMsg(err)
		}
		
		// Invalidate users cache
		m.cache.Delete(m.cacheKeys.IAMUsers())
		
		return IAMSuccessMsg("User created")
	}
}

func (m IAMModel) deleteUser(name string) tea.Cmd {
	return func() tea.Msg {
		client, err := aws.NewIAMClient(context.Background(), m.profile)
		if err != nil {
			return IAMErrorMsg(err)
		}
		err = client.DeleteUser(context.Background(), name)
		if err != nil {
			return IAMErrorMsg(err)
		}
		
		// Invalidate caches
		m.cache.Delete(m.cacheKeys.IAMUsers())
		m.cache.Delete(m.cacheKeys.IAMUserDetails(name))
		
		return IAMSuccessMsg("User deleted")
	}
}

func (m *IAMModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width-4, height-10) // Extra line for header
}

func (m IAMModel) Update(msg tea.Msg) (IAMModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case IAMUsersMsg:
		items := make([]list.Item, len(msg))
		for i, u := range msg {
			lastUsed := ""
			if u.PasswordLastUsed != nil {
				lastUsed = u.PasswordLastUsed.Format("2006-01-02 15:04")
			}
			items[i] = iamItem{
				userName:         u.UserName,
				userID:           u.UserID,
				arn:              u.Arn,
				path:             u.Path,
				createDate:       u.CreateDate.Format("2006-01-02 15:04"),
				passwordLastUsed: lastUsed,
			}
		}
		m.list.SetItems(items)
		m.state = IAMStateUsers

	case IAMUserDetailsMsg:
		m.userDetail = msg.Info
		m.userKeys = msg.Keys
		
		actions := []list.Item{
			iamActionItem{title: "Reset Password", key: "reset"},
		}
		if m.userDetail.PasswordExists {
			actions = append(actions, iamActionItem{title: "Disable Console Access", key: "disable_console"})
		} else {
			actions = append(actions, iamActionItem{title: "Enable Console Access", key: "enable_console"})
		}
		actions = append(actions, iamActionItem{title: "Delete User", key: "delete"})
		
		m.actionList.SetItems(actions)
		m.actionList.SetSize(36, len(actions))
		return m, nil

	case IAMSuccessMsg:
		m.err = nil
		if m.action == IAMActionResetPassword || m.action == IAMActionEnableConsole || m.action == IAMActionDisableConsole {
			m.action = IAMActionNone
			// Stay in Actions state while refreshing details
			return m, m.fetchUserDetails(m.selectedUser.userName)
		}
		// Return to user list and refresh
		m.state = IAMStateUsers
		return m, m.fetchUsers()

	case IAMErrorMsg:
		m.err = msg

	case tea.KeyMsg:
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		if m.state == IAMStateInput {
			switch msg.String() {
			case "enter":
				name := m.input.Value()
				if name == "" {
					m.state = IAMStateUsers
					if m.userDetail != nil {
						m.state = IAMStateActions
					}
					return m, nil
				}
				var actionCmd tea.Cmd
				if m.action == IAMActionCreateUser {
					actionCmd = m.createUser(name)
				} else if m.action == IAMActionResetPassword {
					actionCmd = m.resetPassword(m.selectedUser.userName, name)
				}
				m.input.Reset()
				return m, actionCmd
			case "esc":
				m.state = IAMStateUsers
				if m.userDetail != nil {
					m.state = IAMStateActions
				}
				return m, nil
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		if m.state == IAMStateConfirmDelete {
			switch msg.String() {
			case "y", "Y":
				return m, m.deleteUser(m.selectedUser.userName)
			default:
				m.state = IAMStateActions
				return m, nil
			}
		}

		if m.state == IAMStateConfirmConsoleToggle {
			switch msg.String() {
			case "y", "Y":
				enable := m.action == IAMActionEnableConsole
				return m, m.toggleConsoleAccess(m.selectedUser.userName, enable)
			default:
				m.state = IAMStateActions
				return m, nil
			}
		}

		if m.state == IAMStateActions {
			switch msg.String() {
			case "enter":
				if item, ok := m.actionList.SelectedItem().(iamActionItem); ok {
					switch item.key {
					case "reset":
						m.state = IAMStateInput
						m.action = IAMActionResetPassword
						m.input.Placeholder = "New password"
						m.input.Focus()
					case "enable_console":
						m.state = IAMStateConfirmConsoleToggle
						m.action = IAMActionEnableConsole
					case "disable_console":
						m.state = IAMStateConfirmConsoleToggle
						m.action = IAMActionDisableConsole
					case "delete":
						m.state = IAMStateConfirmDelete
						m.action = IAMActionDeleteUser
					}
					return m, nil
				}
			case "esc", "backspace":
				m.state = IAMStateUsers
				m.userDetail = nil
				return m, nil
			}
			m.actionList, cmd = m.actionList.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "r": // Manual refresh for IAM
			if m.state == IAMStateUsers {
				m.cache.Delete(m.cacheKeys.IAMUsers())
				return m, m.fetchUsers()
			} else if m.state == IAMStateActions && m.userDetail != nil {
				m.cache.Delete(m.cacheKeys.IAMUserDetails(m.selectedUser.userName))
				return m, m.fetchUserDetails(m.selectedUser.userName)
			}
		case "enter":
			if m.state == IAMStateUsers {
				if item, ok := m.list.SelectedItem().(iamItem); ok {
					m.selectedUser = item
					m.state = IAMStateActions
					
					// Pre-populate with basic actions while loading details
					actions := []list.Item{
						iamActionItem{title: "Reset Password", key: "reset"},
						iamActionItem{title: "Toggle Console Access", key: "toggle_console"},
						iamActionItem{title: "Delete User", key: "delete"},
					}
					m.actionList.SetItems(actions)
					m.actionList.SetSize(36, len(actions))
					
					return m, m.fetchUserDetails(item.userName)
				}
			}
		case "esc", "backspace":
			if m.state == IAMStateActions {
				m.state = IAMStateUsers
				m.userDetail = nil
				return m, nil
			}
		case "n":
			if m.state == IAMStateUsers {
				m.state = IAMStateInput
				m.action = IAMActionCreateUser
				m.input.Placeholder = "User name"
				m.input.Focus()
				return m, nil
			}
		case "d":
			if item, ok := m.list.SelectedItem().(iamItem); ok {
				m.selectedUser = item
				m.state = IAMStateConfirmDelete
				m.action = IAMActionDeleteUser
				return m, nil
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m IAMModel) View() string {
	if m.err != nil {
		return m.styles.Error.Render(fmt.Sprintf("✘ Error: %v\n\nPress any key to continue...", m.err))
	}

	_, header := RenderTableHelpers(m.list, m.styles, iamColumns)

	switch m.state {
	case IAMStateLoading:
		if len(m.list.Items()) == 0 {
			return header + "\n\n  " + lipgloss.NewStyle().Foreground(m.styles.Primary).Render("󱎯 Loading...")
		}
		return header + "\n" + m.list.View()

	case IAMStateInput:
		return RenderOverlay(header + "\n" + m.list.View(), m.styles.Popup.Width(40).Render(fmt.Sprintf(
			" %s\n\n %s\n\n %s",
			lipgloss.NewStyle().Foreground(m.styles.Primary).Render(m.input.Placeholder),
			m.input.View(),
			m.styles.StatusMuted.Render("(esc to cancel)"),
		)), m.width, m.height)

	case IAMStateConfirmDelete:
		return RenderOverlay(header + "\n" + m.list.View(), m.styles.Popup.Width(40).BorderForeground(ErrorColor).Render(fmt.Sprintf(
			" %s\n\n %s %s\n\n %s",
			m.styles.Error.Bold(true).Render("⚠ Confirm Deletion"),
			"Are you sure you want to delete user",
			lipgloss.NewStyle().Foreground(m.styles.Primary).Bold(true).Render(m.selectedUser.userName),
			m.styles.StatusMuted.Render("(y/n)"),
		)), m.width, m.height)

	case IAMStateConfirmConsoleToggle:
		action := "enable"
		if m.action == IAMActionDisableConsole {
			action = "disable"
		}
		return RenderOverlay(header + "\n" + m.list.View(), m.styles.Popup.Width(40).Render(fmt.Sprintf(
			" %s\n\n Are you sure you want to %s console access for %s?\n\n %s",
			lipgloss.NewStyle().Foreground(m.styles.Primary).Bold(true).Render("⚠ Confirm Console Access Toggle"),
			action,
			lipgloss.NewStyle().Foreground(m.styles.Accent).Bold(true).Render(m.selectedUser.userName),
			m.styles.StatusMuted.Render("(y/n)"),
		)), m.width, m.height)

	case IAMStateActions:
		title := lipgloss.NewStyle().
			Foreground(m.styles.Primary).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(m.styles.Muted).
			Width(36).
			Align(lipgloss.Center).
			Render("Actions: " + m.selectedUser.userName)

		// Render action items manually to avoid bubbles/list issues in small spaces
		var actionsView strings.Builder
		for i, item := range m.actionList.Items() {
			action := item.(iamActionItem)
			if i == m.actionList.Index() {
				actionsView.WriteString(m.styles.SelectedMenuItem.Render("➜ " + action.title) + "\n")
			} else {
				actionsView.WriteString(m.styles.MenuItem.Render("  " + action.title) + "\n")
			}
		}

		menu := m.styles.Popup.Width(40).Render(
			lipgloss.JoinVertical(lipgloss.Left,
				title,
				actionsView.String(),
			),
		)
		
		return RenderOverlay(header + "\n" + m.list.View(), menu, m.width, m.height)

	default:
		return header + "\n" + m.list.View()
	}
}

func (m IAMModel) renderUserDetails() string {
	var s strings.Builder
	// Header moved to global header

	// Info Grid
	infoStyle := lipgloss.NewStyle().MarginRight(4)
	labelStyle := lipgloss.NewStyle().Foreground(m.styles.Muted)
	valueStyle := lipgloss.NewStyle().Foreground(m.styles.Snow).Bold(true)

	s.WriteString(infoStyle.Render(labelStyle.Render("User ID: ") + valueStyle.Render(m.userDetail.UserID)))
	s.WriteString(labelStyle.Render(" Created: ") + valueStyle.Render(m.userDetail.CreateDate.Format("2006-01-02 15:04")) + "\n")

	// Security Info
	mfaStatus := m.styles.Error.Render("Disabled ✘")
	if m.userDetail.MFAEnabled {
		mfaStatus = m.styles.Success.Render("Enabled ✔")
	}

	consoleStatus := m.styles.Error.Render("Disabled ✘")
	if m.userDetail.PasswordExists {
		consoleStatus = m.styles.Success.Render("Enabled ✔")
	}

	s.WriteString("\n" + labelStyle.Render("MFA Status:      ") + mfaStatus + "\n")
	s.WriteString(labelStyle.Render("Console Access:  ") + consoleStatus + "\n")

	if m.userDetail.PasswordLastUsed != nil {
		s.WriteString(labelStyle.Render("Last Login:      ") + valueStyle.Render(m.userDetail.PasswordLastUsed.Format("2006-01-02 15:04")) + "\n")
	} else {
		s.WriteString(labelStyle.Render("Last Login:      ") + valueStyle.Render("Never") + "\n")
	}

	// Access Keys
	s.WriteString("\n" + lipgloss.NewStyle().Foreground(m.styles.Primary).Bold(true).Render("ACCESS KEYS") + "\n")
	if len(m.userKeys) == 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(m.styles.Muted).Render("  No access keys found.") + "\n")
	} else {
		for _, k := range m.userKeys {
			status := m.styles.Success.Render("Active")
			if k.Status != "Active" {
				status = m.styles.Error.Render(k.Status)
			}
			s.WriteString(fmt.Sprintf("  %s (%s) - %s\n", k.AccessKeyId, k.CreateDate.Format("2006-01-02"), status))
		}
	}

	// Controls
	s.WriteString("\n" + lipgloss.NewStyle().Foreground(m.styles.Primary).Render("Actions: ") +
		lipgloss.NewStyle().Foreground(m.styles.Snow).Render("r: Reset Password • c: Toggle Console Access • esc: Back") + "\n")

	return m.styles.MainContainer.Render(s.String())
}
