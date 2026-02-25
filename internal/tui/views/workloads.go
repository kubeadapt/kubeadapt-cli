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

// WorkloadsView displays workloads in a table.
type WorkloadsView struct {
	client    *api.Client
	data      []types.WorkloadResponse
	total     int
	table     *components.Table
	loading   bool
	err       error
	spinner   spinner.Model
	filterBar *components.FilterBar
	keys      tui.KeyMap
}

func NewWorkloadsView(client *api.Client) *WorkloadsView {
	return &WorkloadsView{
		client:    client,
		loading:   true,
		spinner:   components.NewSpinner(),
		filterBar: components.NewFilterBar(),
		keys:      tui.DefaultKeyMap(),
	}
}

func (v *WorkloadsView) Title() string { return "Workloads" }

func (v *WorkloadsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *WorkloadsView) loadData() tea.Cmd {
	return func() tea.Msg {
		data, err := v.client.GetWorkloads(context.Background(), "", "", "", 50, 0)
		return tui.DataLoadedMsg{View: tui.ViewWorkloads, Data: data, Err: err}
	}
}

func (v *WorkloadsView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.DataLoadedMsg:
		if msg.View == tui.ViewWorkloads {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				resp, ok := msg.Data.(*types.WorkloadListResponse)
				if !ok {
					return v, nil
				}
				v.data = resp.Workloads
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
							w := v.data[i]
							detail := NewWorkloadDetailView(v.client, &w)
							return v, func() tea.Msg {
								return tui.PushDetailMsg{View: detail, Breadcrumb: w.WorkloadName}
							}
						}
					}
				}
			}
		}
	}
	return v, nil
}

func (v *WorkloadsView) buildTable() {
	headers := []string{"Name", "Kind", "Namespace", "Cluster", "Replicas", "Efficiency", "Monthly $", "$/hr"}
	rows := make([][]string, 0, len(v.data))
	rowIDs := make([]string, 0, len(v.data))
	for _, w := range v.data {
		rows = append(rows, []string{
			w.WorkloadName,
			w.WorkloadKind,
			w.Namespace,
			w.ClusterName,
			fmt.Sprintf("%d/%d", w.AvailableReplicas, w.Replicas),
			output.FormatPercentPtr(w.EfficiencyScore),
			output.FormatCostPtr(w.MonthlyCost),
			output.FormatCost(w.HourlyCost),
		})
		rowIDs = append(rowIDs, w.ID)
	}
	v.table = components.NewTable(headers, rows)
	v.table.RowIDs = rowIDs
}

func (v *WorkloadsView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading workloads..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.table == nil {
		return "No workloads found"
	}
	v.table.Width = width
	v.table.ViewportHeight = height - 5 // info_header(1) + gap(1) + header(1) + separator(1) + scroll(1)

	header := tui.HelpStyle.Render(fmt.Sprintf("Showing %d of %d workloads", len(v.data), v.total))
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
