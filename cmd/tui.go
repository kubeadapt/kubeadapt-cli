package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui/components"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui/views"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Long:  `Launch the KubeAdapt interactive terminal UI for browsing clusters, workloads, costs, and recommendations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		app := tui.NewApp(client)

		sb := components.NewSidebar()
		app.SetSidebar(func(activeView tui.ViewID, height int) string {
			sb.ActiveView = activeView
			sb.Height = height
			return sb.Render()
		})

		app.RegisterView(tui.ViewOverview, &viewAdapter{views.NewOverviewView(client)})
		app.RegisterView(tui.ViewClusters, &viewAdapter{views.NewClustersView(client)})
		app.RegisterView(tui.ViewWorkloads, &viewAdapter{views.NewWorkloadsView(client)})
		app.RegisterView(tui.ViewNodes, &viewAdapter{views.NewNodesView(client)})
		app.RegisterView(tui.ViewRecommendations, &viewAdapter{views.NewRecommendationsView(client)})
		app.RegisterView(tui.ViewCosts, &viewAdapter{views.NewCostsView(client)})
		app.RegisterView(tui.ViewNamespaces, &viewAdapter{views.NewNamespacesView(client)})
		app.RegisterView(tui.ViewNodeGroups, &viewAdapter{views.NewNodeGroupsView(client)})
		app.RegisterView(tui.ViewPVs, &viewAdapter{views.NewPVsView(client)})
		app.RegisterView(tui.ViewHelp, &viewAdapter{views.NewHelpView()})

		p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("running TUI: %w", err)
		}
		return nil
	},
}

// viewAdapter bridges views.View to tui.ViewInterface.
type viewAdapter struct {
	view views.View
}

func (a *viewAdapter) Init() tea.Cmd {
	return a.view.Init()
}

func (a *viewAdapter) Update(msg tea.Msg) (tui.ViewInterface, tea.Cmd) {
	newView, cmd := a.view.Update(msg)
	return &viewAdapter{newView}, cmd
}

func (a *viewAdapter) View(width, height int) string {
	return a.view.View(width, height)
}

func (a *viewAdapter) Title() string {
	return a.view.Title()
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
