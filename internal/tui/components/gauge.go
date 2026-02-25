package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// GaugeBar renders a horizontal utilization bar like:
//
//	CPU [████████░░░░] 62%
type GaugeBar struct {
	Label   string
	Percent float64
	Width   int
}

// NewGaugeBar creates a new gauge bar.
func NewGaugeBar(label string, percent float64, width int) GaugeBar {
	if width < 20 {
		width = 20
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return GaugeBar{Label: label, Percent: percent, Width: width}
}

// Render draws the gauge bar.
func (g GaugeBar) Render() string {
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorWhite).Width(8)
	pctStr := fmt.Sprintf("%5.1f%%", g.Percent)

	barWidth := g.Width - 8 - len(pctStr) - 4 // label + pct + brackets/spaces
	if barWidth < 10 {
		barWidth = 10
	}

	filled := int(float64(barWidth) * g.Percent / 100)
	if filled > barWidth {
		filled = barWidth
	}

	filledStr := strings.Repeat("█", filled)
	emptyStr := strings.Repeat("░", barWidth-filled)

	var barColor lipgloss.Color
	switch {
	case g.Percent >= 90:
		barColor = tui.ColorRed
	case g.Percent >= 70:
		barColor = tui.ColorYellow
	default:
		barColor = tui.ColorGreen
	}

	bar := lipgloss.NewStyle().Foreground(barColor).Render(filledStr) +
		lipgloss.NewStyle().Foreground(tui.ColorDarkGray).Render(emptyStr)

	pctStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorWhite)

	return labelStyle.Render(g.Label) + " [" + bar + "] " + pctStyle.Render(pctStr)
}
