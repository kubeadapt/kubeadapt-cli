package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// Panel renders a bordered section container with a title.
type Panel struct {
	Title   string
	Content string
	Width   int
	Style   lipgloss.Style
}

// NewPanel creates a new panel with default styling.
func NewPanel(title string, content string, width int) Panel {
	if width < 10 {
		width = 10
	}
	return Panel{
		Title:   title,
		Content: content,
		Width:   width,
		Style: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(tui.ColorDarkGray).
			Padding(0, 1).
			Width(width),
	}
}

// Render draws the panel.
func (p Panel) Render() string {
	titleStr := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.ColorPrimary).
		Render(p.Title)

	return p.Style.
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		Render(titleStr + "\n" + p.Content)
}

// SideBySide renders two panels side by side.
func SideBySide(left, right Panel) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, left.Render(), " ", right.Render())
}
