package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the global keybindings.
type KeyMap struct {
	Quit    key.Binding
	Help    key.Binding
	Refresh key.Binding
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Filter  key.Binding
	Tab     key.Binding
	Nav1    key.Binding
	Nav2    key.Binding
	Nav3    key.Binding
	Nav4    key.Binding
	Nav5    key.Binding
	Nav6    key.Binding
	Nav7    key.Binding
	Nav8    key.Binding
	Nav9    key.Binding
	Nav0    key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/↓", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		Nav1: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "overview")),
		Nav2: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "clusters")),
		Nav3: key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "workloads")),
		Nav4: key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "nodes")),
		Nav5: key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "recommendations")),
		Nav6: key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "costs")),
		Nav7: key.NewBinding(key.WithKeys("7"), key.WithHelp("7", "namespaces")),
		Nav8: key.NewBinding(key.WithKeys("8"), key.WithHelp("8", "node groups")),
		Nav9: key.NewBinding(key.WithKeys("9"), key.WithHelp("9", "pvs")),
		Nav0: key.NewBinding(key.WithKeys("0"), key.WithHelp("0", "help")),
	}
}
