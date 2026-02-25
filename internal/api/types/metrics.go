package types

// --- Cluster Dashboard ---

// ClusterDashboardResponse is the dashboard summary for a cluster.
type ClusterDashboardResponse struct {
	ClusterID   string  `json:"cluster_id"`
	ClusterName string  `json:"cluster_name"`
	Provider    string  `json:"provider"`
	Region      *string `json:"region"`
	Environment string  `json:"environment"`
	Status      string  `json:"status"`
	Version     *string `json:"version"`

	NodeCount       int `json:"node_count"`
	PodCount        int `json:"pod_count"`
	ContainerCount  int `json:"container_count"`
	DeploymentCount int `json:"deployment_count"`
	NamespaceCount  int `json:"namespace_count"`

	HourlyCost         float64 `json:"hourly_cost"`
	MonthlyCost        float64 `json:"monthly_cost"`
	TotalSavingsHourly float64 `json:"total_savings_hourly"`
	MonthlySavings     float64 `json:"monthly_savings"`

	CPUCores                 float64 `json:"cpu_cores"`
	CPUUsage                 float64 `json:"cpu_usage"`
	CPUUtilizationPercent    float64 `json:"cpu_utilization_percent"`
	MemoryGB                 float64 `json:"memory_gb"`
	MemoryUsageGB            float64 `json:"memory_usage_gb"`
	MemoryUtilizationPercent float64 `json:"memory_utilization_percent"`
	ClusterEfficiency        float64 `json:"cluster_efficiency"`

	RecommendationCount int `json:"recommendation_count"`

	CostBreakdown           map[string]float64          `json:"cost_breakdown"`
	MTDActualCost           *float64                    `json:"mtd_actual_cost"`
	PotentialMonthlySavings *float64                    `json:"potential_monthly_savings"`
	RecommendationSummary   []RecommendationSummaryItem `json:"recommendation_summary"`
}

// RecommendationSummaryItem is a single entry in the recommendation summary.
type RecommendationSummaryItem struct {
	Type             string  `json:"type"`
	Count            int     `json:"count"`
	PotentialSavings float64 `json:"potential_savings"`
}

// --- Cost Distribution ---

// CostDistributionPoint is a single time-series data point.
type CostDistributionPoint struct {
	Timestamp         string  `json:"timestamp"`
	HourlyCost        float64 `json:"hourly_cost"`
	CPUUtilization    float64 `json:"cpu_utilization"`
	MemoryUtilization float64 `json:"memory_utilization"`
	Efficiency        float64 `json:"efficiency"`
}

// CostDistributionResponse is time-series cost/utilization data.
type CostDistributionResponse struct {
	ClusterID  string                  `json:"cluster_id"`
	Timeframe  string                  `json:"timeframe"`
	DataPoints []CostDistributionPoint `json:"data_points"`
}

// --- Node Metrics ---

// NodeMetricPoint is a single time-series data point for node metrics.
type NodeMetricPoint struct {
	Timestamp           string  `json:"timestamp"`
	CPUUsage            float64 `json:"cpu_usage"`
	CPUCapacity         float64 `json:"cpu_capacity"`
	MemoryUsageBytes    float64 `json:"memory_usage_bytes"`
	MemoryCapacityBytes float64 `json:"memory_capacity_bytes"`
	CPUUsagePercent     float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent  float64 `json:"memory_usage_percent"`
}

// NodeMetricsResponse is time-series metrics for a node.
type NodeMetricsResponse struct {
	NodeUID    string            `json:"node_uid"`
	ClusterID  string            `json:"cluster_id"`
	Timeframe  string            `json:"timeframe"`
	DataPoints []NodeMetricPoint `json:"data_points"`
}

// --- Workload Metrics ---

// WorkloadMetricPoint is a single time-series data point for workload metrics.
type WorkloadMetricPoint struct {
	Timestamp          string  `json:"timestamp"`
	CPUUsage           float64 `json:"cpu_usage"`
	CPURequest         float64 `json:"cpu_request"`
	MemoryUsageBytes   float64 `json:"memory_usage_bytes"`
	MemoryRequestBytes float64 `json:"memory_request_bytes"`
	HourlyCost         float64 `json:"hourly_cost"`
}

// WorkloadMetricsResponse is time-series metrics for a workload.
type WorkloadMetricsResponse struct {
	WorkloadUID string                `json:"workload_uid"`
	ClusterID   string                `json:"cluster_id"`
	Timeframe   string                `json:"timeframe"`
	DataPoints  []WorkloadMetricPoint `json:"data_points"`
}

// WorkloadNodeInfo describes a node running a workload's pods.
type WorkloadNodeInfo struct {
	NodeUID          string  `json:"node_uid"`
	NodeName         string  `json:"node_name"`
	PodCount         int     `json:"pod_count"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsageBytes float64 `json:"memory_usage_bytes"`
}

// WorkloadNodesResponse is the node distribution for a workload.
type WorkloadNodesResponse struct {
	WorkloadUID string             `json:"workload_uid"`
	ClusterID   string             `json:"cluster_id"`
	Nodes       []WorkloadNodeInfo `json:"nodes"`
}

// --- Namespace Details ---

// NamespaceWorkload is a workload within a namespace detail view.
type NamespaceWorkload struct {
	ID                string  `json:"id"`
	WorkloadName      string  `json:"workload_name"`
	WorkloadKind      string  `json:"workload_kind"`
	Replicas          int     `json:"replicas"`
	AvailableReplicas int     `json:"available_replicas"`
	CPURequest        float64 `json:"cpu_request"`
	MemoryRequestGB   float64 `json:"memory_request_gb"`
	HourlyCost        float64 `json:"hourly_cost"`
}

// NamespaceDetailResponse is detailed namespace information.
type NamespaceDetailResponse struct {
	Name            string              `json:"name"`
	ClusterID       string              `json:"cluster_id"`
	ClusterName     *string             `json:"cluster_name"`
	Team            *string             `json:"team"`
	Department      *string             `json:"department"`
	PodCount        int                 `json:"pod_count"`
	WorkloadCount   int                 `json:"workload_count"`
	TotalCPUCores   float64             `json:"total_cpu_cores"`
	TotalMemoryGB   float64             `json:"total_memory_gb"`
	HourlyCost      float64             `json:"hourly_cost"`
	EfficiencyScore *float64            `json:"efficiency_score"`
	MonthlyCost     *float64            `json:"monthly_cost"`
	ContainerCount  *int                `json:"container_count"`
	Workloads       []NamespaceWorkload `json:"workloads"`
}

// NamespaceTrendPoint is a single time-series point for namespace trends.
type NamespaceTrendPoint struct {
	Timestamp        string  `json:"timestamp"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsageBytes float64 `json:"memory_usage_bytes"`
	HourlyCost       float64 `json:"hourly_cost"`
}

// NamespaceTrendsResponse is time-series trends for a namespace.
type NamespaceTrendsResponse struct {
	NamespaceName string                `json:"namespace_name"`
	ClusterID     string                `json:"cluster_id"`
	Timeframe     string                `json:"timeframe"`
	DataPoints    []NamespaceTrendPoint `json:"data_points"`
}

// --- Node Group Details ---

// NodeGroupNodeInfo is a node within a node group detail view.
type NodeGroupNodeInfo struct {
	ID                  string  `json:"id"`
	NodeName            string  `json:"node_name"`
	InstanceType        *string `json:"instance_type"`
	AvailabilityZone    *string `json:"availability_zone"`
	IsReady             bool    `json:"is_ready"`
	CPUCapacity         float64 `json:"cpu_capacity"`
	CPUAllocatable      float64 `json:"cpu_allocatable"`
	MemoryCapacityGB    float64 `json:"memory_capacity_gb"`
	MemoryAllocatableGB float64 `json:"memory_allocatable_gb"`
	PodsCapacity        int     `json:"pods_capacity"`
	HourlyCost          float64 `json:"hourly_cost"`
	SpotInstance        bool    `json:"spot_instance"`
}

// NodeGroupDetailResponse is detailed node group information.
type NodeGroupDetailResponse struct {
	Name                 string              `json:"name"`
	ClusterID            string              `json:"cluster_id"`
	ClusterName          *string             `json:"cluster_name"`
	InstanceType         *string             `json:"instance_type"`
	NodeCount            int                 `json:"node_count"`
	TotalCPUCores        float64             `json:"total_cpu_cores"`
	TotalMemoryGB        float64             `json:"total_memory_gb"`
	SpotPercentage       float64             `json:"spot_percentage"`
	HourlyCost           float64             `json:"hourly_cost"`
	AvgCPUUtilization    float64             `json:"avg_cpu_utilization"`
	AvgMemoryUtilization float64             `json:"avg_memory_utilization"`
	Nodes                []NodeGroupNodeInfo `json:"nodes"`
}
