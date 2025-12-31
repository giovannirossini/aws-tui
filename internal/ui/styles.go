package ui

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	// AWS Official-ish Colors
	AWSAmber    = lipgloss.Color("#FF9900") // The classic AWS Orange
	AWSSquid    = lipgloss.Color("#232F3E") // Deep Navy/Squid Ink
	AWSSky      = lipgloss.Color("#00A1C9") // Light Blue
	AWSWhite    = lipgloss.Color("#FFFFFF")
	AWSSnow     = lipgloss.Color("#F2F3F3")
	AWSGray     = lipgloss.Color("#545B64")
	AWSDarkGray = lipgloss.Color("#16191F")

	// Functional Colors
	SuccessColor = lipgloss.Color("#3EB13D")
	ErrorColor   = lipgloss.Color("#D13212")
	WarningColor = lipgloss.Color("#FF9900")
	InfoColor    = lipgloss.Color("#0073BB")
)

type Styles struct {
	// Colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Muted     lipgloss.Color
	White     lipgloss.Color
	Snow      lipgloss.Color
	DarkGray  lipgloss.Color
	Squid     lipgloss.Color

	AppTitle         lipgloss.Style
	Header           lipgloss.Style
	Profile          lipgloss.Style
	SelectedProfile  lipgloss.Style
	MainContainer    lipgloss.Style
	MenuContainer    lipgloss.Style
	Popup            lipgloss.Style
	MenuItem         lipgloss.Style
	SelectedMenuItem lipgloss.Style
	StatusBar        lipgloss.Style
	StatusKey        lipgloss.Style
	StatusMuted      lipgloss.Style
	S3BucketName     lipgloss.Style
	S3Folder         lipgloss.Style
	S3Object         lipgloss.Style
	S3Muted          lipgloss.Style
	Error            lipgloss.Style
	Success          lipgloss.Style
	Warning          lipgloss.Style
	Info             lipgloss.Style
	ViewTitle        lipgloss.Style

	// Delegate styles
	ListSelectedTitle lipgloss.Style
	ListSelectedDesc  lipgloss.Style
}

func DefaultStyles() Styles {
	s := Styles{
		Primary:   AWSAmber,
		Secondary: AWSSquid,
		Accent:    AWSSky,
		Muted:     AWSGray,
		White:     AWSWhite,
		Snow:      AWSSnow,
		DarkGray:  AWSDarkGray,
		Squid:     AWSSquid,
	}

	s.ViewTitle = lipgloss.NewStyle().
		Foreground(s.Snow).
		Padding(0, 1)

	s.AppTitle = lipgloss.NewStyle().
		Foreground(s.White).
		Background(s.Primary).
		Padding(0, 2).
		MarginRight(1)

	s.Header = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Primary).
		Padding(0, 1).
		Margin(0, 2)

	s.Profile = lipgloss.NewStyle().
		Foreground(s.Snow)

	s.SelectedProfile = lipgloss.NewStyle().
		Foreground(s.White).
		Background(InfoColor).
		Padding(0, 1)

	s.MainContainer = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Primary).
		Padding(0, 1).
		Margin(0, 2)

	s.MenuContainer = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Primary).
		Padding(1, 2).
		MarginTop(1)

	s.Popup = s.MenuContainer.Copy()

	s.MenuItem = lipgloss.NewStyle().
		Foreground(s.Snow).
		PaddingLeft(2)

	s.SelectedMenuItem = lipgloss.NewStyle().
		PaddingLeft(0).
		Foreground(s.Primary).
		Bold(true)

	s.StatusBar = lipgloss.NewStyle().
		Foreground(s.Snow).
		Background(s.Squid).
		Padding(0, 1)

	s.StatusKey = lipgloss.NewStyle().
		Foreground(s.Primary).
		Bold(true)

	s.StatusMuted = lipgloss.NewStyle().
		Foreground(s.Muted)

	s.S3BucketName = lipgloss.NewStyle().
		Foreground(s.Accent).
		Bold(true)

	s.S3Folder = lipgloss.NewStyle().
		Foreground(s.Primary)

	s.S3Object = lipgloss.NewStyle().
		Foreground(s.Snow)

	s.S3Muted = lipgloss.NewStyle().
		Foreground(s.Muted).
		Italic(true)

	s.Error = lipgloss.NewStyle().Foreground(ErrorColor)
	s.Success = lipgloss.NewStyle().Foreground(SuccessColor)
	s.Warning = lipgloss.NewStyle().Foreground(WarningColor)
	s.Info = lipgloss.NewStyle().Foreground(InfoColor)

	s.ListSelectedTitle = lipgloss.NewStyle().
		Foreground(s.Primary).
		Bold(true)

	s.ListSelectedDesc = s.ListSelectedTitle.Copy().
		Foreground(s.Muted).
		Bold(false)

	return s
}
