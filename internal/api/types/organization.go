package types

// Organization is the /v1 Organization resource (GET /v1/organization) —
// the tenant-level snapshot. The org-level cost block REJECTS ?cost_mode=
// and OMITS cost.cost_mode — the org bill is one physical number; mode
// applies only to per-cluster/per-team rollups.
type Organization struct {
	ID          string                  `json:"id"`
	Kind        string                  `json:"kind"`
	Metadata    OrganizationMetadata    `json:"metadata"`
	Capacity    OrganizationCapacity    `json:"capacity"`
	Utilization OrganizationUtilization `json:"utilization"`
	Cost        OrganizationCost        `json:"cost"`
}

// OrganizationMetadata is the identity sub-block for an Organization.
type OrganizationMetadata struct {
	Name      string `json:"name"`
	Domain    string `json:"domain,omitempty"`
	PlanType  string `json:"plan_type,omitempty"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at,omitempty"`
}

// OrganizationCapacity is the capacity block for an Organization — the
// sum across every connected cluster.
type OrganizationCapacity struct {
	CPU     CapacityCPU     `json:"cpu"`
	Memory  CapacityMemory  `json:"memory"`
	GPU     CapacityGPU     `json:"gpu"`
	Storage CapacityStorage `json:"storage"`
}

// OrganizationUtilization is the utilization block for an Organization.
type OrganizationUtilization struct {
	CPU    UtilizationCPU     `json:"cpu"`
	Memory UtilizationMemory  `json:"memory"`
	Counts OrganizationCounts `json:"counts"`
}

// OrganizationCounts is the org-level live-count block. Adds
// connected_clusters (subset of clusters whose last heartbeat is within
// the server-side staleness threshold) and recommendations on top of the
// cluster aggregate counts.
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

// OrganizationCost is the org-level cost block. OMITS CostMode — the org
// bill is one physical number, mode-invariant.
type OrganizationCost struct {
	CurrentRunRateHourly Money  `json:"current_run_rate_hourly"`
	LastUpdatedAt        string `json:"last_updated_at,omitempty"`
}

// OrganizationDashboard is the /v1/organization/dashboard body. THE ONLY
// ENDPOINT where MTD and savings figures live — every other resource
// endpoint exposes only current_run_rate_hourly.
type OrganizationDashboard struct {
	OrganizationID string                      `json:"organization_id"`
	Snapshot       Organization                `json:"snapshot"`
	MonthToDate    OrgDashboardMTD             `json:"month_to_date"`
	Savings        OrgDashboardSavings         `json:"savings"`
	TopClusters    []OrgDashboardClusterRollup `json:"top_clusters"`
}

// OrgDashboardMTD is the month-to-date sub-block on the dashboard.
type OrgDashboardMTD struct {
	BilledCost Money                `json:"billed_cost"`
	Calendar   OrgDashboardCalendar `json:"calendar"`
}

// OrgDashboardCalendar is the calendar context for the month-to-date
// figures — useful for explaining "we're 12 days into a 30-day month" UX.
type OrgDashboardCalendar struct {
	Month        string `json:"month"`
	DaysElapsed  int    `json:"days_elapsed"`
	DaysInMonth  int    `json:"days_in_month"`
	MonthStartAt string `json:"month_start_at"`
}

// OrgDashboardSavings is the potential-savings sub-block on the dashboard.
type OrgDashboardSavings struct {
	CurrentHourlyPotential Money `json:"current_hourly_potential"`
	RecommendationCount    int   `json:"recommendation_count"`
}

// OrgDashboardClusterRollup is the per-cluster slice of the dashboard
// top-N list.
type OrgDashboardClusterRollup struct {
	Cluster              NestedRef `json:"cluster"`
	CurrentRunRateHourly Money     `json:"current_run_rate_hourly"`
	MonthToDateCost      Money     `json:"month_to_date_cost"`
}
