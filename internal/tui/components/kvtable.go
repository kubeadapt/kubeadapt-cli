package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// KVTable renders a key-value property display.
type KVTable struct {
	Items []KVItem
}

// KVItem is a single key-value pair.
type KVItem struct {
	Key   string
	Value string
}

// NewKVTable creates a new key-value table.
func NewKVTable() *KVTable {
	return &KVTable{}
}

// Add adds a key-value pair.
func (t *KVTable) Add(key, value string) {
	t.Items = append(t.Items, KVItem{Key: key, Value: value})
}

// Render draws the key-value table.
func (t *KVTable) Render() string {
	if len(t.Items) == 0 {
		return ""
	}

	keyStyle := lipgloss.NewStyle().Foreground(tui.ColorGray)
	valStyle := lipgloss.NewStyle().Foreground(tui.ColorWhite).Bold(true)

	// Find max key width
	maxKeyLen := 0
	for _, item := range t.Items {
		if len(item.Key) > maxKeyLen {
			maxKeyLen = len(item.Key)
		}
	}

	var sb strings.Builder
	for _, item := range t.Items {
		paddedKey := item.Key + strings.Repeat(" ", maxKeyLen-len(item.Key))
		sb.WriteString(keyStyle.Render(paddedKey) + "  " + valStyle.Render(item.Value) + "\n")
	}

	return sb.String()
}
