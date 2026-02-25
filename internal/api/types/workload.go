package types

// WorkloadResponse represents a single workload.
type WorkloadResponse struct {
	ID                string   `json:"id"`
	ClusterID         string   `json:"cluster_id"`
	ClusterName       string   `json:"cluster_name"`
	Namespace         string   `json:"namespace"`
	WorkloadName      string   `json:"workload_name"`
	WorkloadKind      string   `json:"workload_kind"`
	Replicas          int      `json:"replicas"`
	AvailableReplicas int      `json:"available_replicas"`
	CPURequest        float64  `json:"cpu_request"`
	CPULimit          float64  `json:"cpu_limit"`
	MemoryRequestGB   float64  `json:"memory_request_gb"`
	MemoryLimitGB     float64  `json:"memory_limit_gb"`
	HourlyCost        float64  `json:"hourly_cost"`
	EfficiencyScore   *float64 `json:"efficiency_score"`
	MonthlyCost       *float64 `json:"monthly_cost"`
}

// WorkloadListResponse is a list of workloads.
type WorkloadListResponse struct {
	Workloads []WorkloadResponse `json:"workloads"`
	Total     int                `json:"total"`
}
