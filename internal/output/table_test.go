package output

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
)

func TestRenderOverview(t *testing.T) {
	overview := testutil.SampleOverview()
	// Should not panic
	RenderOverview(overview, true)
}

func TestRenderClusters(t *testing.T) {
	clusters := testutil.SampleClusters()
	// Should not panic
	RenderClusters(clusters, true)
}

func TestFormatCost(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "$0.00"},
		{1.5, "$1.50"},
		{1234.56, "$1234.56"},
	}
	for _, tt := range tests {
		got := FormatCost(tt.input)
		if got != tt.want {
			t.Errorf("FormatCost(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "0.0%"},
		{42.5, "42.5%"},
		{100, "100.0%"},
	}
	for _, tt := range tests {
		got := FormatPercent(tt.input)
		if got != tt.want {
			t.Errorf("FormatPercent(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatMemoryGB(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0.5, "512 MB"},
		{1.0, "1.0 GB"},
		{16.5, "16.5 GB"},
	}
	for _, tt := range tests {
		got := FormatMemoryGB(tt.input)
		if got != tt.want {
			t.Errorf("FormatMemoryGB(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatBool(t *testing.T) {
	if FormatBool(true) != "Yes" {
		t.Error("FormatBool(true) should be 'Yes'")
	}
	if FormatBool(false) != "No" {
		t.Error("FormatBool(false) should be 'No'")
	}
}
