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

type CostsView struct {
	client      *api.Client
	teams       []types.TeamCostResponse
	departments []types.DepartmentCostResponse
	teamTable   *components.Table
	deptTable   *components.Table
	activeTab   int // 0=teams, 1=departments
	loading     bool
	err         error
	spinner     spinner.Model
	filterBar   *components.FilterBar
	keys        tui.KeyMap
}

func NewCostsView(client *api.Client) *CostsView {
	return &CostsView{
		client:    client,
		loading:   true,
		spinner:   components.NewSpinner(),
		filterBar: components.NewFilterBar(),
		keys:      tui.DefaultKeyMap(),
	}
}

func (v *CostsView) Title() string { return "Costs" }

func (v *CostsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *CostsView) loadData() tea.Cmd {
	return func() tea.Msg {
		teams, err := v.client.GetCostsTeams(context.Background(), "")
		if err != nil {
			return tui.DataLoadedMsg{View: tui.ViewCosts, Err: err}
		}
		depts, err := v.client.GetCostsDepartments(context.Background(), "")
		if err != nil {
			return tui.DataLoadedMsg{View: tui.ViewCosts, Err: err}
		}
		return tui.DataLoadedMsg{
			View: tui.ViewCosts,
			Data: &tui.CostsData{Teams: teams.Teams, Departments: depts.Departments},
		}
	}
}

func (v *CostsView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.DataLoadedMsg:
		if msg.View == tui.ViewCosts {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				data, ok := msg.Data.(*tui.CostsData)
				if !ok {
					return v, nil
				}
				v.teams = data.Teams
				v.departments = data.Departments
				v.buildTables()
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	case tea.MouseMsg:
		if v.activeTable() != nil {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				v.activeTable().MoveUp()
			case tea.MouseButtonWheelDown:
				v.activeTable().MoveDown()
			}
		}
	case tea.KeyMsg:
		if v.filterBar.IsActive() {
			if v.filterBar.HandleKey(msg) {
				if v.filterBar.IsActive() {
					if v.activeTable() != nil {
						v.activeTable().ApplyFilter(v.filterBar.GetQuery())
					}
				} else {
					if v.activeTable() != nil {
						v.activeTable().ClearFilter()
					}
				}
				return v, nil
			}
		}
		switch {
		case key.Matches(msg, v.keys.Filter):
			if !v.filterBar.IsActive() {
				v.filterBar.Activate()
				return v, nil
			}
		case key.Matches(msg, v.keys.Tab):
			if v.filterBar.IsActive() {
				v.filterBar.Deactivate()
				if v.activeTable() != nil {
					v.activeTable().ClearFilter()
				}
			}
			v.activeTab = (v.activeTab + 1) % 2
		case key.Matches(msg, v.keys.Up):
			if v.activeTable() != nil {
				v.activeTable().MoveUp()
			}
		case key.Matches(msg, v.keys.Down):
			if v.activeTable() != nil {
				v.activeTable().MoveDown()
			}
		}
	}
	return v, nil
}

func (v *CostsView) activeTable() *components.Table {
	if v.activeTab == 0 {
		return v.teamTable
	}
	return v.deptTable
}

func (v *CostsView) buildTables() {
	// Teams table
	headers := []string{"Team", "Namespaces", "Workloads", "Pods", "CPU", "Memory", "$/hr", "$/mo"}
	rows := make([][]string, 0, len(v.teams))
	for _, c := range v.teams {
		rows = append(rows, []string{
			c.Team,
			output.FormatInt(c.NamespaceCount),
			output.FormatInt(c.WorkloadCount),
			output.FormatInt(c.PodCount),
			output.FormatFloat(c.TotalCPUCores, 1),
			output.FormatMemoryGB(c.TotalMemoryGB),
			output.FormatCost(c.HourlyCost),
			output.FormatCost(c.MonthlyCost),
		})
	}
	v.teamTable = components.NewTable(headers, rows)

	// Departments table
	headers = []string{"Department", "Namespaces", "Workloads", "Pods", "CPU", "Memory", "$/hr", "$/mo"}
	deptRows := make([][]string, 0, len(v.departments))
	for _, c := range v.departments {
		deptRows = append(deptRows, []string{
			c.Department,
			output.FormatInt(c.NamespaceCount),
			output.FormatInt(c.WorkloadCount),
			output.FormatInt(c.PodCount),
			output.FormatFloat(c.TotalCPUCores, 1),
			output.FormatMemoryGB(c.TotalMemoryGB),
			output.FormatCost(c.HourlyCost),
			output.FormatCost(c.MonthlyCost),
		})
	}
	v.deptTable = components.NewTable(headers, deptRows)
}

func (v *CostsView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading costs..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}

	// Tab headers
	teamTab := "  Teams  "
	deptTab := "  Departments  "
	if v.activeTab == 0 {
		teamTab = tui.NavbarActiveTab.Render(teamTab)
		deptTab = tui.NavbarInactiveTab.Render(deptTab)
	} else {
		teamTab = tui.NavbarInactiveTab.Render(teamTab)
		deptTab = tui.NavbarActiveTab.Render(deptTab)
	}

	table := v.activeTable()
	if table == nil {
		return teamTab + deptTab + "\n\nNo data"
	}
	table.Width = width
	table.ViewportHeight = height - 5 // tab(1) + gap(1) + header(1) + separator(1) + scroll(1)

	var parts []string
	parts = append(parts, teamTab+deptTab)
	parts = append(parts, "")
	if v.filterBar.IsActive() {
		v.filterBar.Width = width
		parts = append(parts, v.filterBar.Render())
		table.ViewportHeight--
	}
	parts = append(parts, table.Render())
	return strings.Join(parts, "\n")
}
