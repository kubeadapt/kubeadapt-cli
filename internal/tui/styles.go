package tui

import "github.com/charmbracelet/lipgloss"

// Brand colors.
var (
	ColorPrimary      = lipgloss.Color("#3B82F6")
	ColorGreen        = lipgloss.Color("#22C55E")
	ColorYellow       = lipgloss.Color("#EAB308")
	ColorRed          = lipgloss.Color("#EF4444")
	ColorGray         = lipgloss.Color("#6B7280")
	ColorDarkGray     = lipgloss.Color("#374151")
	ColorWhite        = lipgloss.Color("#F9FAFB")
	ColorSurface      = lipgloss.Color("#1F2937")
	ColorPrimaryDark  = lipgloss.Color("#1E40AF")
	ColorPrimaryLight = lipgloss.Color("#60A5FA")
)

// TUI styles.
var (
	AppStyle = lipgloss.NewStyle()

	NavbarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1)

	NavbarActiveTab = lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorWhite).
			Bold(true).
			Padding(0, 2)

	NavbarInactiveTab = lipgloss.NewStyle().
				Background(lipgloss.Color("#1F2937")).
				Foreground(ColorGray).
				Padding(0, 2)

	ContentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(ColorGray).
			Padding(0, 1)

	StatusBarKeyStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#374151")).
				Foreground(ColorWhite).
				Bold(true).
				Padding(0, 1)

	StatusBarDescStyle = lipgloss.NewStyle().
				Foreground(ColorGray).
				Padding(0, 1)

	CardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	CardValueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary).
				Padding(0, 1)

	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TableSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#1E3A5F")).
				Foreground(ColorWhite).
				Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorYellow).
			Padding(1, 2).
			Width(50).
			Align(lipgloss.Center)

	SidebarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			BorderRight(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorDarkGray)

	SidebarItemActive = lipgloss.NewStyle().
				Background(lipgloss.Color("#1E40AF")).
				Foreground(ColorWhite).
				Bold(true).
				Padding(0, 1)

	SidebarItemInactive = lipgloss.NewStyle().
				Background(lipgloss.Color("#1F2937")).
				Foreground(ColorGray).
				Padding(0, 1)

	SidebarTitle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	FilterBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1)

	FilterInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorPrimary).
				Foreground(ColorWhite)

	FilterLabelStyle = lipgloss.NewStyle().
				Foreground(ColorGray)

	ScrollIndicatorStyle = lipgloss.NewStyle().
				Foreground(ColorGray).
				Align(lipgloss.Right)

	SubTabActive   = NavbarActiveTab
	SubTabInactive = NavbarInactiveTab
)
