package types

// NodeResponse represents a single node.
type NodeResponse struct {
	ID                  string   `json:"id"`
	ClusterID           string   `json:"cluster_id"`
	ClusterName         string   `json:"cluster_name"`
	NodeName            string   `json:"node_name"`
	InstanceType        *string  `json:"instance_type"`
	NodeGroup           *string  `json:"node_group"`
	AvailabilityZone    *string  `json:"availability_zone"`
	IsReady             bool     `json:"is_ready"`
	IsSchedulable       bool     `json:"is_schedulable"`
	CPUCapacity         float64  `json:"cpu_capacity"`
	CPUAllocatable      float64  `json:"cpu_allocatable"`
	MemoryCapacityGB    float64  `json:"memory_capacity_gb"`
	MemoryAllocatableGB float64  `json:"memory_allocatable_gb"`
	PodsCapacity        int      `json:"pods_capacity"`
	PodsAllocatable     int      `json:"pods_allocatable"`
	HourlyCost          float64  `json:"hourly_cost"`
	SpotInstance        bool     `json:"spot_instance"`
	PodCount            *int     `json:"pod_count"`
	MonthlyCost         *float64 `json:"monthly_cost"`
}

// NodeListResponse is a list of nodes.
type NodeListResponse struct {
	Nodes []NodeResponse `json:"nodes"`
	Total int            `json:"total"`
}
