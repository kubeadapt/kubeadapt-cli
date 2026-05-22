package types

// Recommendation is the /v1 Recommendation resource. One of the bespoke
// resource shapes — metadata + current + recommended + applied + savings +
// metrics_snapshot — NOT the standard capacity/utilization/cost 4-block
// pattern. The endpoint REJECTS ?cost_mode= (savings are mode-agnostic).
//
// Config sub-blocks (Current.Config, Recommended.Config, Applied.Config)
// vary by recommendation_type and are decoded as map[string]any to mirror
// the backend's polymorphic JSON. Same applies to MetricsSnapshot.
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

// RecommendationMetadata is the identity / classification sub-block for
// a Recommendation. RecommendationType drives the polymorphic shape of
// the Config sub-blocks.
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

// RecommendationSnapshot is the "current state" sub-block — the resource
// configuration and hourly cost as it stands today.
type RecommendationSnapshot struct {
	Config     map[string]any `json:"config,omitempty"`
	HourlyCost Money          `json:"hourly_cost"`
}

// RecommendationProposal is the "recommended state" sub-block — the
// proposed resource configuration if the recommendation is applied.
type RecommendationProposal struct {
	Config map[string]any `json:"config,omitempty"`
}

// RecommendationApplied is the "applied state" sub-block — the resource
// configuration actually applied by the customer (may differ from
// Proposal if the customer partially applied the recommendation).
type RecommendationApplied struct {
	Config map[string]any `json:"config"`
}

// RecommendationSavings is the cost-impact sub-block for a Recommendation.
// Only carries the pre-apply hourly estimate; clients format hourly figures
// as needed.
type RecommendationSavings struct {
	EstimatedHourly Money `json:"estimated_hourly"`
}
