package types

// Namespace is the /v1 Namespace resource. Unlike Cluster, the namespace
// cost block ACCEPTS cost_mode and echoes the applied mode in cost.cost_mode.
// ID is the Kubernetes namespace name (stable within a cluster).
//
// Capacity is OMITTED entirely when no ResourceQuota is set on the
// namespace; emitted with quota_cores / quota_bytes when a quota is
// present.
type Namespace struct {
	ID          string               `json:"id"`
	Kind        string               `json:"kind"`
	Metadata    NamespaceMetadata    `json:"metadata"`
	Capacity    *NamespaceCapacity   `json:"capacity,omitempty"`
	Utilization NamespaceUtilization `json:"utilization"`
	Cost        NamespaceCost        `json:"cost"`

	// WorkloadsTop5 is populated only on the detail endpoint
	// (GET /v1/clusters/{cid}/namespaces/{ns}) — the top 5 workloads in
	// the namespace by current_run_rate_hourly DESC. Omitted on list
	// endpoints to keep list payloads small.
	WorkloadsTop5 []NamespaceTopWorkload `json:"workloads_top_5,omitempty"`
}

// NamespaceTopWorkload is the trimmed workload reference embedded under
// Namespace.WorkloadsTop5 on the detail endpoint — just enough to power
// the "namespace landing page" UX without fanning out to a per-workload
// request.
type NamespaceTopWorkload struct {
	ID   string                   `json:"id"`
	Kind string                   `json:"kind"`
	Name string                   `json:"name"`
	Cost NamespaceTopWorkloadCost `json:"cost"`
}

// NamespaceTopWorkloadCost is the trimmed cost block on top-workload
// references. Only the run rate is exposed — full cost metadata lives
// on the workload's own endpoint.
type NamespaceTopWorkloadCost struct {
	CurrentRunRateHourly Money `json:"current_run_rate_hourly"`
}

// NamespaceMetadata is the identity / governance sub-block for a Namespace.
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

// NamespaceCapacity is the capacity block for a Namespace — populated
// only when a ResourceQuota is set on the namespace.
type NamespaceCapacity struct {
	CPU    NamespaceCapacityCPU    `json:"cpu,omitempty"`
	Memory NamespaceCapacityMemory `json:"memory,omitempty"`
}

// NamespaceCapacityCPU is the namespace-quota CPU sub-block.
type NamespaceCapacityCPU struct {
	QuotaCores float64 `json:"quota_cores"`
}

// NamespaceCapacityMemory is the namespace-quota memory sub-block.
type NamespaceCapacityMemory struct {
	QuotaBytes int64 `json:"quota_bytes"`
}

// NamespaceUtilization is the utilization block for a Namespace.
type NamespaceUtilization struct {
	CPU    UtilizationCPU    `json:"cpu"`
	Memory UtilizationMemory `json:"memory"`
	Counts NamespaceCounts   `json:"counts"`
}

// NamespaceCounts is the live-count sub-block for a Namespace.
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

// NamespaceCost is the cost block for a Namespace. INCLUDES CostMode —
// namespace cost is a sum-of-workload-costs which DOES vary by mode.
type NamespaceCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	CostMode             string `json:"cost_mode,omitempty"`
	LastUpdatedAt        string `json:"last_updated_at,omitempty"`
}
