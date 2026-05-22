package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const workloadFixture = `{
  "id": "22222222-2222-2222-2222-222222222222",
  "kind": "workload",
  "metadata": {
    "name": "checkout-api",
    "workload_kind": "Deployment",
    "namespace": "payments",
    "cluster": {"id": "c-1", "name": "prod-eks-1"},
    "labels": {"app": "checkout-api", "tier": "frontend"},
    "service_account_name": "checkout-sa",
    "status": "Available",
    "is_suspended": false,
    "is_paused": false,
    "has_hpa": true,
    "created_at_k8s": "2026-04-01T08:11:00Z",
    "last_seen_at": "2026-05-20T11:59:30Z"
  },
  "capacity": {
    "cpu":    {"limit_cores": 4.0},
    "memory": {"limit_bytes": 8589934592}
  },
  "utilization": {
    "cpu":      {"requested_cores": 2.0, "used_cores": 1.2, "utilization_percent": 30.0},
    "memory":   {"requested_bytes": 4294967296, "used_bytes": 2147483648, "utilization_percent": 25.0},
    "replicas": {"desired": 3, "available": 3, "unavailable": 0, "updated": 3},
    "counts":   {"pods": 3, "running_pods": 3, "containers": 6}
  },
  "cost": {
    "current_run_rate_hourly": {"amount": "0.4200", "currency": "USD"},
    "cost_mode": "fully_loaded",
    "last_updated_at": "2026-05-20T11:55:00Z"
  }
}`

func TestUnmarshalWorkload(t *testing.T) {
	t.Parallel()

	var got Workload
	require.NoError(t, json.Unmarshal([]byte(workloadFixture), &got))

	assert.Equal(t, "workload", got.Kind)
	assert.Equal(t, "Deployment", got.Metadata.WorkloadKind)
	assert.Equal(t, "c-1", got.Metadata.Cluster.ID)
	assert.Equal(t, "prod-eks-1", got.Metadata.Cluster.Name)
	assert.True(t, got.Metadata.HasHPA, "HasHPA should be true")
	assert.Equal(t, 4.0, got.Capacity.CPU.LimitCores)
	assert.Equal(t, 3, got.Utilization.Replicas.Desired)
	assert.Equal(t, "fully_loaded", got.Cost.CostMode)
	assert.Equal(t, "0.4200", got.Cost.CurrentRunRateHourly.Amount)

	roundtrip, err := json.Marshal(got)
	require.NoError(t, err)
	var second Workload
	require.NoError(t, json.Unmarshal(roundtrip, &second))
	assert.Equal(t, got.Cost.CostMode, second.Cost.CostMode, "round-trip drift cost_mode")
}
