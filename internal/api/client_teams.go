package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// TeamFilter narrows the result set of ListTeams.
type TeamFilter struct {
	PagedOpts
	CostModeOpt
	DepartmentIDs []string // csv
	Origins       []string // csv
}

// TeamGetOpts captures optional query parameters accepted by
// GET /v1/teams/{team_id}.
type TeamGetOpts struct {
	CostModeOpt
}

// AssignmentFilter narrows the result set of ListTeamAssignments.
type AssignmentFilter struct {
	PagedOpts
	EntityType string   // namespace|workload|cluster
	ClusterIDs []string // csv
	Source     string   // manual|label|... (backend-defined)
}

// ListTeams lists teams via GET /v1/teams.
func (c *Client) ListTeams(
	ctx context.Context, f TeamFilter,
) ([]types.Team, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, f.CostMode)
	setCSV(params, "department_id", f.DepartmentIDs)
	setCSV(params, "origin", f.Origins)
	params = appendCursorParams(params, f.Cursor, f.Limit, f.IncludeTotal)
	return DoEnvelopeGet[[]types.Team](ctx, c, "/v1/teams", params)
}

// GetTeam fetches a team by ID via GET /v1/teams/{team_id}.
func (c *Client) GetTeam(
	ctx context.Context, teamID string, opts TeamGetOpts,
) (*types.Team, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, opts.CostMode)
	return DoEnvelopeGet[*types.Team](ctx, c, "/v1/teams/"+teamID, params)
}

// ListTeamAssignments lists assignments for a team via
// GET /v1/teams/{team_id}/assignments.
func (c *Client) ListTeamAssignments(
	ctx context.Context, teamID string, f AssignmentFilter,
) ([]types.TeamAssignment, *types.Meta, error) {
	params := url.Values{}
	if f.EntityType != "" {
		params.Set("entity_type", f.EntityType)
	}
	setCSV(params, "cluster_id", f.ClusterIDs)
	if f.Source != "" {
		params.Set("source", f.Source)
	}
	params = appendCursorParams(params, f.Cursor, f.Limit, f.IncludeTotal)
	return DoEnvelopeGet[[]types.TeamAssignment](ctx, c, "/v1/teams/"+teamID+"/assignments", params)
}
