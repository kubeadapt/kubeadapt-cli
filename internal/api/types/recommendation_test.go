package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const recommendationFixture = `{
  "id": "66666666-6666-6666-6666-666666666666",
  "kind": "recommendation",
  "metadata": {
    "recommendation_type": "workload_rightsizing",
    "resource_type": "Deployment",
    "resource_name": "checkout-api",
    "resource_uid": "w-uid-1",
    "cluster": {"id": "c-1", "name": "prod-eks-1"},
    "namespace": "payments",
    "title": "Reduce CPU requests for checkout-api",
    "description": "P95 CPU usage is 30% of requests over 7d",
    "cause": "over_provisioned",
    "risk_level": "low",
    "priority": "high",
    "status": "pending",
    "data_points_analyzed": 10080,
    "created_at": "2026-05-20T08:00:00Z",
    "updated_at": "2026-05-20T08:00:00Z"
  },
  "current": {
    "config": {"cpu_request_cores": 2.0, "memory_request_bytes": 4294967296},
    "hourly_cost": {"amount": "0.4200", "currency": "USD"}
  },
  "recommended": {
    "config": {"cpu_request_cores": 0.8, "memory_request_bytes": 2147483648}
  },
  "applied": {
    "config": {}
  },
  "savings": {
    "estimated_hourly": {"amount": "0.2100", "currency": "USD"}
  },
  "metrics_snapshot": {"p95_cpu_cores": 0.6, "p99_memory_bytes": 1900000000}
}`

func TestUnmarshalRecommendation(t *testing.T) {
	t.Parallel()

	var got Recommendation
	require.NoError(t, json.Unmarshal([]byte(recommendationFixture), &got))

	assert.Equal(t, "workload_rightsizing", got.Metadata.RecommendationType)
	assert.Equal(t, "pending", got.Metadata.Status)
	assert.Equal(t, "0.4200", got.Current.HourlyCost.Amount)
	cpuRec, ok := got.Recommended.Config["cpu_request_cores"].(float64)
	require.True(t, ok, "Recommended.Config[cpu_request_cores] not a float64")
	assert.Equal(t, 0.8, cpuRec)
	assert.Equal(t, "0.2100", got.Savings.EstimatedHourly.Amount)
	require.NotNil(t, got.MetricsSnapshot, "MetricsSnapshot missing")

	roundtrip, err := json.Marshal(got)
	require.NoError(t, err)
	var second Recommendation
	require.NoError(t, json.Unmarshal(roundtrip, &second))
	assert.Equal(t, got.Metadata.RecommendationType, second.Metadata.RecommendationType, "round-trip drift recommendation_type")
}
