package types

// AssignedWorkloads and AssignedPVs are raw counts; the breakdown lives at
// /v1/teams/{team_id}/assignments. Cost is computed live on every read.
type Team struct {
	ID                string       `json:"id"`
	Kind              string       `json:"kind"`
	Metadata          TeamMetadata `json:"metadata"`
	AssignedWorkloads int          `json:"assigned_workloads"`
	AssignedPVs       int          `json:"assigned_pvs"`
	Cost              TeamCost     `json:"cost"`
}

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

// TeamCost includes CostMode: it sums workload costs across assignments, so it
// varies by mode.
type TeamCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	CostMode             string `json:"cost_mode,omitempty"`
}
