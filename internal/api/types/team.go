package types

// Team is the /v1 Team resource. AssignedWorkloads / AssignedPVs are raw
// counts from team_assignments — for the breakdown call
// /v1/teams/{team_id}/assignments. Cost is built live from CostQL on
// every read.
type Team struct {
	ID                string       `json:"id"`
	Kind              string       `json:"kind"`
	Metadata          TeamMetadata `json:"metadata"`
	AssignedWorkloads int          `json:"assigned_workloads"`
	AssignedPVs       int          `json:"assigned_pvs"`
	Cost              TeamCost     `json:"cost"`
}

// TeamMetadata is the identity / governance sub-block for a Team.
type TeamMetadata struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Origin      string     `json:"origin,omitempty"`
	OwnerEmail  string     `json:"owner_email,omitempty"`
	Department  *NestedRef `json:"department,omitempty"`
	CreatedAt   string     `json:"created_at,omitempty"`
	UpdatedAt   string     `json:"updated_at,omitempty"`
	LastSeenAt  string     `json:"last_seen_at,omitempty"`
}

// TeamCost is the live run-rate cost block on a Team. INCLUDES CostMode —
// team cost is a sum-of-workload-costs across assignments which DOES vary
// by mode.
type TeamCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	CostMode             string `json:"cost_mode,omitempty"`
}
