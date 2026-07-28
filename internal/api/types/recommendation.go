package types

// Recommendation does not follow the standard capacity/utilization/cost shape,
// and its endpoint rejects ?cost_mode=. The Config sub-blocks are map[string]any
// because their shape is polymorphic on recommendation_type.
type Recommendation struct {
	ID              string                 `json:"id"`
	Kind            string                 `json:"kind"`
	Metadata        RecommendationMetadata `json:"metadata"`
	Current         RecommendationSnapshot `json:"current"`
	Recommended     RecommendationProposal `json:"recommended"`
	Applied         RecommendationApplied  `json:"applied"`
	Savings         RecommendationSavings  `json:"savings"`
	MetricsSnapshot map[string]any         `json:"metrics_snapshot,omitempty"`
}

// RecommendationType drives the polymorphic shape of the Config sub-blocks.
type RecommendationMetadata struct {
	RecommendationType string    `json:"recommendation_type"`
	ResourceType       string    `json:"resource_type,omitempty"`
	ResourceName       string    `json:"resource_name,omitempty"`
	ResourceUID        string    `json:"resource_uid,omitempty"`
	Cluster            NestedRef `json:"cluster"`
	Namespace          string    `json:"namespace,omitempty"`
	Title              string    `json:"title,omitempty"`
	Description        string    `json:"description,omitempty"`
	Cause              string    `json:"cause,omitempty"`
	RiskLevel          string    `json:"risk_level,omitempty"`
	Priority           string    `json:"priority,omitempty"`
	Status             string    `json:"status"`
	DataPointsAnalyzed int       `json:"data_points_analyzed,omitempty"`
	CreatedAt          string    `json:"created_at,omitempty"`
	UpdatedAt          string    `json:"updated_at,omitempty"`
}

type RecommendationSnapshot struct {
	Config     map[string]any `json:"config,omitempty"`
	HourlyCost Money          `json:"hourly_cost"`
}

type RecommendationProposal struct {
	Config map[string]any `json:"config,omitempty"`
}

// RecommendationApplied may differ from the proposal when the customer applied
// it only partially.
type RecommendationApplied struct {
	Config map[string]any `json:"config"`
}

type RecommendationSavings struct {
	EstimatedHourly Money `json:"estimated_hourly"`
}
