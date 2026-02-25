package views

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui/components"
)

// ClusterDetailView displays detailed cluster information with charts.
type ClusterDetailView struct {
	client        *api.Client
	clusterID     string
	dashboard     *types.ClusterDashboardResponse
	costDist      *types.CostDistributionResponse
	loading       bool
	err           error
	spinner       spinner.Model
	viewport      viewport.Model
	viewportReady bool
}

// NewClusterDetailView creates a new cluster detail view.
func NewClusterDetailView(client *api.Client, clusterID string) *ClusterDetailView {
	return &ClusterDetailView{
		client:    client,
		clusterID: clusterID,
		loading:   true,
		spinner:   components.NewSpinner(),
	}
}

func (v *ClusterDetailView) Title() string { return "Cluster Detail" }

func (v *ClusterDetailView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *ClusterDetailView) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		dash, err := v.client.GetClusterDashboard(ctx, v.clusterID)
		if err != nil {
			return tui.DetailDataLoadedMsg{EntityType: "cluster", EntityID: v.clusterID, Err: err}
		}

		costDist, _ := v.client.GetClusterCostDistribution(ctx, v.clusterID, "7d")

		return tui.DetailDataLoadedMsg{
			EntityType: "cluster",
			EntityID:   v.clusterID,
			Data: &clusterDetailData{
				Dashboard: dash,
				CostDist:  costDist,
			},
		}
	}
}

type clusterDetailData struct {
	Dashboard *types.ClusterDashboardResponse
	CostDist  *types.CostDistributionResponse
}

func (v *ClusterDetailView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle detail-specific keys BEFORE viewport
		if msg.String() == "c" {
			return v, func() tea.Msg {
				return tui.PushDetailMsg{
					View:       NewCapacityPlanningView(v.client, v.clusterID),
					Breadcrumb: "Capacity",
				}
			}
		}
		// Forward to viewport for j/k scrolling
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
		if msg.EntityType == "cluster" && msg.EntityID == v.clusterID {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				data, ok := msg.Data.(*clusterDetailData)
				if !ok {
					return v, nil
				}
				v.dashboard = data.Dashboard
				v.costDist = data.CostDist
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *ClusterDetailView) renderContent(width int) string {
	d := v.dashboard
	var sections []string

	halfWidth := (width - 3) / 2
	if halfWidth < 25 {
		halfWidth = 25
	}

	propsKV := components.NewKVTable()
	propsKV.Add("Name", d.ClusterName)
	propsKV.Add("Provider", d.Provider)
	propsKV.Add("Region", output.FormatOptionalString(d.Region))
	propsKV.Add("Environment", d.Environment)
	propsKV.Add("Status", d.Status)
	propsKV.Add("Version", output.FormatOptionalString(d.Version))
	propsPanel := components.NewPanel("Properties", propsKV.Render(), halfWidth)

	resourceKV := components.NewKVTable()
	resourceKV.Add("Nodes", fmt.Sprintf("%d", d.NodeCount))
	resourceKV.Add("Pods", fmt.Sprintf("%d", d.PodCount))
	resourceKV.Add("Containers", fmt.Sprintf("%d", d.ContainerCount))
	resourceKV.Add("Deployments", fmt.Sprintf("%d", d.DeploymentCount))
	resourceKV.Add("Namespaces", fmt.Sprintf("%d", d.NamespaceCount))
	resourceKV.Add("Pending Recs", fmt.Sprintf("%d", d.RecommendationCount))
	resourcePanel := components.NewPanel("Resources", resourceKV.Render(), halfWidth)

	sections = append(sections, components.SideBySide(propsPanel, resourcePanel))

	costKV := components.NewKVTable()
	costKV.Add("Hourly", output.FormatCost(d.HourlyCost))
	costKV.Add("Monthly", output.FormatCost(d.MonthlyCost))
	costKV.Add("Savings/hr", output.FormatCost(d.TotalSavingsHourly))

	if d.MTDActualCost != nil {
		costKV.Add("MTD Actual", output.FormatCost(*d.MTDActualCost))
	}
	if d.PotentialMonthlySavings != nil {
		costKV.Add("Pot. Savings/mo", output.FormatCost(*d.PotentialMonthlySavings))
	} else {
		costKV.Add("Savings/mo", output.FormatCost(d.MonthlySavings))
	}
	costKV.Add("Efficiency", output.FormatPercent(d.ClusterEfficiency))

	costPanel := components.NewPanel("Costs", costKV.Render(), halfWidth)

	cpuGauge := components.NewGaugeBar("CPU", d.CPUUtilizationPercent, halfWidth-4)
	memGauge := components.NewGaugeBar("Memory", d.MemoryUtilizationPercent, halfWidth-4)
	effGauge := components.NewGaugeBar("Eff.", d.ClusterEfficiency, halfWidth-4)

	gaugeContent := cpuGauge.Render() + "\n" + memGauge.Render() + "\n" + effGauge.Render() +
		"\n\n" + fmt.Sprintf("CPU: %.2f / %.2f cores", d.CPUUsage, d.CPUCores) +
		"\n" + fmt.Sprintf("Mem: %.2f / %.2f GB", d.MemoryUsageGB, d.MemoryGB)
	utilPanel := components.NewPanel("Utilization & Efficiency", gaugeContent, halfWidth)

	sections = append(sections, components.SideBySide(costPanel, utilPanel))

	var breakdownPanel components.Panel
	if len(d.CostBreakdown) > 0 {
		bc := components.NewBarChart("Cost Breakdown", halfWidth-4)
		keys := make([]string, 0, len(d.CostBreakdown))
		for k := range d.CostBreakdown {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			label := strings.ReplaceAll(k, "_cost", "")
			bc.AddItem(label, d.CostBreakdown[k])
		}
		breakdownPanel = components.NewPanel("Cost Breakdown", bc.Render(), halfWidth)
	} else {
		breakdownPanel = components.NewPanel("Cost Breakdown", "No breakdown data", halfWidth)
	}

	var recPanel components.Panel
	if len(d.RecommendationSummary) > 0 {
		recLines := make([]string, 0, len(d.RecommendationSummary))
		sort.Slice(d.RecommendationSummary, func(i, j int) bool {
			return d.RecommendationSummary[i].PotentialSavings > d.RecommendationSummary[j].PotentialSavings
		})

		for _, rec := range d.RecommendationSummary {
			rType := rec.Type
			if len(rType) > 25 {
				rType = rType[:22] + "..."
			}
			recLines = append(recLines, fmt.Sprintf("• %s (%d): %s",
				rType, rec.Count, output.FormatCost(rec.PotentialSavings)))
		}
		recPanel = components.NewPanel("Recommendation Summary", strings.Join(recLines, "\n"), halfWidth)
	} else {
		recPanel = components.NewPanel("Recommendation Summary", "No recommendations", halfWidth)
	}

	sections = append(sections, components.SideBySide(breakdownPanel, recPanel))

	if v.costDist != nil && len(v.costDist.DataPoints) > 0 {
		cpuData := make([]float64, len(v.costDist.DataPoints))
		memData := make([]float64, len(v.costDist.DataPoints))
		for i, dp := range v.costDist.DataPoints {
			cpuData[i] = dp.CPUUtilization
			memData[i] = dp.MemoryUtilization
		}

		chartWidth := width - 4
		if chartWidth > 120 {
			chartWidth = 120
		}
		chart := components.NewLineChart("Resource Utilization (7d)", chartWidth, 8)
		chart.AddSeries("CPU %", cpuData)
		chart.AddSeries("Memory %", memData)
		sections = append(sections, chart.Render())
	}

	hint := tui.HelpStyle.Render("Press 'c' for Capacity Planning")
	sections = append(sections, hint)

	return strings.Join(sections, "\n\n")
}

func (v *ClusterDetailView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading cluster details..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.dashboard == nil {
		return "No data"
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
