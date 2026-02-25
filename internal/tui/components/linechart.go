package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// LineChart wraps asciigraph to render ASCII line charts.
type LineChart struct {
	Title  string
	Series [][]float64
	Labels []string // legend labels for each series
	Width  int
	Height int
}

// NewLineChart creates a new line chart.
func NewLineChart(title string, width, height int) *LineChart {
	if width < 40 {
		width = 40
	}
	if height < 5 {
		height = 5
	}
	return &LineChart{
		Title:  title,
		Width:  width,
		Height: height,
	}
}

// AddSeries adds a data series with a label.
func (c *LineChart) AddSeries(label string, data []float64) {
	c.Labels = append(c.Labels, label)
	c.Series = append(c.Series, data)
}

// Render draws the chart.
func (c *LineChart) Render() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorPrimary)
	header := titleStyle.Render(c.Title)

	if len(c.Series) == 0 {
		return header + "\n" + lipgloss.NewStyle().Foreground(tui.ColorGray).Render("  No data available")
	}

	chartWidth := c.Width - 10 // leave room for y-axis labels
	if chartWidth < 20 {
		chartWidth = 20
	}

	var chart string
	if len(c.Series) == 1 {
		chart = asciigraph.Plot(c.Series[0],
			asciigraph.Width(chartWidth),
			asciigraph.Height(c.Height),
		)
	} else {
		chart = asciigraph.PlotMany(c.Series,
			asciigraph.Width(chartWidth),
			asciigraph.Height(c.Height),
			asciigraph.SeriesColors(
				asciigraph.Blue,
				asciigraph.Green,
				asciigraph.Yellow,
				asciigraph.Red,
			),
		)
	}

	// Build legend
	legend := ""
	if len(c.Labels) > 0 {
		colors := []lipgloss.Color{tui.ColorPrimary, tui.ColorGreen, tui.ColorYellow, tui.ColorRed}
		for i, label := range c.Labels {
			colorIdx := i % len(colors)
			s := lipgloss.NewStyle().Foreground(colors[colorIdx])
			if i > 0 {
				legend += "  "
			}
			legend += s.Render("━ " + label)
		}
		legend = "\n" + legend
	}

	return header + "\n" + chart + legend
}
