package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const podFixture = `{
  "id": "55555555-5555-5555-5555-555555555555",
  "kind": "pod",
  "metadata": {
    "name": "checkout-api-7d4f-abc12",
    "namespace": "payments",
    "cluster": {"id": "c-1", "name": "prod-eks-1"},
    "workload": {"uid": "w-uid-1", "kind": "Deployment", "name": "checkout-api"},
    "node": {"id": "n-1", "name": "ip-10-0-1-42.ec2.internal"},
    "phase": "Running",
    "qos_class": "Burstable",
    "pod_ip": "10.0.4.21",
    "host_ip": "10.0.1.42",
    "host_network": false,
    "priority_class": "default",
    "has_hostpath": false,
    "has_emptydir": true,
    "labels": {"app": "checkout-api"},
    "created_at_k8s": "2026-05-19T08:11:00Z",
    "last_seen_at": "2026-05-20T11:59:30Z"
  },
  "capacity": {
    "cpu":    {"limit_cores": 1.0},
    "memory": {"limit_bytes": 1073741824}
  },
  "utilization": {
    "cpu":    {"requested_cores": 0.5, "used_cores": 0.3, "utilization_percent": 30.0},
    "memory": {"requested_bytes": 268435456, "used_bytes": 134217728, "utilization_percent": 12.5},
    "counts": {"containers": 2, "running_containers": 2, "ready_containers": 2},
    "restarts_total": 0,
    "oom_kills_total": 0
  },
  "cost": {
    "current_run_rate_hourly": {"amount": "0.1400", "currency": "USD"},
    "cost_mode": "fully_loaded",
    "last_updated_at": "2026-05-20T11:55:00Z"
  }
}`

func TestUnmarshalPod(t *testing.T) {
	t.Parallel()

	var got Pod
	require.NoError(t, json.Unmarshal([]byte(podFixture), &got))

	require.NotNil(t, got.Metadata.Workload)
	assert.Equal(t, "w-uid-1", got.Metadata.Workload.UID)
	assert.Equal(t, "Deployment", got.Metadata.Workload.Kind)
	require.NotNil(t, got.Metadata.Node)
	assert.Equal(t, "ip-10-0-1-42.ec2.internal", got.Metadata.Node.Name)
	assert.True(t, got.Metadata.HasEmptyDir, "HasEmptyDir should be true")
	assert.Equal(t, 2, got.Utilization.Counts.ReadyContainers)
	assert.Equal(t, "fully_loaded", got.Cost.CostMode)

	roundtrip, err := json.Marshal(got)
	require.NoError(t, err)
	var second Pod
	require.NoError(t, json.Unmarshal(roundtrip, &second))
	assert.Equal(t, "0.1400", second.Cost.CurrentRunRateHourly.Amount)
}
