package output

import (
	"fmt"
	"io"
	"reflect"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"sigs.k8s.io/yaml"
)

// Routes through sigs.k8s.io/yaml (JSON first) because the resource types carry
// `json:` tags only - a direct YAML encoder emits k8sversion, not k8s_version.
// The cost is alphabetical key order, which is not part of the contract.
func RenderYAML(w io.Writer, v any) error {
	out, err := yaml.Marshal(emptyForNil(v))
	if err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	if _, err := w.Write(out); err != nil {
		return fmt.Errorf("writing YAML: %w", err)
	}
	return nil
}

// The envelope carries `json:` tags only, since RenderYAML derives YAML keys
// from them; a `yaml:` tag here would read as authoritative but be dead weight.
func RenderYAMLWithMeta(w io.Writer, data any, meta *types.Meta) error {
	if meta == nil {
		return RenderYAML(w, data)
	}
	return RenderYAML(w, struct {
		Data any         `json:"data"`
		Meta *types.Meta `json:"meta"`
	}{Data: emptyForNil(data), Meta: meta})
}

// Commands build into `var items []T`, so an empty result stays nil and
// json.Marshal turns it into null, which survives into the YAML. Substituting an
// empty slice/map here renders [] without every caller pre-seeding.
func emptyForNil(v any) any {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsNil() {
			return reflect.MakeSlice(rv.Type(), 0, 0).Interface()
		}
	case reflect.Map:
		if rv.IsNil() {
			return reflect.MakeMap(rv.Type()).Interface()
		}
	}
	return v
}
