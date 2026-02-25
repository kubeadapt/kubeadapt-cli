package output

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// YAML writes the value as YAML to stdout.
func YAML(v interface{}) error {
	return YAMLTo(os.Stdout, v)
}

// YAMLTo writes the value as YAML to the given writer.
func YAMLTo(w io.Writer, v interface{}) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	return nil
}
