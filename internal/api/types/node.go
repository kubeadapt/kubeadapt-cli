package types

// Node has one physical bill, so the endpoint rejects ?cost_mode= and the cost
// block omits cost.cost_mode.
type Node struct {
	ID          string          `json:"id"`
	Kind        string          `json:"kind"`
	Metadata    NodeMetadata    `json:"metadata"`
	Capacity    NodeCapacity    `json:"capacity"`
	Utilization NodeUtilization `json:"utilization"`
	Cost        NodeCost        `json:"cost"`
}

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

type NodeCapacity struct {
	CPU              CapacityCPU     `json:"cpu"`
	Memory           CapacityMemory  `json:"memory"`
	GPU              CapacityGPU     `json:"gpu"`
	EphemeralStorage CapacityStorage `json:"ephemeral_storage"`
	Pods             CapacityPods    `json:"pods"`
}

// Uses Node-specific CPU/memory sub-blocks because request roll-ups are
// meaningful only at the workload/namespace level.
type NodeUtilization struct {
	CPU    NodeUtilizationCPU    `json:"cpu"`
	Memory NodeUtilizationMemory `json:"memory"`
	GPU    UtilizationGPU        `json:"gpu"`
	Counts NodeCounts            `json:"counts"`
}

type NodeUtilizationCPU struct {
	UsedCores          float64 `json:"used_cores"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

type NodeUtilizationMemory struct {
	UsedBytes          int64   `json:"used_bytes"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

type NodeCounts struct {
	Pods        int `json:"pods"`
	RunningPods int `json:"running_pods"`
}

// OnDemandEquivalentHourly is populated only on spot nodes, as the baseline for
// the spot-savings number.
type NodeCost struct {
	CurrentRunRateHourly     Money  `json:"current_run_rate_hourly"`
	OnDemandEquivalentHourly *Money `json:"on_demand_equivalent_hourly,omitempty"`
	PricingSource            string `json:"pricing_source,omitempty"`
	LastUpdatedAt            string `json:"last_updated_at,omitempty"`
}
