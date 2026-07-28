package types

// Teams counts direct member teams; AssignedWorkloads and AssignedPVs are the
// transitive rollup across those teams.
type Department struct {
	ID                string             `json:"id"`
	Kind              string             `json:"kind"`
	Metadata          DepartmentMetadata `json:"metadata"`
	Teams             int                `json:"teams"`
	AssignedWorkloads int                `json:"assigned_workloads"`
	AssignedPVs       int                `json:"assigned_pvs"`
	Cost              DepartmentCost     `json:"cost"`
}

type DepartmentMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Origin      string `json:"origin,omitempty"`
	OwnerEmail  string `json:"owner_email,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// DepartmentCost includes CostMode: it rolls up every workload whose team
// belongs to this department, so it varies by mode.
type DepartmentCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	CostMode             string `json:"cost_mode,omitempty"`
}
