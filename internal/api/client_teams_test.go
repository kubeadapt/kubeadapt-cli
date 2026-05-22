package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTeams(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.ListTeams(t.Context(), api.TeamFilter{
		DepartmentIDs: []string{"dept-1"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one team")
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Equal(t, "dept-1", last.Query.Get("department_id"))
}

func TestGetTeam(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.GetTeam(t.Context(), "team-1", api.TeamGetOpts{})
	require.NoError(t, err)
	require.NotNil(t, got, "expected team payload")
	assert.NotEmpty(t, got.ID, "expected team payload")
}

func TestListTeamAssignments(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.ListTeamAssignments(t.Context(), "team-1", api.AssignmentFilter{
		EntityType: "namespace",
		Source:     "manual",
	})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one assignment")
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Equal(t, "/v1/teams/team-1/assignments", last.Path)
}
