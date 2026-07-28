package types

// Cluster omits cost.cost_mode - a cluster has one physical bill, so mode does
// not apply.
type Cluster struct {
	ID          string             `json:"id"`
	Kind        string             `json:"kind"`
	Metadata    ClusterMetadata    `json:"metadata"`
	Capacity    ClusterCapacity    `json:"capacity"`
	Utilization ClusterUtilization `json:"utilization"`
	Cost        ClusterCost        `json:"cost"`
}

// IsStale is derived from LastSeenAt against the server-side heartbeat
// threshold.
type ClusterMetadata struct {
	Name              string   `json:"name"`
	Provider          string   `json:"provider"`
	Service           string   `json:"service,omitempty"`
	Region            string   `json:"region,omitempty"`
	AvailabilityZones []string `json:"availability_zones,omitempty"`
	Environment       string   `json:"environment"`
	Status            string   `json:"status"`
	IsStale           bool     `json:"is_stale"`
	K8sVersion        string   `json:"k8s_version,omitempty"`
	AgentVersion      string   `json:"agent_version,omitempty"`
	DiscoverySource   string   `json:"discovery_source,omitempty"`
	CreatedAt         string   `json:"created_at,omitempty"`
	LastSeenAt        string   `json:"last_seen_at,omitempty"`
}

type ClusterCapacity struct {
	CPU     CapacityCPU     `json:"cpu"`
	Memory  CapacityMemory  `json:"memory"`
	GPU     CapacityGPU     `json:"gpu"`
	Storage CapacityStorage `json:"storage"`
	Pods    CapacityPods    `json:"pods"`
}

type ClusterUtilization struct {
	CPU    UtilizationCPU    `json:"cpu"`
	Memory UtilizationMemory `json:"memory"`
	GPU    UtilizationGPU    `json:"gpu"`
	Counts ClusterCounts     `json:"counts"`
}

type ClusterCounts struct {
	Nodes             int `json:"nodes"`
	Namespaces        int `json:"namespaces"`
	Workloads         int `json:"workloads"`
	Deployments       int `json:"deployments"`
	StatefulSets      int `json:"statefulsets"`
	DaemonSets        int `json:"daemonsets"`
	Jobs              int `json:"jobs"`
	CronJobs          int `json:"cronjobs"`
	Pods              int `json:"pods"`
	RunningPods       int `json:"running_pods"`
	Containers        int `json:"containers"`
	RunningContainers int `json:"running_containers"`
	PersistentVolumes int `json:"persistent_volumes"`
}

type ClusterCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	LastUpdatedAt        string `json:"last_updated_at,omitempty"`
}
