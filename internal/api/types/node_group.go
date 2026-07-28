package types

// NodeGroup aggregates nodes by group name. The endpoint rejects ?cost_mode=
// because its cost sums underlying physical bills.
type NodeGroup struct {
	ID          string               `json:"id"`
	Kind        string               `json:"kind"`
	Metadata    NodeGroupMetadata    `json:"metadata"`
	Capacity    NodeGroupCapacity    `json:"capacity"`
	Utilization NodeGroupUtilization `json:"utilization"`
	Cost        NodeGroupCost        `json:"cost"`

	// Populated only on the detail endpoint; omitted on list endpoints.
	Nodes []Node `json:"nodes,omitempty"`
}

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

type NodeGroupCapacity struct {
	CPU    CapacityCPU    `json:"cpu"`
	Memory CapacityMemory `json:"memory"`
}

type NodeGroupUtilization struct {
	CPU    NodeUtilizationCPU    `json:"cpu"`
	Memory NodeUtilizationMemory `json:"memory"`
	Counts NodeGroupCounts       `json:"counts"`
}

type NodeGroupCounts struct {
	Nodes      int `json:"nodes"`
	ReadyNodes int `json:"ready_nodes"`
	Pods       int `json:"pods"`
}

// SpotSavingsVsOndemandHourly is populated when the group contains spot nodes;
// it is the delta vs the on-demand pricing baseline.
type NodeGroupCost struct {
	CurrentRunRateHourly        Money  `json:"current_run_rate_hourly"`
	SpotSavingsVsOndemandHourly *Money `json:"spot_savings_vs_ondemand_hourly,omitempty"`
	LastUpdatedAt               string `json:"last_updated_at,omitempty"`
}
