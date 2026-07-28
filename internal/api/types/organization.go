package types

// Organization is the tenant-level snapshot. Its cost block rejects ?cost_mode=
// and omits cost.cost_mode - the org bill is one physical number; mode applies
// only to per-cluster/per-team rollups.
type Organization struct {
	ID          string                  `json:"id"`
	Kind        string                  `json:"kind"`
	Metadata    OrganizationMetadata    `json:"metadata"`
	Capacity    OrganizationCapacity    `json:"capacity"`
	Utilization OrganizationUtilization `json:"utilization"`
	Cost        OrganizationCost        `json:"cost"`
}

type OrganizationMetadata struct {
	Name      string `json:"name"`
	Domain    string `json:"domain,omitempty"`
	PlanType  string `json:"plan_type,omitempty"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at,omitempty"`
}

// OrganizationCapacity sums across every connected cluster.
type OrganizationCapacity struct {
	CPU     CapacityCPU     `json:"cpu"`
	Memory  CapacityMemory  `json:"memory"`
	GPU     CapacityGPU     `json:"gpu"`
	Storage CapacityStorage `json:"storage"`
}

type OrganizationUtilization struct {
	CPU    UtilizationCPU     `json:"cpu"`
	Memory UtilizationMemory  `json:"memory"`
	Counts OrganizationCounts `json:"counts"`
}

// ConnectedClusters counts only clusters whose last heartbeat is within the
// server-side staleness threshold.
type OrganizationCounts struct {
	Clusters          int `json:"clusters"`
	ConnectedClusters int `json:"connected_clusters"`
	Nodes             int `json:"nodes"`
	Namespaces        int `json:"namespaces"`
	Workloads         int `json:"workloads"`
	Pods              int `json:"pods"`
	Containers        int `json:"containers"`
	PersistentVolumes int `json:"persistent_volumes"`
	Recommendations   int `json:"recommendations"`
}

type OrganizationCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	LastUpdatedAt        string `json:"last_updated_at,omitempty"`
}

// OrganizationDashboard is the only endpoint exposing MTD and savings figures;
// every other resource endpoint exposes only current_run_rate_hourly.
type OrganizationDashboard struct {
	OrganizationID string                      `json:"organization_id"`
	Snapshot       Organization                `json:"snapshot"`
	MonthToDate    OrgDashboardMTD             `json:"month_to_date"`
	Savings        OrgDashboardSavings         `json:"savings"`
	TopClusters    []OrgDashboardClusterRollup `json:"top_clusters"`
}

type OrgDashboardMTD struct {
	BilledCost Money                `json:"billed_cost"`
	Calendar   OrgDashboardCalendar `json:"calendar"`
}

type OrgDashboardCalendar struct {
	Month        string `json:"month"`
	DaysElapsed  int    `json:"days_elapsed"`
	DaysInMonth  int    `json:"days_in_month"`
	MonthStartAt string `json:"month_start_at"`
}

type OrgDashboardSavings struct {
	CurrentHourlyPotential Money `json:"current_hourly_potential"`
	RecommendationCount    int   `json:"recommendation_count"`
}

type OrgDashboardClusterRollup struct {
	Cluster              NestedRef `json:"cluster"`
	CurrentRunRateHourly Money     `json:"current_run_rate_hourly"`
	MonthToDateCost      Money     `json:"month_to_date_cost"`
}
