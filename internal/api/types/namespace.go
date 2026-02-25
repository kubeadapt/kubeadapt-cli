package types

// NamespaceResponse represents a single namespace.
type NamespaceResponse struct {
	Name            string   `json:"name"`
	ClusterID       string   `json:"cluster_id"`
	ClusterName     *string  `json:"cluster_name"`
	PodCount        int      `json:"pod_count"`
	WorkloadCount   int      `json:"workload_count"`
	TotalCPUCores   float64  `json:"total_cpu_cores"`
	TotalMemoryGB   float64  `json:"total_memory_gb"`
	HourlyCost      float64  `json:"hourly_cost"`
	Team            *string  `json:"team"`
	Department      *string  `json:"department"`
	EfficiencyScore *float64 `json:"efficiency_score"`
	MonthlyCost     *float64 `json:"monthly_cost"`
	ContainerCount  *int     `json:"container_count"`
}

// NamespaceListResponse is a list of namespaces.
type NamespaceListResponse struct {
	Namespaces []NamespaceResponse `json:"namespaces"`
	Total      int                 `json:"total"`
	Summary    NamespaceSummary    `json:"summary"`
}
