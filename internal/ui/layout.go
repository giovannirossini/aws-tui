package ui

import "github.com/charmbracelet/lipgloss"

// Layout constants for consistent UI sizing across the app
const (
	// Width Offsets
	AppWidthOffset          = 6  // Total width - AppWidthOffset = Header/MainContainer width
	MainContentPadding      = 4  // Padding inside the main container
	InnerContentWidthOffset = 10 // width - InnerContentWidthOffset = List/Table width

	// Height Offsets
	AppHeaderHeight         = 3 // Header content + border
	AppFooterHeight         = 2 // Newlines at the bottom
	MainContainerMargins    = 1 // Newlines between header and container
	TableColumnHeaderHeight = 1 // Height of the table column titles
	AppInternalFooterHeight = 0 // Height of the footer inside the box (1 line border + 1 line text)

	// Total height offset for a standard list inside the main container
	// height - AppHeaderHeight - AppFooterHeight - MainContainerMargins - TableColumnHeaderHeight - AppInternalFooterHeight - (Container Border 2)
	StandardListHeightOffset = 10
)

// GetMainContainerSize returns the width and height for the MainContainer
func GetMainContainerSize(width, height int) (int, int) {
	w := width - AppWidthOffset
	h := height - AppHeaderHeight - AppFooterHeight - MainContainerMargins
	return w, h
}

// GetInnerListSize returns the width and height for a list/table inside the MainContainer
func GetInnerListSize(width, height int) (int, int) {
	w := width - InnerContentWidthOffset
	h := height - StandardListHeightOffset
	return w, h
}

// RenderBoxedContainer renders a boxed container with header and an internal footer
func RenderBoxedContainer(styles Styles, content, footer string, width, height int) string {
	w, h := GetMainContainerSize(width, height)

	// Join content and internal footer
	fullContent := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Height(h-AppInternalFooterHeight-2).Render(content),
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(styles.Muted).
			Faint(true).
			Width(w-2).
			Render(footer),
	)

	return styles.MainContainer.
		Width(w).
		Height(h).
		Render(fullContent)
}
