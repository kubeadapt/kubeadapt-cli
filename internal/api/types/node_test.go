package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const nodeFixture = `{
  "id": "33333333-3333-3333-3333-333333333333",
  "kind": "node",
  "metadata": {
    "name": "ip-10-0-1-42.ec2.internal",
    "cluster": {"id": "c-1", "name": "prod-eks-1"},
    "node_role": "worker",
    "instance_type": "m5.4xlarge",
    "node_group": "general",
    "availability_zone": "us-east-1a",
    "region": "us-east-1",
    "is_spot": true,
    "capacity_type": "SPOT",
    "architecture": "amd64",
    "operating_system": "linux",
    "kubelet_version": "v1.30.4",
    "provider_id": "aws:///us-east-1a/i-0abcdef",
    "is_ready": true,
    "is_schedulable": true,
    "labels": {"node.kubernetes.io/instance-type": "m5.4xlarge"},
    "created_at_k8s": "2026-04-01T08:11:00Z",
    "last_seen_at": "2026-05-20T11:59:30Z"
  },
  "capacity": {
    "cpu":               {"total_cores": 16.0, "allocatable_cores": 15.5},
    "memory":            {"total_bytes": 68719476736, "allocatable_bytes": 67645734912},
    "gpu":               {"total": 0, "allocatable": 0},
    "ephemeral_storage": {"total_bytes": 107374182400},
    "pods":              {"allocatable": 110}
  },
  "utilization": {
    "cpu":    {"used_cores": 8.4, "utilization_percent": 52.5},
    "memory": {"used_bytes": 36507222016, "utilization_percent": 53.1},
    "gpu":    {"utilization_percent": 0, "memory_used_bytes": 0, "memory_total_bytes": 0},
    "counts": {"pods": 42, "running_pods": 41}
  },
  "cost": {
    "current_run_rate_hourly": {"amount": "0.2576", "currency": "USD"},
    "on_demand_equivalent_hourly": {"amount": "0.7680", "currency": "USD"},
    "pricing_source": "aws-spot-feed",
    "last_updated_at": "2026-05-20T11:55:00Z"
  }
}`

func TestUnmarshalNode(t *testing.T) {
	t.Parallel()

	var got Node
	require.NoError(t, json.Unmarshal([]byte(nodeFixture), &got))

	assert.Equal(t, "node", got.Kind)
	assert.Equal(t, "m5.4xlarge", got.Metadata.InstanceType)
	assert.True(t, got.Metadata.IsSpot, "IsSpot should be true")
	assert.Equal(t, int64(107374182400), got.Capacity.EphemeralStorage.TotalBytes)
	require.NotNil(t, got.Cost.OnDemandEquivalentHourly, "OnDemandEquivalentHourly should be populated for spot node")
	assert.Equal(t, "0.7680", got.Cost.OnDemandEquivalentHourly.Amount)

	roundtrip, err := json.Marshal(got)
	require.NoError(t, err)
	var second Node
	require.NoError(t, json.Unmarshal(roundtrip, &second))

	var raw map[string]any
	require.NoError(t, json.Unmarshal(roundtrip, &raw))
	costRaw, ok := raw["cost"].(map[string]any)
	require.True(t, ok, "cost block missing in round-trip")
	_, exists := costRaw["cost_mode"]
	assert.False(t, exists, "node cost block must NOT contain cost_mode, found in round-trip: %v", costRaw)
}
