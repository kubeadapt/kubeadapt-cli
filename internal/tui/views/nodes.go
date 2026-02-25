package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui/components"
)

type NodesView struct {
	client    *api.Client
	data      []types.NodeResponse
	total     int
	table     *components.Table
	loading   bool
	err       error
	spinner   spinner.Model
	filterBar *components.FilterBar
	keys      tui.KeyMap
}

func NewNodesView(client *api.Client) *NodesView {
	return &NodesView{
		client:    client,
		loading:   true,
		spinner:   components.NewSpinner(),
		filterBar: components.NewFilterBar(),
		keys:      tui.DefaultKeyMap(),
	}
}

func (v *NodesView) Title() string { return "Nodes" }

func (v *NodesView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *NodesView) loadData() tea.Cmd {
	return func() tea.Msg {
		data, err := v.client.GetNodes(context.Background(), "", "", 50, 0)
		return tui.DataLoadedMsg{View: tui.ViewNodes, Data: data, Err: err}
	}
}

func (v *NodesView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.DataLoadedMsg:
		if msg.View == tui.ViewNodes {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				resp, ok := msg.Data.(*types.NodeListResponse)
				if !ok {
					return v, nil
				}
				v.data = resp.Nodes
				v.total = resp.Total
				v.buildTable()
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	case tea.MouseMsg:
		if v.table != nil {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				v.table.MoveUp()
			case tea.MouseButtonWheelDown:
				v.table.MoveDown()
			}
		}
	case tea.KeyMsg:
		if v.filterBar.IsActive() {
			if v.filterBar.HandleKey(msg) {
				if v.filterBar.IsActive() {
					if v.table != nil {
						v.table.ApplyFilter(v.filterBar.GetQuery())
					}
				} else {
					if v.table != nil {
						v.table.ClearFilter()
					}
				}
				return v, nil
			}
		}
		if v.table != nil {
			switch {
			case key.Matches(msg, v.keys.Filter):
				if !v.filterBar.IsActive() {
					v.filterBar.Activate()
					return v, nil
				}
			case key.Matches(msg, v.keys.Up):
				v.table.MoveUp()
			case key.Matches(msg, v.keys.Down):
				v.table.MoveDown()
			case key.Matches(msg, v.keys.Enter):
				if id := v.table.SelectedID(); id != "" {
					for i := range v.data {
						if v.data[i].ID == id {
							n := v.data[i]
							detail := NewNodeDetailView(v.client, &n)
							return v, func() tea.Msg {
								return tui.PushDetailMsg{View: detail, Breadcrumb: n.NodeName}
							}
						}
					}
				}
			}
		}
	}
	return v, nil
}

func (v *NodesView) buildTable() {
	headers := []string{"Name", "Cluster", "Instance", "Ready", "CPU", "Memory", "Pods", "Spot", "$/hr"}
	rows := make([][]string, 0, len(v.data))
	rowIDs := make([]string, 0, len(v.data))
	for _, n := range v.data {
		rows = append(rows, []string{
			n.NodeName,
			n.ClusterName,
			output.FormatOptionalString(n.InstanceType),
			output.FormatBool(n.IsReady),
			output.FormatFloat(n.CPUAllocatable, 1),
			output.FormatMemoryGB(n.MemoryAllocatableGB),
			output.FormatIntPtr(n.PodCount),
			output.FormatBool(n.SpotInstance),
			output.FormatCost(n.HourlyCost),
		})
		rowIDs = append(rowIDs, n.ID)
	}
	v.table = components.NewTable(headers, rows)
	v.table.RowIDs = rowIDs
}

func (v *NodesView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading nodes..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.table == nil {
		return "No nodes found"
	}
	v.table.Width = width
	v.table.ViewportHeight = height - 5

	header := tui.HelpStyle.Render(fmt.Sprintf("Showing %d of %d nodes", len(v.data), v.total))
	var parts []string
	parts = append(parts, header)
	parts = append(parts, "")
	if v.filterBar.IsActive() {
		v.filterBar.Width = width
		parts = append(parts, v.filterBar.Render())
		v.table.ViewportHeight--
	}
	parts = append(parts, v.table.Render())
	return strings.Join(parts, "\n")
}
