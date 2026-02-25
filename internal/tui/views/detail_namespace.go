package views

import (
	"context"
	"fmt"
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

// NamespaceDetailView displays detailed namespace information with charts.
type NamespaceDetailView struct {
	client        *api.Client
	namespaceName string
	clusterID     string
	detail        *types.NamespaceDetailResponse
	trends        *types.NamespaceTrendsResponse
	loading       bool
	err           error
	spinner       spinner.Model
	viewport      viewport.Model
	viewportReady bool
}

// NewNamespaceDetailView creates a new namespace detail view.
func NewNamespaceDetailView(client *api.Client, namespaceName, clusterID string) *NamespaceDetailView {
	return &NamespaceDetailView{
		client:        client,
		namespaceName: namespaceName,
		clusterID:     clusterID,
		loading:       true,
		spinner:       components.NewSpinner(),
	}
}

func (v *NamespaceDetailView) Title() string { return "Namespace Detail" }

func (v *NamespaceDetailView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *NamespaceDetailView) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		detail, err := v.client.GetNamespaceDetails(ctx, v.namespaceName, v.clusterID)
		if err != nil {
			return tui.DetailDataLoadedMsg{
				EntityType: "namespace",
				EntityID:   v.namespaceName,
				Err:        err,
			}
		}

		trends, _ := v.client.GetNamespaceTrends(ctx, v.namespaceName, v.clusterID, "30d")

		return tui.DetailDataLoadedMsg{
			EntityType: "namespace",
			EntityID:   v.namespaceName,
			Data: &namespaceDetailData{
				Detail: detail,
				Trends: trends,
			},
		}
	}
}

type namespaceDetailData struct {
	Detail *types.NamespaceDetailResponse
	Trends *types.NamespaceTrendsResponse
}

func (v *NamespaceDetailView) Update(msg tea.Msg) (View, tea.Cmd) {
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
		if msg.EntityType == "namespace" && msg.EntityID == v.namespaceName {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				data, ok := msg.Data.(*namespaceDetailData)
				if !ok {
					return v, nil
				}
				v.detail = data.Detail
				v.trends = data.Trends
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *NamespaceDetailView) renderContent(width int) string {
	d := v.detail
	var sections []string

	halfWidth := (width - 3) / 2
	if halfWidth < 25 {
		halfWidth = 25
	}

	propsKV := components.NewKVTable()
	propsKV.Add("Name", d.Name)
	if d.ClusterName != nil {
		propsKV.Add("Cluster", *d.ClusterName)
	}
	if d.Team != nil {
		propsKV.Add("Team", *d.Team)
	}
	if d.Department != nil {
		propsKV.Add("Department", *d.Department)
	}
	propsPanel := components.NewPanel("Properties", propsKV.Render(), halfWidth)

	resourceKV := components.NewKVTable()
	resourceKV.Add("Pods", fmt.Sprintf("%d", d.PodCount))
	resourceKV.Add("Workloads", fmt.Sprintf("%d", d.WorkloadCount))
	resourceKV.Add("Containers", output.FormatIntPtr(d.ContainerCount))
	resourceKV.Add("Total CPU", output.FormatFloat(d.TotalCPUCores, 2)+" cores")
	resourceKV.Add("Total Memory", output.FormatMemoryGB(d.TotalMemoryGB))
	resourceKV.Add("Efficiency", output.FormatPercentPtr(d.EfficiencyScore))
	resourceKV.Add("Hourly Cost", output.FormatCost(d.HourlyCost))
	if d.MonthlyCost != nil {
		resourceKV.Add("Monthly Cost", output.FormatCostPtr(d.MonthlyCost))
	} else {
		resourceKV.Add("Monthly Cost", output.FormatCost(d.HourlyCost*730))
	}
	resourcePanel := components.NewPanel("Resources", resourceKV.Render(), halfWidth)

	sections = append(sections, components.SideBySide(propsPanel, resourcePanel))

	if v.trends != nil && len(v.trends.DataPoints) > 0 {
		costData := make([]float64, len(v.trends.DataPoints))
		for i, dp := range v.trends.DataPoints {
			costData[i] = dp.HourlyCost
		}

		chartWidth := width - 4
		if chartWidth > 120 {
			chartWidth = 120
		}
		chart := components.NewLineChart("Cost Trend (30d, $/hr)", chartWidth, 8)
		chart.AddSeries("Cost", costData)
		sections = append(sections, chart.Render())
	}

	if v.trends != nil && len(v.trends.DataPoints) > 0 {
		cpuValues := make([]float64, len(v.trends.DataPoints))
		memValues := make([]float64, len(v.trends.DataPoints))
		for i, dp := range v.trends.DataPoints {
			cpuValues[i] = dp.CPUUsage
			memValues[i] = dp.MemoryUsageBytes / (1024 * 1024 * 1024)
		}

		lastCPU := cpuValues[len(cpuValues)-1]
		lastMem := memValues[len(memValues)-1]
		cpuSpark := components.NewSparkline("CPU Usage", cpuValues, fmt.Sprintf("%.2f cores", lastCPU))
		memSpark := components.NewSparkline("Memory Usage", memValues, fmt.Sprintf("%.1f GB", lastMem))
		sections = append(sections, cpuSpark.Render()+"\n"+memSpark.Render())
	}

	if len(d.Workloads) > 0 {
		headers := []string{"Name", "Kind", "Replicas", "CPU Req", "Mem Req", "$/hr"}
		rows := make([][]string, 0, len(d.Workloads))
		for _, w := range d.Workloads {
			rows = append(rows, []string{
				w.WorkloadName,
				w.WorkloadKind,
				fmt.Sprintf("%d/%d", w.AvailableReplicas, w.Replicas),
				output.FormatFloat(w.CPURequest, 2),
				output.FormatMemoryGB(w.MemoryRequestGB),
				output.FormatCost(w.HourlyCost),
			})
		}
		table := components.NewTable(headers, rows)
		table.Width = width - 4
		sections = append(sections, tui.CardTitleStyle.Render(fmt.Sprintf("Workloads (%d)", len(d.Workloads)))+"\n"+table.Render())
	}

	return strings.Join(sections, "\n\n")
}

func (v *NamespaceDetailView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading namespace details..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.detail == nil {
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
