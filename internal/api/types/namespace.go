package types

// Namespace is the /v1 Namespace resource. ID is the Kubernetes namespace name.
// Unlike Cluster, its cost block accepts cost_mode. Capacity is omitted entirely
// unless a ResourceQuota is set.
type Namespace struct {
	ID          string               `json:"id"`
	Kind        string               `json:"kind"`
	Metadata    NamespaceMetadata    `json:"metadata"`
	Capacity    *NamespaceCapacity   `json:"capacity,omitempty"`
	Utilization NamespaceUtilization `json:"utilization"`
	Cost        NamespaceCost        `json:"cost"`

	// Populated only on the detail endpoint, ordered by
	// current_run_rate_hourly DESC. Omitted on list endpoints.
	WorkloadsTop5 []NamespaceTopWorkload `json:"workloads_top_5,omitempty"`
}

type NamespaceTopWorkload struct {
	ID   string                   `json:"id"`
	Kind string                   `json:"kind"`
	Name string                   `json:"name"`
	Cost NamespaceTopWorkloadCost `json:"cost"`
}

type NamespaceTopWorkloadCost struct {
	CurrentRunRateHourly Money `json:"current_run_rate_hourly"`
}

type NamespaceMetadata struct {
	Name         string            `json:"name"`
	Cluster      NestedRef         `json:"cluster"`
	UIDK8s       string            `json:"uid_k8s,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
	Team         *NestedRef        `json:"team,omitempty"`
	Department   *NestedRef        `json:"department,omitempty"`
	CreatedAtK8s string            `json:"created_at_k8s,omitempty"`
	LastSeenAt   string            `json:"last_seen_at,omitempty"`
}

type NamespaceCapacity struct {
	CPU    NamespaceCapacityCPU    `json:"cpu,omitempty"`
	Memory NamespaceCapacityMemory `json:"memory,omitempty"`
}

type NamespaceCapacityCPU struct {
	QuotaCores float64 `json:"quota_cores"`
}

type NamespaceCapacityMemory struct {
	QuotaBytes int64 `json:"quota_bytes"`
}

type NamespaceUtilization struct {
	CPU    UtilizationCPU    `json:"cpu"`
	Memory UtilizationMemory `json:"memory"`
	Counts NamespaceCounts   `json:"counts"`
}

type NamespaceCounts struct {
	Workloads         int `json:"workloads"`
	Deployments       int `json:"deployments"`
	StatefulSets      int `json:"statefulsets"`
	DaemonSets        int `json:"daemonsets"`
	Jobs              int `json:"jobs"`
	CronJobs          int `json:"cronjobs"`
	Pods              int `json:"pods"`
	RunningPods       int `json:"running_pods"`
	Containers        int `json:"containers"`
	PersistentVolumes int `json:"persistent_volumes"`
}

// NamespaceCost includes CostMode: it sums workload costs, so it varies by mode.
type NamespaceCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	CostMode             string `json:"cost_mode,omitempty"`
	LastUpdatedAt        string `json:"last_updated_at,omitempty"`
}
