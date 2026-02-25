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

type NamespacesView struct {
	client    *api.Client
	data      []types.NamespaceResponse
	table     *components.Table
	loading   bool
	err       error
	spinner   spinner.Model
	filterBar *components.FilterBar
	keys      tui.KeyMap
}

func NewNamespacesView(client *api.Client) *NamespacesView {
	return &NamespacesView{
		client:    client,
		loading:   true,
		spinner:   components.NewSpinner(),
		filterBar: components.NewFilterBar(),
		keys:      tui.DefaultKeyMap(),
	}
}

func (v *NamespacesView) Title() string { return "Namespaces" }

func (v *NamespacesView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *NamespacesView) loadData() tea.Cmd {
	return func() tea.Msg {
		data, err := v.client.GetNamespaces(context.Background(), "", "", "")
		return tui.DataLoadedMsg{View: tui.ViewNamespaces, Data: data, Err: err}
	}
}

func (v *NamespacesView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.DataLoadedMsg:
		if msg.View == tui.ViewNamespaces {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				resp, ok := msg.Data.(*types.NamespaceListResponse)
				if !ok {
					return v, nil
				}
				v.data = resp.Namespaces
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
				if idx := v.table.SelectedIndex(); idx >= 0 && idx < len(v.data) {
					ns := v.data[idx]
					detail := NewNamespaceDetailView(v.client, ns.Name, ns.ClusterID)
					return v, func() tea.Msg {
						return tui.PushDetailMsg{View: detail, Breadcrumb: ns.Name}
					}
				}
			}
		}
	}
	return v, nil
}

func (v *NamespacesView) buildTable() {
	headers := []string{"Name", "Cluster", "Pods", "Workloads", "Containers", "Efficiency", "Monthly $", "Team", "$/hr"}
	rows := make([][]string, 0, len(v.data))
	for _, n := range v.data {
		rows = append(rows, []string{
			n.Name,
			output.FormatOptionalString(n.ClusterName),
			output.FormatInt(n.PodCount),
			output.FormatInt(n.WorkloadCount),
			output.FormatIntPtr(n.ContainerCount),
			output.FormatPercentPtr(n.EfficiencyScore),
			output.FormatCostPtr(n.MonthlyCost),
			output.FormatOptionalString(n.Team),
			output.FormatCost(n.HourlyCost),
		})
	}
	v.table = components.NewTable(headers, rows)
}

func (v *NamespacesView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading namespaces..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.table == nil {
		return "No namespaces found"
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
