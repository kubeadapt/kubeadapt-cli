package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui/components"
)

const noDataMsg = "No data"

// OverviewView displays the dashboard overview with panels, charts, and metrics.
type OverviewView struct {
	client        *api.Client
	data          *types.DashboardResponse
	loading       bool
	err           error
	spinner       spinner.Model
	viewport      viewport.Model
	viewportReady bool
	contentWidth  int
	contentHeight int
}

// NewOverviewView creates a new overview view.
func NewOverviewView(client *api.Client) *OverviewView {
	return &OverviewView{
		client:  client,
		loading: true,
		spinner: components.NewSpinner(),
	}
}

func (v *OverviewView) Title() string { return "Dashboard" }

func (v *OverviewView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadData())
}

func (v *OverviewView) loadData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		// Load dashboard with default 30 days
		data, err := v.client.GetDashboard(ctx, 30)
		return tui.DataLoadedMsg{View: tui.ViewOverview, Data: data, Err: err}
	}
}

func (v *OverviewView) Update(msg tea.Msg) (View, tea.Cmd) {
	switch msg := msg.(type) {
	case tui.DataLoadedMsg:
		if msg.View == tui.ViewOverview {
			v.loading = false
			if msg.Err != nil {
				v.err = msg.Err
			} else {
				if d, ok := msg.Data.(*types.DashboardResponse); ok {
					v.data = d
				}
			}
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	case tea.MouseMsg:
		if v.viewportReady {
			var cmd tea.Cmd
			v.viewport, cmd = v.viewport.Update(msg)
			return v, cmd
		}
	case tea.KeyMsg:
		if v.viewportReady {
			var cmd tea.Cmd
			v.viewport, cmd = v.viewport.Update(msg)
			return v, cmd
		}
	}
	return v, nil
}

func (v *OverviewView) renderContent(width int) string {
	d := v.data
	var sections []string

	// --- Top row: Hero Metrics Panels ---
	// Responsive: 3-column above 80 chars, 2-column below
	panelWidth := (width - 6) / 3
	if panelWidth < 22 {
		panelWidth = 22
	}
	narrowLayout := width < 80
	if narrowLayout {
		panelWidth = (width - 3) / 2
		if panelWidth < 22 {
			panelWidth = 22
		}
	}

	// Infrastructure panel
	infraKV := components.NewKVTable()
	infraKV.Add("Clusters", fmt.Sprintf("%d", d.ClusterCount))
	infraKV.Add("Nodes", fmt.Sprintf("%d", d.NodeCount))
	infraKV.Add("Pods", fmtThousands(d.PodCount))
	infraKV.Add("Days Elapsed", fmt.Sprintf("%d/%d", d.DaysElapsed, d.DaysInMonth))
	infraPanel := components.NewPanel("Infrastructure", infraKV.Render(), panelWidth)

	// Costs panel
	costsKV := components.NewKVTable()
	costsKV.Add("Monthly Cost", output.FormatCost(d.TotalMonthlyCost))
	costsKV.Add("Hourly Cost", output.FormatCost(d.TotalHourlyCost))
	costsKV.Add("MTD Cost", output.FormatCost(d.MTDActualCost))
	costsKV.Add("Run Rate", output.FormatCost(d.RunRate))
	costsKV.Add("Savings", output.FormatCost(d.PotentialMonthlySavings))
	costsPanel := components.NewPanel("Financials", costsKV.Render(), panelWidth)

	// Efficiency panel
	effContent := ""
	if d.EfficiencyScore != nil {
		effGauge := components.NewGaugeBar("Score", *d.EfficiencyScore, panelWidth-4)
		effContent = effGauge.Render() + "\n\n"
	}
	effContent += fmt.Sprintf("Recommendations: %d", d.TotalRecommendations)
	effPanel := components.NewPanel("Efficiency", effContent, panelWidth)

	if narrowLayout {
		// Stack 2 panels on first row, 1 below
		row1 := lipgloss.JoinHorizontal(lipgloss.Top,
			infraPanel.Render(), " ", costsPanel.Render())
		sections = append(sections, row1)
		sections = append(sections, effPanel.Render())
	} else {
		topRow := lipgloss.JoinHorizontal(lipgloss.Top,
			infraPanel.Render(), " ", costsPanel.Render(), " ", effPanel.Render())
		sections = append(sections, topRow)
	}

	// --- Middle row: Cost Trends Line Chart ---
	if len(d.CostTrends) > 0 {
		chart := components.NewLineChart("Cost Trends (30 Days)", width-4, 10)
		series := make([]float64, 0, len(d.CostTrends))
		for _, p := range d.CostTrends {
			series = append(series, p.TotalCost)
		}
		chart.AddSeries("Total Cost", series)
		sections = append(sections, chart.Render())
	}

	// --- Bottom row: Top Clusters Bar Chart ---
	if len(d.TopClusters) > 0 {
		barChart := components.NewBarChart("Top Clusters by Cost", width-4)
		for _, c := range d.TopClusters {
			barChart.AddItem(c.ClusterName, c.HourlyCost)
		}
		sections = append(sections, barChart.Render())
	}

	return strings.Join(sections, "\n\n")
}

func (v *OverviewView) View(width, height int) string {
	if v.loading {
		return v.spinner.View() + " Loading dashboard..."
	}
	if v.err != nil {
		return tui.ErrorStyle.Render("Error: " + v.err.Error())
	}
	if v.data == nil {
		return noDataMsg
	}

	if !v.viewportReady {
		v.viewport = viewport.New(width, height)
		v.viewportReady = true
		v.contentWidth = width
		v.contentHeight = height
	}
	if v.contentWidth != width || v.contentHeight != height {
		v.viewport.Width = width
		v.viewport.Height = height
		v.contentWidth = width
		v.contentHeight = height
	}

	content := v.renderContent(width)
	v.viewport.SetContent(content)

	return v.viewport.View()
}

func fmtThousands(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d", n)
}
