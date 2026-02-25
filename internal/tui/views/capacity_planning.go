package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui/components"
)

// CapacityPlanningView displays capacity planning metrics for a cluster.
type CapacityPlanningView struct {
	client        *api.Client
	clusterID     string
	data          *types.CapacityPlanningResponse
	loading       bool
	err           error
	spinner       spinner.Model
	viewport      viewport.Model
	viewportReady bool
}

// NewCapacityPlanningView creates a new capacity planning view.
func NewCapacityPlanningView(client *api.Client, clusterID string) *CapacityPlanningView {
	return &CapacityPlanningView{
		client:    client,
		clusterID: clusterID,
		loading:   true,
		spinner:   components.NewSpinner(),
	}
}

func (v *CapacityPlanningView) Title() string { return "Capacity Planning" }

func (v *CapacityPlanningView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *CapacityPlanningView) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		data, err := v.client.GetCapacityPlanning(ctx, v.clusterID)
		return tui.DetailDataLoadedMsg{
			EntityType: "capacity-planning",
			EntityID:   v.clusterID,
			Data:       data,
			Err:        err,
		}
	}
}

func (v *CapacityPlanningView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if v.viewportReady {
			var cmd tea.Cmd
			v.viewport, cmd = v.viewport.Update(msg)
			return v, cmd
		}
	case tea.MouseMsg:
		if v.viewportReady {
			var cmd tea.Cmd
			v.viewport, cmd = v.viewport.Update(msg)
			return v, cmd
		}
	case tui.DetailDataLoadedMsg:
		if msg.EntityType == "capacity-planning" && msg.EntityID == v.clusterID {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				if d, ok := msg.Data.(*types.CapacityPlanningResponse); ok {
					v.data = d
				}
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *CapacityPlanningView) renderContent(width int) string {
	d := v.data
	var sections []string

	halfWidth := (width - 3) / 2
	if halfWidth < 25 {
		halfWidth = 25
	}

	osChart := components.NewBarChart("Nodes by OS", halfWidth-4)
	for _, osNode := range d.NodesByOS {
		osChart.AddItem(osNode.OS, float64(osNode.Count))
	}

	spotGauge := components.NewGaugeBar("Spot %", d.SpotVsOnDemand.SpotPercent, halfWidth-4)
	spotText := fmt.Sprintf("Spot: %d (%.1f%%)\nOn-Demand: %d (%.1f%%)\nTotal: %d",
		d.SpotVsOnDemand.SpotCount, d.SpotVsOnDemand.SpotPercent,
		d.SpotVsOnDemand.OnDemandCount, d.SpotVsOnDemand.OnDemandPercent,
		d.SpotVsOnDemand.Total)

	infraContent := osChart.Render() + "\n\n" + spotGauge.Render() + "\n" + spotText
	infraPanel := components.NewPanel("Infrastructure", infraContent, halfWidth)

	podKV := components.NewKVTable()
	podKV.Add("Total Pods", fmt.Sprintf("%d", d.PodDensity.TotalPods))
	podKV.Add("Total Nodes", fmt.Sprintf("%d", d.PodDensity.TotalNodes))
	podKV.Add("Avg Pods/Node", fmt.Sprintf("%.1f", d.PodDensity.AvgPodsPerNode))
	podKV.Add("Max Pods/Node", fmt.Sprintf("%d", d.PodDensity.MaxPodsPerNode))
	podPanel := components.NewPanel("Pod Density", podKV.Render(), halfWidth)

	sections = append(sections, components.SideBySide(infraPanel, podPanel))

	if len(d.CostByAZ) > 0 {
		azChart := components.NewBarChart("Hourly Cost by AZ", width-4)
		for _, az := range d.CostByAZ {
			azChart.AddItem(az.Zone, az.HourlyCost)
		}
		sections = append(sections, azChart.Render())
	}

	if len(d.NodeGroups) > 0 {
		rows := make([][]string, 0, len(d.NodeGroups))
		for _, ng := range d.NodeGroups {
			instType := "-"
			if ng.InstanceType != nil {
				instType = *ng.InstanceType
			}
			rows = append(rows, []string{
				ng.Name,
				instType,
				fmt.Sprintf("%d", ng.Count),
				fmt.Sprintf("%.1f%%", ng.SpotPercent),
				output.FormatCost(ng.HourlyCost),
			})
		}

		tableStr := "Node Groups:\n"

		nameW, instW, countW, spotW, costW := 20, 15, 8, 10, 10
		header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s", nameW, "Name", instW, "Instance", countW, "Count", spotW, "Spot %", costW, "$/hr")
		tableStr += lipgloss.NewStyle().Bold(true).Render(header) + "\n"

		for _, row := range rows {
			line := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s", nameW, row[0], instW, row[1], countW, row[2], spotW, row[3], costW, row[4])
			tableStr += line + "\n"
		}

		sections = append(sections, components.NewPanel("Node Groups", tableStr, width-4).Render())
	}

	return strings.Join(sections, "\n\n")
}

func (v *CapacityPlanningView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading capacity planning..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.data == nil {
		return noDataMsg
	}

	if !v.viewportReady {
		v.viewport = viewport.New(width, height)
		v.viewportReady = true
	}
	v.viewport.Width = width
	v.viewport.Height = height

	v.viewport.SetContent(v.renderContent(width))

	return v.viewport.View()
}
