package types

// CostSummary is the aggregate summary in cost-related responses.
type CostSummary struct {
	TotalHourlyCost  float64 `json:"total_hourly_cost"`
	TotalMonthlyCost float64 `json:"total_monthly_cost"`
}

// NamespaceSummary is the aggregate summary in namespace responses.
type NamespaceSummary struct {
	TotalHourlyCost float64 `json:"total_hourly_cost"`
	TotalPods       int     `json:"total_pods"`
	TotalWorkloads  int     `json:"total_workloads"`
}

// PVSummary is the aggregate summary in persistent volume responses.
type PVSummary struct {
	TotalCapacityGB float64 `json:"total_capacity_gb"`
	TotalHourlyCost float64 `json:"total_hourly_cost"`
}
