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

// WorkloadDetailView displays detailed workload information with charts.
type WorkloadDetailView struct {
	client        *api.Client
	workloadUID   string
	clusterID     string
	workload      *types.WorkloadResponse // base data from list
	metrics       *types.WorkloadMetricsResponse
	nodes         *types.WorkloadNodesResponse
	loading       bool
	err           error
	spinner       spinner.Model
	viewport      viewport.Model
	viewportReady bool
}

// NewWorkloadDetailView creates a new workload detail view.
func NewWorkloadDetailView(client *api.Client, workload *types.WorkloadResponse) *WorkloadDetailView {
	return &WorkloadDetailView{
		client:      client,
		workloadUID: workload.ID,
		clusterID:   workload.ClusterID,
		workload:    workload,
		loading:     true,
		spinner:     components.NewSpinner(),
	}
}

func (v *WorkloadDetailView) Title() string { return "Workload Detail" }

func (v *WorkloadDetailView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *WorkloadDetailView) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		metrics, _ := v.client.GetWorkloadMetrics(ctx, v.workloadUID, v.clusterID, "7d")
		nodes, _ := v.client.GetWorkloadNodes(ctx, v.workloadUID, v.clusterID)

		return tui.DetailDataLoadedMsg{
			EntityType: "workload",
			EntityID:   v.workloadUID,
			Data: &workloadDetailData{
				Metrics: metrics,
				Nodes:   nodes,
			},
		}
	}
}

type workloadDetailData struct {
	Metrics *types.WorkloadMetricsResponse
	Nodes   *types.WorkloadNodesResponse
}

func (v *WorkloadDetailView) Update(msg tea.Msg) (View, tea.Cmd) {
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
		if msg.EntityType == "workload" && msg.EntityID == v.workloadUID {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				data := msg.Data.(*workloadDetailData)
				v.metrics = data.Metrics
				v.nodes = data.Nodes
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *WorkloadDetailView) renderContent(width int) string {
	w := v.workload
	var sections []string

	halfWidth := (width - 3) / 2
	if halfWidth < 25 {
		halfWidth = 25
	}

	propsKV := components.NewKVTable()
	propsKV.Add("Name", w.WorkloadName)
	propsKV.Add("Kind", w.WorkloadKind)
	propsKV.Add("Namespace", w.Namespace)
	propsKV.Add("Cluster", w.ClusterName)
	propsKV.Add("Replicas", fmt.Sprintf("%d/%d", w.AvailableReplicas, w.Replicas))
	propsPanel := components.NewPanel("Properties", propsKV.Render(), halfWidth)

	resourceKV := components.NewKVTable()
	resourceKV.Add("CPU Request", output.FormatFloat(w.CPURequest, 3))
	resourceKV.Add("CPU Limit", output.FormatFloat(w.CPULimit, 3))
	resourceKV.Add("Mem Request", output.FormatMemoryGB(w.MemoryRequestGB))
	resourceKV.Add("Mem Limit", output.FormatMemoryGB(w.MemoryLimitGB))
	resourceKV.Add("Hourly Cost", output.FormatCost(w.HourlyCost))
	resourceKV.Add("Efficiency", output.FormatPercentPtr(w.EfficiencyScore))
	resourceKV.Add("Monthly Cost", output.FormatCostPtr(w.MonthlyCost))
	resourcePanel := components.NewPanel("Resources", resourceKV.Render(), halfWidth)

	sections = append(sections, components.SideBySide(propsPanel, resourcePanel))

	if v.metrics != nil && len(v.metrics.DataPoints) > 0 {
		lastN := v.metrics.DataPoints
		if len(lastN) > 10 {
			lastN = lastN[len(lastN)-10:]
		}
		var avgCPU, avgCPUReq, avgMem, avgMemReq float64
		for _, dp := range lastN {
			avgCPU += dp.CPUUsage
			avgCPUReq += dp.CPURequest
			avgMem += dp.MemoryUsageBytes
			avgMemReq += dp.MemoryRequestBytes
		}
		n := float64(len(lastN))
		avgCPU /= n
		avgCPUReq /= n
		avgMem /= n
		avgMemReq /= n

		cpuPct := 0.0
		if avgCPUReq > 0 {
			cpuPct = (avgCPU / avgCPUReq) * 100
		}
		memPct := 0.0
		if avgMemReq > 0 {
			memPct = (avgMem / avgMemReq) * 100
		}

		cpuGauge := components.NewGaugeBar("CPU", cpuPct, width-4)
		memGauge := components.NewGaugeBar("Memory", memPct, width-4)
		sections = append(sections, cpuGauge.Render()+"\n"+memGauge.Render())
	}

	if v.metrics != nil && len(v.metrics.DataPoints) > 0 {
		cpuUsage := make([]float64, len(v.metrics.DataPoints))
		cpuReq := make([]float64, len(v.metrics.DataPoints))
		for i, dp := range v.metrics.DataPoints {
			cpuUsage[i] = dp.CPUUsage * 1000
			cpuReq[i] = dp.CPURequest * 1000
		}

		chartWidth := width - 4
		if chartWidth > 120 {
			chartWidth = 120
		}
		chart := components.NewLineChart("CPU Usage (7d, millicores)", chartWidth, 8)
		chart.AddSeries("Usage", cpuUsage)
		chart.AddSeries("Request", cpuReq)
		sections = append(sections, chart.Render())
	}

	if v.metrics != nil && len(v.metrics.DataPoints) > 0 {
		memUsage := make([]float64, len(v.metrics.DataPoints))
		memReq := make([]float64, len(v.metrics.DataPoints))
		for i, dp := range v.metrics.DataPoints {
			memUsage[i] = dp.MemoryUsageBytes / (1024 * 1024)
			memReq[i] = dp.MemoryRequestBytes / (1024 * 1024)
		}

		chartWidth := width - 4
		if chartWidth > 120 {
			chartWidth = 120
		}
		chart := components.NewLineChart("Memory Usage (7d, MB)", chartWidth, 8)
		chart.AddSeries("Usage", memUsage)
		chart.AddSeries("Request", memReq)
		sections = append(sections, chart.Render())
	}

	if v.nodes != nil && len(v.nodes.Nodes) > 0 {
		headers := []string{"Node", "Pods", "CPU Usage", "Memory"}
		var rows [][]string
		for _, n := range v.nodes.Nodes {
			rows = append(rows, []string{
				n.NodeName,
				fmt.Sprintf("%d", n.PodCount),
				output.FormatFloat(n.CPUUsage, 4),
				output.FormatMemoryGB(n.MemoryUsageBytes / (1024 * 1024 * 1024)),
			})
		}
		table := components.NewTable(headers, rows)
		table.Width = width - 4
		sections = append(sections, tui.CardTitleStyle.Render("Node Distribution")+"\n"+table.Render())
	}

	return strings.Join(sections, "\n\n")
}

func (v *WorkloadDetailView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading workload details..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
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
