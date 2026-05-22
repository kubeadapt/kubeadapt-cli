package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const namespaceFixture = `{
  "id": "payments",
  "kind": "namespace",
  "metadata": {
    "name": "payments",
    "cluster": {"id": "c-1", "name": "prod-eks-1"},
    "uid_k8s": "abc-uid",
    "labels": {"owner": "platform"},
    "team": {"id": "t-1", "name": "platform"},
    "department": {"id": "d-1", "name": "engineering"},
    "created_at_k8s": "2026-01-04T08:11:00Z",
    "last_seen_at": "2026-05-20T11:59:30Z"
  },
  "capacity": {
    "cpu":    {"quota_cores": 50.0},
    "memory": {"quota_bytes": 107374182400}
  },
  "utilization": {
    "cpu":    {"requested_cores": 24.0, "used_cores": 12.4, "utilization_percent": 24.8},
    "memory": {"requested_bytes": 53687091200, "used_bytes": 26843545600, "utilization_percent": 25.0},
    "counts": {
      "workloads": 24, "deployments": 18, "statefulsets": 4, "daemonsets": 0,
      "jobs": 1, "cronjobs": 1, "pods": 76, "running_pods": 75,
      "containers": 120, "persistent_volumes": 6
    }
  },
  "cost": {
    "current_run_rate_hourly": {"amount": "3.2100", "currency": "USD"},
    "cost_mode": "workload_only",
    "last_updated_at": "2026-05-20T11:55:00Z"
  },
  "workloads_top_5": [
    {
      "id": "w-1", "kind": "workload", "name": "checkout-api",
      "cost": {"current_run_rate_hourly": {"amount": "0.4200", "currency": "USD"}}
    }
  ]
}`

func TestUnmarshalNamespace(t *testing.T) {
	t.Parallel()

	var got Namespace
	require.NoError(t, json.Unmarshal([]byte(namespaceFixture), &got))

	assert.Equal(t, "payments", got.ID)
	require.NotNil(t, got.Metadata.Team)
	assert.Equal(t, "t-1", got.Metadata.Team.ID)
	require.NotNil(t, got.Metadata.Department)
	assert.Equal(t, "engineering", got.Metadata.Department.Name)
	require.NotNil(t, got.Capacity, "Capacity should be populated (quota present)")
	assert.Equal(t, 50.0, got.Capacity.CPU.QuotaCores)
	assert.Equal(t, 6, got.Utilization.Counts.PersistentVolumes)
	assert.Equal(t, "workload_only", got.Cost.CostMode)
	assert.Len(t, got.WorkloadsTop5, 1)
	assert.Equal(t, "0.4200", got.WorkloadsTop5[0].Cost.CurrentRunRateHourly.Amount)
}

func TestUnmarshalNamespace_NoQuota(t *testing.T) {
	t.Parallel()

	payload := `{
		"id": "default", "kind": "namespace",
		"metadata": {"name": "default", "cluster": {"id": "c-1", "name": "prod-eks-1"}},
		"utilization": {
			"cpu":    {"requested_cores": 0, "used_cores": 0, "utilization_percent": 0},
			"memory": {"requested_bytes": 0, "used_bytes": 0, "utilization_percent": 0},
			"counts": {"workloads": 0, "deployments": 0, "statefulsets": 0, "daemonsets": 0,
			           "jobs": 0, "cronjobs": 0, "pods": 0, "running_pods": 0,
			           "containers": 0, "persistent_volumes": 0}
		},
		"cost": {"current_run_rate_hourly": {"amount": "0.0000", "currency": "USD"}}
	}`

	var got Namespace
	require.NoError(t, json.Unmarshal([]byte(payload), &got))
	assert.Nil(t, got.Capacity, "Capacity should be nil when no quota")
}
