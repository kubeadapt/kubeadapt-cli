package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFilterBarActivateDeactivate(t *testing.T) {
	fb := NewFilterBar()

	if fb.IsActive() {
		t.Error("Expected new FilterBar to be inactive")
	}

	fb.Activate()

	if !fb.IsActive() {
		t.Error("Expected FilterBar to be active after Activate()")
	}

	fb.Deactivate()

	if fb.IsActive() {
		t.Error("Expected FilterBar to be inactive after Deactivate()")
	}

	if fb.GetQuery() != "" {
		t.Errorf("Expected empty query after Deactivate(), got %q", fb.GetQuery())
	}
}

func TestFilterBarHandleKeyRunes(t *testing.T) {
	fb := NewFilterBar()
	fb.Activate()

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	query := fb.GetQuery()
	if query != "abc" {
		t.Errorf("Expected query 'abc', got %q", query)
	}
}

func TestFilterBarHandleKeyBackspace(t *testing.T) {
	fb := NewFilterBar()
	fb.Activate()

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})

	query := fb.GetQuery()
	if query != "ab" {
		t.Errorf("Expected query 'ab' after backspace, got %q", query)
	}
}

func TestFilterBarHandleKeyEsc(t *testing.T) {
	fb := NewFilterBar()
	fb.Activate()

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})

	if fb.IsActive() {
		t.Error("Expected FilterBar to be inactive after Esc")
	}

	if fb.GetQuery() != "" {
		t.Errorf("Expected empty query after Esc, got %q", fb.GetQuery())
	}
}

func TestFilterBarRenderInactive(t *testing.T) {
	fb := NewFilterBar()

	output := fb.Render()

	if output != "" {
		t.Errorf("Expected empty output when inactive, got %q", output)
	}
}

func TestFilterBarRenderActive(t *testing.T) {
	fb := NewFilterBar()
	fb.Activate()

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})

	output := fb.Render()

	if !strings.Contains(output, "Filter") {
		t.Error("Expected output to contain 'Filter' when active")
	}

	if !strings.Contains(output, "test") {
		t.Error("Expected output to contain 'test' query text")
	}
}

func TestFilterBarHandleKeyDelete(t *testing.T) {
	fb := NewFilterBar()
	fb.Activate()

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyDelete})

	query := fb.GetQuery()
	if query != "ac" {
		t.Errorf("Expected query 'ac' after delete at position 1, got %q", query)
	}
}

func TestFilterBarHandleKeyNavigation(t *testing.T) {
	fb := NewFilterBar()
	fb.Activate()

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyHome})
	if fb.cursorPos != 0 {
		t.Errorf("Expected cursor at position 0 after Home, got %d", fb.cursorPos)
	}

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyEnd})
	if fb.cursorPos != 5 {
		t.Errorf("Expected cursor at position 5 after End, got %d", fb.cursorPos)
	}

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if fb.cursorPos != 4 {
		t.Errorf("Expected cursor at position 4 after Left, got %d", fb.cursorPos)
	}

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRight})
	if fb.cursorPos != 5 {
		t.Errorf("Expected cursor at position 5 after Right, got %d", fb.cursorPos)
	}
}

func TestFilterBarHandleKeyEnter(t *testing.T) {
	fb := NewFilterBar()
	fb.Activate()

	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	fb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})

	consumed := fb.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})

	if !consumed {
		t.Error("Expected HandleKey to return true for Enter key")
	}

	if !fb.IsActive() {
		t.Error("Expected FilterBar to remain active after Enter")
	}

	if fb.GetQuery() != "test" {
		t.Errorf("Expected query to remain 'test' after Enter, got %q", fb.GetQuery())
	}
}
