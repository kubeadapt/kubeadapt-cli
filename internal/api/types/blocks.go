package types

// NestedRef is the canonical cross-resource reference: every reference uses
// {id, name}, never flat <ref>_id + <ref>_name pairs.
type NestedRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CapacityCPU struct {
	TotalCores       float64 `json:"total_cores"`
	AllocatableCores float64 `json:"allocatable_cores"`
}

type CapacityMemory struct {
	TotalBytes       int64 `json:"total_bytes"`
	AllocatableBytes int64 `json:"allocatable_bytes"`
}

type CapacityGPU struct {
	Total       int    `json:"total"`
	Allocatable int    `json:"allocatable"`
	Model       string `json:"model,omitempty"`
}

type CapacityStorage struct {
	TotalBytes int64 `json:"total_bytes"`
}

type CapacityPods struct {
	Allocatable int `json:"allocatable"`
}

type UtilizationCPU struct {
	RequestedCores     float64 `json:"requested_cores"`
	UsedCores          float64 `json:"used_cores"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

type UtilizationMemory struct {
	RequestedBytes     int64   `json:"requested_bytes"`
	UsedBytes          int64   `json:"used_bytes"`
	UtilizationPercent float64 `json:"utilization_percent"`
}

type UtilizationGPU struct {
	UtilizationPercent float64 `json:"utilization_percent"`
	MemoryUsedBytes    int64   `json:"memory_used_bytes"`
	MemoryTotalBytes   int64   `json:"memory_total_bytes"`
}
