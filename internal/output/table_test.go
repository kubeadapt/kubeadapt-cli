package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

func init() {
	SetNoColor(true)
}

func usdMoney(amount string) types.Money {
	return types.Money{Amount: amount, Currency: "USD"}
}

func sampleCluster(name string) types.Cluster {
	return types.Cluster{
		ID:   "cluster-" + name,
		Kind: "cluster",
		Metadata: types.ClusterMetadata{
			Name:        name,
			Provider:    "aws",
			Region:      "us-east-1",
			Environment: "production",
			Status:      "connected",
			K8sVersion:  "1.29",
		},
		Capacity: types.ClusterCapacity{
			CPU:    types.CapacityCPU{TotalCores: 96, AllocatableCores: 90},
			Memory: types.CapacityMemory{TotalBytes: 412316860416, AllocatableBytes: 400000000000},
		},
		Utilization: types.ClusterUtilization{
			CPU:    types.UtilizationCPU{UsedCores: 48, UtilizationPercent: 50.0},
			Memory: types.UtilizationMemory{UsedBytes: 200000000000, UtilizationPercent: 50.0},
			Counts: types.ClusterCounts{Nodes: 6, Pods: 120, Workloads: 30},
		},
		Cost: types.ClusterCost{
			CurrentRunRateHourly: usdMoney("12.5000"),
			LastUpdatedAt:        "2025-05-20T12:00:00Z",
		},
	}
}

func TestRenderClusters(t *testing.T) {
	var buf bytes.Buffer
	items := []types.Cluster{sampleCluster("prod-east"), sampleCluster("prod-west")}
	require.NoError(t, RenderClusters(&buf, items, nil))
	out := buf.String()
	for _, want := range []string{"prod-east", "prod-west", "$/hr", "$12.5000", "aws"} {
		assert.Contains(t, out, want)
	}
}

func TestRenderClusters_Empty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, RenderClusters(&buf, nil, nil))
	assert.Contains(t, buf.String(), "No clusters")
}

func TestRenderClusters_WithPagination(t *testing.T) {
	var buf bytes.Buffer
	items := []types.Cluster{sampleCluster("prod-east")}
	longCursor := "eyJ2IjoxLCJjIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5In0="
	meta := &types.Meta{
		Pagination: &types.Pagination{
			NextCursor: longCursor,
			HasMore:    true,
			Limit:      50,
		},
	}
	require.NoError(t, RenderClusters(&buf, items, meta))
	out := buf.String()
	assert.Contains(t, out, "--cursor="+longCursor)
	assert.NotContains(t, out, "...")
	assert.Contains(t, out, "Showing 1")
	assert.Contains(t, out, "(limit 50)")
	assert.Contains(t, out, "--paginate")
}

func TestRenderClusters_EndOfResults(t *testing.T) {
	var buf bytes.Buffer
	items := []types.Cluster{sampleCluster("only")}
	meta := &types.Meta{
		Pagination: &types.Pagination{HasMore: false, Limit: 100},
	}
	require.NoError(t, RenderClusters(&buf, items, meta))
	assert.Contains(t, buf.String(), "End of results.")
}

func TestRenderOrganization(t *testing.T) {
	var buf bytes.Buffer
	org := types.Organization{
		ID:   "org-1",
		Kind: "organization",
		Metadata: types.OrganizationMetadata{
			Name:     "Acme",
			Domain:   "acme.io",
			PlanType: "enterprise",
			IsActive: true,
		},
		Capacity: types.OrganizationCapacity{
			CPU:    types.CapacityCPU{TotalCores: 192, AllocatableCores: 180},
			Memory: types.CapacityMemory{TotalBytes: 824633720832},
		},
		Utilization: types.OrganizationUtilization{
			CPU:    types.UtilizationCPU{UsedCores: 96, UtilizationPercent: 50.0},
			Counts: types.OrganizationCounts{Clusters: 3, ConnectedClusters: 2},
		},
		Cost: types.OrganizationCost{CurrentRunRateHourly: usdMoney("42.1234")},
	}
	require.NoError(t, RenderOrganization(&buf, org))
	out := buf.String()
	for _, want := range []string{"Acme", "acme.io", "enterprise", "$42.1234", "Clusters"} {
		assert.Contains(t, out, want)
	}
}

func TestRenderRecommendation(t *testing.T) {
	var buf bytes.Buffer
	rec := types.Recommendation{
		ID:   "rec-1",
		Kind: "recommendation",
		Metadata: types.RecommendationMetadata{
			RecommendationType: "rightsize_workload",
			ResourceType:       "workload",
			ResourceName:       "checkout-api",
			Cluster:            types.NestedRef{ID: "c1", Name: "prod-east"},
			Namespace:          "shop",
			Priority:           "high",
			RiskLevel:          "low",
			Status:             "open",
		},
		Current: types.RecommendationSnapshot{HourlyCost: usdMoney("3.4500")},
		Savings: types.RecommendationSavings{
			EstimatedHourly: usdMoney("1.0000"),
		},
	}
	require.NoError(t, RenderRecommendation(&buf, rec))
	out := buf.String()
	for _, want := range []string{"rec-1", "rightsize_workload", "checkout-api", "$1.0000", "$3.4500", "high"} {
		assert.Contains(t, out, want)
	}
}

func TestFormatMoney_Wrappers(t *testing.T) {
	tests := []struct {
		name string
		in   types.Money
		want string
	}{
		{"zero value", types.Money{}, "-"},
		{"usd", usdMoney("12.5000"), "$12.5000"},
		{"eur", types.Money{Amount: "9.9999", Currency: "EUR"}, "EUR 9.9999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatMoney(tt.in))
		})
	}
}

func TestFormatMoneyPtr(t *testing.T) {
	assert.Equal(t, "-", FormatMoneyPtr(nil))
	m := usdMoney("1.0000")
	assert.Equal(t, "$1.0000", FormatMoneyPtr(&m))
}

func TestFormatCursor(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", "(none)"},
		{"short", "abc123", "abc123"},
		{"long verbatim", "abcdefghijklmnopqrstuvwxyz0123456789", "abcdefghijklmnopqrstuvwxyz0123456789"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatCursor(tt.in))
		})
	}
}

func TestPaginationFooter(t *testing.T) {
	total := 1234
	tests := []struct {
		name    string
		shown   int
		meta    *types.Meta
		want    []string
		notWant []string
		empty   bool
	}{
		{name: "nil meta", shown: 3, meta: nil, empty: true},
		{name: "nil pagination", shown: 3, meta: &types.Meta{}, empty: true},
		{
			name:  "end of results",
			shown: 3,
			meta:  &types.Meta{Pagination: &types.Pagination{Limit: 100, HasMore: false}},
			want:  []string{"Showing 3", "(limit 100)", "End of results."},
		},
		{
			name:  "has more with cursor",
			shown: 50,
			meta: &types.Meta{Pagination: &types.Pagination{
				Limit: 50, HasMore: true,
				NextCursor: "eyJ2IjoxLCJjIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5In0=",
			}},
			want: []string{
				"Showing 50",
				"(limit 50)",
				"--paginate",
				"--cursor=eyJ2IjoxLCJjIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXowMTIzNDU2Nzg5In0=",
			},
			notWant: []string{"..."},
		},
		{
			name:  "with total count",
			shown: 50,
			meta: &types.Meta{Pagination: &types.Pagination{
				Limit: 50, HasMore: true, TotalCount: &total,
				NextCursor: "shortcursor",
			}},
			want: []string{"Showing 50 of 1234", "shortcursor"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PaginationFooter(tt.shown, tt.meta)
			if tt.empty {
				assert.Empty(t, got)
				return
			}
			for _, w := range tt.want {
				assert.Contains(t, got, w)
			}
			for _, nw := range tt.notWant {
				assert.NotContains(t, got, nw)
			}
		})
	}
}

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0.0%"},
		{73.5, "73.5%"},
		{100, "100.0%"},
		{-1, "-"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, FormatPercentage(tt.in))
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{int64(1024) * 1024 * 1024, "1.0 GiB"},
		{-5, "-"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, FormatBytes(tt.in))
	}
}

func TestFormatCores(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "0.00"},
		{12.5, "12.50"},
		{0.5, "0.50"},
		{-1, "-"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, FormatCores(tt.in))
	}
}

func TestRenderEmpty_Smoke(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, RenderWorkloads(&buf, nil, nil))
	require.NoError(t, RenderPods(&buf, nil, nil))
	require.NoError(t, RenderNodes(&buf, nil, nil))
	require.NoError(t, RenderNodeGroups(&buf, nil, nil))
	require.NoError(t, RenderNamespaces(&buf, nil, nil))
	require.NoError(t, RenderRecommendations(&buf, nil, nil))
	require.NoError(t, RenderTeams(&buf, nil, nil))
	require.NoError(t, RenderTeamAssignments(&buf, nil, nil))
	require.NoError(t, RenderDepartments(&buf, nil, nil))
}
