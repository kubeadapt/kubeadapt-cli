package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/api"

	"github.com/charmbracelet/bubbles/key"
)

// ViewInterface is the interface that views must implement.
type ViewInterface interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (ViewInterface, tea.Cmd)
	View(width, height int) string
	Title() string
}

// SidebarRenderFn renders the sidebar for a given active view and height.
type SidebarRenderFn func(activeView ViewID, height int) string

// App is the root bubbletea model.
type App struct {
	client      *api.Client
	activeView  ViewID
	views       map[ViewID]ViewInterface
	width       int
	height      int
	keys        KeyMap
	detailStack []ViewInterface // Navigation stack for drill-down views
	breadcrumbs []string        // Breadcrumb labels for the stack
	sidebar     SidebarRenderFn
}

// NewApp creates a new App model.
func NewApp(client *api.Client) *App {
	return &App{
		client:     client,
		activeView: ViewOverview,
		views:      make(map[ViewID]ViewInterface),
		keys:       DefaultKeyMap(),
	}
}

// SetSidebar sets the sidebar render function, injected to avoid circular imports.
func (a *App) SetSidebar(fn SidebarRenderFn) {
	a.sidebar = fn
}

// RegisterView registers a view with the app.
func (a *App) RegisterView(id ViewID, view ViewInterface) {
	a.views[id] = view
}

// InDetailMode returns true if a detail view is on the stack.
func (a *App) InDetailMode() bool {
	return len(a.detailStack) > 0
}

func (a *App) Init() tea.Cmd {
	// Initialize the active view
	if v, ok := a.views[a.activeView]; ok {
		return v.Init()
	}
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case PushDetailMsg:
		a.detailStack = append(a.detailStack, msg.View)
		a.breadcrumbs = append(a.breadcrumbs, msg.Breadcrumb)
		return a, msg.View.Init()

	case PopDetailMsg:
		if len(a.detailStack) > 0 {
			a.detailStack = a.detailStack[:len(a.detailStack)-1]
			a.breadcrumbs = a.breadcrumbs[:len(a.breadcrumbs)-1]
		}
		return a, nil

	case tea.KeyMsg:
		// Esc pops the detail stack
		if a.InDetailMode() && key.Matches(msg, a.keys.Back) {
			a.detailStack = a.detailStack[:len(a.detailStack)-1]
			a.breadcrumbs = a.breadcrumbs[:len(a.breadcrumbs)-1]
			return a, nil
		}

		// Global keybindings
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.Nav1):
			a.clearDetailStack()
			return a, a.switchView(ViewOverview)
		case key.Matches(msg, a.keys.Nav2):
			a.clearDetailStack()
			return a, a.switchView(ViewClusters)
		case key.Matches(msg, a.keys.Nav3):
			a.clearDetailStack()
			return a, a.switchView(ViewWorkloads)
		case key.Matches(msg, a.keys.Nav4):
			a.clearDetailStack()
			return a, a.switchView(ViewNodes)
		case key.Matches(msg, a.keys.Nav5):
			a.clearDetailStack()
			return a, a.switchView(ViewRecommendations)
		case key.Matches(msg, a.keys.Nav6):
			a.clearDetailStack()
			return a, a.switchView(ViewCosts)
		case key.Matches(msg, a.keys.Nav7):
			a.clearDetailStack()
			return a, a.switchView(ViewNamespaces)
		case key.Matches(msg, a.keys.Nav8):
			a.clearDetailStack()
			return a, a.switchView(ViewNodeGroups)
		case key.Matches(msg, a.keys.Nav9):
			a.clearDetailStack()
			return a, a.switchView(ViewPVs)
		case key.Matches(msg, a.keys.Nav0):
			a.clearDetailStack()
			return a, a.switchView(ViewHelp)
		case key.Matches(msg, a.keys.Help):
			a.clearDetailStack()
			return a, a.switchView(ViewHelp)
		case key.Matches(msg, a.keys.Refresh):
			if a.InDetailMode() {
				top := a.detailStack[len(a.detailStack)-1]
				return a, top.Init()
			}
			if v, ok := a.views[a.activeView]; ok {
				return a, v.Init()
			}
		}
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			break
		}
		sidebarWidth := 20
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.X < sidebarWidth {
				itemIdx := msg.Y - 2
				if itemIdx >= 0 && itemIdx < 10 {
					viewIDs := []ViewID{
						ViewOverview, ViewClusters, ViewWorkloads, ViewNodes,
						ViewRecommendations, ViewCosts, ViewNamespaces,
						ViewNodeGroups, ViewPVs, ViewHelp,
					}
					a.clearDetailStack()
					return a, a.switchView(viewIDs[itemIdx])
				}
			}
		case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown:
		}
	}

	// Forward to detail view if in detail mode
	if a.InDetailMode() {
		top := a.detailStack[len(a.detailStack)-1]
		newView, cmd := top.Update(msg)
		a.detailStack[len(a.detailStack)-1] = newView
		return a, cmd
	}

	// Forward to active view
	if v, ok := a.views[a.activeView]; ok {
		newView, cmd := v.Update(msg)
		a.views[a.activeView] = newView
		return a, cmd
	}

	return a, nil
}

func (a *App) clearDetailStack() {
	a.detailStack = nil
	a.breadcrumbs = nil
}

func (a *App) switchView(id ViewID) tea.Cmd {
	if a.activeView == id && !a.InDetailMode() {
		return nil
	}
	a.activeView = id

	// Initialize view if not yet loaded
	if v, ok := a.views[id]; ok {
		return v.Init()
	}
	return nil
}

func (a *App) View() string {
	if a.width < 40 || a.height < 10 {
		return "Terminal too small. Minimum: 40x10"
	}

	sidebarWidth := 20
	contentWidth := a.width - sidebarWidth
	contentHeight := a.height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}
	if contentWidth < 1 {
		contentWidth = 1
	}

	sidebarStr := ""
	if a.sidebar != nil {
		sidebarStr = a.sidebar(a.activeView, contentHeight)
	}

	content := ""
	if a.InDetailMode() {
		top := a.detailStack[len(a.detailStack)-1]
		var crumbs []string
		if v, ok := a.views[a.activeView]; ok {
			crumbs = append(crumbs, v.Title())
		}
		crumbs = append(crumbs, a.breadcrumbs...)
		breadcrumbStr := renderBreadcrumb(crumbs)
		content = breadcrumbStr + "\n" + top.View(contentWidth-2, contentHeight-2)
	} else if v, ok := a.views[a.activeView]; ok {
		content = v.View(contentWidth-2, contentHeight)
	} else {
		content = "View not available"
	}

	contentBox := lipgloss.NewStyle().
		Width(contentWidth).
		Height(contentHeight).
		Padding(0, 1).
		Render(content)

	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebarStr, contentBox)
	statusbar := renderStatusBar(a.width, a.InDetailMode())

	return mainArea + "\n" + statusbar
}

func renderBreadcrumb(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	sep := lipgloss.NewStyle().Foreground(ColorDarkGray).Render(" > ")
	activeStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorWhite)
	inactiveStyle := lipgloss.NewStyle().Foreground(ColorGray)

	var rendered []string
	for i, part := range parts {
		if i == len(parts)-1 {
			rendered = append(rendered, activeStyle.Render(part))
		} else {
			rendered = append(rendered, inactiveStyle.Render(part))
		}
	}
	return strings.Join(rendered, sep)
}

func renderStatusBar(width int, inDetailMode bool) string {
	var hints []struct{ key, desc string }

	if inDetailMode {
		hints = []struct{ key, desc string }{
			{"esc", "back"},
			{"j/k", "scroll"},
			{"r", "refresh"},
			{"1-0", "views"},
			{"q", "quit"},
		}
	} else {
		hints = []struct{ key, desc string }{
			{"j/k", "navigate"},
			{"enter", "details"},
			{"/", "filter"},
			{"r", "refresh"},
			{"tab", "sub-tab"},
			{"1-0", "views"},
			{"q", "quit"},
		}
	}

	var parts []string
	for _, h := range hints {
		k := StatusBarKeyStyle.Render(h.key)
		d := StatusBarDescStyle.Render(h.desc)
		parts = append(parts, k+d)
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	return StatusBarStyle.Width(width).Render(bar)
}
