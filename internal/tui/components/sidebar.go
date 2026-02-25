package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

type sidebarItem struct {
	Key    string
	Label  string
	ViewID tui.ViewID
}

var sidebarItems = []sidebarItem{
	{"1", "Overview", tui.ViewOverview},
	{"2", "Clusters", tui.ViewClusters},
	{"3", "Workloads", tui.ViewWorkloads},
	{"4", "Nodes", tui.ViewNodes},
	{"5", "Recomm.", tui.ViewRecommendations},
	{"6", "Costs", tui.ViewCosts},
	{"7", "Namespc.", tui.ViewNamespaces},
	{"8", "NodeGrps", tui.ViewNodeGroups},
	{"9", "PVs", tui.ViewPVs},
	{"0", "Help", tui.ViewHelp},
}

// Sidebar is a stateless vertical navigation renderer (no tea.Model methods).
type Sidebar struct {
	ActiveView tui.ViewID
	Width      int
	Height     int
}

// NewSidebar returns a Sidebar with default width of 20.
func NewSidebar() *Sidebar {
	return &Sidebar{
		Width: 20,
	}
}

// Render draws the sidebar with title, separator, and navigation items.
func (s *Sidebar) Render() string {
	// -1 accounts for the right border character added by SidebarStyle.
	contentWidth := s.Width - 1

	title := tui.SidebarTitle.
		Width(contentWidth).
		Align(lipgloss.Center).
		Render("KubeAdapt")

	sep := lipgloss.NewStyle().
		Foreground(tui.ColorDarkGray).
		Width(contentWidth).
		Render(strings.Repeat("─", contentWidth))

	lines := make([]string, 0, len(sidebarItems)+2)
	lines = append(lines, title, sep)

	for _, item := range sidebarItems {
		var indicator string
		var style lipgloss.Style

		if item.ViewID == s.ActiveView {
			indicator = "●"
			style = tui.SidebarItemActive
		} else {
			indicator = "○"
			style = tui.SidebarItemInactive
		}

		text := fmt.Sprintf("%s %s %s", item.Key, indicator, item.Label)
		lines = append(lines, style.Width(contentWidth).Render(text))
	}

	content := strings.Join(lines, "\n")

	sidebarStyle := tui.SidebarStyle.Width(contentWidth)
	if s.Height > 0 {
		sidebarStyle = sidebarStyle.Height(s.Height)
	}

	return sidebarStyle.Render(content)
}
