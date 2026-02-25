package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// BarChart renders horizontal bar charts for cost breakdowns.
type BarChart struct {
	Title   string
	Items   []BarChartItem
	Width   int
	MaxBars int
}

// BarChartItem represents a single bar.
type BarChartItem struct {
	Label string
	Value float64
}

// NewBarChart creates a new horizontal bar chart.
func NewBarChart(title string, width int) *BarChart {
	if width < 30 {
		width = 30
	}
	return &BarChart{Title: title, Width: width, MaxBars: 5}
}

// AddItem adds a bar item.
func (c *BarChart) AddItem(label string, value float64) {
	c.Items = append(c.Items, BarChartItem{Label: label, Value: value})
}

// Render draws the bar chart.
func (c *BarChart) Render() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorPrimary)
	header := titleStyle.Render(c.Title)

	if len(c.Items) == 0 {
		return header + "\n" + lipgloss.NewStyle().Foreground(tui.ColorGray).Render("  No data")
	}

	items := c.Items
	if c.MaxBars > 0 && len(items) > c.MaxBars {
		items = items[:c.MaxBars]
	}

	// Find max value and max label width
	maxVal := 0.0
	maxLabelLen := 0
	for _, item := range items {
		if item.Value > maxVal {
			maxVal = item.Value
		}
		if len(item.Label) > maxLabelLen {
			maxLabelLen = len(item.Label)
		}
	}
	if maxLabelLen > 20 {
		maxLabelLen = 20
	}

	barMaxWidth := c.Width - maxLabelLen - 15 // label + value + padding
	if barMaxWidth < 10 {
		barMaxWidth = 10
	}

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")

	colors := []lipgloss.Color{tui.ColorPrimary, tui.ColorGreen, tui.ColorYellow, tui.ColorRed, tui.ColorGray}

	for i, item := range items {
		label := item.Label
		if len(label) > maxLabelLen {
			label = label[:maxLabelLen-3] + "..."
		}

		barLen := 0
		if maxVal > 0 {
			barLen = int(float64(barMaxWidth) * item.Value / maxVal)
		}
		if barLen < 1 && item.Value > 0 {
			barLen = 1
		}

		colorIdx := i % len(colors)
		bar := lipgloss.NewStyle().Foreground(colors[colorIdx]).Render(strings.Repeat("█", barLen))
		valStr := lipgloss.NewStyle().Foreground(tui.ColorWhite).Render(fmt.Sprintf(" $%.0f", item.Value))
		labelStr := lipgloss.NewStyle().Foreground(tui.ColorGray).Width(maxLabelLen + 1).Render(label)

		sb.WriteString("  " + labelStr + bar + valStr + "\n")
	}

	return sb.String()
}
