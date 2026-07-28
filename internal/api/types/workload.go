package types

// Workload covers Deployment, StatefulSet, DaemonSet, Job and CronJob. ID is the
// Kubernetes metadata.uid. Its cost block accepts cost_mode.
type Workload struct {
	ID          string              `json:"id"`
	Kind        string              `json:"kind"`
	Metadata    WorkloadMetadata    `json:"metadata"`
	Capacity    WorkloadCapacity    `json:"capacity"`
	Utilization WorkloadUtilization `json:"utilization"`
	Cost        WorkloadCost        `json:"cost"`
}

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

// WorkloadCapacity aggregates container limits across the whole pod template.
type WorkloadCapacity struct {
	CPU    WorkloadCapacityCPU    `json:"cpu"`
	Memory WorkloadCapacityMemory `json:"memory"`
}

type WorkloadCapacityCPU struct {
	LimitCores float64 `json:"limit_cores"`
}

type WorkloadCapacityMemory struct {
	LimitBytes int64 `json:"limit_bytes"`
}

type WorkloadUtilization struct {
	CPU      UtilizationCPU    `json:"cpu"`
	Memory   UtilizationMemory `json:"memory"`
	Replicas WorkloadReplicas  `json:"replicas"`
	Counts   WorkloadCounts    `json:"counts"`
}

type WorkloadReplicas struct {
	Desired     int `json:"desired"`
	Available   int `json:"available"`
	Unavailable int `json:"unavailable"`
	Updated     int `json:"updated"`
}

type WorkloadCounts struct {
	Pods        int `json:"pods"`
	RunningPods int `json:"running_pods"`
	Containers  int `json:"containers"`
}

// WorkloadCost includes CostMode: it sums container costs, so it varies by mode.
type WorkloadCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	CostMode             string `json:"cost_mode,omitempty"`
	LastUpdatedAt        string `json:"last_updated_at,omitempty"`
}
