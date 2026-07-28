package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

func RenderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(emptyForNil(v)); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}

// Emits an envelope so callers can read next_cursor programmatically. Falls
// back to RenderJSON when meta is nil.
func RenderJSONWithMeta(w io.Writer, data any, meta *types.Meta) error {
	if meta == nil {
		return RenderJSON(w, data)
	}
	return RenderJSON(w, struct {
		Data any         `json:"data"`
		Meta *types.Meta `json:"meta"`
	}{Data: emptyForNil(data), Meta: meta})
}
