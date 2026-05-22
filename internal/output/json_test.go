package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testJSONData struct {
	Name string   `json:"name"`
	Cost *float64 `json:"cost"`
}

func TestRenderJSON_ValidStruct(t *testing.T) {
	cost := 42.5
	data := testJSONData{Name: "test", Cost: &cost}
	var buf bytes.Buffer
	require.NoError(t, RenderJSON(&buf, data))
	got := buf.String()
	for _, want := range []string{`"name"`, `"test"`, `"cost"`, `42.5`} {
		assert.Contains(t, got, want)
	}
}

func TestRenderJSON_NilFields(t *testing.T) {
	data := testJSONData{Name: "test", Cost: nil}
	var buf bytes.Buffer
	require.NoError(t, RenderJSON(&buf, data))
	assert.Contains(t, buf.String(), "null")
}

func TestRenderJSON_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, RenderJSON(&buf, []string{}))
	assert.Contains(t, buf.String(), "[]")
}
