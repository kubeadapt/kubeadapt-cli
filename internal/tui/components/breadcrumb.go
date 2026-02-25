package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// Breadcrumb renders a navigation path like: Clusters > prod-us-east
type Breadcrumb struct {
	Parts []string
}

// NewBreadcrumb creates a new breadcrumb from path parts.
func NewBreadcrumb(parts ...string) Breadcrumb {
	return Breadcrumb{Parts: parts}
}

// Render draws the breadcrumb.
func (b Breadcrumb) Render() string {
	if len(b.Parts) == 0 {
		return ""
	}

	sep := lipgloss.NewStyle().Foreground(tui.ColorDarkGray).Render(" > ")
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorWhite)
	inactiveStyle := lipgloss.NewStyle().Foreground(tui.ColorGray)

	var parts []string
	for i, part := range b.Parts {
		if i == len(b.Parts)-1 {
			parts = append(parts, activeStyle.Render(part))
		} else {
			parts = append(parts, inactiveStyle.Render(part))
		}
	}

	return strings.Join(parts, sep)
}
