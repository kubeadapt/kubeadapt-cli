package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const organizationFixture = `{
  "id": "org-1",
  "kind": "organization",
  "metadata": {
    "name": "Acme Corp",
    "domain": "acme.com",
    "plan_type": "enterprise",
    "is_active": true,
    "created_at": "2025-09-01T00:00:00Z"
  },
  "capacity": {
    "cpu":     {"total_cores": 1024.0, "allocatable_cores": 1010.0},
    "memory":  {"total_bytes": 4398046511104, "allocatable_bytes": 4362076094464},
    "gpu":     {"total": 0, "allocatable": 0},
    "storage": {"total_bytes": 109951162777600}
  },
  "utilization": {
    "cpu":    {"requested_cores": 700.0, "used_cores": 540.0, "utilization_percent": 52.7},
    "memory": {"requested_bytes": 2199023255552, "used_bytes": 1500000000000, "utilization_percent": 34.1},
    "counts": {
      "clusters": 12, "connected_clusters": 11, "nodes": 96, "namespaces": 124,
      "workloads": 1420, "pods": 4824, "containers": 6500, "persistent_volumes": 142,
      "recommendations": 38
    }
  },
  "cost": {
    "current_run_rate_hourly": {"amount": "84.5600", "currency": "USD"},
    "last_updated_at": "2026-05-20T11:55:00Z"
  }
}`

func TestUnmarshalOrganization(t *testing.T) {
	t.Parallel()

	var got Organization
	require.NoError(t, json.Unmarshal([]byte(organizationFixture), &got))

	assert.Equal(t, "Acme Corp", got.Metadata.Name)
	assert.True(t, got.Metadata.IsActive, "IsActive should be true")
	assert.Equal(t, 11, got.Utilization.Counts.ConnectedClusters)
	assert.Equal(t, 38, got.Utilization.Counts.Recommendations)
	assert.Equal(t, "84.5600", got.Cost.CurrentRunRateHourly.Amount)

	roundtrip, err := json.Marshal(got)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(roundtrip, &raw))
	costRaw := raw["cost"].(map[string]any)
	_, exists := costRaw["cost_mode"]
	assert.False(t, exists, "organization cost block must NOT contain cost_mode")
}

func TestUnmarshalOrganizationDashboard(t *testing.T) {
	t.Parallel()

	payload := `{
		"organization_id": "org-1",
		"snapshot": ` + organizationFixture + `,
		"month_to_date": {
			"billed_cost": {"amount": "12500.0000", "currency": "USD"},
			"calendar": {
				"month": "2026-05",
				"days_elapsed": 20,
				"days_in_month": 31,
				"month_start_at": "2026-05-01T00:00:00Z"
			}
		},
		"savings": {
			"current_hourly_potential": {"amount": "8.4000", "currency": "USD"},
			"recommendation_count": 38
		},
		"top_clusters": [
			{
				"cluster": {"id": "c-1", "name": "prod-eks-1"},
				"current_run_rate_hourly": {"amount": "12.4700", "currency": "USD"},
				"month_to_date_cost":      {"amount": "5984.0000", "currency": "USD"}
			}
		]
	}`

	var got OrganizationDashboard
	require.NoError(t, json.Unmarshal([]byte(payload), &got))
	assert.Equal(t, "org-1", got.OrganizationID)
	assert.Equal(t, 31, got.MonthToDate.Calendar.DaysInMonth)
	assert.Equal(t, 38, got.Savings.RecommendationCount)
	require.Len(t, got.TopClusters, 1)
	assert.Equal(t, "c-1", got.TopClusters[0].Cluster.ID)
}
