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
		{"MTD Actual Cost", FormatCostPtr(o.MTDActualCost)},
		{"Run Rate (monthly)", FormatCostPtr(o.RunRate)},
		{"Efficiency Score", FormatPercentPtr(o.EfficiencyScore)},
		{"Recommendations", FormatInt(o.RecommendationCount)},
	}
	renderTable(headers, rows, noColor)
}

// RenderClusters renders a list of clusters as a styled table.
func RenderClusters(clusters []types.ClusterResponse, noColor bool) {
	headers := []string{"ID", "Name", "Provider", "Region", "Status", "Nodes", "Efficiency", "Monthly $", "Potential Savings", "$/hr"}
	rows := make([][]string, 0, len(clusters))
	for _, c := range clusters {
		rows = append(rows, []string{
			c.ID,
			c.Name,
			c.Provider,
			FormatOptionalString(c.Region),
			c.Status,
			FormatInt(c.NodeCount),
			FormatPercentPtr(c.EfficiencyScore),
			FormatCostPtr(c.MonthlyCost),
			FormatCostPtr(c.PotentialMonthlySavings),
			FormatCost(c.HourlyCost),
		})
	}
	renderTable(headers, rows, noColor)
	fmt.Fprintln(os.Stdout, StyleMuted.Render("Potential Savings reflects workload rightsizing only. For all recommendations visit https://app.kubeadapt.io"))
}

// RenderCluster renders a single cluster as a detail table.
func RenderCluster(c *types.ClusterResponse, noColor bool) {
	headers := []string{"Field", "Value"}
	rows := [][]string{
		{"ID", c.ID},
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
func RenderWorkloads(workloads []types.WorkloadResponse, noColor bool) {
	headers := []string{"Name", "Kind", "Namespace", "Cluster", "Replicas", "Efficiency", "Monthly $", "$/hr"}
	rows := make([][]string, 0, len(workloads))
	for _, w := range workloads {
		rows = append(rows, []string{
			w.WorkloadName,
			w.WorkloadKind,
			w.Namespace,
			w.ClusterName,
			fmt.Sprintf("%d/%d", w.AvailableReplicas, w.Replicas),
			FormatPercentPtr(w.EfficiencyScore),
			FormatCostPtr(w.MonthlyCost),
			FormatCost(w.HourlyCost),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderNodes renders a list of nodes as a styled table.
func RenderNodes(nodes []types.NodeResponse, noColor bool) {
	headers := []string{"Name", "Cluster", "Instance", "Ready", "CPU", "Memory", "Pods", "Spot", "$/hr"}
	rows := make([][]string, 0, len(nodes))
	for _, n := range nodes {
		rows = append(rows, []string{
			n.NodeName,
			n.ClusterName,
			FormatOptionalString(n.InstanceType),
			FormatBool(n.IsReady),
			FormatFloat(n.CPUAllocatable, 1),
			FormatMemoryGB(n.MemoryAllocatableGB),
			FormatIntPtr(n.PodCount),
			FormatBool(n.SpotInstance),
			FormatCost(n.HourlyCost),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderRecommendations renders a list of recommendations as a styled table.
func RenderRecommendations(recs []types.RecommendationResponse, noColor bool) {
	headers := []string{"ID", "Type", "Cluster", "Resource", "Priority", "Status", "Monthly Savings"}
	rows := make([][]string, 0, len(recs))
	for _, r := range recs {
		resource := FormatOptionalString(r.ResourceName)
		if ns := FormatOptionalString(r.Namespace); ns != "-" && ns != "" {
			resource = ns + "/" + resource
		}
		rows = append(rows, []string{
			r.ID,
			r.RecommendationType,
			r.ClusterName,
			resource,
			FormatOptionalString(r.Priority),
			r.Status,
			FormatCost(r.EstimatedMonthlySavings),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderTeamCosts renders team cost data as a styled table.
func RenderTeamCosts(costs []types.TeamCostResponse, noColor bool) {
	headers := []string{"Team", "Namespaces", "Workloads", "Pods", "CPU", "Memory", "$/hr", "$/mo"}
	rows := make([][]string, 0, len(costs))
	for _, c := range costs {
		rows = append(rows, []string{
			c.Team,
			FormatInt(c.NamespaceCount),
			FormatInt(c.WorkloadCount),
			FormatInt(c.PodCount),
			FormatFloat(c.TotalCPUCores, 1),
			FormatMemoryGB(c.TotalMemoryGB),
			FormatCost(c.HourlyCost),
			FormatCost(c.MonthlyCost),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderDepartmentCosts renders department cost data as a styled table.
func RenderDepartmentCosts(costs []types.DepartmentCostResponse, noColor bool) {
	headers := []string{"Department", "Namespaces", "Workloads", "Pods", "CPU", "Memory", "$/hr", "$/mo"}
	rows := make([][]string, 0, len(costs))
	for _, c := range costs {
		rows = append(rows, []string{
			c.Department,
			FormatInt(c.NamespaceCount),
			FormatInt(c.WorkloadCount),
			FormatInt(c.PodCount),
			FormatFloat(c.TotalCPUCores, 1),
			FormatMemoryGB(c.TotalMemoryGB),
			FormatCost(c.HourlyCost),
			FormatCost(c.MonthlyCost),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderNodeGroups renders node groups as a styled table.
func RenderNodeGroups(groups []types.NodeGroupResponse, noColor bool) {
	headers := []string{"Name", "Cluster", "Instance", "Nodes", "CPU", "Memory", "Spot %", "$/hr"}
	rows := make([][]string, 0, len(groups))
	for _, g := range groups {
		rows = append(rows, []string{
			g.Name,
			FormatOptionalString(g.ClusterName),
			FormatOptionalString(g.InstanceType),
			FormatInt(g.NodeCount),
			FormatFloatPtr(g.TotalCPUCores, 1),
			FormatMemoryGBPtr(g.TotalMemoryGB),
			FormatPercentPtr(g.SpotPercentage),
			FormatCostPtr(g.HourlyCost),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderNamespaces renders namespaces as a styled table.
func RenderNamespaces(namespaces []types.NamespaceResponse, noColor bool) {
	headers := []string{"Name", "Cluster", "Pods", "Workloads", "Containers", "Efficiency", "Monthly $", "Team", "$/hr"}
	rows := make([][]string, 0, len(namespaces))
	for _, n := range namespaces {
		rows = append(rows, []string{
			n.Name,
			FormatOptionalString(n.ClusterName),
			FormatInt(n.PodCount),
			FormatInt(n.WorkloadCount),
			FormatIntPtr(n.ContainerCount),
			FormatPercentPtr(n.EfficiencyScore),
			FormatCostPtr(n.MonthlyCost),
			FormatOptionalString(n.Team),
			FormatCost(n.HourlyCost),
		})
	}
	renderTable(headers, rows, noColor)
}

// RenderPersistentVolumes renders persistent volumes as a styled table.
func RenderPersistentVolumes(pvs []types.PersistentVolumeResponse, noColor bool) {
	headers := []string{"Name", "Cluster", "Namespace", "PVC", "Storage Class", "Capacity", "Type", "$/hr"}
	rows := make([][]string, 0, len(pvs))
	for _, pv := range pvs {
		rows = append(rows, []string{
			pv.Name,
			FormatOptionalString(pv.ClusterName),
			FormatOptionalString(pv.Namespace),
			FormatOptionalString(pv.PVCName),
			FormatOptionalString(pv.StorageClass),
			FormatMemoryGB(pv.CapacityGB),
			FormatOptionalString(pv.VolumeType),
			FormatCostPtr(pv.HourlyCost),
		})
	}
	renderTable(headers, rows, noColor)
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
		{"Potential Savings", FormatCost(d.PotentialMonthlySavings)},
		{"Efficiency", FormatPercentPtr(d.EfficiencyScore)},
		{"MTD Actual Cost", FormatCost(d.MTDActualCost)},
		{"Run Rate", FormatCost(d.RunRate)},
		{"Recommendations", FormatInt(d.TotalRecommendations)},
	}
	renderTable(headers, rows, noColor)
}
