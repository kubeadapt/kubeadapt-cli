package views

import (
	"context"
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

type NodeGroupsView struct {
	client    *api.Client
	data      []types.NodeGroupResponse
	table     *components.Table
	loading   bool
	err       error
	spinner   spinner.Model
	filterBar *components.FilterBar
	keys      tui.KeyMap
}

func NewNodeGroupsView(client *api.Client) *NodeGroupsView {
	return &NodeGroupsView{
		client:    client,
		loading:   true,
		spinner:   components.NewSpinner(),
		filterBar: components.NewFilterBar(),
		keys:      tui.DefaultKeyMap(),
	}
}

func (v *NodeGroupsView) Title() string { return "Node Groups" }

func (v *NodeGroupsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *NodeGroupsView) loadData() tea.Cmd {
	return func() tea.Msg {
		data, err := v.client.GetNodeGroups(context.Background(), "")
		return tui.DataLoadedMsg{View: tui.ViewNodeGroups, Data: data, Err: err}
	}
}

func (v *NodeGroupsView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.DataLoadedMsg:
		if msg.View == tui.ViewNodeGroups {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				resp := msg.Data.(*types.NodeGroupListResponse)
				v.data = resp.NodeGroups
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
					for _, g := range v.data {
						if g.ID == id {
							detail := NewNodeGroupDetailView(v.client, g.Name, g.ClusterID)
							return v, func() tea.Msg {
								return tui.PushDetailMsg{View: detail, Breadcrumb: g.Name}
							}
						}
					}
				}
			}
		}
	}
	return v, nil
}

func (v *NodeGroupsView) buildTable() {
	headers := []string{"Name", "Cluster", "Instance", "Nodes", "CPU", "Memory", "Spot %", "$/hr"}
	var rows [][]string
	var rowIDs []string
	for _, g := range v.data {
		rows = append(rows, []string{
			g.Name,
			output.FormatOptionalString(g.ClusterName),
			output.FormatOptionalString(g.InstanceType),
			output.FormatInt(g.NodeCount),
			output.FormatFloatPtr(g.TotalCPUCores, 1),
			output.FormatMemoryGBPtr(g.TotalMemoryGB),
			output.FormatPercentPtr(g.SpotPercentage),
			output.FormatCostPtr(g.HourlyCost),
		})
		rowIDs = append(rowIDs, g.ID)
	}
	v.table = components.NewTable(headers, rows)
	v.table.RowIDs = rowIDs
}

func (v *NodeGroupsView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading node groups..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.table == nil {
		return "No node groups found"
	}
	v.table.Width = width
	v.table.ViewportHeight = height - 3

	var parts []string
	if v.filterBar.IsActive() {
		v.filterBar.Width = width
		parts = append(parts, v.filterBar.Render())
		v.table.ViewportHeight--
	}
	parts = append(parts, v.table.Render())
	return strings.Join(parts, "\n")
}
