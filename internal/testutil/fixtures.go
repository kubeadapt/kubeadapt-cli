package testutil

import "github.com/kubeadapt/kubeadapt-cli/internal/api/types"

// Deterministic UUIDs and timestamps used across fixtures. Tests can match on
// these literal values; do NOT swap them for time.Now() or uuid.New() — every
// fixture must produce byte-identical output across runs.
const (
	clusterIDProd    = "00000000-0000-0000-0000-000000000001"
	clusterIDStaging = "00000000-0000-0000-0000-000000000002"
	clusterIDDev     = "00000000-0000-0000-0000-000000000003"

	workloadUIDAPI       = "11111111-1111-1111-1111-000000000001"
	workloadUIDPrometheus = "11111111-1111-1111-1111-000000000002"
	workloadUIDETL        = "11111111-1111-1111-1111-000000000003"

	podUIDOne   = "22222222-2222-2222-2222-000000000001"
	podUIDTwo   = "22222222-2222-2222-2222-000000000002"
	podUIDThree = "22222222-2222-2222-2222-000000000003"
	podUIDFour  = "22222222-2222-2222-2222-000000000004"
	podUIDFive  = "22222222-2222-2222-2222-000000000005"

	nodeUIDOne   = "33333333-3333-3333-3333-000000000001"
	nodeUIDTwo   = "33333333-3333-3333-3333-000000000002"
	nodeUIDThree = "33333333-3333-3333-3333-000000000003"

	recIDOne   = "44444444-4444-4444-4444-000000000001"
	recIDTwo   = "44444444-4444-4444-4444-000000000002"
	recIDThree = "44444444-4444-4444-4444-000000000003"

	teamIDPlatform = "55555555-5555-5555-5555-000000000001"
	teamIDSRE      = "55555555-5555-5555-5555-000000000002"

	deptIDEngineering = "66666666-6666-6666-6666-000000000001"
	deptIDData        = "66666666-6666-6666-6666-000000000002"

	orgID = "77777777-7777-7777-7777-000000000001"

	assignmentIDOne   = "88888888-8888-8888-8888-000000000001"
	assignmentIDTwo   = "88888888-8888-8888-8888-000000000002"
	assignmentIDThree = "88888888-8888-8888-8888-000000000003"

	tsCreated  = "2024-06-01T10:00:00Z"
	tsLastSeen = "2025-05-20T14:30:00Z"
	tsUpdated  = "2025-05-19T09:15:00Z"

	currencyUSD = "USD"

	nameProdCluster    = "prod-cluster"
	nameAPIGateway     = "api-gateway"
	nameMonitoringNS   = "monitoring"
	nameDataNS         = "data"
	nameDefaultNS      = "default"
	nameGeneralPurpose = "general-purpose"
	nameComputeOpt     = "compute-optimized"
	nameMxlarge        = "m5.xlarge"
	nameZoneEast1A     = "us-east-1a"
	nameTeamPlatform   = "platform"
	nameTeamSRE        = "sre"
	nameDeptEng        = "engineering"
	kindDeployment     = "Deployment"
	costFullyLoaded    = "fully_loaded"
	originManual       = "manual"
	labelTeam          = "team"
	nameZoneEast1B     = "us-east-1b"
)

func money(amount string) types.Money {
	return types.Money{Amount: amount, Currency: currencyUSD}
}

func stringPtr(s string) *string { return &s }

// SampleCluster returns a single deterministic Cluster fixture (the production
// cluster). Use SampleClusters() for the list endpoint.
func SampleCluster() types.Cluster {
	return types.Cluster{
		ID:   clusterIDProd,
		Kind: "cluster",
		Metadata: types.ClusterMetadata{
			Name:              nameProdCluster,
			Provider:          "aws",
			Service:           "eks",
			Region:            "us-east-1",
			AvailabilityZones: []string{nameZoneEast1A, nameZoneEast1B},
			Environment:       "production",
			Status:            "active",
			IsStale:           false,
			K8sVersion:        "1.30.0",
			AgentVersion:      "0.5.0",
			DiscoverySource:   "agent",
			CreatedAt:         tsCreated,
			LastSeenAt:        tsLastSeen,
		},
		Capacity: types.ClusterCapacity{
			CPU:     types.CapacityCPU{TotalCores: 256.0, AllocatableCores: 248.0},
			Memory:  types.CapacityMemory{TotalBytes: 1099511627776, AllocatableBytes: 1090519040000},
			GPU:     types.CapacityGPU{Total: 8, Allocatable: 8, Model: "nvidia-a100"},
			Storage: types.CapacityStorage{TotalBytes: 21474836480000},
			Pods:    types.CapacityPods{Allocatable: 660},
		},
		Utilization: types.ClusterUtilization{
			CPU:    types.UtilizationCPU{RequestedCores: 160.0, UsedCores: 128.0, UtilizationPercent: 50.0},
			Memory: types.UtilizationMemory{RequestedBytes: 700000000000, UsedBytes: 549755813888, UtilizationPercent: 50.0},
			GPU:    types.UtilizationGPU{UtilizationPercent: 65.0, MemoryUsedBytes: 60000000000, MemoryTotalBytes: 80000000000},
			Counts: types.ClusterCounts{
				Nodes: 12, Namespaces: 8, Workloads: 45, Deployments: 30,
				StatefulSets: 5, DaemonSets: 8, Jobs: 2, CronJobs: 0,
				Pods: 180, RunningPods: 175, Containers: 220, RunningContainers: 215,
				PersistentVolumes: 24,
			},
		},
		Cost: types.ClusterCost{
			CurrentRunRateHourly: money("12.4700"),
			LastUpdatedAt:        tsLastSeen,
		},
	}
}

// SampleClusters returns a deterministic slice of three Cluster fixtures
// (production, staging, dev) used by the list endpoint and pagination tests.
func SampleClusters() []types.Cluster {
	prod := SampleCluster()

	staging := prod
	staging.ID = clusterIDStaging
	staging.Metadata.Name = "staging-cluster"
	staging.Metadata.Environment = "staging"
	staging.Cost.CurrentRunRateHourly = money("4.2100")

	dev := prod
	dev.ID = clusterIDDev
	dev.Metadata.Name = "dev-cluster"
	dev.Metadata.Environment = "development"
	dev.Cost.CurrentRunRateHourly = money("1.1500")

	return []types.Cluster{prod, staging, dev}
}

// SampleWorkload returns a single Workload fixture (the api-gateway deployment
// in the production cluster).
func SampleWorkload() types.Workload {
	return types.Workload{
		ID:   workloadUIDAPI,
		Kind: "workload",
		Metadata: types.WorkloadMetadata{
			Name:               nameAPIGateway,
			WorkloadKind:       kindDeployment,
			Namespace:          nameDefaultNS,
			Cluster:            types.NestedRef{ID: clusterIDProd, Name: nameProdCluster},
			Labels:             map[string]string{"app": nameAPIGateway, "tier": "edge"},
			ServiceAccountName: nameAPIGateway,
			Status:             "available",
			IsSuspended:        false,
			IsPaused:           false,
			HasHPA:             true,
			CreatedAtK8s:       tsCreated,
			LastSeenAt:         tsLastSeen,
		},
		Capacity: types.WorkloadCapacity{
			CPU:    types.WorkloadCapacityCPU{LimitCores: 3.0},
			Memory: types.WorkloadCapacityMemory{LimitBytes: 3221225472},
		},
		Utilization: types.WorkloadUtilization{
			CPU:      types.UtilizationCPU{RequestedCores: 1.5, UsedCores: 0.9, UtilizationPercent: 60.0},
			Memory:   types.UtilizationMemory{RequestedBytes: 1610612736, UsedBytes: 1073741824, UtilizationPercent: 66.7},
			Replicas: types.WorkloadReplicas{Desired: 3, Available: 3, Unavailable: 0, Updated: 3},
			Counts:   types.WorkloadCounts{Pods: 3, RunningPods: 3, Containers: 6},
		},
		Cost: types.WorkloadCost{
			CurrentRunRateHourly: money("0.4800"),
			CostMode:             costFullyLoaded,
			LastUpdatedAt:        tsLastSeen,
		},
	}
}

// SampleWorkloads returns three Workload fixtures across multiple namespaces
// and workload kinds.
func SampleWorkloads() []types.Workload {
	apiGateway := SampleWorkload()

	prometheus := apiGateway
	prometheus.ID = workloadUIDPrometheus
	prometheus.Metadata.Name = "prometheus"
	prometheus.Metadata.WorkloadKind = "StatefulSet"
	prometheus.Metadata.Namespace = nameMonitoringNS
	prometheus.Metadata.HasHPA = false
	prometheus.Cost.CurrentRunRateHourly = money("0.2200")

	etl := apiGateway
	etl.ID = workloadUIDETL
	etl.Metadata.Name = "nightly-etl"
	etl.Metadata.WorkloadKind = "CronJob"
	etl.Metadata.Namespace = "data"
	etl.Metadata.HasHPA = false
	etl.Cost.CurrentRunRateHourly = money("0.0900")

	return []types.Workload{apiGateway, prometheus, etl}
}

// SamplePod returns a single Pod fixture from the api-gateway workload.
func SamplePod() types.Pod {
	return types.Pod{
		ID:   podUIDOne,
		Kind: "pod",
		Metadata: types.PodMetadata{
			Name:      "api-gateway-7f9b8c8d6f-aaaa1",
			Namespace: nameDefaultNS,
			Cluster:   types.NestedRef{ID: clusterIDProd, Name: nameProdCluster},
			Workload: &types.PodWorkloadRef{
				UID:  workloadUIDAPI,
				Kind: kindDeployment,
				Name: nameAPIGateway,
			},
			Node:          &types.NestedRef{ID: nodeUIDOne, Name: "ip-10-0-1-42.ec2.internal"},
			Phase:         "Running",
			QOSClass:      "Burstable",
			PodIP:         "10.244.1.10",
			HostIP:        "10.0.1.42",
			HostNetwork:   false,
			PriorityClass: "default-priority",
			HasHostPath:   false,
			HasEmptyDir:   true,
			Labels:        map[string]string{"app": nameAPIGateway},
			CreatedAtK8s:  tsCreated,
			LastSeenAt:    tsLastSeen,
		},
		Capacity: types.PodCapacity{
			CPU:    types.WorkloadCapacityCPU{LimitCores: 1.0},
			Memory: types.WorkloadCapacityMemory{LimitBytes: 1073741824},
		},
		Utilization: types.PodUtilization{
			CPU:           types.UtilizationCPU{RequestedCores: 0.5, UsedCores: 0.3, UtilizationPercent: 60.0},
			Memory:        types.UtilizationMemory{RequestedBytes: 536870912, UsedBytes: 357913941, UtilizationPercent: 66.7},
			Counts:        types.PodCounts{Containers: 2, RunningContainers: 2, ReadyContainers: 2},
			RestartsTotal: 0,
			OOMKillsTotal: 0,
		},
		Cost: types.PodCost{
			CurrentRunRateHourly: money("0.1600"),
			CostMode:             costFullyLoaded,
			LastUpdatedAt:        tsLastSeen,
		},
	}
}

// SamplePods returns five deterministic Pod fixtures with distinct UIDs.
func SamplePods() []types.Pod {
	base := SamplePod()
	uids := []string{podUIDOne, podUIDTwo, podUIDThree, podUIDFour, podUIDFive}
	out := make([]types.Pod, 0, len(uids))
	for i, uid := range uids {
		p := base
		p.ID = uid
		p.Metadata.Name = "api-gateway-7f9b8c8d6f-aaaa" + string(rune('1'+i))
		out = append(out, p)
	}
	return out
}

// SampleNode returns a single Node fixture (on-demand m5.xlarge in us-east-1a).
func SampleNode() types.Node {
	return types.Node{
		ID:   nodeUIDOne,
		Kind: "node",
		Metadata: types.NodeMetadata{
			Name:             "ip-10-0-1-42.ec2.internal",
			Cluster:          types.NestedRef{ID: clusterIDProd, Name: nameProdCluster},
			NodeRole:         "worker",
			InstanceType:     nameMxlarge,
			NodeGroup:        nameGeneralPurpose,
			AvailabilityZone: nameZoneEast1A,
			Region:           "us-east-1",
			IsSpot:           false,
			CapacityType:     "ON_DEMAND",
			Architecture:     "amd64",
			OperatingSystem:  "linux",
			KubeletVersion:   "1.30.0",
			ProviderID:       "aws:///us-east-1a/i-0123456789abcdef0",
			IsReady:          true,
			IsSchedulable:    true,
			Labels:           map[string]string{"node.kubernetes.io/instance-type": nameMxlarge},
			CreatedAtK8s:     tsCreated,
			LastSeenAt:       tsLastSeen,
		},
		Capacity: types.NodeCapacity{
			CPU:              types.CapacityCPU{TotalCores: 4.0, AllocatableCores: 3.92},
			Memory:           types.CapacityMemory{TotalBytes: 17179869184, AllocatableBytes: 16320000000},
			GPU:              types.CapacityGPU{Total: 0, Allocatable: 0},
			EphemeralStorage: types.CapacityStorage{TotalBytes: 107374182400},
			Pods:             types.CapacityPods{Allocatable: 58},
		},
		Utilization: types.NodeUtilization{
			CPU:    types.NodeUtilizationCPU{UsedCores: 2.0, UtilizationPercent: 50.0},
			Memory: types.NodeUtilizationMemory{UsedBytes: 8589934592, UtilizationPercent: 50.0},
			GPU:    types.UtilizationGPU{},
			Counts: types.NodeCounts{Pods: 12, RunningPods: 12},
		},
		Cost: types.NodeCost{
			CurrentRunRateHourly: money("0.1920"),
			PricingSource:        "aws-on-demand",
			LastUpdatedAt:        tsLastSeen,
		},
	}
}

// SampleNodes returns three Node fixtures: an on-demand m5.xlarge, a spot
// c5.2xlarge (with the on-demand baseline populated), and a second on-demand.
func SampleNodes() []types.Node {
	first := SampleNode()

	spot := first
	spot.ID = nodeUIDTwo
	spot.Metadata.Name = "ip-10-0-2-87.ec2.internal"
	spot.Metadata.InstanceType = "c5.2xlarge"
	spot.Metadata.NodeGroup = nameComputeOpt
	spot.Metadata.AvailabilityZone = nameZoneEast1B
	spot.Metadata.IsSpot = true
	spot.Metadata.CapacityType = "SPOT"
	onDemand := money("0.3400")
	spot.Cost = types.NodeCost{
		CurrentRunRateHourly:     money("0.1200"),
		OnDemandEquivalentHourly: &onDemand,
		PricingSource:            "aws-spot",
		LastUpdatedAt:            tsLastSeen,
	}

	third := first
	third.ID = nodeUIDThree
	third.Metadata.Name = "ip-10-0-3-12.ec2.internal"
	third.Metadata.AvailabilityZone = "us-east-1c"

	return []types.Node{first, spot, third}
}

// SampleNodeGroup returns a single NodeGroup fixture for the general-purpose
// pool. Includes the embedded member-node detail (Nodes slice) populated.
func SampleNodeGroup() types.NodeGroup {
	spotSavings := money("0.2200")
	return types.NodeGroup{
		ID:   nameGeneralPurpose,
		Kind: "node_group",
		Metadata: types.NodeGroupMetadata{
			Name:                   nameGeneralPurpose,
			Cluster:                types.NestedRef{ID: clusterIDProd, Name: nameProdCluster},
			InstanceTypes:          []string{nameMxlarge, "m5.2xlarge"},
			Zones:                  []string{nameZoneEast1A, nameZoneEast1B, "us-east-1c"},
			SpotCount:              1,
			OnDemandCount:          2,
			SpotPercentage:         33.3,
			OldestNodeCreatedAtK8s: tsCreated,
			Status:                 "healthy",
		},
		Capacity: types.NodeGroupCapacity{
			CPU:    types.CapacityCPU{TotalCores: 16.0, AllocatableCores: 15.7},
			Memory: types.CapacityMemory{TotalBytes: 51539607552, AllocatableBytes: 48960000000},
		},
		Utilization: types.NodeGroupUtilization{
			CPU:    types.NodeUtilizationCPU{UsedCores: 8.0, UtilizationPercent: 50.0},
			Memory: types.NodeUtilizationMemory{UsedBytes: 25769803776, UtilizationPercent: 50.0},
			Counts: types.NodeGroupCounts{Nodes: 3, ReadyNodes: 3, Pods: 36},
		},
		Cost: types.NodeGroupCost{
			CurrentRunRateHourly:        money("0.7040"),
			SpotSavingsVsOndemandHourly: &spotSavings,
			LastUpdatedAt:               tsLastSeen,
		},
		Nodes: SampleNodes(),
	}
}

// SampleNodeGroups returns two NodeGroup fixtures: general-purpose and
// compute-optimized.
func SampleNodeGroups() []types.NodeGroup {
	first := SampleNodeGroup()

	second := first
	second.ID = nameComputeOpt
	second.Metadata.Name = nameComputeOpt
	second.Metadata.InstanceTypes = []string{"c5.2xlarge", "c5.4xlarge"}
	second.Cost.CurrentRunRateHourly = money("1.3600")
	second.Nodes = nil

	return []types.NodeGroup{first, second}
}

// SampleNamespace returns a single Namespace fixture (default, in production
// cluster) with the top-5 workload embed populated.
func SampleNamespace() types.Namespace {
	quotaCPU := types.NamespaceCapacityCPU{QuotaCores: 16.0}
	quotaMem := types.NamespaceCapacityMemory{QuotaBytes: 34359738368}
	return types.Namespace{
		ID:   nameDefaultNS,
		Kind: "namespace",
		Metadata: types.NamespaceMetadata{
			Name:         nameDefaultNS,
			Cluster:      types.NestedRef{ID: clusterIDProd, Name: nameProdCluster},
			UIDK8s:       "ns-uid-default",
			Labels:       map[string]string{labelTeam: nameTeamPlatform},
			Team:         &types.NestedRef{ID: teamIDPlatform, Name: nameTeamPlatform},
			Department:   &types.NestedRef{ID: deptIDEngineering, Name: nameDeptEng},
			CreatedAtK8s: tsCreated,
			LastSeenAt:   tsLastSeen,
		},
		Capacity: &types.NamespaceCapacity{
			CPU:    quotaCPU,
			Memory: quotaMem,
		},
		Utilization: types.NamespaceUtilization{
			CPU:    types.UtilizationCPU{RequestedCores: 4.5, UsedCores: 3.0, UtilizationPercent: 66.7},
			Memory: types.UtilizationMemory{RequestedBytes: 8589934592, UsedBytes: 6442450944, UtilizationPercent: 75.0},
			Counts: types.NamespaceCounts{
				Workloads: 5, Deployments: 4, StatefulSets: 1, DaemonSets: 0,
				Jobs: 0, CronJobs: 0, Pods: 12, RunningPods: 12, Containers: 24,
				PersistentVolumes: 2,
			},
		},
		Cost: types.NamespaceCost{
			CurrentRunRateHourly: money("0.4500"),
			CostMode:             costFullyLoaded,
			LastUpdatedAt:        tsLastSeen,
		},
		WorkloadsTop5: []types.NamespaceTopWorkload{
			{
				ID:   workloadUIDAPI,
				Kind: kindDeployment,
				Name: nameAPIGateway,
				Cost: types.NamespaceTopWorkloadCost{CurrentRunRateHourly: money("0.4800")},
			},
		},
	}
}

// SampleNamespaces returns two Namespace fixtures (default and monitoring).
func SampleNamespaces() []types.Namespace {
	first := SampleNamespace()

	second := first
	second.ID = nameMonitoringNS
	second.Metadata.Name = nameMonitoringNS
	second.Metadata.UIDK8s = "ns-uid-monitoring"
	second.Metadata.Labels = map[string]string{labelTeam: nameTeamSRE}
	second.Metadata.Team = &types.NestedRef{ID: teamIDSRE, Name: nameTeamSRE}
	second.Cost.CurrentRunRateHourly = money("0.3200")
	second.WorkloadsTop5 = nil

	return []types.Namespace{first, second}
}

// SampleRecommendation returns a single Recommendation fixture — a CPU
// right-size proposal for the api-gateway workload.
func SampleRecommendation() types.Recommendation {
	return types.Recommendation{
		ID:   recIDOne,
		Kind: "recommendation",
		Metadata: types.RecommendationMetadata{
			RecommendationType: "workload_cpu_rightsize",
			ResourceType:       kindDeployment,
			ResourceName:       nameAPIGateway,
			ResourceUID:        workloadUIDAPI,
			Cluster:            types.NestedRef{ID: clusterIDProd, Name: nameProdCluster},
			Namespace:          nameDefaultNS,
			Title:              "Right-size api-gateway CPU request",
			Description:        "CPU request is 3.3x P95 usage; reduce from 500m to 150m.",
			Cause:              "over_provisioned_cpu_request",
			RiskLevel:          "low",
			Priority:           "medium",
			Status:             "open",
			DataPointsAnalyzed: 20160,
			CreatedAt:          tsCreated,
			UpdatedAt:          tsUpdated,
		},
		Current: types.RecommendationSnapshot{
			Config:     map[string]any{"cpu_request_cores": 0.5},
			HourlyCost: money("0.4800"),
		},
		Recommended: types.RecommendationProposal{
			Config: map[string]any{"cpu_request_cores": 0.15},
		},
		Applied: types.RecommendationApplied{
			Config: map[string]any{},
		},
		Savings: types.RecommendationSavings{
			EstimatedHourly: money("0.0120"),
		},
		MetricsSnapshot: map[string]any{
			"cpu_p95_cores": 0.14,
			"cpu_p99_cores": 0.18,
		},
	}
}

// SampleRecommendations returns three Recommendation fixtures.
func SampleRecommendations() []types.Recommendation {
	first := SampleRecommendation()

	second := first
	second.ID = recIDTwo
	second.Metadata.RecommendationType = "workload_memory_rightsize"
	second.Metadata.Title = "Right-size prometheus memory limit"
	second.Metadata.ResourceName = "prometheus"
	second.Metadata.ResourceUID = workloadUIDPrometheus
	second.Metadata.Namespace = nameMonitoringNS
	second.Savings.EstimatedHourly = money("0.0250")

	third := first
	third.ID = recIDThree
	third.Metadata.RecommendationType = "node_consolidate"
	third.Metadata.ResourceType = "Node"
	third.Metadata.ResourceName = "ip-10-0-3-12.ec2.internal"
	third.Metadata.Title = "Consolidate underutilized node"
	third.Metadata.Priority = "high"
	third.Savings.EstimatedHourly = money("0.1920")

	return []types.Recommendation{first, second, third}
}

// SampleTeam returns a single Team fixture (the platform team).
func SampleTeam() types.Team {
	return types.Team{
		ID:   teamIDPlatform,
		Kind: "team",
		Metadata: types.TeamMetadata{
			Name:        nameTeamPlatform,
			Description: "Platform engineering",
			Origin:      originManual,
			OwnerEmail:  "platform@example.com",
			Department:  &types.NestedRef{ID: deptIDEngineering, Name: nameDeptEng},
			CreatedAt:   tsCreated,
			UpdatedAt:   tsUpdated,
			LastSeenAt:  tsLastSeen,
		},
		AssignedWorkloads: 12,
		AssignedPVs:       4,
		Cost: types.TeamCost{
			CurrentRunRateHourly: money("0.2200"),
			CostMode:             costFullyLoaded,
		},
	}
}

// SampleTeams returns two Team fixtures (platform and sre).
func SampleTeams() []types.Team {
	first := SampleTeam()

	second := first
	second.ID = teamIDSRE
	second.Metadata.Name = nameTeamSRE
	second.Metadata.Description = "Site reliability engineering"
	second.Metadata.OwnerEmail = "sre@example.com"
	second.AssignedWorkloads = 8
	second.AssignedPVs = 2
	second.Cost.CurrentRunRateHourly = money("0.1600")

	return []types.Team{first, second}
}

// SampleDepartment returns a single Department fixture (engineering).
func SampleDepartment() types.Department {
	return types.Department{
		ID:   deptIDEngineering,
		Kind: "department",
		Metadata: types.DepartmentMetadata{
			Name:        nameDeptEng,
			Description: "Engineering org",
			Origin:      originManual,
			OwnerEmail:  "eng@example.com",
			CreatedAt:   tsCreated,
			UpdatedAt:   tsUpdated,
		},
		Teams:             2,
		AssignedWorkloads: 20,
		AssignedPVs:       6,
		Cost: types.DepartmentCost{
			CurrentRunRateHourly: money("0.3800"),
			CostMode:             costFullyLoaded,
		},
	}
}

// SampleDepartments returns two Department fixtures (engineering and data).
func SampleDepartments() []types.Department {
	first := SampleDepartment()

	second := first
	second.ID = deptIDData
	second.Metadata.Name = "data"
	second.Metadata.Description = "Data platform"
	second.Metadata.OwnerEmail = "data@example.com"
	second.Teams = 1
	second.AssignedWorkloads = 5
	second.AssignedPVs = 2
	second.Cost.CurrentRunRateHourly = money("0.1500")

	return []types.Department{first, second}
}

// SampleOrganization returns the Organization fixture used by GET
// /v1/organization. Counts and rollups are populated to match the cluster
// fixtures so dashboards stay consistent.
func SampleOrganization() types.Organization {
	return types.Organization{
		ID:   orgID,
		Kind: "organization",
		Metadata: types.OrganizationMetadata{
			Name:      "Example Inc",
			Domain:    "example.com",
			PlanType:  "enterprise",
			IsActive:  true,
			CreatedAt: tsCreated,
		},
		Capacity: types.OrganizationCapacity{
			CPU:     types.CapacityCPU{TotalCores: 320.0, AllocatableCores: 308.0},
			Memory:  types.CapacityMemory{TotalBytes: 1374389534720, AllocatableBytes: 1320000000000},
			GPU:     types.CapacityGPU{Total: 8, Allocatable: 8, Model: "nvidia-a100"},
			Storage: types.CapacityStorage{TotalBytes: 26843545600000},
		},
		Utilization: types.OrganizationUtilization{
			CPU:    types.UtilizationCPU{RequestedCores: 200.0, UsedCores: 160.0, UtilizationPercent: 50.0},
			Memory: types.UtilizationMemory{RequestedBytes: 880000000000, UsedBytes: 687194767360, UtilizationPercent: 52.0},
			Counts: types.OrganizationCounts{
				Clusters: 3, ConnectedClusters: 3, Nodes: 18, Namespaces: 14,
				Workloads: 60, Pods: 230, Containers: 280, PersistentVolumes: 30,
				Recommendations: 8,
			},
		},
		Cost: types.OrganizationCost{
			CurrentRunRateHourly: money("17.8300"),
			LastUpdatedAt:        tsLastSeen,
		},
	}
}

// SampleOrganizationDashboard returns the OrganizationDashboard fixture used
// by GET /v1/organization/dashboard. TopClusters is populated with five
// entries so tests can exercise the top_clusters_limit query parameter.
func SampleOrganizationDashboard() types.OrganizationDashboard {
	return types.OrganizationDashboard{
		OrganizationID: orgID,
		Snapshot:       SampleOrganization(),
		MonthToDate: types.OrgDashboardMTD{
			BilledCost: money("4250.0000"),
			Calendar: types.OrgDashboardCalendar{
				Month:        "2025-05",
				DaysElapsed:  20,
				DaysInMonth:  31,
				MonthStartAt: "2025-05-01T00:00:00Z",
			},
		},
		Savings: types.OrgDashboardSavings{
			CurrentHourlyPotential: money("0.5200"),
			RecommendationCount:    8,
		},
		TopClusters: []types.OrgDashboardClusterRollup{
			{
				Cluster:              types.NestedRef{ID: clusterIDProd, Name: nameProdCluster},
				CurrentRunRateHourly: money("12.4700"),
				MonthToDateCost:      money("2990.0000"),
			},
			{
				Cluster:              types.NestedRef{ID: clusterIDStaging, Name: "staging-cluster"},
				CurrentRunRateHourly: money("4.2100"),
				MonthToDateCost:      money("1010.0000"),
			},
			{
				Cluster:              types.NestedRef{ID: clusterIDDev, Name: "dev-cluster"},
				CurrentRunRateHourly: money("1.1500"),
				MonthToDateCost:      money("250.0000"),
			},
			{
				Cluster:              types.NestedRef{ID: "00000000-0000-0000-0000-000000000004", Name: "qa-cluster"},
				CurrentRunRateHourly: money("0.8700"),
				MonthToDateCost:      money("180.0000"),
			},
			{
				Cluster:              types.NestedRef{ID: "00000000-0000-0000-0000-000000000005", Name: "sandbox-cluster"},
				CurrentRunRateHourly: money("0.3300"),
				MonthToDateCost:      money("75.0000"),
			},
		},
	}
}

// SampleTeamAssignment returns a single TeamAssignment fixture binding the
// platform team to a namespace.
func SampleTeamAssignment() types.TeamAssignment {
	return types.TeamAssignment{
		ID:   assignmentIDOne,
		Kind: "team_assignment",
		Metadata: types.TeamAssignmentMetadata{
			Team:             types.NestedRef{ID: teamIDPlatform, Name: nameTeamPlatform},
			Cluster:          types.NestedRef{ID: clusterIDProd, Name: nameProdCluster},
			EntityType:       "namespace",
			EntityIdentifier: nameDefaultNS,
			EntityName:       nameDefaultNS,
			EntityNamespace:  nameDefaultNS,
			WeightPercentage: 100.0,
			Source:           originManual,
			AssignedByUserID: stringPtr("user_admin_1"),
			CreatedAt:        tsCreated,
			UpdatedAt:        tsUpdated,
		},
	}
}

// SampleTeamAssignments returns three TeamAssignment fixtures across
// namespace, workload, and cluster entity types.
func SampleTeamAssignments() []types.TeamAssignment {
	first := SampleTeamAssignment()

	second := first
	second.ID = assignmentIDTwo
	second.Metadata.EntityType = "workload"
	second.Metadata.EntityIdentifier = workloadUIDAPI
	second.Metadata.EntityName = nameAPIGateway
	second.Metadata.WeightPercentage = 50.0

	third := first
	third.ID = assignmentIDThree
	third.Metadata.EntityType = "cluster"
	third.Metadata.EntityIdentifier = clusterIDProd
	third.Metadata.EntityName = nameProdCluster
	third.Metadata.EntityNamespace = ""
	third.Metadata.WeightPercentage = 25.0
	third.Metadata.AssignedByUserID = nil
	third.Metadata.Source = "auto"

	return []types.TeamAssignment{first, second, third}
}
