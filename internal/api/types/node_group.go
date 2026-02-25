package types

// NodeGroupResponse represents a single node group.
type NodeGroupResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ClusterID      string   `json:"cluster_id"`
	ClusterName    *string  `json:"cluster_name"`
	NodeCount      int      `json:"node_count"`
	InstanceType   *string  `json:"instance_type"`
	TotalCPUCores  *float64 `json:"total_cpu_cores"`
	TotalMemoryGB  *float64 `json:"total_memory_gb"`
	SpotPercentage *float64 `json:"spot_percentage"`
	HourlyCost     *float64 `json:"hourly_cost"`
}

// NodeGroupListResponse is a list of node groups.
type NodeGroupListResponse struct {
	NodeGroups []NodeGroupResponse `json:"node_groups"`
	Total      int                 `json:"total"`
}
