package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// RenderJSON writes the value as indented JSON to w.
func RenderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}

// RenderJSONWithMeta writes an envelope-shaped JSON document containing both
// the data payload and the pagination metadata, so callers can extract the
// next_cursor programmatically (jq '.meta.pagination.next_cursor'). When meta
// is nil it falls back to plain RenderJSON over data.
func RenderJSONWithMeta(w io.Writer, data any, meta *types.Meta) error {
	if meta == nil {
		return RenderJSON(w, data)
	}
	return RenderJSON(w, struct {
		Data any         `json:"data"`
		Meta *types.Meta `json:"meta"`
	}{Data: data, Meta: meta})
}
