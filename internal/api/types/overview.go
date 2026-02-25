package types

// OverviewResponse represents the organization dashboard overview.
type OverviewResponse struct {
	OrganizationID          string   `json:"organization_id"`
	ClusterCount            int      `json:"cluster_count"`
	ConnectedClusterCount   int      `json:"connected_cluster_count"`
	TotalNodes              int      `json:"total_nodes"`
	TotalPods               int      `json:"total_pods"`
	TotalWorkloads          int      `json:"total_workloads"`
	TotalHourlyCost         *float64 `json:"total_hourly_cost"`
	TotalMonthlyCost        *float64 `json:"total_monthly_cost"`
	PotentialMonthlySavings *float64 `json:"potential_monthly_savings"`
	AvgCPUUtilization       *float64 `json:"avg_cpu_utilization"`
	AvgMemoryUtilization    *float64 `json:"avg_memory_utilization"`
	RecommendationCount     int      `json:"recommendation_count"`
	MTDActualCost           *float64 `json:"mtd_actual_cost"`
	RunRate                 *float64 `json:"run_rate"`
	EfficiencyScore         *float64 `json:"efficiency_score"`
}
