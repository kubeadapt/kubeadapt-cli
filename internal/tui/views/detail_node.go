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

// NodeDetailView displays detailed node information with charts.
type NodeDetailView struct {
	client        *api.Client
	nodeUID       string
	clusterID     string
	node          *types.NodeResponse // base data from list
	metrics       *types.NodeMetricsResponse
	loading       bool
	err           error
	spinner       spinner.Model
	viewport      viewport.Model
	viewportReady bool
}

// NewNodeDetailView creates a new node detail view.
func NewNodeDetailView(client *api.Client, node *types.NodeResponse) *NodeDetailView {
	return &NodeDetailView{
		client:    client,
		nodeUID:   node.ID,
		clusterID: node.ClusterID,
		node:      node,
		loading:   true,
		spinner:   components.NewSpinner(),
	}
}

func (v *NodeDetailView) Title() string { return "Node Detail" }

func (v *NodeDetailView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *NodeDetailView) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		metrics, _ := v.client.GetNodeMetrics(ctx, v.nodeUID, v.clusterID, "24h")

		return tui.DetailDataLoadedMsg{
			EntityType: "node",
			EntityID:   v.nodeUID,
			Data: &nodeDetailData{
				Metrics: metrics,
			},
		}
	}
}

type nodeDetailData struct {
	Metrics *types.NodeMetricsResponse
}

func (v *NodeDetailView) Update(msg tea.Msg) (View, tea.Cmd) {
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
		if msg.EntityType == "node" && msg.EntityID == v.nodeUID {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				data := msg.Data.(*nodeDetailData)
				v.metrics = data.Metrics
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *NodeDetailView) renderContent(width int) string {
	n := v.node
	var sections []string

	halfWidth := (width - 3) / 2
	if halfWidth < 25 {
		halfWidth = 25
	}

	propsKV := components.NewKVTable()
	propsKV.Add("Name", n.NodeName)
	propsKV.Add("Cluster", n.ClusterName)
	propsKV.Add("Instance", output.FormatOptionalString(n.InstanceType))
	propsKV.Add("Node Group", output.FormatOptionalString(n.NodeGroup))
	propsKV.Add("AZ", output.FormatOptionalString(n.AvailabilityZone))
	propsKV.Add("Ready", output.FormatBool(n.IsReady))
	propsKV.Add("Schedulable", output.FormatBool(n.IsSchedulable))
	propsKV.Add("Spot", output.FormatBool(n.SpotInstance))
	propsPanel := components.NewPanel("Properties", propsKV.Render(), halfWidth)

	capKV := components.NewKVTable()
	capKV.Add("CPU Capacity", output.FormatFloat(n.CPUCapacity, 1))
	capKV.Add("CPU Allocatable", output.FormatFloat(n.CPUAllocatable, 1))
	capKV.Add("Mem Capacity", output.FormatMemoryGB(n.MemoryCapacityGB))
	capKV.Add("Mem Allocatable", output.FormatMemoryGB(n.MemoryAllocatableGB))
	capKV.Add("Pods Capacity", fmt.Sprintf("%d", n.PodsCapacity))
	capKV.Add("Pods Allocatable", fmt.Sprintf("%d", n.PodsAllocatable))
	capKV.Add("Pod Count", output.FormatIntPtr(n.PodCount))
	capKV.Add("Hourly Cost", output.FormatCost(n.HourlyCost))
	capKV.Add("Monthly Cost", output.FormatCostPtr(n.MonthlyCost))
	capPanel := components.NewPanel("Capacity", capKV.Render(), halfWidth)

	sections = append(sections, components.SideBySide(propsPanel, capPanel))

	if v.metrics != nil && len(v.metrics.DataPoints) > 0 {
		lastN := v.metrics.DataPoints
		if len(lastN) > 10 {
			lastN = lastN[len(lastN)-10:]
		}
		var avgCPUPct, avgMemPct float64
		for _, dp := range lastN {
			avgCPUPct += dp.CPUUsagePercent
			avgMemPct += dp.MemoryUsagePercent
		}
		count := float64(len(lastN))
		avgCPUPct /= count
		avgMemPct /= count

		cpuGauge := components.NewGaugeBar("CPU", avgCPUPct, width-4)
		memGauge := components.NewGaugeBar("Memory", avgMemPct, width-4)
		sections = append(sections, cpuGauge.Render()+"\n"+memGauge.Render())
	}

	if v.metrics != nil && len(v.metrics.DataPoints) > 0 {
		cpuUsage := make([]float64, len(v.metrics.DataPoints))
		cpuCap := make([]float64, len(v.metrics.DataPoints))
		for i, dp := range v.metrics.DataPoints {
			cpuUsage[i] = dp.CPUUsage
			cpuCap[i] = dp.CPUCapacity
		}

		chartWidth := width - 4
		if chartWidth > 120 {
			chartWidth = 120
		}
		chart := components.NewLineChart("CPU Usage (24h, cores)", chartWidth, 8)
		chart.AddSeries("Usage", cpuUsage)
		chart.AddSeries("Capacity", cpuCap)
		sections = append(sections, chart.Render())
	}

	if v.metrics != nil && len(v.metrics.DataPoints) > 0 {
		memUsage := make([]float64, len(v.metrics.DataPoints))
		memCap := make([]float64, len(v.metrics.DataPoints))
		for i, dp := range v.metrics.DataPoints {
			memUsage[i] = dp.MemoryUsageBytes / (1024 * 1024 * 1024)
			memCap[i] = dp.MemoryCapacityBytes / (1024 * 1024 * 1024)
		}

		chartWidth := width - 4
		if chartWidth > 120 {
			chartWidth = 120
		}
		chart := components.NewLineChart("Memory Usage (24h, GB)", chartWidth, 8)
		chart.AddSeries("Usage", memUsage)
		chart.AddSeries("Capacity", memCap)
		sections = append(sections, chart.Render())
	}

	return strings.Join(sections, "\n\n")
}

func (v *NodeDetailView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading node details..."
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
