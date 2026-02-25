package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

type HelpView struct{}

func NewHelpView() *HelpView {
	return &HelpView{}
}

func (v *HelpView) Title() string { return "Help" }

func (v *HelpView) Init() tea.Cmd { return nil }

func (v *HelpView) Update(msg tea.Msg) (View, tea.Cmd) {
	return v, nil
}

func (v *HelpView) View(width, height int) string {
	help := tui.CardTitleStyle.Render("KubeAdapt TUI - Keyboard Shortcuts") + "\n\n"

	sections := []struct {
		title string
		keys  []struct{ key, desc string }
	}{
		{
			"Navigation",
			[]struct{ key, desc string }{
				{"1-0", "Switch between views (clears detail stack)"},
				{"Tab", "Switch sub-tabs (where available)"},
				{"j/k or Arrow keys", "Navigate rows in tables"},
				{"j/k or PageUp/PageDown", "Scroll detail views with viewport"},
				{"Enter", "Drill down into selected item"},
				{"Esc", "Go back to previous view"},
			},
		},
		{
			"Actions",
			[]struct{ key, desc string }{
				{"r", "Refresh current view or detail"},
				{"/", "Filter (where available)"},
				{"?", "Toggle this help screen"},
				{"q / Ctrl+C", "Quit"},
				{"Mouse click", "Click sidebar to switch views, click table rows"},
				{"Mouse wheel", "Scroll tables and detail views"},
			},
		},
		{
			"Views",
			[]struct{ key, desc string }{
				{"1", "Overview - Dashboard with gauges and charts"},
				{"2", "Clusters - All connected clusters"},
				{"3", "Workloads - Deployments, StatefulSets, etc."},
				{"4", "Nodes - Cluster nodes"},
				{"5", "Recommendations - Cost optimization suggestions"},
				{"6", "Costs - Team and department cost breakdown"},
				{"7", "Namespaces - Namespace details"},
				{"8", "Node Groups - Node group configuration"},
				{"9", "PVs - Persistent Volumes"},
				{"0", "Help - This screen"},
			},
		},
		{
			"Detail Views (Enter on a list row)",
			[]struct{ key, desc string }{
				{"Cluster", "Properties, costs, utilization gauges, 7d charts"},
				{"Workload", "Resources, CPU/memory charts, node distribution"},
				{"Node", "Capacity, 24h CPU/memory history charts"},
				{"Node Group", "Aggregate stats, per-node table"},
				{"Namespace", "Cost trends, sparklines, workload table"},
			},
		},
	}

	for _, section := range sections {
		help += tui.CardTitleStyle.Render(section.title) + "\n"
		for _, k := range section.keys {
			help += "  " + tui.StatusBarKeyStyle.Render(k.key) + "  " + k.desc + "\n"
		}
		help += "\n"
	}

	return help
}
