package types

// NodeGroup is the /v1 NodeGroup resource — an aggregate over node_metadata
// GROUP BY node_group. The endpoint REJECTS ?cost_mode= because the cost
// is a sum of underlying physical bills (mode-invariant).
type NodeGroup struct {
	ID          string               `json:"id"`
	Kind        string               `json:"kind"`
	Metadata    NodeGroupMetadata    `json:"metadata"`
	Capacity    NodeGroupCapacity    `json:"capacity"`
	Utilization NodeGroupUtilization `json:"utilization"`
	Cost        NodeGroupCost        `json:"cost"`

	// Nodes is populated only on the detail endpoint
	// (GET /v1/clusters/{cid}/node-groups/{name}) — the member nodes of
	// this group. Omitted on list endpoints.
	Nodes []Node `json:"nodes,omitempty"`
}

// NodeGroupMetadata is the identity / fleet-composition sub-block.
type NodeGroupMetadata struct {
	Name                   string    `json:"name"`
	Cluster                NestedRef `json:"cluster"`
	InstanceTypes          []string  `json:"instance_types,omitempty"`
	Zones                  []string  `json:"zones,omitempty"`
	SpotCount              int       `json:"spot_count"`
	OnDemandCount          int       `json:"ondemand_count"`
	SpotPercentage         float64   `json:"spot_percentage"`
	OldestNodeCreatedAtK8s string    `json:"oldest_node_created_at_k8s,omitempty"`
	Status                 string    `json:"status,omitempty"`
}

// NodeGroupCapacity is the capacity block for a NodeGroup — the sum of
// member-node capacities.
type NodeGroupCapacity struct {
	CPU    CapacityCPU    `json:"cpu"`
	Memory CapacityMemory `json:"memory"`
}

// NodeGroupUtilization is the utilization block for a NodeGroup.
type NodeGroupUtilization struct {
	CPU    NodeUtilizationCPU    `json:"cpu"`
	Memory NodeUtilizationMemory `json:"memory"`
	Counts NodeGroupCounts       `json:"counts"`
}

// NodeGroupCounts is the live-count sub-block for a NodeGroup.
type NodeGroupCounts struct {
	Nodes      int `json:"nodes"`
	ReadyNodes int `json:"ready_nodes"`
	Pods       int `json:"pods"`
}

// NodeGroupCost is the cost block for a NodeGroup. OMITS CostMode — a
// node group's bill is the sum of its member-node physical bills.
//
// SpotSavingsVsOndemandHourly is populated when the group contains spot
// nodes and represents the delta vs. the on-demand pricing baseline.
type NodeGroupCost struct {
	CurrentRunRateHourly        Money  `json:"current_run_rate_hourly"`
	SpotSavingsVsOndemandHourly *Money `json:"spot_savings_vs_ondemand_hourly,omitempty"`
	LastUpdatedAt               string `json:"last_updated_at,omitempty"`
}
