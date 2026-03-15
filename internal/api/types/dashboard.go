package types

// TopCluster represents a cluster entry in the top clusters list.
type TopCluster struct {
	ClusterID   string   `json:"cluster_id"`
	ClusterName string   `json:"cluster_name"`
	HourlyCost  float64  `json:"hourly_cost"`
	Efficiency  *float64 `json:"efficiency"`
}

// DashboardResponse is the organization-level dashboard response.
type DashboardResponse struct {
	OrganizationID          string       `json:"organization_id"`
	TotalMonthlyCost        float64      `json:"total_monthly_cost"`
	TotalHourlyCost         float64      `json:"total_hourly_cost"`
	PotentialMonthlySavings float64      `json:"potential_monthly_savings"`
	EfficiencyScore         *float64     `json:"efficiency_score"`
	ClusterCount            int          `json:"cluster_count"`
	NodeCount               int          `json:"node_count"`
	PodCount                int          `json:"pod_count"`
	MTDActualCost           float64      `json:"mtd_actual_cost"`
	RunRate                 float64      `json:"run_rate"`
	DaysElapsed             int          `json:"days_elapsed"`
	DaysInMonth             int          `json:"days_in_month"`
	TopClusters             []TopCluster `json:"top_clusters"`
	TotalRecommendations    int          `json:"total_recommendations"`
}
