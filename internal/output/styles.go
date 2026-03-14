package output

import "github.com/charmbracelet/lipgloss"

// Kubeadapt brand colors.
var (
	ColorPrimary = lipgloss.Color("#3B82F6") // Blue
	ColorGreen   = lipgloss.Color("#22C55E") // Green - savings
	ColorYellow  = lipgloss.Color("#EAB308") // Yellow - warnings
	ColorRed     = lipgloss.Color("#EF4444") // Red - errors
	ColorGray    = lipgloss.Color("#6B7280") // Gray - muted text
	ColorWhite   = lipgloss.Color("#F9FAFB") // White
)

// Common styles.
var (
	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	StyleSuccess = lipgloss.NewStyle().
			Foreground(ColorGreen)

	StyleWarning = lipgloss.NewStyle().
			Foreground(ColorYellow)

	StyleError = lipgloss.NewStyle().
			Foreground(ColorRed)

	StyleMuted = lipgloss.NewStyle().
			Foreground(ColorGray)

	StyleBold = lipgloss.NewStyle().
			Bold(true)
)
