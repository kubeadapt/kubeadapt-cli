package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const assignmentFixture = `{
  "id": "a-1",
  "kind": "team_assignment",
  "metadata": {
    "team":    {"id": "t-1", "name": "platform"},
    "cluster": {"id": "c-1", "name": "prod-eks-1"},
    "entity_type": "namespace",
    "entity_identifier": "payments",
    "entity_name": "payments",
    "entity_namespace": "",
    "weight_percentage": 100.0,
    "source": "manual",
    "assigned_by_user_id": "user_2abc",
    "created_at": "2026-04-12T00:00:00Z",
    "updated_at": "2026-04-12T00:00:00Z"
  }
}`

func TestUnmarshalTeamAssignment(t *testing.T) {
	t.Parallel()

	var got TeamAssignment
	require.NoError(t, json.Unmarshal([]byte(assignmentFixture), &got))

	assert.Equal(t, "team_assignment", got.Kind)
	assert.Equal(t, "t-1", got.Metadata.Team.ID)
	assert.Equal(t, "namespace", got.Metadata.EntityType)
	assert.Equal(t, 100.0, got.Metadata.WeightPercentage)
	require.NotNil(t, got.Metadata.AssignedByUserID)
	assert.Equal(t, "user_2abc", *got.Metadata.AssignedByUserID)
}

func TestUnmarshalTeamAssignment_SystemAssigned(t *testing.T) {
	t.Parallel()

	payload := `{
		"id": "a-2", "kind": "team_assignment",
		"metadata": {
			"team":    {"id": "t-1", "name": "platform"},
			"cluster": {"id": "c-1", "name": "prod-eks-1"},
			"entity_type": "workload",
			"entity_identifier": "w-uid-1",
			"weight_percentage": 100.0,
			"source": "label",
			"assigned_by_user_id": null
		}
	}`

	var got TeamAssignment
	require.NoError(t, json.Unmarshal([]byte(payload), &got))
	assert.Nil(t, got.Metadata.AssignedByUserID, "AssignedByUserID should be nil for system-assigned")
}
