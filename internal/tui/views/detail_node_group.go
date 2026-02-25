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

// NodeGroupDetailView displays detailed node group information.
type NodeGroupDetailView struct {
	client        *api.Client
	groupName     string
	clusterID     string
	detail        *types.NodeGroupDetailResponse
	loading       bool
	err           error
	spinner       spinner.Model
	viewport      viewport.Model
	viewportReady bool
}

// NewNodeGroupDetailView creates a new node group detail view.
func NewNodeGroupDetailView(client *api.Client, groupName, clusterID string) *NodeGroupDetailView {
	return &NodeGroupDetailView{
		client:    client,
		groupName: groupName,
		clusterID: clusterID,
		loading:   true,
		spinner:   components.NewSpinner(),
	}
}

func (v *NodeGroupDetailView) Title() string { return "Node Group Detail" }

func (v *NodeGroupDetailView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *NodeGroupDetailView) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		detail, err := v.client.GetNodeGroupDetails(ctx, v.groupName, v.clusterID)

		return tui.DetailDataLoadedMsg{
			EntityType: "node_group",
			EntityID:   v.groupName,
			Data:       detail,
			Err:        err,
		}
	}
}

func (v *NodeGroupDetailView) Update(msg tea.Msg) (View, tea.Cmd) {
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
		if msg.EntityType == "node_group" && msg.EntityID == v.groupName {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				v.detail = msg.Data.(*types.NodeGroupDetailResponse)
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v *NodeGroupDetailView) renderContent(width int) string {
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
	if d.InstanceType != nil {
		propsKV.Add("Instance Type", *d.InstanceType)
	}
	propsKV.Add("Node Count", fmt.Sprintf("%d", d.NodeCount))
	propsKV.Add("Spot %", output.FormatPercent(d.SpotPercentage))
	propsPanel := components.NewPanel("Properties", propsKV.Render(), halfWidth)

	resourceKV := components.NewKVTable()
	resourceKV.Add("Total CPU", output.FormatFloat(d.TotalCPUCores, 1)+" cores")
	resourceKV.Add("Total Memory", output.FormatMemoryGB(d.TotalMemoryGB))
	resourceKV.Add("Hourly Cost", output.FormatCost(d.HourlyCost))
	resourceKV.Add("Monthly Cost", output.FormatCost(d.HourlyCost*730))
	resourcePanel := components.NewPanel("Resources", resourceKV.Render(), halfWidth)

	sections = append(sections, components.SideBySide(propsPanel, resourcePanel))

	cpuGauge := components.NewGaugeBar("Avg CPU", d.AvgCPUUtilization, width-4)
	memGauge := components.NewGaugeBar("Avg Memory", d.AvgMemoryUtilization, width-4)
	sections = append(sections, cpuGauge.Render()+"\n"+memGauge.Render())

	if len(d.Nodes) > 0 {
		headers := []string{"Node", "Instance", "AZ", "Ready", "CPU Cap", "Mem Cap", "Spot", "$/hr"}
		var rows [][]string
		for _, n := range d.Nodes {
			rows = append(rows, []string{
				n.NodeName,
				output.FormatOptionalString(n.InstanceType),
				output.FormatOptionalString(n.AvailabilityZone),
				output.FormatBool(n.IsReady),
				output.FormatFloat(n.CPUCapacity, 1),
				output.FormatMemoryGB(n.MemoryCapacityGB),
				output.FormatBool(n.SpotInstance),
				output.FormatCost(n.HourlyCost),
			})
		}
		table := components.NewTable(headers, rows)
		table.Width = width - 4
		sections = append(sections, tui.CardTitleStyle.Render(fmt.Sprintf("Nodes (%d)", len(d.Nodes)))+"\n"+table.Render())
	}

	return strings.Join(sections, "\n\n")
}

func (v *NodeGroupDetailView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading node group details..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.detail == nil {
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
