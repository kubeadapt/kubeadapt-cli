package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// Dialog renders a centered confirmation dialog.
type Dialog struct {
	Title   string
	Message string
	Width   int
	Height  int
}

// Render renders the dialog centered on the screen.
func (d *Dialog) Render() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.ColorYellow).
		Render(d.Title)

	content := title + "\n\n" + d.Message + "\n\n" +
		tui.StatusBarKeyStyle.Render("y") + " confirm  " +
		tui.StatusBarKeyStyle.Render("n") + " cancel"

	dialog := tui.DialogStyle.Render(content)

	if d.Width > 0 && d.Height > 0 {
		return lipgloss.Place(d.Width, d.Height, lipgloss.Center, lipgloss.Center, dialog)
	}

	return dialog
}
