package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const clusterFixture = `{
  "id": "11111111-1111-1111-1111-111111111111",
  "kind": "cluster",
  "metadata": {
    "name": "prod-eks-1",
    "provider": "aws",
    "service": "eks",
    "region": "us-east-1",
    "availability_zones": ["us-east-1a", "us-east-1b"],
    "environment": "production",
    "status": "active",
    "is_stale": false,
    "k8s_version": "1.30.4",
    "agent_version": "0.14.2",
    "discovery_source": "agent",
    "created_at": "2026-01-04T08:11:00Z",
    "last_seen_at": "2026-05-20T11:59:30Z"
  },
  "capacity": {
    "cpu":     {"total_cores": 240.0, "allocatable_cores": 232.5},
    "memory":  {"total_bytes": 1099511627776, "allocatable_bytes": 1090519040000},
    "gpu":     {"total": 0, "allocatable": 0},
    "storage": {"total_bytes": 21990232555520},
    "pods":    {"allocatable": 1100}
  },
  "utilization": {
    "cpu":    {"requested_cores": 184.0, "used_cores": 120.4, "utilization_percent": 51.81},
    "memory": {"requested_bytes": 824633720832, "used_bytes": 605590700032, "utilization_percent": 55.55},
    "gpu":    {"utilization_percent": 0, "memory_used_bytes": 0, "memory_total_bytes": 0},
    "counts": {
      "nodes": 8, "namespaces": 23, "workloads": 142,
      "deployments": 96, "statefulsets": 14, "daemonsets": 12,
      "jobs": 14, "cronjobs": 6, "pods": 412, "running_pods": 405,
      "containers": 612, "running_containers": 600, "persistent_volumes": 18
    }
  },
  "cost": {
    "current_run_rate_hourly": {"amount": "12.4700", "currency": "USD"},
    "last_updated_at": "2026-05-20T11:55:00Z"
  }
}`

func TestUnmarshalCluster(t *testing.T) {
	t.Parallel()

	var got Cluster
	require.NoError(t, json.Unmarshal([]byte(clusterFixture), &got))

	assert.Equal(t, "11111111-1111-1111-1111-111111111111", got.ID)
	assert.Equal(t, "cluster", got.Kind)
	assert.Equal(t, "prod-eks-1", got.Metadata.Name)
	assert.Equal(t, "aws", got.Metadata.Provider)
	assert.False(t, got.Metadata.IsStale, "Metadata.IsStale should be false")
	assert.Len(t, got.Metadata.AvailabilityZones, 2)
	assert.Equal(t, 240.0, got.Capacity.CPU.TotalCores)
	assert.Equal(t, int64(1090519040000), got.Capacity.Memory.AllocatableBytes)
	assert.Equal(t, 8, got.Utilization.Counts.Nodes)
	assert.Equal(t, 18, got.Utilization.Counts.PersistentVolumes)
	assert.Equal(t, "12.4700", got.Cost.CurrentRunRateHourly.Amount)
	assert.Equal(t, "USD", got.Cost.CurrentRunRateHourly.Currency)

	roundtrip, err := json.Marshal(got)
	require.NoError(t, err)
	var second Cluster
	require.NoError(t, json.Unmarshal(roundtrip, &second))
	assert.Equal(t, got.ID, second.ID)
	assert.Equal(t, got.Cost.CurrentRunRateHourly.Amount, second.Cost.CurrentRunRateHourly.Amount)
}

func TestUnmarshalCluster_InEnvelope(t *testing.T) {
	t.Parallel()

	payload := `{
		"data": ` + clusterFixture + `,
		"meta": {"request_id": "req-cluster-1", "applied_at": "2026-05-20T12:00:00Z"}
	}`

	var env Envelope[Cluster]
	require.NoError(t, json.Unmarshal([]byte(payload), &env))
	require.NotEmpty(t, env.Data.ID, "envelope Data ID is empty")
	assert.Empty(t, env.Meta.CostMode, "Meta.CostMode should be empty (cluster endpoint rejects cost_mode)")
}
