package types

// Node is the /v1 Node resource. Node has ONE physical bill — the endpoint
// REJECTS ?cost_mode= and the cost block OMITS cost.cost_mode.
type Node struct {
	ID          string          `json:"id"`
	Kind        string          `json:"kind"`
	Metadata    NodeMetadata    `json:"metadata"`
	Capacity    NodeCapacity    `json:"capacity"`
	Utilization NodeUtilization `json:"utilization"`
	Cost        NodeCost        `json:"cost"`
}

// NodeMetadata is the identity / scheduling sub-block for a Node.
type NodeMetadata struct {
	Name             string            `json:"name"`
	Cluster          NestedRef         `json:"cluster"`
	NodeRole         string            `json:"node_role,omitempty"`
	InstanceType     string            `json:"instance_type,omitempty"`
	NodeGroup        string            `json:"node_group,omitempty"`
	AvailabilityZone string            `json:"availability_zone,omitempty"`
	Region           string            `json:"region,omitempty"`
	IsSpot           bool              `json:"is_spot"`
	CapacityType     string            `json:"capacity_type,omitempty"`
	Architecture     string            `json:"architecture,omitempty"`
	OperatingSystem  string            `json:"operating_system,omitempty"`
	KubeletVersion   string            `json:"kubelet_version,omitempty"`
	ProviderID       string            `json:"provider_id,omitempty"`
	IsReady          bool              `json:"is_ready"`
	IsSchedulable    bool              `json:"is_schedulable"`
	Labels           map[string]string `json:"labels,omitempty"`
	CreatedAtK8s     string            `json:"created_at_k8s,omitempty"`
	LastSeenAt       string            `json:"last_seen_at,omitempty"`
}

// NodeCapacity is the capacity block for a Node.
type NodeCapacity struct {
	CPU              CapacityCPU     `json:"cpu"`
	Memory           CapacityMemory  `json:"memory"`
	GPU              CapacityGPU     `json:"gpu"`
	EphemeralStorage CapacityStorage `json:"ephemeral_storage"`
	Pods             CapacityPods    `json:"pods"`
}

// NodeUtilization is the utilization block for a Node. Node-level CPU and
// memory utilization do NOT include requested_cores / requested_bytes
// (the request roll-up is meaningful only on the workload / namespace
// level), so this uses Node-specific sub-blocks.
type NodeUtilization struct {
	CPU    NodeUtilizationCPU    `json:"cpu"`
	Memory NodeUtilizationMemory `json:"memory"`
	GPU    UtilizationGPU        `json:"gpu"`
	Counts NodeCounts            `json:"counts"`
}

// NodeUtilizationCPU is the Node-level CPU utilization sub-block. Lacks
// the requested_cores field present on the shared UtilizationCPU.
type NodeUtilizationCPU struct {
	UsedCores          float64 `json:"used_cores"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

// NodeUtilizationMemory is the Node-level memory utilization sub-block.
type NodeUtilizationMemory struct {
	UsedBytes          int64   `json:"used_bytes"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

// NodeCounts is the live-count sub-block for a Node.
type NodeCounts struct {
	Pods        int `json:"pods"`
	RunningPods int `json:"running_pods"`
}

// NodeCost is the node-resource cost block. Like ClusterCost, OMITS the
// CostMode field — node has one physical bill (mode-invariant).
//
// OnDemandEquivalentHourly is populated only on spot nodes and serves as
// the comparison baseline for the spot-savings number.
type NodeCost struct {
	CurrentRunRateHourly     Money  `json:"current_run_rate_hourly"`
	OnDemandEquivalentHourly *Money `json:"on_demand_equivalent_hourly,omitempty"`
	PricingSource            string `json:"pricing_source,omitempty"`
	LastUpdatedAt            string `json:"last_updated_at,omitempty"`
}
