package output

import (
	"bytes"
	"strings"
	"testing"
)

type testYAMLData struct {
	Name string   `yaml:"name"`
	Cost *float64 `yaml:"cost"`
}

func TestYAMLTo_ValidStruct(t *testing.T) {
	cost := 42.5
	data := testYAMLData{Name: "test", Cost: &cost}
	var buf bytes.Buffer
	if err := YAMLTo(&buf, data); err != nil {
		t.Fatalf("YAMLTo() error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "name:") {
		t.Errorf("expected YAML to contain 'name:', got: %s", got)
	}
	if !strings.Contains(got, "test") {
		t.Errorf("expected YAML to contain 'test', got: %s", got)
	}
	if !strings.Contains(got, "cost:") {
		t.Errorf("expected YAML to contain 'cost:', got: %s", got)
	}
}

func TestYAMLTo_NilFields(t *testing.T) {
	data := testYAMLData{Name: "test", Cost: nil}
	var buf bytes.Buffer
	if err := YAMLTo(&buf, data); err != nil {
		t.Fatalf("YAMLTo() error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "name:") {
		t.Errorf("expected YAML to contain 'name:', got: %s", got)
	}
	if got == "" {
		t.Error("expected non-empty YAML output")
	}
}
