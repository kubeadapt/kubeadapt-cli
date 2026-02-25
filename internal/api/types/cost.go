package types

// TeamCostResponse represents cost breakdown by team.
type TeamCostResponse struct {
	Team           string  `json:"team"`
	NamespaceCount int     `json:"namespace_count"`
	WorkloadCount  int     `json:"workload_count"`
	PodCount       int     `json:"pod_count"`
	TotalCPUCores  float64 `json:"total_cpu_cores"`
	TotalMemoryGB  float64 `json:"total_memory_gb"`
	HourlyCost     float64 `json:"hourly_cost"`
	MonthlyCost    float64 `json:"monthly_cost"`
}

// TeamCostListResponse is a list of team costs.
type TeamCostListResponse struct {
	Teams   []TeamCostResponse `json:"teams"`
	Total   int                `json:"total"`
	Summary CostSummary        `json:"summary"`
}

// DepartmentCostResponse represents cost breakdown by department.
type DepartmentCostResponse struct {
	Department     string  `json:"department"`
	TeamCount      int     `json:"team_count"`
	NamespaceCount int     `json:"namespace_count"`
	WorkloadCount  int     `json:"workload_count"`
	PodCount       int     `json:"pod_count"`
	TotalCPUCores  float64 `json:"total_cpu_cores"`
	TotalMemoryGB  float64 `json:"total_memory_gb"`
	HourlyCost     float64 `json:"hourly_cost"`
	MonthlyCost    float64 `json:"monthly_cost"`
}

// DepartmentCostListResponse is a list of department costs.
type DepartmentCostListResponse struct {
	Departments []DepartmentCostResponse `json:"departments"`
	Total       int                      `json:"total"`
	Summary     CostSummary              `json:"summary"`
}
