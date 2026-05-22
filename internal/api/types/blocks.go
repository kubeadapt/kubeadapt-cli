package types

// NestedRef is the canonical cross-resource reference shape used by /v1
// resource bodies. Every reference between resources uses {id, name} —
// never flat <ref>_id + <ref>_name pairs.
type NestedRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CapacityCPU is the shared CPU capacity sub-block on aggregate resources
// (Cluster, Node, NodeGroup, Organization).
type CapacityCPU struct {
	TotalCores       float64 `json:"total_cores"`
	AllocatableCores float64 `json:"allocatable_cores"`
}

// CapacityMemory is the shared memory capacity sub-block on aggregate
// resources.
type CapacityMemory struct {
	TotalBytes       int64 `json:"total_bytes"`
	AllocatableBytes int64 `json:"allocatable_bytes"`
}

// CapacityGPU is the shared GPU capacity sub-block.
type CapacityGPU struct {
	Total       int    `json:"total"`
	Allocatable int    `json:"allocatable"`
	Model       string `json:"model,omitempty"`
}

// CapacityStorage is the shared storage capacity sub-block.
type CapacityStorage struct {
	TotalBytes int64 `json:"total_bytes"`
}

// CapacityPods is the shared pod-count capacity sub-block.
type CapacityPods struct {
	Allocatable int `json:"allocatable"`
}

// UtilizationCPU is the shared CPU utilization sub-block.
type UtilizationCPU struct {
	RequestedCores     float64 `json:"requested_cores"`
	UsedCores          float64 `json:"used_cores"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

// UtilizationMemory is the shared memory utilization sub-block.
type UtilizationMemory struct {
	RequestedBytes     int64   `json:"requested_bytes"`
	UsedBytes          int64   `json:"used_bytes"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

// UtilizationGPU is the shared GPU utilization sub-block.
type UtilizationGPU struct {
	UtilizationPercent float64 `json:"utilization_percent"`
	MemoryUsedBytes    int64   `json:"memory_used_bytes"`
	MemoryTotalBytes   int64   `json:"memory_total_bytes"`
}
