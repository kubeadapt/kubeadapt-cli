package output

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
)

func TestRenderOverview(t *testing.T) {
	overview := testutil.SampleOverview()
	// Should not panic with noColor=true
	RenderOverview(overview, true)
	// Should not panic with noColor=false
	RenderOverview(overview, false)
}

func TestRenderClusters(t *testing.T) {
	clusters := testutil.SampleClusters()
	// Should not panic with normal data
	RenderClusters(clusters, true)
	// Should not panic with empty slice
	RenderClusters([]types.ClusterResponse{}, true)
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

func TestFormatCostPtr(t *testing.T) {
	f := 42.5
	tests := []struct {
		input *float64
		want  string
	}{
		{nil, "-"},
		{&f, "$42.50"},
	}
	for _, tt := range tests {
		got := FormatCostPtr(tt.input)
		if got != tt.want {
			t.Errorf("FormatCostPtr(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatPercentPtr(t *testing.T) {
	tests := []struct {
		input *float64
		want  string
	}{
		{nil, "-"},
		{func() *float64 { v := 42.5; return &v }(), "42.5%"},
		{func() *float64 { v := 0.0; return &v }(), "0.0%"},
	}
	for _, tt := range tests {
		got := FormatPercentPtr(tt.input)
		if got != tt.want {
			t.Errorf("FormatPercentPtr(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatOptionalString(t *testing.T) {
	tests := []struct {
		input *string
		want  string
	}{
		{nil, "-"},
		{func() *string { s := "hello"; return &s }(), "hello"},
		{func() *string { s := ""; return &s }(), ""},
	}
	for _, tt := range tests {
		got := FormatOptionalString(tt.input)
		if got != tt.want {
			t.Errorf("FormatOptionalString(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{1000000, "1000000"},
	}
	for _, tt := range tests {
		got := FormatInt(tt.input)
		if got != tt.want {
			t.Errorf("FormatInt(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input    float64
		decimals int
		want     string
	}{
		{1.0, 1, "1.0"},
		{3.14159, 2, "3.14"},
		{0.0, 0, "0"},
	}
	for _, tt := range tests {
		got := FormatFloat(tt.input, tt.decimals)
		if got != tt.want {
			t.Errorf("FormatFloat(%v, %d) = %q, want %q", tt.input, tt.decimals, got, tt.want)
		}
	}
}

func TestFormatFloatPtr(t *testing.T) {
	tests := []struct {
		input    *float64
		decimals int
		want     string
	}{
		{nil, 2, "-"},
		{func() *float64 { v := 3.14; return &v }(), 2, "3.14"},
	}
	for _, tt := range tests {
		got := FormatFloatPtr(tt.input, tt.decimals)
		if got != tt.want {
			t.Errorf("FormatFloatPtr(%v, %d) = %q, want %q", tt.input, tt.decimals, got, tt.want)
		}
	}
}

func TestFormatMemoryGBPtr(t *testing.T) {
	tests := []struct {
		input *float64
		want  string
	}{
		{nil, "-"},
		{func() *float64 { v := 1.0; return &v }(), "1.0 GB"},
		{func() *float64 { v := 0.5; return &v }(), "512 MB"},
	}
	for _, tt := range tests {
		got := FormatMemoryGBPtr(tt.input)
		if got != tt.want {
			t.Errorf("FormatMemoryGBPtr(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatIntPtr(t *testing.T) {
	tests := []struct {
		input *int
		want  string
	}{
		{nil, "-"},
		{func() *int { v := 5; return &v }(), "5"},
		{func() *int { v := 0; return &v }(), "0"},
	}
	for _, tt := range tests {
		got := FormatIntPtr(tt.input)
		if got != tt.want {
			t.Errorf("FormatIntPtr(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestShortID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc", "abc"},
		{"12345678", "12345678"},
		{"123456789", "12345678"},
		{"abcdef012345", "abcdef01"},
	}
	for _, tt := range tests {
		got := ShortID(tt.input)
		if got != tt.want {
			t.Errorf("ShortID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRenderNodes(t *testing.T) {
	nodes := testutil.SampleNodes()
	// Should not panic with normal data
	RenderNodes(nodes, true)
	// Should not panic with noColor=false
	RenderNodes(nodes, false)
}

func TestRenderWorkloads(t *testing.T) {
	workloads := testutil.SampleWorkloads()
	RenderWorkloads(workloads, true)
}

func TestRenderRecommendations(t *testing.T) {
	recs := testutil.SampleRecommendations()
	RenderRecommendations(recs, true)
}

func TestRenderNamespaces(t *testing.T) {
	namespaces := testutil.SampleNamespaces()
	RenderNamespaces(namespaces, true)
}

func TestRenderDashboard(t *testing.T) {
	dashboard := testutil.SampleDashboard()
	RenderDashboard(dashboard, true)
}

func TestRender_EmptyData(t *testing.T) {
	// All render functions should handle empty slices without panic
	RenderClusters([]types.ClusterResponse{}, true)
	RenderNodes([]types.NodeResponse{}, true)
	RenderWorkloads([]types.WorkloadResponse{}, true)
	RenderRecommendations([]types.RecommendationResponse{}, true)
	RenderNamespaces([]types.NamespaceResponse{}, true)
	RenderTeamCosts([]types.TeamCostResponse{}, true)
	RenderDepartmentCosts([]types.DepartmentCostResponse{}, true)
	RenderNodeGroups([]types.NodeGroupResponse{}, true)
	RenderPersistentVolumes([]types.PersistentVolumeResponse{}, true)
}
