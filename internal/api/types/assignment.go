package types

// TeamAssignment binds a Team to a cluster, namespace or workload, with an
// optional weight for shared-cost allocation.
type TeamAssignment struct {
	ID       string                 `json:"id"`
	Kind     string                 `json:"kind"`
	Metadata TeamAssignmentMetadata `json:"metadata"`
}

// AssignedByUserID is a pointer to distinguish system-assigned (null) from
// human-assigned. Source is "manual", "label" or "auto"; more may be added.
type TeamAssignmentMetadata struct {
	Team             NestedRef `json:"team"`
	Cluster          NestedRef `json:"cluster"`
	EntityType       string    `json:"entity_type"`
	EntityIdentifier string    `json:"entity_identifier"`
	EntityName       string    `json:"entity_name,omitempty"`
	EntityNamespace  string    `json:"entity_namespace,omitempty"`
	WeightPercentage float64   `json:"weight_percentage"`
	Source           string    `json:"source,omitempty"`
	AssignedByUserID *string   `json:"assigned_by_user_id"`
	CreatedAt        string    `json:"created_at,omitempty"`
	UpdatedAt        string    `json:"updated_at,omitempty"`
}
