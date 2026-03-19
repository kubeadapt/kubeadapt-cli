package testutil

import "github.com/kubeadapt/kubeadapt-cli/internal/api/types"

// StringPtr returns a pointer to the given string value.
func StringPtr(s string) *string { return &s }

// Float64Ptr returns a pointer to the given float64 value.
func Float64Ptr(f float64) *float64 { return &f }

// IntPtr returns a pointer to the given int value.
func IntPtr(i int) *int { return &i }

// SampleOverview returns a sample overview response.
func SampleOverview() *types.OverviewResponse {
	return &types.OverviewResponse{
		OrganizationID:          "org-123",
		ClusterCount:            3,
		ConnectedClusterCount:   2,
		TotalNodes:              15,
		TotalPods:               120,
		TotalWorkloads:          45,
		TotalHourlyCost:         Float64Ptr(12.50),
		TotalMonthlyCost:        Float64Ptr(9000.00),
		PotentialMonthlySavings: Float64Ptr(1500.00),
		AvgCPUUtilization:       Float64Ptr(42.5),
		AvgMemoryUtilization:    Float64Ptr(65.3),
		RecommendationCount:     8,
		MTDActualCost:           Float64Ptr(4250.00),
		RunRate:                 Float64Ptr(9125.00),
		EfficiencyScore:         Float64Ptr(62.8),
	}
}

// SampleClusters returns sample cluster data.
func SampleClusters() []types.ClusterResponse {
	return []types.ClusterResponse{
		{
			ID:                       "cls-001",
			Name:                     "production-us",
			Provider:                 "aws",
			Region:                   StringPtr("us-east-1"),
			Environment:              "production",
			Status:                   "connected",
			Version:                  StringPtr("1.28"),
			NodeCount:                10,
			PodCount:                 85,
			CPUCores:                 40,
			MemoryGB:                 160,
			CPUUtilizationPercent:    45.2,
			MemoryUtilizationPercent: 68.7,
			HourlyCost:               8.50,
			EfficiencyScore:          Float64Ptr(72.5),
			MonthlyCost:              Float64Ptr(6205.00),
			PotentialMonthlySavings:  Float64Ptr(450.00),
			RecommendationCount:      IntPtr(3),
		},
		{
			ID:                       "cls-002",
			Name:                     "staging-eu",
			Provider:                 "aws",
			Region:                   StringPtr("eu-west-1"),
			Environment:              "staging",
			Status:                   "connected",
			Version:                  StringPtr("1.28"),
			NodeCount:                5,
			PodCount:                 35,
			CPUCores:                 20,
			MemoryGB:                 80,
			CPUUtilizationPercent:    22.1,
			MemoryUtilizationPercent: 45.3,
			HourlyCost:               4.00,
			EfficiencyScore:          Float64Ptr(45.8),
			MonthlyCost:              Float64Ptr(2920.00),
			PotentialMonthlySavings:  Float64Ptr(180.00),
			RecommendationCount:      IntPtr(2),
		},
	}
}

// SampleNodes returns sample node data.
func SampleNodes() []types.NodeResponse {
	return []types.NodeResponse{
		{
			ID:                  "node-001",
			ClusterID:           "cls-001",
			ClusterName:         "production-us",
			NodeName:            "ip-10-0-1-42.ec2.internal",
			InstanceType:        StringPtr("m5.xlarge"),
			NodeGroup:           StringPtr("general-purpose"),
			AvailabilityZone:    StringPtr("us-east-1a"),
			IsReady:             true,
			IsSchedulable:       true,
			CPUCapacity:         4,
			CPUAllocatable:      3.92,
			MemoryCapacityGB:    16,
			MemoryAllocatableGB: 15.2,
			PodsCapacity:        58,
			PodsAllocatable:     58,
			HourlyCost:          0.192,
			SpotInstance:        false,
			PodCount:            IntPtr(12),
			MonthlyCost:         Float64Ptr(140.16),
		},
		{
			ID:                  "node-002",
			ClusterID:           "cls-001",
			ClusterName:         "production-us",
			NodeName:            "ip-10-0-2-87.ec2.internal",
			InstanceType:        StringPtr("c5.2xlarge"),
			NodeGroup:           StringPtr("compute-optimized"),
			AvailabilityZone:    StringPtr("us-east-1b"),
			IsReady:             true,
			IsSchedulable:       true,
			CPUCapacity:         8,
			CPUAllocatable:      7.91,
			MemoryCapacityGB:    16,
			MemoryAllocatableGB: 15.0,
			PodsCapacity:        58,
			PodsAllocatable:     58,
			HourlyCost:          0.34,
			SpotInstance:        true,
			PodCount:            IntPtr(8),
			MonthlyCost:         Float64Ptr(248.20),
		},
	}
}

// SampleWorkloads returns sample workload data.
func SampleWorkloads() []types.WorkloadResponse {
	return []types.WorkloadResponse{
		{
			ID:                "wl-001",
			ClusterID:         "cls-001",
			ClusterName:       "production-us",
			Namespace:         "default",
			WorkloadName:      "api-gateway",
			WorkloadKind:      "Deployment",
			Replicas:          3,
			AvailableReplicas: 3,
			CPURequest:        0.5,
			CPULimit:          1.0,
			MemoryRequestGB:   0.512,
			MemoryLimitGB:     1.0,
			HourlyCost:        0.085,
			EfficiencyScore:   Float64Ptr(68.5),
			MonthlyCost:       Float64Ptr(62.05),
		},
		{
			ID:                "wl-002",
			ClusterID:         "cls-001",
			ClusterName:       "production-us",
			Namespace:         "monitoring",
			WorkloadName:      "prometheus-server",
			WorkloadKind:      "StatefulSet",
			Replicas:          1,
			AvailableReplicas: 1,
			CPURequest:        1.0,
			CPULimit:          2.0,
			MemoryRequestGB:   2.0,
			MemoryLimitGB:     4.0,
			HourlyCost:        0.142,
			EfficiencyScore:   Float64Ptr(42.3),
			MonthlyCost:       Float64Ptr(103.66),
		},
	}
}

// SampleNamespaces returns sample namespace data.
func SampleNamespaces() []types.NamespaceResponse {
	return []types.NamespaceResponse{
		{
			Name:            "default",
			ClusterID:       "cls-001",
			ClusterName:     StringPtr("production-us"),
			PodCount:        12,
			WorkloadCount:   5,
			TotalCPUCores:   4.5,
			TotalMemoryGB:   8.0,
			HourlyCost:      0.45,
			Team:            StringPtr("platform"),
			Department:      StringPtr("engineering"),
			EfficiencyScore: Float64Ptr(65.2),
			MonthlyCost:     Float64Ptr(328.50),
			ContainerCount:  IntPtr(24),
		},
		{
			Name:            "monitoring",
			ClusterID:       "cls-001",
			ClusterName:     StringPtr("production-us"),
			PodCount:        6,
			WorkloadCount:   3,
			TotalCPUCores:   3.0,
			TotalMemoryGB:   6.0,
			HourlyCost:      0.32,
			Team:            StringPtr("sre"),
			Department:      StringPtr("engineering"),
			EfficiencyScore: Float64Ptr(52.8),
			MonthlyCost:     Float64Ptr(233.60),
			ContainerCount:  IntPtr(12),
		},
	}
}

// SampleDashboard returns a sample dashboard response.
func SampleDashboard() *types.DashboardResponse {
	return &types.DashboardResponse{
		OrganizationID:          "org-123",
		TotalMonthlyCost:        9125.00,
		TotalHourlyCost:         12.50,
		PotentialMonthlySavings: 1500.00,
		EfficiencyScore:         Float64Ptr(62.8),
		ClusterCount:            3,
		NodeCount:               15,
		PodCount:                120,
		MTDActualCost:           4250.00,
		RunRate:                 9125.00,
		DaysElapsed:             15,
		DaysInMonth:             31,
		TopClusters: []types.TopCluster{
			{ClusterID: "cls-001", ClusterName: "production-us", HourlyCost: 8.50, Efficiency: Float64Ptr(72.5)},
			{ClusterID: "cls-002", ClusterName: "staging-eu", HourlyCost: 4.00, Efficiency: Float64Ptr(45.8)},
		},
		TotalRecommendations: 8,
	}
}

// SampleClusterDashboard returns a sample cluster dashboard response.
func SampleClusterDashboard() *types.ClusterDashboardResponse {
	return &types.ClusterDashboardResponse{
		ClusterID:                "cls-001",
		ClusterName:              "production-us",
		Provider:                 "aws",
		Region:                   StringPtr("us-east-1"),
		Environment:              "production",
		Status:                   "connected",
		Version:                  StringPtr("1.28"),
		NodeCount:                10,
		PodCount:                 85,
		ContainerCount:           120,
		DeploymentCount:          15,
		NamespaceCount:           8,
		HourlyCost:               8.50,
		MonthlyCost:              6205.00,
		TotalSavingsHourly:       0.037,
		MonthlySavings:           27.01,
		CPUCores:                 40,
		CPUUsage:                 18.5,
		CPUUtilizationPercent:    45.2,
		MemoryGB:                 160,
		MemoryUsageGB:            109.9,
		MemoryUtilizationPercent: 68.7,
		ClusterEfficiency:        72.5,
		RecommendationCount:      5,
		CostBreakdown: map[string]float64{
			"cpu_cost": 5.20, "memory_cost": 2.80, "storage_cost": 0.40, "gpu_cost": 0.10,
		},
		MTDActualCost:           Float64Ptr(3100.00),
		PotentialMonthlySavings: Float64Ptr(27.01),
		RecommendationSummary: []types.RecommendationSummaryItem{
			{Type: "rightsize", Count: 3, PotentialSavings: 18.50},
			{Type: "scale_down", Count: 2, PotentialSavings: 8.51},
		},
	}
}

// SampleRecommendations returns sample recommendation data.
func SampleRecommendations() []types.RecommendationResponse {
	return []types.RecommendationResponse{
		{
			ID:                      "rec-001",
			ClusterID:               "cls-001",
			ClusterName:             "production-us",
			RecommendationType:      "rightsize",
			Namespace:               StringPtr("default"),
			ResourceName:            StringPtr("api-gateway"),
			ResourceType:            StringPtr("Deployment"),
			Title:                   StringPtr("Right-size api-gateway CPU request"),
			Description:             StringPtr("CPU request is 3x higher than P95 usage. Reduce from 500m to 150m."),
			EstimatedHourlySavings:  0.012,
			EstimatedMonthlySavings: 8.64,
			CurrentHourlyCost:       0.085,
			Status:                  "open",
			CreatedAt:               StringPtr("2025-01-15T10:30:00Z"),
			Priority:                StringPtr("low"),
		},
		{
			ID:                      "rec-002",
			ClusterID:               "cls-001",
			ClusterName:             "production-us",
			RecommendationType:      "scale_down",
			Namespace:               StringPtr("monitoring"),
			ResourceName:            StringPtr("prometheus-server"),
			ResourceType:            StringPtr("StatefulSet"),
			Title:                   StringPtr("Scale down prometheus-server memory limit"),
			Description:             StringPtr("Memory limit is 2x higher than peak usage. Reduce from 4GB to 2.5GB."),
			EstimatedHourlySavings:  0.025,
			EstimatedMonthlySavings: 18.00,
			CurrentHourlyCost:       0.142,
			Status:                  "open",
			CreatedAt:               StringPtr("2025-01-16T14:00:00Z"),
			Priority:                StringPtr("medium"),
		},
	}
}

// SampleTeamCosts returns sample team cost data.
func SampleTeamCosts() []types.TeamCostResponse {
	return []types.TeamCostResponse{
		{
			Team:           "platform",
			NamespaceCount: 2,
			WorkloadCount:  5,
			PodCount:       12,
			TotalCPUCores:  4.5,
			TotalMemoryGB:  8.0,
			HourlyCost:     0.22,
			MonthlyCost:    160.60,
		},
		{
			Team:           "sre",
			NamespaceCount: 1,
			WorkloadCount:  3,
			PodCount:       6,
			TotalCPUCores:  3.0,
			TotalMemoryGB:  6.0,
			HourlyCost:     0.16,
			MonthlyCost:    116.80,
		},
	}
}

// SampleDepartmentCosts returns sample department cost data.
func SampleDepartmentCosts() []types.DepartmentCostResponse {
	return []types.DepartmentCostResponse{
		{
			Department:     "engineering",
			TeamCount:      2,
			NamespaceCount: 3,
			WorkloadCount:  8,
			PodCount:       18,
			TotalCPUCores:  7.5,
			TotalMemoryGB:  14.0,
			HourlyCost:     0.77,
			MonthlyCost:    562.10,
		},
		{
			Department:     "data",
			TeamCount:      1,
			NamespaceCount: 1,
			WorkloadCount:  2,
			PodCount:       4,
			TotalCPUCores:  2.0,
			TotalMemoryGB:  4.0,
			HourlyCost:     0.18,
			MonthlyCost:    131.40,
		},
	}
}

// SampleNodeGroups returns sample node group data.
func SampleNodeGroups() []types.NodeGroupResponse {
	return []types.NodeGroupResponse{
		{
			ID:             "ng-001",
			Name:           "general-purpose",
			ClusterID:      "cls-001",
			ClusterName:    StringPtr("production-us"),
			NodeCount:      6,
			InstanceType:   StringPtr("m5.xlarge"),
			TotalCPUCores:  Float64Ptr(24.0),
			TotalMemoryGB:  Float64Ptr(96.0),
			SpotPercentage: Float64Ptr(0.0),
			HourlyCost:     Float64Ptr(1.152),
		},
		{
			ID:             "ng-002",
			Name:           "compute-optimized",
			ClusterID:      "cls-001",
			ClusterName:    StringPtr("production-us"),
			NodeCount:      4,
			InstanceType:   StringPtr("c5.2xlarge"),
			TotalCPUCores:  Float64Ptr(32.0),
			TotalMemoryGB:  Float64Ptr(64.0),
			SpotPercentage: Float64Ptr(50.0),
			HourlyCost:     Float64Ptr(0.68),
		},
	}
}

// SamplePersistentVolumes returns sample persistent volume data.
func SamplePersistentVolumes() []types.PersistentVolumeResponse {
	return []types.PersistentVolumeResponse{
		{
			ID:           "pv-001",
			Name:         "pvc-prometheus-data",
			ClusterID:    "cls-001",
			ClusterName:  StringPtr("production-us"),
			Namespace:    StringPtr("monitoring"),
			PVCName:      StringPtr("prometheus-data"),
			StorageClass: StringPtr("gp3"),
			CapacityGB:   100.0,
			AccessModes:  []string{"ReadWriteOnce"},
			VolumeType:   StringPtr("gp3"),
			Zone:         StringPtr("us-east-1a"),
			HourlyCost:   Float64Ptr(0.011),
		},
		{
			ID:           "pv-002",
			Name:         "pvc-postgres-data",
			ClusterID:    "cls-001",
			ClusterName:  StringPtr("production-us"),
			Namespace:    StringPtr("default"),
			PVCName:      StringPtr("postgres-data"),
			StorageClass: StringPtr("gp3"),
			CapacityGB:   20.0,
			AccessModes:  []string{"ReadWriteOnce"},
			VolumeType:   StringPtr("gp3"),
			Zone:         StringPtr("us-east-1b"),
			HourlyCost:   Float64Ptr(0.005),
		},
	}
}
