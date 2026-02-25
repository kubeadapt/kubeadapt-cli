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

type RecommendationsView struct {
	client    *api.Client
	data      []types.RecommendationResponse
	total     int
	table     *components.Table
	loading   bool
	err       error
	spinner   spinner.Model
	filterBar *components.FilterBar
	keys      tui.KeyMap
}

func NewRecommendationsView(client *api.Client) *RecommendationsView {
	return &RecommendationsView{
		client:    client,
		loading:   true,
		spinner:   components.NewSpinner(),
		filterBar: components.NewFilterBar(),
		keys:      tui.DefaultKeyMap(),
	}
}

func (v *RecommendationsView) Title() string { return "Recommendations" }

func (v *RecommendationsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *RecommendationsView) loadData() tea.Cmd {
	return func() tea.Msg {
		data, err := v.client.GetRecommendations(context.Background(), "", "", "", 50, 0)
		return tui.DataLoadedMsg{View: tui.ViewRecommendations, Data: data, Err: err}
	}
}

func (v *RecommendationsView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.DataLoadedMsg:
		if msg.View == tui.ViewRecommendations {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				resp := msg.Data.(*types.RecommendationListResponse)
				v.data = resp.Recommendations
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
			}
		}
	}
	return v, nil
}

func (v *RecommendationsView) buildTable() {
	headers := []string{"ID", "Type", "Cluster", "Resource", "Priority", "Status", "Monthly Savings"}
	var rows [][]string
	for _, r := range v.data {
		resource := output.FormatOptionalString(r.ResourceName)
		if ns := output.FormatOptionalString(r.Namespace); ns != "-" && ns != "" {
			resource = ns + "/" + resource
		}
		rows = append(rows, []string{
			output.ShortID(r.ID),
			r.RecommendationType,
			r.ClusterName,
			resource,
			output.FormatOptionalString(r.Priority),
			r.Status,
			output.FormatCost(r.EstimatedMonthlySavings),
		})
	}
	v.table = components.NewTable(headers, rows)
}

func (v *RecommendationsView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading recommendations..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.table == nil {
		return "No recommendations found"
	}
	v.table.Width = width
	v.table.ViewportHeight = height - 5

	header := tui.HelpStyle.Render(fmt.Sprintf("Showing %d of %d recommendations", len(v.data), v.total))
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
