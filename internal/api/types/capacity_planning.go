package types

// NodeByOS represents node count grouped by operating system.
type NodeByOS struct {
	OS         string  `json:"os"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// SpotVsOnDemand summarizes spot vs on-demand node distribution.
type SpotVsOnDemand struct {
	SpotCount       int     `json:"spot_count"`
	SpotPercent     float64 `json:"spot_percent"`
	OnDemandCount   int     `json:"ondemand_count"`
	OnDemandPercent float64 `json:"ondemand_percent"`
	Total           int     `json:"total"`
}

// PodDensity summarizes pod distribution across nodes.
type PodDensity struct {
	TotalPods      int     `json:"total_pods"`
	TotalNodes     int     `json:"total_nodes"`
	AvgPodsPerNode float64 `json:"avg_pods_per_node"`
	MaxPodsPerNode int     `json:"max_pods_per_node"`
}

// CostByAZ represents cost breakdown by availability zone.
type CostByAZ struct {
	Zone       string  `json:"zone"`
	HourlyCost float64 `json:"hourly_cost"`
	NodeCount  int     `json:"node_count"`
}

// NodeGroupSummary is a summary of a node group in capacity planning.
type NodeGroupSummary struct {
	Name         string  `json:"name"`
	InstanceType *string `json:"instance_type"`
	Count        int     `json:"count"`
	HourlyCost   float64 `json:"hourly_cost"`
	SpotPercent  float64 `json:"spot_percent"`
}

// CapacityPlanningResponse is the capacity planning data for a cluster.
type CapacityPlanningResponse struct {
	ClusterID      string             `json:"cluster_id"`
	NodesByOS      []NodeByOS         `json:"nodes_by_os"`
	SpotVsOnDemand SpotVsOnDemand     `json:"spot_vs_ondemand"`
	PodDensity     PodDensity         `json:"pod_density"`
	CostByAZ       []CostByAZ         `json:"cost_by_az"`
	NodeGroups     []NodeGroupSummary `json:"node_groups"`
}
