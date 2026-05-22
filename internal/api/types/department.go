package types

// Department is the /v1 Department resource. Teams is the count of member
// teams (direct FK via teams.department_id). AssignedWorkloads /
// AssignedPVs are the transitive rollup across those member teams.
type Department struct {
	ID                string             `json:"id"`
	Kind              string             `json:"kind"`
	Metadata          DepartmentMetadata `json:"metadata"`
	Teams             int                `json:"teams"`
	AssignedWorkloads int                `json:"assigned_workloads"`
	AssignedPVs       int                `json:"assigned_pvs"`
	Cost              DepartmentCost     `json:"cost"`
}

// DepartmentMetadata is the identity / governance sub-block for a
// Department.
type DepartmentMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Origin      string `json:"origin,omitempty"`
	OwnerEmail  string `json:"owner_email,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// DepartmentCost is the department analog of TeamCost. INCLUDES CostMode —
// the rollup is over every workload whose team belongs to this department,
// which varies by mode.
type DepartmentCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	CostMode             string `json:"cost_mode,omitempty"`
}
