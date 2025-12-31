package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	Title string
	Width float64 // Percentage of total width (0.0 to 1.0)
}

func RenderTableHelpers(m list.Model, styles Styles, columns []Column) ([]lipgloss.Style, string) {
	width := m.Width() - 4 // Account for list padding
	
	columnStyles := make([]lipgloss.Style, len(columns))
	headerStrings := make([]string, len(columns))
	
	totalWidthUsed := 0
	for i, col := range columns {
		colWidth := int(float64(width) * col.Width)
		if i == len(columns)-1 {
			// Last column takes the remaining space
			colWidth = width - totalWidthUsed
		}
		totalWidthUsed += colWidth
		
		columnStyles[i] = lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).MaxHeight(1).PaddingRight(2)
		headerStrings[i] = columnStyles[i].Copy().Foreground(styles.Muted).Bold(true).Render(strings.ToUpper(col.Title))
	}
	
	header := lipgloss.JoinHorizontal(lipgloss.Top, headerStrings...)
	return columnStyles, header
}

func RenderTableRow(w io.Writer, m list.Model, styles Styles, columnStyles []lipgloss.Style, values []string, isSelected bool) {
	rowValues := make([]string, len(columnStyles))
	
	contentColor := styles.Snow
	if isSelected {
		contentColor = styles.Primary
	}
	
	for i, val := range values {
		style := columnStyles[i].Copy().Foreground(contentColor)
		if isSelected {
			style = style.Bold(true)
		}
		rowValues[i] = style.Render(val)
	}
	
	row := lipgloss.JoinHorizontal(lipgloss.Top, rowValues...)
	
	itemStyle := lipgloss.NewStyle().PaddingLeft(2).Width(m.Width() - 2)
	fmt.Fprintf(w, "%s", itemStyle.Render(row))
}

// RenderOverlay places the overlay text on top of the base text, centered.
// It uses pure string manipulation to avoid ANSI cursor movement issues in Bubble Tea.
func RenderOverlay(base, overlay string, width, height int) string {
	overlayHeight := lipgloss.Height(overlay)
	overlayWidth := lipgloss.Width(overlay)

	// Calculate center position
	x := (width - overlayWidth) / 2
	y := (height - overlayHeight) / 2
	
	if x < 0 { x = 0 }
	if y < 0 { y = 0 }

	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Ensure base has enough lines to reach the bottom of the overlay
	for len(baseLines) < y+overlayHeight {
		baseLines = append(baseLines, "")
	}

	var sb strings.Builder
	for i := 0; i < len(baseLines); i++ {
		if i >= y && i < y+overlayHeight {
			// For lines where the overlay is, we render the overlay line centered.
			// This obscures the background on those specific lines but is 100% stable
			// because it uses lipgloss's own centering logic which respects ANSI.
			sb.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center, overlayLines[i-y]))
		} else {
			// Otherwise just render the background line
			sb.WriteString(baseLines[i])
		}
		if i < len(baseLines)-1 {
			sb.WriteString("\n")
		}
	}
	
	return sb.String()
}
