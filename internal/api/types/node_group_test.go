package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const nodeGroupFixture = `{
  "id": "44444444-4444-4444-4444-444444444444",
  "kind": "node_group",
  "metadata": {
    "name": "general",
    "cluster": {"id": "c-1", "name": "prod-eks-1"},
    "instance_types": ["m5.4xlarge", "m5.2xlarge"],
    "zones": ["us-east-1a", "us-east-1b"],
    "spot_count": 4,
    "ondemand_count": 2,
    "spot_percentage": 66.67,
    "oldest_node_created_at_k8s": "2026-04-01T08:11:00Z",
    "status": "active"
  },
  "capacity": {
    "cpu":    {"total_cores": 96.0, "allocatable_cores": 93.0},
    "memory": {"total_bytes": 412316860416, "allocatable_bytes": 405874409472}
  },
  "utilization": {
    "cpu":    {"used_cores": 48.0, "utilization_percent": 50.0},
    "memory": {"used_bytes": 200000000000, "utilization_percent": 48.5},
    "counts": {"nodes": 6, "ready_nodes": 6, "pods": 240}
  },
  "cost": {
    "current_run_rate_hourly": {"amount": "1.5400", "currency": "USD"},
    "spot_savings_vs_ondemand_hourly": {"amount": "3.0600", "currency": "USD"},
    "last_updated_at": "2026-05-20T11:55:00Z"
  }
}`

func TestUnmarshalNodeGroup(t *testing.T) {
	t.Parallel()

	var got NodeGroup
	require.NoError(t, json.Unmarshal([]byte(nodeGroupFixture), &got))

	assert.Equal(t, "general", got.Metadata.Name)
	want := []string{"m5.4xlarge", "m5.2xlarge"}
	assert.Len(t, got.Metadata.InstanceTypes, len(want))
	for i, it := range want {
		assert.Equal(t, it, got.Metadata.InstanceTypes[i])
	}
	assert.Equal(t, 4, got.Metadata.SpotCount)
	assert.Equal(t, 2, got.Metadata.OnDemandCount)
	assert.Equal(t, 6, got.Utilization.Counts.Nodes)
	require.NotNil(t, got.Cost.SpotSavingsVsOndemandHourly, "SpotSavingsVsOndemandHourly missing")
	assert.Equal(t, "3.0600", got.Cost.SpotSavingsVsOndemandHourly.Amount)

	roundtrip, err := json.Marshal(got)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(roundtrip, &raw))
	costRaw := raw["cost"].(map[string]any)
	_, exists := costRaw["cost_mode"]
	assert.False(t, exists, "node_group cost block must NOT contain cost_mode, found: %v", costRaw)
}
