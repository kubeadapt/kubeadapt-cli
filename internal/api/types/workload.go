package types

// Workload is the /v1 Workload resource covering all 5 kinds:
// Deployment, StatefulSet, DaemonSet, Job, CronJob. ID is the Kubernetes
// metadata.uid. Cost block ACCEPTS cost_mode and echoes the applied mode
// in cost.cost_mode.
type Workload struct {
	ID          string              `json:"id"`
	Kind        string              `json:"kind"`
	Metadata    WorkloadMetadata    `json:"metadata"`
	Capacity    WorkloadCapacity    `json:"capacity"`
	Utilization WorkloadUtilization `json:"utilization"`
	Cost        WorkloadCost        `json:"cost"`
}

// WorkloadMetadata is the identity / status sub-block for a Workload.
type WorkloadMetadata struct {
	Name               string            `json:"name"`
	WorkloadKind       string            `json:"workload_kind"`
	Namespace          string            `json:"namespace"`
	Cluster            NestedRef         `json:"cluster"`
	Labels             map[string]string `json:"labels,omitempty"`
	ServiceAccountName string            `json:"service_account_name,omitempty"`
	Status             string            `json:"status,omitempty"`
	StatusReason       string            `json:"status_reason,omitempty"`
	IsSuspended        bool              `json:"is_suspended"`
	IsPaused           bool              `json:"is_paused"`
	HasHPA             bool              `json:"has_hpa"`
	CreatedAtK8s       string            `json:"created_at_k8s,omitempty"`
	LastSeenAt         string            `json:"last_seen_at,omitempty"`
}

// WorkloadCapacity is the capacity block for a Workload — the aggregate
// of container limits across all containers in the pod template.
type WorkloadCapacity struct {
	CPU    WorkloadCapacityCPU    `json:"cpu"`
	Memory WorkloadCapacityMemory `json:"memory"`
}

// WorkloadCapacityCPU is the CPU limit roll-up for a Workload.
type WorkloadCapacityCPU struct {
	LimitCores float64 `json:"limit_cores"`
}

// WorkloadCapacityMemory is the memory limit roll-up for a Workload.
type WorkloadCapacityMemory struct {
	LimitBytes int64 `json:"limit_bytes"`
}

// WorkloadUtilization is the utilization block for a Workload.
type WorkloadUtilization struct {
	CPU      UtilizationCPU    `json:"cpu"`
	Memory   UtilizationMemory `json:"memory"`
	Replicas WorkloadReplicas  `json:"replicas"`
	Counts   WorkloadCounts    `json:"counts"`
}

// WorkloadReplicas is the replica-state breakdown for a Workload.
type WorkloadReplicas struct {
	Desired     int `json:"desired"`
	Available   int `json:"available"`
	Unavailable int `json:"unavailable"`
	Updated     int `json:"updated"`
}

// WorkloadCounts is the live-count sub-block for a Workload.
type WorkloadCounts struct {
	Pods        int `json:"pods"`
	RunningPods int `json:"running_pods"`
	Containers  int `json:"containers"`
}

// WorkloadCost is the cost block for a Workload. INCLUDES CostMode —
// workload cost is a sum-of-container-costs which DOES vary by mode.
type WorkloadCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	CostMode             string `json:"cost_mode,omitempty"`
	LastUpdatedAt        string `json:"last_updated_at,omitempty"`
}
