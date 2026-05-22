package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testYAMLData struct {
	Name string   `yaml:"name"`
	Cost *float64 `yaml:"cost"`
}

func TestRenderYAML_ValidStruct(t *testing.T) {
	cost := 42.5
	data := testYAMLData{Name: "test", Cost: &cost}
	var buf bytes.Buffer
	require.NoError(t, RenderYAML(&buf, data))
	got := buf.String()
	for _, want := range []string{"name:", "test", "cost:"} {
		assert.Contains(t, got, want)
	}
}

func TestRenderYAML_NilFields(t *testing.T) {
	data := testYAMLData{Name: "test", Cost: nil}
	var buf bytes.Buffer
	require.NoError(t, RenderYAML(&buf, data))
	assert.Contains(t, buf.String(), "name:")
}
