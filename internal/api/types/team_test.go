package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const teamFixture = `{
  "id": "t-1",
  "kind": "team",
  "metadata": {
    "name": "platform",
    "description": "Platform engineering team",
    "origin": "manual",
    "owner_email": "platform@acme.com",
    "department": {"id": "d-1", "name": "engineering"},
    "created_at": "2025-09-15T00:00:00Z",
    "updated_at": "2026-04-12T00:00:00Z",
    "last_seen_at": "2026-05-20T11:55:00Z"
  },
  "assigned_workloads": 142,
  "assigned_pvs": 18,
  "cost": {
    "current_run_rate_hourly": {"amount": "12.4700", "currency": "USD"},
    "cost_mode": "fully_loaded"
  }
}`

func TestUnmarshalTeam(t *testing.T) {
	t.Parallel()

	var got Team
	require.NoError(t, json.Unmarshal([]byte(teamFixture), &got))

	assert.Equal(t, "t-1", got.ID)
	assert.Equal(t, "team", got.Kind)
	require.NotNil(t, got.Metadata.Department)
	assert.Equal(t, "engineering", got.Metadata.Department.Name)
	assert.Equal(t, 142, got.AssignedWorkloads)
	assert.Equal(t, 18, got.AssignedPVs)
	assert.Equal(t, "fully_loaded", got.Cost.CostMode)
	assert.Equal(t, "12.4700", got.Cost.CurrentRunRateHourly.Amount)

	roundtrip, err := json.Marshal(got)
	require.NoError(t, err)
	var second Team
	require.NoError(t, json.Unmarshal(roundtrip, &second))
	assert.Equal(t, got.Cost.CostMode, second.Cost.CostMode, "round-trip drift")
}
