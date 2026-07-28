package types

// Pod accepts cost_mode like Workload. ID is the Kubernetes metadata.uid.
type Pod struct {
	ID          string         `json:"id"`
	Kind        string         `json:"kind"`
	Metadata    PodMetadata    `json:"metadata"`
	Capacity    PodCapacity    `json:"capacity"`
	Utilization PodUtilization `json:"utilization"`
	Cost        PodCost        `json:"cost"`
}

type PodMetadata struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	Cluster       NestedRef         `json:"cluster"`
	Workload      *PodWorkloadRef   `json:"workload,omitempty"`
	Node          *NestedRef        `json:"node,omitempty"`
	Phase         string            `json:"phase,omitempty"`
	Reason        string            `json:"reason,omitempty"`
	QOSClass      string            `json:"qos_class,omitempty"`
	PodIP         string            `json:"pod_ip,omitempty"`
	HostIP        string            `json:"host_ip,omitempty"`
	HostNetwork   bool              `json:"host_network"`
	PriorityClass string            `json:"priority_class,omitempty"`
	HasHostPath   bool              `json:"has_hostpath"`
	HasEmptyDir   bool              `json:"has_emptydir"`
	Labels        map[string]string `json:"labels,omitempty"`
	CreatedAtK8s  string            `json:"created_at_k8s,omitempty"`
	LastSeenAt    string            `json:"last_seen_at,omitempty"`
}

// PodWorkloadRef deviates from the NestedRef {id, name} shape and uses uid,
// because workloads are uid-keyed.
type PodWorkloadRef struct {
	UID  string `json:"uid"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type PodCapacity struct {
	CPU    WorkloadCapacityCPU    `json:"cpu"`
	Memory WorkloadCapacityMemory `json:"memory"`
}

type PodUtilization struct {
	CPU           UtilizationCPU    `json:"cpu"`
	Memory        UtilizationMemory `json:"memory"`
	Counts        PodCounts         `json:"counts"`
	RestartsTotal int64             `json:"restarts_total"`
	OOMKillsTotal int64             `json:"oom_kills_total"`
}

type PodCounts struct {
	Containers        int `json:"containers"`
	RunningContainers int `json:"running_containers"`
	ReadyContainers   int `json:"ready_containers"`
}

// PodCost includes CostMode: it sums container costs, so it varies by mode.
type PodCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	CostMode             string `json:"cost_mode,omitempty"`
	LastUpdatedAt        string `json:"last_updated_at,omitempty"`
}
