package output

import (
	"fmt"
	"io"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"gopkg.in/yaml.v3"
)

// RenderYAML writes the value as YAML to w.
func RenderYAML(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	return nil
}

// RenderYAMLWithMeta is the YAML counterpart of RenderJSONWithMeta. See that
// function's doc comment for the envelope shape and nil-meta fallback.
func RenderYAMLWithMeta(w io.Writer, data any, meta *types.Meta) error {
	if meta == nil {
		return RenderYAML(w, data)
	}
	return RenderYAML(w, struct {
		Data any         `json:"data" yaml:"data"`
		Meta *types.Meta `json:"meta" yaml:"meta"`
	}{Data: data, Meta: meta})
}
