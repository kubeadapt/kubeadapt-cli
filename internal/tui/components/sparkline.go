package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// Sparkline renders an inline trend visualization like:
//
//	Cost 7d: ▁▂▃▅▇█▅ $1,240/day
type Sparkline struct {
	Label  string
	Values []float64
	Suffix string
}

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// NewSparkline creates a new sparkline.
func NewSparkline(label string, values []float64, suffix string) Sparkline {
	return Sparkline{Label: label, Values: values, Suffix: suffix}
}

// Render draws the sparkline.
func (s Sparkline) Render() string {
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(tui.ColorPrimary)
	suffixStyle := lipgloss.NewStyle().Foreground(tui.ColorGray)

	if len(s.Values) == 0 {
		return labelStyle.Render(s.Label+": ") + suffixStyle.Render("no data")
	}

	// Find min/max
	min, max := s.Values[0], s.Values[0]
	for _, v := range s.Values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	var sb strings.Builder
	for _, v := range s.Values {
		idx := 0
		if max > min {
			idx = int((v - min) / (max - min) * float64(len(sparkChars)-1))
		}
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		sb.WriteRune(sparkChars[idx])
	}

	sparkStyle := lipgloss.NewStyle().Foreground(tui.ColorGreen)
	result := labelStyle.Render(s.Label+": ") + sparkStyle.Render(sb.String())
	if s.Suffix != "" {
		result += " " + suffixStyle.Render(s.Suffix)
	}
	return result
}

// FormatSparkSuffix formats a value with appropriate unit for sparkline suffix.
func FormatSparkSuffix(value float64, unit string) string {
	if value >= 1000 {
		return fmt.Sprintf("$%.0f%s", value, unit)
	}
	return fmt.Sprintf("$%.2f%s", value, unit)
}
