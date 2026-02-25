package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// FilterBar provides an inline text input for filtering table data.
type FilterBar struct {
	active    bool
	query     string
	cursorPos int
	Width     int
}

// NewFilterBar creates a new filter bar.
func NewFilterBar() *FilterBar {
	return &FilterBar{}
}

// Activate enables the filter bar and resets its state.
func (f *FilterBar) Activate() {
	f.active = true
	f.query = ""
	f.cursorPos = 0
}

// Deactivate disables the filter bar and clears input.
func (f *FilterBar) Deactivate() {
	f.active = false
	f.query = ""
	f.cursorPos = 0
}

// IsActive returns whether the filter bar is currently active.
func (f *FilterBar) IsActive() bool {
	return f.active
}

// GetQuery returns the current filter query string.
func (f *FilterBar) GetQuery() string {
	return f.query
}

// HandleKey processes a key message and returns true if the key was consumed.
func (f *FilterBar) HandleKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyEsc:
		f.Deactivate()
		return true

	case tea.KeyEnter:
		return true

	case tea.KeyBackspace:
		if f.cursorPos > 0 {
			before := f.query[:f.cursorPos-1]
			after := f.query[f.cursorPos:]
			f.query = before + after
			f.cursorPos--
		}
		return true

	case tea.KeyDelete:
		if f.cursorPos < len(f.query) {
			before := f.query[:f.cursorPos]
			after := f.query[f.cursorPos+1:]
			f.query = before + after
		}
		return true

	case tea.KeyLeft:
		if f.cursorPos > 0 {
			f.cursorPos--
		}
		return true

	case tea.KeyRight:
		if f.cursorPos < len(f.query) {
			f.cursorPos++
		}
		return true

	case tea.KeyHome, tea.KeyCtrlA:
		f.cursorPos = 0
		return true

	case tea.KeyEnd, tea.KeyCtrlE:
		f.cursorPos = len(f.query)
		return true

	case tea.KeyRunes:
		before := f.query[:f.cursorPos]
		after := f.query[f.cursorPos:]
		f.query = before + string(msg.Runes) + after
		f.cursorPos += len(msg.Runes)
		return true
	}

	return false
}

// Render returns the filter bar display string. Returns empty when inactive.
func (f *FilterBar) Render() string {
	if !f.active {
		return ""
	}

	label := tui.FilterLabelStyle.Render("/ Filter: ")
	input := f.query[:f.cursorPos] + "_" + f.query[f.cursorPos:]

	line := label + input

	style := tui.FilterBarStyle
	if f.Width > 0 {
		style = style.Width(f.Width)
	}

	return style.Render(line)
}
