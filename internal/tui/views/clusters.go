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

// ClustersView displays clusters in a table.
type ClustersView struct {
	client    *api.Client
	clusters  []types.ClusterResponse
	table     *components.Table
	loading   bool
	err       error
	spinner   spinner.Model
	filterBar *components.FilterBar
	keys      tui.KeyMap
}

// NewClustersView creates a new clusters view.
func NewClustersView(client *api.Client) *ClustersView {
	return &ClustersView{
		client:    client,
		loading:   true,
		spinner:   components.NewSpinner(),
		filterBar: components.NewFilterBar(),
		keys:      tui.DefaultKeyMap(),
	}
}

func (v *ClustersView) Title() string { return "Clusters" }

func (v *ClustersView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *ClustersView) loadData() tea.Cmd {
	return func() tea.Msg {
		data, err := v.client.GetClusters(context.Background())
		return tui.DataLoadedMsg{View: tui.ViewClusters, Data: data, Err: err}
	}
}

func (v *ClustersView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.DataLoadedMsg:
		if msg.View == tui.ViewClusters {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				resp := msg.Data.(*types.ClusterListResponse)
				v.clusters = resp.Clusters
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
					// Find the cluster by ID
					for _, c := range v.clusters {
						if c.ID == id {
							detail := NewClusterDetailView(v.client, c.ID)
							return v, func() tea.Msg {
								return tui.PushDetailMsg{View: detail, Breadcrumb: c.Name}
							}
						}
					}
				}
			}
		}
	}
	return v, nil
}

func (v *ClustersView) buildTable() {
	headers := []string{"ID", "Name", "Provider", "Region", "Status", "Nodes", "Efficiency", "Monthly $", "$/hr"}
	var rows [][]string
	var rowIDs []string
	for _, c := range v.clusters {
		rows = append(rows, []string{
			output.ShortID(c.ID),
			c.Name,
			c.Provider,
			output.FormatOptionalString(c.Region),
			c.Status,
			output.FormatInt(c.NodeCount),
			output.FormatPercentPtr(c.EfficiencyScore),
			output.FormatCostPtr(c.MonthlyCost),
			output.FormatCost(c.HourlyCost),
		})
		rowIDs = append(rowIDs, c.ID)
	}
	v.table = components.NewTable(headers, rows)
	v.table.RowIDs = rowIDs
}

func (v *ClustersView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading clusters..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.table == nil {
		return "No clusters found"
	}
	v.table.Width = width
	v.table.ViewportHeight = height - 3 // header(1) + separator(1) + scroll indicator(1)

	var parts []string
	if v.filterBar.IsActive() {
		v.filterBar.Width = width
		parts = append(parts, v.filterBar.Render())
		v.table.ViewportHeight--
	}
	parts = append(parts, v.table.Render())
	return strings.Join(parts, "\n")
}
