package types

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The four members the API adds beside every cost block's total. partial is
// false and components non-empty here on purpose: both are the values a naive
// omitempty would drop on the way back out.
const vectorJSON = `"components":{"compute.cpu":{"amount":"0.7500","currency":"USD"},"storage.pv":{"amount":"0.0000","currency":"USD"}},` +
	`"unallocated":{"amount":"0.2500","currency":"USD"},"basis":"legacy","partial":false`

var wantVector = map[string]any{
	"components": map[string]any{
		"compute.cpu": map[string]any{"amount": "0.7500", "currency": "USD"},
		"storage.pv":  map[string]any{"amount": "0.0000", "currency": "USD"},
	},
	"unallocated": map[string]any{"amount": "0.2500", "currency": "USD"},
	"basis":       "legacy",
	"partial":     false,
}

var vectorKeys = []string{"components", "unallocated", "basis", "partial"}

// costBlocks lists every cost block the API sends, as the path to it inside
// the resource that carries it. `-o json` decodes into these types and
// re-encodes, so a block missing here silently loses its breakdown.
func costBlocks() []struct {
	name  string
	json  string
	into  func() any
	paths []string
} {
	block := `{"current_run_rate_hourly":{"amount":"1.0000","currency":"USD"},` + vectorJSON + `}`
	return []struct {
		name  string
		json  string
		into  func() any
		paths []string
	}{
		{"cluster", `{"cost":` + block + `}`, func() any { return &Cluster{} }, []string{"cost"}},
		{"organization", `{"cost":` + block + `}`, func() any { return &Organization{} }, []string{"cost"}},
		{"workload", `{"cost":` + block + `}`, func() any { return &Workload{} }, []string{"cost"}},
		{"pod", `{"cost":` + block + `}`, func() any { return &Pod{} }, []string{"cost"}},
		{"node", `{"cost":` + block + `}`, func() any { return &Node{} }, []string{"cost"}},
		{"node group", `{"cost":` + block + `,"nodes":[{"cost":` + block + `}]}`,
			func() any { return &NodeGroup{} }, []string{"cost", "nodes.0.cost"}},
		{"namespace", `{"cost":` + block + `,"workloads_top_5":[{"cost":` + block + `}]}`,
			func() any { return &Namespace{} }, []string{"cost", "workloads_top_5.0.cost"}},
		{"team", `{"cost":` + block + `}`, func() any { return &Team{} }, []string{"cost"}},
		{"department", `{"cost":` + block + `}`, func() any { return &Department{} }, []string{"cost"}},
		{"recommendation",
			`{"current":{"hourly_cost":{"amount":"1.0000","currency":"USD"},` + vectorJSON + `},` +
				`"savings":{"estimated_hourly":{"amount":"0.5000","currency":"USD"},` + vectorJSON + `}}`,
			func() any { return &Recommendation{} }, []string{"current", "savings"}},
		{"organization dashboard",
			`{"month_to_date":{"billed_cost":{"amount":"9.0000","currency":"USD"},` + vectorJSON + `},` +
				`"savings":{"current_hourly_potential":{"amount":"0.5000","currency":"USD"},` + vectorJSON + `},` +
				`"top_clusters":[{"month_to_date_cost":{"amount":"9.0000","currency":"USD"},` + vectorJSON + `}]}`,
			func() any { return &OrganizationDashboard{} }, []string{"month_to_date", "savings", "top_clusters.0"}},
	}
}

func roundTrip(t *testing.T, raw string, into any) map[string]any {
	t.Helper()
	require.NoError(t, json.Unmarshal([]byte(raw), into))
	out, err := json.Marshal(into)
	require.NoError(t, err)
	var generic map[string]any
	require.NoError(t, json.Unmarshal(out, &generic))
	return generic
}

func at(t *testing.T, doc map[string]any, path string) map[string]any {
	t.Helper()
	var cur any = doc
	for _, key := range strings.Split(path, ".") {
		switch node := cur.(type) {
		case map[string]any:
			cur = node[key]
		case []any:
			i, err := strconv.Atoi(key)
			require.NoError(t, err)
			require.Less(t, i, len(node), "path %s", path)
			cur = node[i]
		}
	}
	m, ok := cur.(map[string]any)
	require.True(t, ok, "path %s is not an object: %#v", path, cur)
	return m
}

func TestCostBlocks_CarryTheComponentVector(t *testing.T) {
	for _, tc := range costBlocks() {
		t.Run(tc.name, func(t *testing.T) {
			doc := roundTrip(t, tc.json, tc.into())
			for _, path := range tc.paths {
				block := at(t, doc, path)
				for _, key := range vectorKeys {
					assert.Equal(t, wantVector[key], block[key], "%s.%s", path, key)
				}
			}
		})
	}
}
