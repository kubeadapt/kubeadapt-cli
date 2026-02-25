package types

// ClusterResponse represents a single cluster.
type ClusterResponse struct {
	ID                       string   `json:"id"`
	Name                     string   `json:"name"`
	Provider                 string   `json:"provider"`
	Region                   *string  `json:"region"`
	Environment              string   `json:"environment"`
	Status                   string   `json:"status"`
	Version                  *string  `json:"version"`
	NodeCount                int      `json:"node_count"`
	PodCount                 int      `json:"pod_count"`
	CPUCores                 float64  `json:"cpu_cores"`
	MemoryGB                 float64  `json:"memory_gb"`
	CPUUtilizationPercent    float64  `json:"cpu_utilization_percent"`
	MemoryUtilizationPercent float64  `json:"memory_utilization_percent"`
	HourlyCost               float64  `json:"hourly_cost"`
	EfficiencyScore          *float64 `json:"efficiency_score"`
	MonthlyCost              *float64 `json:"monthly_cost"`
	PotentialMonthlySavings  *float64 `json:"potential_monthly_savings"`
	RecommendationCount      *int     `json:"recommendation_count"`
}

// ClusterListResponse is a list of clusters.
type ClusterListResponse struct {
	Clusters []ClusterResponse `json:"clusters"`
	Total    int               `json:"total"`
}
