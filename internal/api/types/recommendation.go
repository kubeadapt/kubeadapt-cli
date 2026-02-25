package types

// RecommendationResponse represents a single recommendation.
type RecommendationResponse struct {
	ID                      string  `json:"id"`
	ClusterID               string  `json:"cluster_id"`
	ClusterName             string  `json:"cluster_name"`
	RecommendationType      string  `json:"recommendation_type"`
	Namespace               *string `json:"namespace"`
	ResourceName            *string `json:"resource_name"`
	ResourceType            *string `json:"resource_type"`
	Title                   *string `json:"title"`
	Description             *string `json:"description"`
	EstimatedHourlySavings  float64 `json:"estimated_hourly_savings"`
	EstimatedMonthlySavings float64 `json:"estimated_monthly_savings"`
	CurrentHourlyCost       float64 `json:"current_hourly_cost"`
	Status                  string  `json:"status"`
	CreatedAt               *string `json:"created_at"`
	Priority                *string `json:"priority"`
}

// RecommendationListResponse is a list of recommendations.
type RecommendationListResponse struct {
	Recommendations              []RecommendationResponse `json:"recommendations"`
	Total                        int                      `json:"total"`
	TotalPotentialSavingsMonthly *float64                 `json:"total_potential_savings_monthly"`
}
