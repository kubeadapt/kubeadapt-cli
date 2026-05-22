package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const departmentFixture = `{
  "id": "d-1",
  "kind": "department",
  "metadata": {
    "name": "engineering",
    "description": "All engineering teams",
    "origin": "manual",
    "owner_email": "vp-eng@acme.com",
    "created_at": "2025-09-01T00:00:00Z",
    "updated_at": "2026-04-12T00:00:00Z"
  },
  "teams": 6,
  "assigned_workloads": 480,
  "assigned_pvs": 42,
  "cost": {
    "current_run_rate_hourly": {"amount": "48.1200", "currency": "USD"},
    "cost_mode": "workload_only"
  }
}`

func TestUnmarshalDepartment(t *testing.T) {
	t.Parallel()

	var got Department
	require.NoError(t, json.Unmarshal([]byte(departmentFixture), &got))

	assert.Equal(t, "department", got.Kind)
	assert.Equal(t, 6, got.Teams)
	assert.Equal(t, 480, got.AssignedWorkloads)
	assert.Equal(t, "workload_only", got.Cost.CostMode)

	roundtrip, err := json.Marshal(got)
	require.NoError(t, err)
	var second Department
	require.NoError(t, json.Unmarshal(roundtrip, &second))
	assert.Equal(t, "48.1200", second.Cost.CurrentRunRateHourly.Amount, "round-trip drift cost amount")
}
