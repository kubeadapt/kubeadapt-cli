package output

import (
	"bytes"
	"strings"
	"testing"
)

type testJSONData struct {
	Name string   `json:"name"`
	Cost *float64 `json:"cost"`
}

func TestJSONTo_ValidStruct(t *testing.T) {
	cost := 42.5
	data := testJSONData{Name: "test", Cost: &cost}
	var buf bytes.Buffer
	if err := JSONTo(&buf, data); err != nil {
		t.Fatalf("JSONTo() error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, `"name"`) {
		t.Errorf("expected JSON to contain 'name' key, got: %s", got)
	}
	if !strings.Contains(got, `"test"`) {
		t.Errorf("expected JSON to contain 'test' value, got: %s", got)
	}
	if !strings.Contains(got, `"cost"`) {
		t.Errorf("expected JSON to contain 'cost' key, got: %s", got)
	}
	if !strings.Contains(got, `42.5`) {
		t.Errorf("expected JSON to contain '42.5', got: %s", got)
	}
}

func TestJSONTo_NilFields(t *testing.T) {
	data := testJSONData{Name: "test", Cost: nil}
	var buf bytes.Buffer
	if err := JSONTo(&buf, data); err != nil {
		t.Fatalf("JSONTo() error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "null") {
		t.Errorf("expected JSON to contain 'null' for nil field, got: %s", got)
	}
}

func TestJSONTo_EmptySlice(t *testing.T) {
	data := []string{}
	var buf bytes.Buffer
	if err := JSONTo(&buf, data); err != nil {
		t.Fatalf("JSONTo() error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "[]") {
		t.Errorf("expected JSON to contain '[]' for empty slice, got: %s", got)
	}
}
