package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

func newTable(headers []string, rows [][]string, noColor bool) *table.Table {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		Headers(headers...).
		Rows(rows...)

	if !noColor {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

		evenRowStyle := lipgloss.NewStyle().
			Padding(0, 1)

		oddRowStyle := lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#D1D5DB"))

		t = t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if row%2 == 0 {
				return evenRowStyle
			}
			return oddRowStyle
		})
	} else {
		t = t.StyleFunc(func(row, col int) lipgloss.Style {
			return lipgloss.NewStyle().Padding(0, 1)
		})
	}

	return t
}

func renderTable(headers []string, rows [][]string, noColor bool) {
	t := newTable(headers, rows, noColor)
	fmt.Fprintln(os.Stdout, t.Render())
}

func renderPaginationFooter(shown, total int) {
	if total > shown {
		fmt.Fprintln(os.Stdout, StyleMuted.Render(fmt.Sprintf("Showing %d of %d results. Use --limit and --offset to paginate.", shown, total)))
	}
}

func renderList[T any](items []T, emptyMsg string, headers []string, rowFn func(T) []string, noColor bool, total int) {
	if len(items) == 0 {
		fmt.Fprintln(os.Stdout, StyleMuted.Render(emptyMsg))
		return
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, rowFn(item))
	}
	renderTable(headers, rows, noColor)
	renderPaginationFooter(len(items), total)
}

// RenderOverview renders the overview as a styled table.
func RenderOverview(o *types.OverviewResponse, noColor bool) {
	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Clusters", FormatInt(o.ClusterCount)},
		{"Connected Clusters", FormatInt(o.ConnectedClusterCount)},
		{"Total Nodes", FormatInt(o.TotalNodes)},
		{"Total Pods", FormatInt(o.TotalPods)},
		{"Total Workloads", FormatInt(o.TotalWorkloads)},
		{"Hourly Cost", FormatCostPtr(o.TotalHourlyCost)},
		{"Monthly Cost", FormatCostPtr(o.TotalMonthlyCost)},
		{"Potential Monthly Savings", FormatCostPtr(o.PotentialMonthlySavings)},
		{"Avg CPU Utilization", FormatPercentPtr(o.AvgCPUUtilization)},
		{"Avg Memory Utilization", FormatPercentPtr(o.AvgMemoryUtilization)},
		{"Month-to-Date Spend", FormatCostPtr(o.MTDActualCost)},
		{"Run Rate (monthly)", FormatCostPtr(o.RunRate)},
		{"Efficiency Score", colorPercent(o.EfficiencyScore, noColor)},
		{"Recommendations", FormatInt(o.RecommendationCount)},
	}
	renderTable(headers, rows, noColor)
	fmt.Fprintln(os.Stdout, StyleMuted.Render("Potential Savings reflects workload rightsizing only. For all recommendations visit https://app.kubeadapt.io"))
}

// RenderClusters renders a list of clusters as a styled table.
func RenderClusters(clusters []types.ClusterResponse, total int, noColor bool) {
	if len(clusters) == 0 {
		fmt.Fprintln(os.Stdout, StyleMuted.Render("No clusters found. Connect a cluster at https://app.kubeadapt.io"))
		return
	}
	headers := []string{"ID", "Name", "Provider", "Region", "Status", "Nodes", "Efficiency", "Monthly $", "Potential Savings", "$/hr"}
	rows := make([][]string, 0, len(clusters))
	for _, c := range clusters {
		status := c.Status
		if !noColor {
			switch c.Status {
			case "connected":
				status = StyleSuccess.Render(c.Status)
			case "disconnected":
				status = StyleError.Render(c.Status)
			default:
				status = StyleWarning.Render(c.Status)
			}
		}
		rows = append(rows, []string{
			ShortID(c.ID),
			c.Name,
			c.Provider,
			FormatOptionalString(c.Region),
			status,
			FormatInt(c.NodeCount),
			FormatPercentPtr(c.EfficiencyScore),
			FormatCostPtr(c.MonthlyCost),
			FormatCostPtr(c.PotentialMonthlySavings),
			FormatCost(c.HourlyCost),
		})
	}
	renderTable(headers, rows, noColor)
	fmt.Fprintln(os.Stdout, StyleMuted.Render("Potential Savings reflects workload rightsizing only. For all recommendations visit https://app.kubeadapt.io"))
	renderPaginationFooter(len(clusters), total)
}

// RenderCluster renders a single cluster as a detail table.
func RenderCluster(c *types.ClusterResponse, noColor bool) {
	headers := []string{"Field", "Value"}
	rows := [][]string{
		{"ID", ShortID(c.ID)},
		{"Name", c.Name},
		{"Provider", c.Provider},
		{"Region", FormatOptionalString(c.Region)},
		{"Environment", c.Environment},
		{"Status", c.Status},
		{"Version", FormatOptionalString(c.Version)},
		{"Nodes", FormatInt(c.NodeCount)},
		{"Pods", FormatInt(c.PodCount)},
		{"CPU Cores", FormatFloat(c.CPUCores, 1)},
		{"Memory", FormatMemoryGB(c.MemoryGB)},
		{"CPU Utilization", FormatPercent(c.CPUUtilizationPercent)},
		{"Memory Utilization", FormatPercent(c.MemoryUtilizationPercent)},
		{"Hourly Cost", FormatCost(c.HourlyCost)},
		{"Efficiency Score", FormatPercentPtr(c.EfficiencyScore)},
		{"Monthly Cost", FormatCostPtr(c.MonthlyCost)},
		{"Potential Savings", FormatCostPtr(c.PotentialMonthlySavings)},
		{"Recommendations", FormatIntPtr(c.RecommendationCount)},
	}
	renderTable(headers, rows, noColor)
}

// RenderWorkloads renders a list of workloads as a styled table.
func RenderWorkloads(workloads []types.WorkloadResponse, total int, noColor bool) {
	renderList(workloads, "No workloads found. Try adjusting your filters or check 'kubeadapt get clusters' first.",
		[]string{"Name", "Kind", "Namespace", "Cluster", "Replicas", "Efficiency", "Monthly $", "$/hr"},
		func(w types.WorkloadResponse) []string {
			return []string{
				w.WorkloadName, w.WorkloadKind, w.Namespace, w.ClusterName,
				fmt.Sprintf("%d/%d", w.AvailableReplicas, w.Replicas),
				FormatPercentPtr(w.EfficiencyScore), FormatCostPtr(w.MonthlyCost), FormatCost(w.HourlyCost),
			}
		}, noColor, total)
}

// RenderNodes renders a list of nodes as a styled table.
func RenderNodes(nodes []types.NodeResponse, total int, noColor bool) {
	renderList(nodes, "No nodes found.",
		[]string{"Name", "Cluster", "Instance", "Ready", "CPU", "Memory", "Pods", "Spot", "$/hr"},
		func(n types.NodeResponse) []string {
			ready := FormatBool(n.IsReady)
			if !noColor {
				if n.IsReady {
					ready = StyleSuccess.Render("Yes")
				} else {
					ready = StyleError.Render("No")
				}
			}
			return []string{
				n.NodeName, n.ClusterName, FormatOptionalString(n.InstanceType), ready,
				FormatFloat(n.CPUAllocatable, 1), FormatMemoryGB(n.MemoryAllocatableGB),
				FormatIntPtr(n.PodCount), FormatBool(n.SpotInstance), FormatCost(n.HourlyCost),
			}
		}, noColor, total)
}

// RenderRecommendations renders a list of recommendations as a styled table.
func RenderRecommendations(recs []types.RecommendationResponse, total int, noColor bool) {
	renderList(recs, "No recommendations found. Your resources may already be optimized!",
		[]string{"ID", "Type", "Cluster", "Resource", "Priority", "Status", "Monthly Savings"},
		func(r types.RecommendationResponse) []string {
			resource := FormatOptionalString(r.ResourceName)
			if ns := FormatOptionalString(r.Namespace); ns != "-" && ns != "" {
				resource = ns + "/" + resource
			}
			status := r.Status
			priority := FormatOptionalString(r.Priority)
			if !noColor {
				switch r.Status {
				case "applied":
					status = StyleSuccess.Render(r.Status)
				case "dismissed":
					status = StyleMuted.Render(r.Status)
				case "open":
					status = StyleWarning.Render(r.Status)
				}
				if r.Priority != nil {
					switch *r.Priority {
					case "critical", "high":
						priority = StyleError.Render(*r.Priority)
					case "medium":
						priority = StyleWarning.Render(*r.Priority)
					case "low":
						priority = StyleSuccess.Render(*r.Priority)
					}
				}
			}
			return []string{ShortID(r.ID), r.RecommendationType, r.ClusterName, resource, priority, status, FormatCost(r.EstimatedMonthlySavings)}
		}, noColor, total)
}

// RenderTeamCosts renders team cost data as a styled table.
func RenderTeamCosts(costs []types.TeamCostResponse, total int, noColor bool) {
	renderList(costs, "No team cost data available.",
		[]string{"Team", "Namespaces", "Workloads", "Pods", "CPU", "Memory", "$/hr", "$/mo"},
		func(c types.TeamCostResponse) []string {
			return []string{
				c.Team, FormatInt(c.NamespaceCount), FormatInt(c.WorkloadCount), FormatInt(c.PodCount),
				FormatFloat(c.TotalCPUCores, 1), FormatMemoryGB(c.TotalMemoryGB),
				FormatCost(c.HourlyCost), FormatCost(c.MonthlyCost),
			}
		}, noColor, total)
}

// RenderDepartmentCosts renders department cost data as a styled table.
func RenderDepartmentCosts(costs []types.DepartmentCostResponse, total int, noColor bool) {
	renderList(costs, "No department cost data available.",
		[]string{"Department", "Namespaces", "Workloads", "Pods", "CPU", "Memory", "$/hr", "$/mo"},
		func(c types.DepartmentCostResponse) []string {
			return []string{
				c.Department, FormatInt(c.NamespaceCount), FormatInt(c.WorkloadCount), FormatInt(c.PodCount),
				FormatFloat(c.TotalCPUCores, 1), FormatMemoryGB(c.TotalMemoryGB),
				FormatCost(c.HourlyCost), FormatCost(c.MonthlyCost),
			}
		}, noColor, total)
}

// RenderNodeGroups renders node groups as a styled table.
func RenderNodeGroups(groups []types.NodeGroupResponse, total int, noColor bool) {
	renderList(groups, "No node groups found.",
		[]string{"Name", "Cluster", "Instance", "Nodes", "CPU", "Memory", "Spot %", "$/hr"},
		func(g types.NodeGroupResponse) []string {
			return []string{
				g.Name, FormatOptionalString(g.ClusterName), FormatOptionalString(g.InstanceType),
				FormatInt(g.NodeCount), FormatFloatPtr(g.TotalCPUCores, 1), FormatMemoryGBPtr(g.TotalMemoryGB),
				FormatPercentPtr(g.SpotPercentage), FormatCostPtr(g.HourlyCost),
			}
		}, noColor, total)
}

// RenderNamespaces renders namespaces as a styled table.
func RenderNamespaces(namespaces []types.NamespaceResponse, total int, noColor bool) {
	renderList(namespaces, "No namespaces found.",
		[]string{"Name", "Cluster", "Pods", "Workloads", "Containers", "Efficiency", "Monthly $", "Team", "$/hr"},
		func(n types.NamespaceResponse) []string {
			return []string{
				n.Name, FormatOptionalString(n.ClusterName),
				FormatInt(n.PodCount), FormatInt(n.WorkloadCount), FormatIntPtr(n.ContainerCount),
				FormatPercentPtr(n.EfficiencyScore), FormatCostPtr(n.MonthlyCost),
				FormatOptionalString(n.Team), FormatCost(n.HourlyCost),
			}
		}, noColor, total)
}

// RenderPersistentVolumes renders persistent volumes as a styled table.
func RenderPersistentVolumes(pvs []types.PersistentVolumeResponse, total int, noColor bool) {
	renderList(pvs, "No persistent volumes found.",
		[]string{"Name", "Cluster", "Namespace", "PVC", "Storage Class", "Capacity", "Type", "$/hr"},
		func(pv types.PersistentVolumeResponse) []string {
			return []string{
				pv.Name, FormatOptionalString(pv.ClusterName), FormatOptionalString(pv.Namespace),
				FormatOptionalString(pv.PVCName), FormatOptionalString(pv.StorageClass),
				FormatMemoryGB(pv.CapacityGB), FormatOptionalString(pv.VolumeType), FormatCostPtr(pv.HourlyCost),
			}
		}, noColor, total)
}

// RenderDashboard renders the dashboard as a styled table.
func RenderDashboard(d *types.DashboardResponse, noColor bool) {
	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Clusters", FormatInt(d.ClusterCount)},
		{"Nodes", FormatInt(d.NodeCount)},
		{"Pods", FormatInt(d.PodCount)},
		{"Monthly Cost", FormatCost(d.TotalMonthlyCost)},
		{"Hourly Cost", FormatCost(d.TotalHourlyCost)},
		{"Potential Savings (monthly)", FormatCost(d.PotentialMonthlySavings)},
		{"Efficiency", colorPercent(d.EfficiencyScore, noColor)},
		{"Month-to-Date Spend", FormatCost(d.MTDActualCost)},
		{"Run Rate (monthly)", FormatCost(d.RunRate)},
		{"Recommendations", FormatInt(d.TotalRecommendations)},
	}
	renderTable(headers, rows, noColor)
	fmt.Fprintln(os.Stdout, StyleMuted.Render("Potential Savings reflects workload rightsizing only. For all recommendations visit https://app.kubeadapt.io"))
}

// RenderClusterDashboard renders a cluster dashboard summary as a key-value table.
func RenderClusterDashboard(d *types.ClusterDashboardResponse, noColor bool) {
	headers := []string{"Metric", "Value"}
	rows := [][]string{
		{"Cluster ID", d.ClusterID},
		{"Cluster Name", d.ClusterName},
		{"Provider", d.Provider},
		{"Region", FormatOptionalString(d.Region)},
		{"Environment", d.Environment},
		{"Status", d.Status},
		{"Version", FormatOptionalString(d.Version)},
		{"Nodes", FormatInt(d.NodeCount)},
		{"Pods", FormatInt(d.PodCount)},
		{"Containers", FormatInt(d.ContainerCount)},
		{"Deployments", FormatInt(d.DeploymentCount)},
		{"Namespaces", FormatInt(d.NamespaceCount)},
		{"Hourly Cost", FormatCost(d.HourlyCost)},
		{"Monthly Cost", FormatCost(d.MonthlyCost)},
		{"Total Savings (hourly)", FormatCost(d.TotalSavingsHourly)},
		{"Monthly Savings", FormatCost(d.MonthlySavings)},
		{"CPU Cores", FormatFloat(d.CPUCores, 1)},
		{"CPU Usage", FormatFloat(d.CPUUsage, 2)},
		{"CPU Utilization", FormatPercent(d.CPUUtilizationPercent)},
		{"Memory", FormatMemoryGB(d.MemoryGB)},
		{"Memory Usage", FormatMemoryGB(d.MemoryUsageGB)},
		{"Memory Utilization", FormatPercent(d.MemoryUtilizationPercent)},
		{"Cluster Efficiency", FormatPercent(d.ClusterEfficiency)},
		{"Recommendations", FormatInt(d.RecommendationCount)},
		{"MTD Actual Cost", FormatCostPtr(d.MTDActualCost)},
		{"Potential Monthly Savings", FormatCostPtr(d.PotentialMonthlySavings)},
	}
	renderTable(headers, rows, noColor)
}

// RenderCostDistribution renders time-series cost distribution data.
func RenderCostDistribution(d *types.CostDistributionResponse, noColor bool) {
	headers := []string{"Timestamp", "Hourly Cost", "CPU Util", "Memory Util", "Efficiency"}
	rows := make([][]string, 0, len(d.DataPoints))
	for _, p := range d.DataPoints {
		rows = append(rows, []string{
			p.Timestamp,
			FormatCost(p.HourlyCost),
			FormatPercent(p.CPUUtilization),
			FormatPercent(p.MemoryUtilization),
			FormatPercent(p.Efficiency),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderNodeMetrics renders time-series node metrics.
func RenderNodeMetrics(d *types.NodeMetricsResponse, noColor bool) {
	headers := []string{"Timestamp", "CPU Usage", "CPU Capacity", "Memory Usage %", "CPU Usage %"}
	rows := make([][]string, 0, len(d.DataPoints))
	for _, p := range d.DataPoints {
		rows = append(rows, []string{
			p.Timestamp,
			FormatFloat(p.CPUUsage, 2),
			FormatFloat(p.CPUCapacity, 1),
			FormatPercent(p.MemoryUsagePercent),
			FormatPercent(p.CPUUsagePercent),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderNodeGroupDetails renders node group details with a header and nodes sub-table.
func RenderNodeGroupDetails(d *types.NodeGroupDetailResponse, noColor bool) {
	headers := []string{"Field", "Value"}
	rows := [][]string{
		{"Name", d.Name},
		{"Cluster ID", d.ClusterID},
		{"Cluster Name", FormatOptionalString(d.ClusterName)},
		{"Instance Type", FormatOptionalString(d.InstanceType)},
		{"Node Count", FormatInt(d.NodeCount)},
		{"Total CPU Cores", FormatFloat(d.TotalCPUCores, 1)},
		{"Total Memory", FormatMemoryGB(d.TotalMemoryGB)},
		{"Spot %", FormatPercent(d.SpotPercentage)},
		{"Hourly Cost", FormatCost(d.HourlyCost)},
		{"Avg CPU Utilization", FormatPercent(d.AvgCPUUtilization)},
		{"Avg Memory Utilization", FormatPercent(d.AvgMemoryUtilization)},
	}
	renderTable(headers, rows, noColor)

	if len(d.Nodes) > 0 {
		fmt.Fprintln(os.Stdout)
		nodeHeaders := []string{"Name", "Instance", "Ready", "CPU", "Memory", "Spot", "$/hr"}
		nodeRows := make([][]string, 0, len(d.Nodes))
		for _, n := range d.Nodes {
			nodeRows = append(nodeRows, []string{
				n.NodeName,
				FormatOptionalString(n.InstanceType),
				FormatBool(n.IsReady),
				FormatFloat(n.CPUAllocatable, 1),
				FormatMemoryGB(n.MemoryAllocatableGB),
				FormatBool(n.SpotInstance),
				FormatCost(n.HourlyCost),
			})
		}
		renderTable(nodeHeaders, nodeRows, noColor)
	}
}

// RenderWorkloadMetrics renders time-series workload metrics.
func RenderWorkloadMetrics(d *types.WorkloadMetricsResponse, noColor bool) {
	headers := []string{"Timestamp", "CPU Usage", "CPU Request", "Memory Usage", "Hourly Cost"}
	rows := make([][]string, 0, len(d.DataPoints))
	for _, p := range d.DataPoints {
		rows = append(rows, []string{
			p.Timestamp,
			FormatFloat(p.CPUUsage, 3),
			FormatFloat(p.CPURequest, 3),
			FormatMemoryGB(p.MemoryUsageBytes / (1024 * 1024 * 1024)),
			FormatCost(p.HourlyCost),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderWorkloadNodes renders node distribution for a workload.
func RenderWorkloadNodes(d *types.WorkloadNodesResponse, noColor bool) {
	headers := []string{"Node Name", "Pods", "CPU Usage", "Memory Usage"}
	rows := make([][]string, 0, len(d.Nodes))
	for _, n := range d.Nodes {
		rows = append(rows, []string{
			n.NodeName,
			FormatInt(n.PodCount),
			FormatFloat(n.CPUUsage, 3),
			FormatMemoryGB(n.MemoryUsageBytes / (1024 * 1024 * 1024)),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderNamespaceDetails renders namespace details with a header and workloads sub-table.
func RenderNamespaceDetails(d *types.NamespaceDetailResponse, noColor bool) {
	headers := []string{"Field", "Value"}
	rows := [][]string{
		{"Name", d.Name},
		{"Cluster ID", d.ClusterID},
		{"Cluster Name", FormatOptionalString(d.ClusterName)},
		{"Team", FormatOptionalString(d.Team)},
		{"Department", FormatOptionalString(d.Department)},
		{"Pods", FormatInt(d.PodCount)},
		{"Workloads", FormatInt(d.WorkloadCount)},
		{"Total CPU Cores", FormatFloat(d.TotalCPUCores, 1)},
		{"Total Memory", FormatMemoryGB(d.TotalMemoryGB)},
		{"Hourly Cost", FormatCost(d.HourlyCost)},
		{"Efficiency Score", FormatPercentPtr(d.EfficiencyScore)},
		{"Monthly Cost", FormatCostPtr(d.MonthlyCost)},
		{"Containers", FormatIntPtr(d.ContainerCount)},
	}
	renderTable(headers, rows, noColor)

	if len(d.Workloads) > 0 {
		fmt.Fprintln(os.Stdout)
		wlHeaders := []string{"Name", "Kind", "Replicas", "CPU Req", "Mem Req", "$/hr"}
		wlRows := make([][]string, 0, len(d.Workloads))
		for _, w := range d.Workloads {
			wlRows = append(wlRows, []string{
				w.WorkloadName,
				w.WorkloadKind,
				fmt.Sprintf("%d/%d", w.AvailableReplicas, w.Replicas),
				FormatFloat(w.CPURequest, 3),
				FormatMemoryGB(w.MemoryRequestGB),
				FormatCost(w.HourlyCost),
			})
		}
		renderTable(wlHeaders, wlRows, noColor)
	}
}

// RenderNamespaceTrends renders time-series namespace trends.
func RenderNamespaceTrends(d *types.NamespaceTrendsResponse, noColor bool) {
	headers := []string{"Timestamp", "CPU Usage", "Memory Usage", "Hourly Cost"}
	rows := make([][]string, 0, len(d.DataPoints))
	for _, p := range d.DataPoints {
		rows = append(rows, []string{
			p.Timestamp,
			FormatFloat(p.CPUUsage, 3),
			FormatMemoryGB(p.MemoryUsageBytes / (1024 * 1024 * 1024)),
			FormatCost(p.HourlyCost),
		})
	}
	renderTable(headers, rows, noColor)
}
