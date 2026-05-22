package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

const (
	colDollarHr       = "$/hr"
	colCluster        = "Cluster"
	colType           = "Type"
	colTeam           = "Team"
	colOrigin         = "Origin"
	colField          = "Field"
	colValue          = "Value"
	colName           = "Name"
	colID             = "ID"
	colKind           = "Kind"
	colRegion         = "Region"
	colNamespace      = "Namespace"
	lblStatus         = "Status"
	lblLastSeenAt     = "Last Seen At"
	lblCPUCores       = "CPU Cores"
	lblCPUAllocatable = "CPU Allocatable"
	lblCPUUsed        = "CPU Used"
	lblCPUUtil        = "CPU Utilization"
	lblMemory         = "Memory"
	lblMemoryAlloc    = "Memory Allocatable"
	lblMemoryUsed     = "Memory Used"
	lblMemoryUtil     = "Memory Utilization"
	lblGPUTotal       = "GPU Total"
	lblNodes          = "Nodes"
	lblWorkloads      = "Workloads"
	lblPods           = "Pods"
	lblRunningPods    = "Running Pods"
	lblContainers     = "Containers"
	lblCostUpdatedAt  = "Cost Updated At"
	lblCostMode       = "Cost Mode"
	lblDescription    = "Description"
	lblCreatedAt      = "Created At"
)

var detailHeaders = []string{colField, colValue}

var noColor bool

// SetNoColor toggles whether table styling emits ANSI color codes. cmd/*
// flips this from the --no-color global flag at root command bind time.
func SetNoColor(v bool) {
	noColor = v
}

func newTable(headers []string, rows [][]string) *table.Table {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		Headers(headers...).
		Rows(rows...)

	if !noColor {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)
		evenRowStyle := lipgloss.NewStyle().Padding(0, 1)
		oddRowStyle := lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#D1D5DB"))

		t = t.StyleFunc(func(row, _ int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			case row%2 == 0:
				return evenRowStyle
			default:
				return oddRowStyle
			}
		})
	} else {
		t = t.StyleFunc(func(_, _ int) lipgloss.Style {
			return lipgloss.NewStyle().Padding(0, 1)
		})
	}
	return t
}

func writeTable(w io.Writer, headers []string, rows [][]string) error {
	if _, err := fmt.Fprintln(w, newTable(headers, rows).Render()); err != nil {
		return fmt.Errorf("writing table: %w", err)
	}
	return nil
}

func writeFooter(w io.Writer, itemsShown int, meta *types.Meta) error {
	footer := PaginationFooter(itemsShown, meta)
	if footer == "" {
		return nil
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("writing footer: %w", err)
	}
	rendered := footer
	if !noColor {
		rendered = StyleMuted.Render(footer)
	}
	if _, err := fmt.Fprintln(w, rendered); err != nil {
		return fmt.Errorf("writing footer: %w", err)
	}
	return nil
}

func writeEmpty(w io.Writer, msg string) error {
	rendered := msg
	if !noColor {
		rendered = StyleMuted.Render(msg)
	}
	if _, err := fmt.Fprintln(w, rendered); err != nil {
		return fmt.Errorf("writing empty message: %w", err)
	}
	return nil
}

func styledStatus(s string) string {
	if noColor || s == "" {
		return s
	}
	switch strings.ToLower(s) {
	case "connected", "ready", "running", "active", "applied":
		return StyleSuccess.Render(s)
	case "disconnected", "failed", "error", "crashloopbackoff":
		return StyleError.Render(s)
	case "pending", "stale", "open", "warning":
		return StyleWarning.Render(s)
	default:
		return s
	}
}

func styledPriority(p string) string {
	if noColor || p == "" {
		return p
	}
	switch strings.ToLower(p) {
	case "critical", "high":
		return StyleError.Render(p)
	case "medium":
		return StyleWarning.Render(p)
	case "low":
		return StyleSuccess.Render(p)
	default:
		return p
	}
}

func joinList(items []string, max int) string {
	if len(items) == 0 {
		return noValue
	}
	if max <= 0 || len(items) <= max {
		return strings.Join(items, ",")
	}
	return strings.Join(items[:max], ",") + fmt.Sprintf(" (+%d)", len(items)-max)
}

// RenderClusters writes a list table of clusters plus a pagination footer.
func RenderClusters(w io.Writer, items []types.Cluster, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No clusters found.")
	}
	headers := []string{colID, colName, "Provider", colRegion, "Environment", lblStatus, "CPU%", "Mem%", colDollarHr}
	rows := make([][]string, 0, len(items))
	for _, c := range items {
		rows = append(rows, []string{
			c.ID,
			c.Metadata.Name,
			formatStr(c.Metadata.Provider),
			formatStr(c.Metadata.Region),
			formatStr(c.Metadata.Environment),
			styledStatus(c.Metadata.Status),
			FormatPercentage(c.Utilization.CPU.UtilizationPercent),
			FormatPercentage(c.Utilization.Memory.UtilizationPercent),
			FormatMoney(c.Cost.CurrentRunRateHourly),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderCluster writes a detail key-value table for one cluster.
func RenderCluster(w io.Writer, c types.Cluster) error {
	rows := [][]string{
		{"ID", c.ID},
		{colName, c.Metadata.Name},
		{"Provider", formatStr(c.Metadata.Provider)},
		{"Service", formatStr(c.Metadata.Service)},
		{colRegion, formatStr(c.Metadata.Region)},
		{"Availability Zones", joinList(c.Metadata.AvailabilityZones, 4)},
		{"Environment", formatStr(c.Metadata.Environment)},
		{lblStatus, styledStatus(c.Metadata.Status)},
		{"Stale", formatBool(c.Metadata.IsStale)},
		{"K8s Version", formatStr(c.Metadata.K8sVersion)},
		{"Agent Version", formatStr(c.Metadata.AgentVersion)},
		{"Discovery Source", formatStr(c.Metadata.DiscoverySource)},
		{lblLastSeenAt, formatStr(c.Metadata.LastSeenAt)},
		{lblCPUCores, FormatCores(c.Capacity.CPU.TotalCores)},
		{lblCPUAllocatable, FormatCores(c.Capacity.CPU.AllocatableCores)},
		{lblCPUUsed, FormatCores(c.Utilization.CPU.UsedCores)},
		{lblCPUUtil, FormatPercentage(c.Utilization.CPU.UtilizationPercent)},
		{lblMemory, FormatBytes(c.Capacity.Memory.TotalBytes)},
		{lblMemoryAlloc, FormatBytes(c.Capacity.Memory.AllocatableBytes)},
		{lblMemoryUsed, FormatBytes(c.Utilization.Memory.UsedBytes)},
		{lblMemoryUtil, FormatPercentage(c.Utilization.Memory.UtilizationPercent)},
		{lblGPUTotal, formatIntPlain(c.Capacity.GPU.Total)},
		{"Storage", FormatBytes(c.Capacity.Storage.TotalBytes)},
		{"Pods Allocatable", formatIntPlain(c.Capacity.Pods.Allocatable)},
		{lblNodes, formatIntPlain(c.Utilization.Counts.Nodes)},
		{"Namespaces", formatIntPlain(c.Utilization.Counts.Namespaces)},
		{lblWorkloads, formatIntPlain(c.Utilization.Counts.Workloads)},
		{lblPods, formatIntPlain(c.Utilization.Counts.Pods)},
		{lblRunningPods, formatIntPlain(c.Utilization.Counts.RunningPods)},
		{lblContainers, formatIntPlain(c.Utilization.Counts.Containers)},
		{colDollarHr, FormatMoney(c.Cost.CurrentRunRateHourly)},
		{lblCostUpdatedAt, formatStr(c.Cost.LastUpdatedAt)},
	}
	return writeTable(w, detailHeaders, rows)
}

// RenderWorkloads writes a list table of workloads plus a pagination footer.
func RenderWorkloads(w io.Writer, items []types.Workload, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No workloads found.")
	}
	headers := []string{colID, colCluster, colNamespace, colName, colKind, "Replicas", "CPU req", "Mem req", colDollarHr}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		replicas := fmt.Sprintf("%d/%d", item.Utilization.Replicas.Available, item.Utilization.Replicas.Desired)
		rows = append(rows, []string{
			item.ID,
			formatStr(item.Metadata.Cluster.Name),
			formatStr(item.Metadata.Namespace),
			item.Metadata.Name,
			formatStr(item.Metadata.WorkloadKind),
			replicas,
			FormatCores(item.Utilization.CPU.RequestedCores),
			FormatBytes(item.Utilization.Memory.RequestedBytes),
			FormatMoney(item.Cost.CurrentRunRateHourly),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderWorkload writes a detail key-value table for one workload.
func RenderWorkload(w io.Writer, item types.Workload) error {
	rows := [][]string{
		{"ID", item.ID},
		{colName, item.Metadata.Name},
		{colKind, formatStr(item.Metadata.WorkloadKind)},
		{colCluster, formatStr(item.Metadata.Cluster.Name)},
		{colNamespace, formatStr(item.Metadata.Namespace)},
		{"Service Account", formatStr(item.Metadata.ServiceAccountName)},
		{lblStatus, styledStatus(item.Metadata.Status)},
		{"Status Reason", formatStr(item.Metadata.StatusReason)},
		{"Suspended", formatBool(item.Metadata.IsSuspended)},
		{"Paused", formatBool(item.Metadata.IsPaused)},
		{"Has HPA", formatBool(item.Metadata.HasHPA)},
		{"Created (k8s)", formatStr(item.Metadata.CreatedAtK8s)},
		{lblLastSeenAt, formatStr(item.Metadata.LastSeenAt)},
		{"CPU Limit", FormatCores(item.Capacity.CPU.LimitCores)},
		{"Memory Limit", FormatBytes(item.Capacity.Memory.LimitBytes)},
		{"CPU Requested", FormatCores(item.Utilization.CPU.RequestedCores)},
		{lblCPUUsed, FormatCores(item.Utilization.CPU.UsedCores)},
		{lblCPUUtil, FormatPercentage(item.Utilization.CPU.UtilizationPercent)},
		{"Memory Requested", FormatBytes(item.Utilization.Memory.RequestedBytes)},
		{lblMemoryUsed, FormatBytes(item.Utilization.Memory.UsedBytes)},
		{lblMemoryUtil, FormatPercentage(item.Utilization.Memory.UtilizationPercent)},
		{"Replicas Desired", formatIntPlain(item.Utilization.Replicas.Desired)},
		{"Replicas Available", formatIntPlain(item.Utilization.Replicas.Available)},
		{"Replicas Updated", formatIntPlain(item.Utilization.Replicas.Updated)},
		{lblPods, formatIntPlain(item.Utilization.Counts.Pods)},
		{lblRunningPods, formatIntPlain(item.Utilization.Counts.RunningPods)},
		{lblContainers, formatIntPlain(item.Utilization.Counts.Containers)},
		{colDollarHr, FormatMoney(item.Cost.CurrentRunRateHourly)},
		{lblCostMode, formatStr(item.Cost.CostMode)},
		{lblCostUpdatedAt, formatStr(item.Cost.LastUpdatedAt)},
	}
	return writeTable(w, detailHeaders, rows)
}

// RenderPods writes a list table of pods plus a pagination footer.
func RenderPods(w io.Writer, items []types.Pod, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No pods found.")
	}
	headers := []string{colNamespace, colName, "Workload", "Node", "Phase", "QoS", colDollarHr}
	rows := make([][]string, 0, len(items))
	for _, p := range items {
		workload := noValue
		if p.Metadata.Workload != nil {
			workload = formatStr(p.Metadata.Workload.Name)
		}
		rows = append(rows, []string{
			formatStr(p.Metadata.Namespace),
			p.Metadata.Name,
			workload,
			formatRefName(p.Metadata.Node),
			styledStatus(p.Metadata.Phase),
			formatStr(p.Metadata.QOSClass),
			FormatMoney(p.Cost.CurrentRunRateHourly),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderNodes writes a list table of nodes plus a pagination footer.
func RenderNodes(w io.Writer, items []types.Node, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No nodes found.")
	}
	headers := []string{colID, colCluster, colName, "InstanceType", "Zone", "Group", "Ready", "Spot", colDollarHr}
	rows := make([][]string, 0, len(items))
	for _, n := range items {
		ready := formatBool(n.Metadata.IsReady)
		if !noColor {
			if n.Metadata.IsReady {
				ready = StyleSuccess.Render("Yes")
			} else {
				ready = StyleError.Render("No")
			}
		}
		rows = append(rows, []string{
			n.ID,
			formatStr(n.Metadata.Cluster.Name),
			n.Metadata.Name,
			formatStr(n.Metadata.InstanceType),
			formatStr(n.Metadata.AvailabilityZone),
			formatStr(n.Metadata.NodeGroup),
			ready,
			formatBool(n.Metadata.IsSpot),
			FormatMoney(n.Cost.CurrentRunRateHourly),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderNode writes a detail key-value table for one node.
func RenderNode(w io.Writer, n types.Node) error {
	rows := [][]string{
		{"ID", n.ID},
		{colName, n.Metadata.Name},
		{colCluster, formatStr(n.Metadata.Cluster.Name)},
		{"Node Role", formatStr(n.Metadata.NodeRole)},
		{"Instance Type", formatStr(n.Metadata.InstanceType)},
		{"Node Group", formatStr(n.Metadata.NodeGroup)},
		{"Availability Zone", formatStr(n.Metadata.AvailabilityZone)},
		{colRegion, formatStr(n.Metadata.Region)},
		{"Spot", formatBool(n.Metadata.IsSpot)},
		{"Capacity Type", formatStr(n.Metadata.CapacityType)},
		{"Architecture", formatStr(n.Metadata.Architecture)},
		{"OS", formatStr(n.Metadata.OperatingSystem)},
		{"Kubelet Version", formatStr(n.Metadata.KubeletVersion)},
		{"Ready", formatBool(n.Metadata.IsReady)},
		{"Schedulable", formatBool(n.Metadata.IsSchedulable)},
		{lblLastSeenAt, formatStr(n.Metadata.LastSeenAt)},
		{lblCPUCores, FormatCores(n.Capacity.CPU.TotalCores)},
		{lblCPUAllocatable, FormatCores(n.Capacity.CPU.AllocatableCores)},
		{lblCPUUsed, FormatCores(n.Utilization.CPU.UsedCores)},
		{lblCPUUtil, FormatPercentage(n.Utilization.CPU.UtilizationPercent)},
		{lblMemory, FormatBytes(n.Capacity.Memory.TotalBytes)},
		{lblMemoryAlloc, FormatBytes(n.Capacity.Memory.AllocatableBytes)},
		{lblMemoryUsed, FormatBytes(n.Utilization.Memory.UsedBytes)},
		{lblMemoryUtil, FormatPercentage(n.Utilization.Memory.UtilizationPercent)},
		{lblGPUTotal, formatIntPlain(n.Capacity.GPU.Total)},
		{"Ephemeral Storage", FormatBytes(n.Capacity.EphemeralStorage.TotalBytes)},
		{"Pods Allocatable", formatIntPlain(n.Capacity.Pods.Allocatable)},
		{"Pods Running", formatIntPlain(n.Utilization.Counts.RunningPods)},
		{colDollarHr, FormatMoney(n.Cost.CurrentRunRateHourly)},
		{"On-Demand Equivalent $/hr", FormatMoneyPtr(n.Cost.OnDemandEquivalentHourly)},
		{"Pricing Source", formatStr(n.Cost.PricingSource)},
		{lblCostUpdatedAt, formatStr(n.Cost.LastUpdatedAt)},
	}
	return writeTable(w, detailHeaders, rows)
}

// RenderNodeGroups writes a list table of node groups plus a pagination footer.
func RenderNodeGroups(w io.Writer, items []types.NodeGroup, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No node groups found.")
	}
	headers := []string{colCluster, colName, "InstanceTypes", lblNodes, "CPU cores", lblMemory, "Spot%", colDollarHr}
	rows := make([][]string, 0, len(items))
	for _, g := range items {
		rows = append(rows, []string{
			formatStr(g.Metadata.Cluster.Name),
			g.Metadata.Name,
			joinList(g.Metadata.InstanceTypes, 3),
			formatIntPlain(g.Utilization.Counts.Nodes),
			FormatCores(g.Capacity.CPU.TotalCores),
			FormatBytes(g.Capacity.Memory.TotalBytes),
			FormatPercentage(g.Metadata.SpotPercentage),
			FormatMoney(g.Cost.CurrentRunRateHourly),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderNodeGroup writes a detail key-value table for one node group, plus
// a member-nodes sub-table when populated.
func RenderNodeGroup(w io.Writer, g types.NodeGroup) error {
	rows := [][]string{
		{"ID", g.ID},
		{colName, g.Metadata.Name},
		{colCluster, formatStr(g.Metadata.Cluster.Name)},
		{"Instance Types", joinList(g.Metadata.InstanceTypes, 8)},
		{"Zones", joinList(g.Metadata.Zones, 8)},
		{"Spot Count", formatIntPlain(g.Metadata.SpotCount)},
		{"On-Demand Count", formatIntPlain(g.Metadata.OnDemandCount)},
		{"Spot %", FormatPercentage(g.Metadata.SpotPercentage)},
		{lblStatus, styledStatus(g.Metadata.Status)},
		{"Oldest Node Created (k8s)", formatStr(g.Metadata.OldestNodeCreatedAtK8s)},
		{lblCPUCores, FormatCores(g.Capacity.CPU.TotalCores)},
		{lblCPUAllocatable, FormatCores(g.Capacity.CPU.AllocatableCores)},
		{lblCPUUsed, FormatCores(g.Utilization.CPU.UsedCores)},
		{lblCPUUtil, FormatPercentage(g.Utilization.CPU.UtilizationPercent)},
		{lblMemory, FormatBytes(g.Capacity.Memory.TotalBytes)},
		{lblMemoryAlloc, FormatBytes(g.Capacity.Memory.AllocatableBytes)},
		{lblMemoryUsed, FormatBytes(g.Utilization.Memory.UsedBytes)},
		{lblMemoryUtil, FormatPercentage(g.Utilization.Memory.UtilizationPercent)},
		{lblNodes, formatIntPlain(g.Utilization.Counts.Nodes)},
		{"Ready Nodes", formatIntPlain(g.Utilization.Counts.ReadyNodes)},
		{lblPods, formatIntPlain(g.Utilization.Counts.Pods)},
		{colDollarHr, FormatMoney(g.Cost.CurrentRunRateHourly)},
		{"Spot Savings vs On-Demand $/hr", FormatMoneyPtr(g.Cost.SpotSavingsVsOndemandHourly)},
		{lblCostUpdatedAt, formatStr(g.Cost.LastUpdatedAt)},
	}
	if err := writeTable(w, detailHeaders, rows); err != nil {
		return err
	}
	if len(g.Nodes) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("writing node group nodes header: %w", err)
		}
		return RenderNodes(w, g.Nodes, nil)
	}
	return nil
}

// RenderNamespaces writes a list table of namespaces plus a pagination footer.
func RenderNamespaces(w io.Writer, items []types.Namespace, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No namespaces found.")
	}
	headers := []string{colCluster, colName, lblPods, "CPU cores", lblMemory, colTeam, "Dept", colDollarHr}
	rows := make([][]string, 0, len(items))
	for _, n := range items {
		rows = append(rows, []string{
			formatStr(n.Metadata.Cluster.Name),
			n.Metadata.Name,
			formatIntPlain(n.Utilization.Counts.Pods),
			FormatCores(n.Utilization.CPU.UsedCores),
			FormatBytes(n.Utilization.Memory.UsedBytes),
			formatRefName(n.Metadata.Team),
			formatRefName(n.Metadata.Department),
			FormatMoney(n.Cost.CurrentRunRateHourly),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderNamespace writes a detail key-value table for one namespace, plus
// a top-5 workloads sub-table when populated.
func RenderNamespace(w io.Writer, n types.Namespace) error {
	rows := [][]string{
		{"ID", n.ID},
		{colName, n.Metadata.Name},
		{colCluster, formatStr(n.Metadata.Cluster.Name)},
		{"UID (k8s)", formatStr(n.Metadata.UIDK8s)},
		{colTeam, formatRefName(n.Metadata.Team)},
		{"Department", formatRefName(n.Metadata.Department)},
		{"Created (k8s)", formatStr(n.Metadata.CreatedAtK8s)},
		{lblLastSeenAt, formatStr(n.Metadata.LastSeenAt)},
	}
	if n.Capacity != nil {
		rows = append(rows,
			[]string{"Quota CPU Cores", FormatCores(n.Capacity.CPU.QuotaCores)},
			[]string{"Quota Memory", FormatBytes(n.Capacity.Memory.QuotaBytes)},
		)
	}
	rows = append(rows,
		[]string{"CPU Requested", FormatCores(n.Utilization.CPU.RequestedCores)},
		[]string{lblCPUUsed, FormatCores(n.Utilization.CPU.UsedCores)},
		[]string{lblCPUUtil, FormatPercentage(n.Utilization.CPU.UtilizationPercent)},
		[]string{"Memory Requested", FormatBytes(n.Utilization.Memory.RequestedBytes)},
		[]string{lblMemoryUsed, FormatBytes(n.Utilization.Memory.UsedBytes)},
		[]string{lblMemoryUtil, FormatPercentage(n.Utilization.Memory.UtilizationPercent)},
		[]string{lblWorkloads, formatIntPlain(n.Utilization.Counts.Workloads)},
		[]string{lblPods, formatIntPlain(n.Utilization.Counts.Pods)},
		[]string{lblRunningPods, formatIntPlain(n.Utilization.Counts.RunningPods)},
		[]string{lblContainers, formatIntPlain(n.Utilization.Counts.Containers)},
		[]string{colDollarHr, FormatMoney(n.Cost.CurrentRunRateHourly)},
		[]string{lblCostMode, formatStr(n.Cost.CostMode)},
		[]string{lblCostUpdatedAt, formatStr(n.Cost.LastUpdatedAt)},
	)
	if err := writeTable(w, detailHeaders, rows); err != nil {
		return err
	}
	if len(n.WorkloadsTop5) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("writing top workloads header: %w", err)
		}
		topHeaders := []string{"ID", colKind, colName, colDollarHr}
		topRows := make([][]string, 0, len(n.WorkloadsTop5))
		for _, wl := range n.WorkloadsTop5 {
			topRows = append(topRows, []string{
				wl.ID, formatStr(wl.Kind), formatStr(wl.Name),
				FormatMoney(wl.Cost.CurrentRunRateHourly),
			})
		}
		return writeTable(w, topHeaders, topRows)
	}
	return nil
}

// RenderRecommendations writes a list table of recommendations plus a
// pagination footer.
func RenderRecommendations(w io.Writer, items []types.Recommendation, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No recommendations found.")
	}
	headers := []string{colID, colCluster, colType, "Resource", "Priority", "Risk", "Savings/hr"}
	rows := make([][]string, 0, len(items))
	for _, r := range items {
		resource := formatStr(r.Metadata.ResourceName)
		if r.Metadata.Namespace != "" && r.Metadata.ResourceName != "" {
			resource = r.Metadata.Namespace + "/" + r.Metadata.ResourceName
		}
		rows = append(rows, []string{
			r.ID,
			formatStr(r.Metadata.Cluster.Name),
			formatStr(r.Metadata.RecommendationType),
			resource,
			styledPriority(r.Metadata.Priority),
			formatStr(r.Metadata.RiskLevel),
			FormatMoney(r.Savings.EstimatedHourly),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderRecommendation writes a detail key-value table for one recommendation.
func RenderRecommendation(w io.Writer, r types.Recommendation) error {
	rows := [][]string{
		{"ID", r.ID},
		{colType, formatStr(r.Metadata.RecommendationType)},
		{"Resource Type", formatStr(r.Metadata.ResourceType)},
		{"Resource Name", formatStr(r.Metadata.ResourceName)},
		{"Resource UID", formatStr(r.Metadata.ResourceUID)},
		{colCluster, formatStr(r.Metadata.Cluster.Name)},
		{colNamespace, formatStr(r.Metadata.Namespace)},
		{"Title", formatStr(r.Metadata.Title)},
		{lblDescription, formatStr(r.Metadata.Description)},
		{"Cause", formatStr(r.Metadata.Cause)},
		{"Risk Level", formatStr(r.Metadata.RiskLevel)},
		{"Priority", styledPriority(r.Metadata.Priority)},
		{lblStatus, styledStatus(r.Metadata.Status)},
		{"Data Points Analyzed", formatIntPlain(r.Metadata.DataPointsAnalyzed)},
		{"Current $/hr", FormatMoney(r.Current.HourlyCost)},
		{"Estimated $/hr Savings", FormatMoney(r.Savings.EstimatedHourly)},
	}
	return writeTable(w, detailHeaders, rows)
}

// RenderTeams writes a list table of teams plus a pagination footer.
func RenderTeams(w io.Writer, items []types.Team, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No teams found.")
	}
	headers := []string{colID, colName, "Owner", colOrigin, "Dept", lblWorkloads, colDollarHr}
	rows := make([][]string, 0, len(items))
	for _, t := range items {
		rows = append(rows, []string{
			t.ID,
			t.Metadata.Name,
			formatStr(t.Metadata.OwnerEmail),
			formatStr(t.Metadata.Origin),
			formatRefName(t.Metadata.Department),
			formatIntPlain(t.AssignedWorkloads),
			FormatMoney(t.Cost.CurrentRunRateHourly),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderTeam writes a detail key-value table for one team.
func RenderTeam(w io.Writer, t types.Team) error {
	rows := [][]string{
		{"ID", t.ID},
		{colName, t.Metadata.Name},
		{lblDescription, formatStr(t.Metadata.Description)},
		{colOrigin, formatStr(t.Metadata.Origin)},
		{"Owner Email", formatStr(t.Metadata.OwnerEmail)},
		{"Department", formatRefName(t.Metadata.Department)},
		{"Assigned Workloads", formatIntPlain(t.AssignedWorkloads)},
		{"Assigned PVs", formatIntPlain(t.AssignedPVs)},
		{colDollarHr, FormatMoney(t.Cost.CurrentRunRateHourly)},
		{lblCostMode, formatStr(t.Cost.CostMode)},
		{lblCreatedAt, formatStr(t.Metadata.CreatedAt)},
		{"Updated At", formatStr(t.Metadata.UpdatedAt)},
		{lblLastSeenAt, formatStr(t.Metadata.LastSeenAt)},
	}
	return writeTable(w, detailHeaders, rows)
}

// RenderTeamAssignments writes a list table of team assignments plus a
// pagination footer.
func RenderTeamAssignments(w io.Writer, items []types.TeamAssignment, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No team assignments found.")
	}
	headers := []string{colTeam, "EntityType", "EntityID", colCluster, "Source"}
	rows := make([][]string, 0, len(items))
	for _, a := range items {
		rows = append(rows, []string{
			formatStr(a.Metadata.Team.Name),
			formatStr(a.Metadata.EntityType),
			formatStr(a.Metadata.EntityIdentifier),
			formatStr(a.Metadata.Cluster.Name),
			formatStr(a.Metadata.Source),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderDepartments writes a list table of departments plus a pagination footer.
func RenderDepartments(w io.Writer, items []types.Department, meta *types.Meta) error {
	if len(items) == 0 {
		return writeEmpty(w, "No departments found.")
	}
	headers := []string{colID, colName, colOrigin, "Teams", colDollarHr}
	rows := make([][]string, 0, len(items))
	for _, d := range items {
		rows = append(rows, []string{
			d.ID,
			d.Metadata.Name,
			formatStr(d.Metadata.Origin),
			formatIntPlain(d.Teams),
			FormatMoney(d.Cost.CurrentRunRateHourly),
		})
	}
	if err := writeTable(w, headers, rows); err != nil {
		return err
	}
	return writeFooter(w, len(items), meta)
}

// RenderDepartment writes a detail key-value table for one department.
func RenderDepartment(w io.Writer, d types.Department) error {
	rows := [][]string{
		{"ID", d.ID},
		{colName, d.Metadata.Name},
		{lblDescription, formatStr(d.Metadata.Description)},
		{colOrigin, formatStr(d.Metadata.Origin)},
		{"Owner Email", formatStr(d.Metadata.OwnerEmail)},
		{"Teams", formatIntPlain(d.Teams)},
		{"Assigned Workloads", formatIntPlain(d.AssignedWorkloads)},
		{"Assigned PVs", formatIntPlain(d.AssignedPVs)},
		{colDollarHr, FormatMoney(d.Cost.CurrentRunRateHourly)},
		{lblCostMode, formatStr(d.Cost.CostMode)},
		{lblCreatedAt, formatStr(d.Metadata.CreatedAt)},
		{"Updated At", formatStr(d.Metadata.UpdatedAt)},
	}
	return writeTable(w, detailHeaders, rows)
}

// RenderOrganization writes a detail key-value table for the org snapshot.
func RenderOrganization(w io.Writer, o types.Organization) error {
	rows := [][]string{
		{"ID", o.ID},
		{colName, o.Metadata.Name},
		{"Domain", formatStr(o.Metadata.Domain)},
		{"Plan", formatStr(o.Metadata.PlanType)},
		{"Active", formatBool(o.Metadata.IsActive)},
		{lblCreatedAt, formatStr(o.Metadata.CreatedAt)},
		{lblCPUCores, FormatCores(o.Capacity.CPU.TotalCores)},
		{lblCPUAllocatable, FormatCores(o.Capacity.CPU.AllocatableCores)},
		{lblCPUUsed, FormatCores(o.Utilization.CPU.UsedCores)},
		{lblCPUUtil, FormatPercentage(o.Utilization.CPU.UtilizationPercent)},
		{lblMemory, FormatBytes(o.Capacity.Memory.TotalBytes)},
		{lblMemoryAlloc, FormatBytes(o.Capacity.Memory.AllocatableBytes)},
		{lblMemoryUsed, FormatBytes(o.Utilization.Memory.UsedBytes)},
		{lblMemoryUtil, FormatPercentage(o.Utilization.Memory.UtilizationPercent)},
		{lblGPUTotal, formatIntPlain(o.Capacity.GPU.Total)},
		{"Storage", FormatBytes(o.Capacity.Storage.TotalBytes)},
		{"Clusters", formatIntPlain(o.Utilization.Counts.Clusters)},
		{"Connected Clusters", formatIntPlain(o.Utilization.Counts.ConnectedClusters)},
		{lblNodes, formatIntPlain(o.Utilization.Counts.Nodes)},
		{"Namespaces", formatIntPlain(o.Utilization.Counts.Namespaces)},
		{lblWorkloads, formatIntPlain(o.Utilization.Counts.Workloads)},
		{lblPods, formatIntPlain(o.Utilization.Counts.Pods)},
		{lblContainers, formatIntPlain(o.Utilization.Counts.Containers)},
		{"Persistent Volumes", formatIntPlain(o.Utilization.Counts.PersistentVolumes)},
		{"Recommendations", formatIntPlain(o.Utilization.Counts.Recommendations)},
		{colDollarHr, FormatMoney(o.Cost.CurrentRunRateHourly)},
		{lblCostUpdatedAt, formatStr(o.Cost.LastUpdatedAt)},
	}
	return writeTable(w, detailHeaders, rows)
}

// RenderOrganizationDashboard writes the dashboard view: snapshot summary,
// MTD/savings/calendar, top clusters, and recommendation summary tables.
func RenderOrganizationDashboard(w io.Writer, d types.OrganizationDashboard) error {
	rows := [][]string{
		{"Organization ID", d.OrganizationID},
		{colName, d.Snapshot.Metadata.Name},
		{"Plan", formatStr(d.Snapshot.Metadata.PlanType)},
		{"Run Rate $/hr", FormatMoney(d.Snapshot.Cost.CurrentRunRateHourly)},
		{"MTD Billed Cost", FormatMoney(d.MonthToDate.BilledCost)},
		{"Calendar Month", formatStr(d.MonthToDate.Calendar.Month)},
		{"Days Elapsed", formatIntPlain(d.MonthToDate.Calendar.DaysElapsed)},
		{"Days In Month", formatIntPlain(d.MonthToDate.Calendar.DaysInMonth)},
		{"Month Start At", formatStr(d.MonthToDate.Calendar.MonthStartAt)},
		{"Current Hourly Savings Potential", FormatMoney(d.Savings.CurrentHourlyPotential)},
		{"Recommendation Count", formatIntPlain(d.Savings.RecommendationCount)},
		{"Clusters", formatIntPlain(d.Snapshot.Utilization.Counts.Clusters)},
		{"Connected Clusters", formatIntPlain(d.Snapshot.Utilization.Counts.ConnectedClusters)},
	}
	if err := writeTable(w, detailHeaders, rows); err != nil {
		return err
	}

	if len(d.TopClusters) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("writing top clusters header: %w", err)
		}
		topHeaders := []string{colCluster, colDollarHr, "MTD Cost"}
		topRows := make([][]string, 0, len(d.TopClusters))
		for _, c := range d.TopClusters {
			topRows = append(topRows, []string{
				formatStr(c.Cluster.Name),
				FormatMoney(c.CurrentRunRateHourly),
				FormatMoney(c.MonthToDateCost),
			})
		}
		if err := writeTable(w, topHeaders, topRows); err != nil {
			return err
		}
	}

	return nil
}
