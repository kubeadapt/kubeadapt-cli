package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// ClusterIDs selects the scoped vs flat endpoint via pickScopedOrFlat.
type WorkloadFilter struct {
	PagedOpts
	CostModeOpt
	ClusterIDs    []string // single -> /v1/clusters/{cid}/workloads; multi/empty -> /v1/workloads
	Namespaces    []string // csv
	Kinds         []string // csv: Deployment, StatefulSet, DaemonSet
	Teams         []string // csv
	Departments   []string // csv
	HasHPA        *bool    // tri-state
	MinCostHourly string
}

// ClusterID picks one cluster when the workload UID exists in more than one.
type WorkloadGetOpts struct {
	CostModeOpt
	ClusterID string
}

// ClusterID keeps one cluster's pods when the workload UID exists in several.
type PodFilter struct {
	PagedOpts
	CostModeOpt
	ClusterID   string
	Namespaces  []string // csv
	NodeUIDs    []string // csv
	Phase       string   // Pending|Running|Succeeded|Failed|Unknown
	QoSClass    string   // Guaranteed|Burstable|BestEffort
	HasHostPath *bool
	HasEmptyDir *bool
	HostNetwork *bool
}

// One ClusterID uses the scoped path; otherwise the flat path with the
// cluster_id list forwarded as CSV.
func (c *Client) ListWorkloads(
	ctx context.Context, f WorkloadFilter,
) ([]types.Workload, *types.Meta, error) {
	path, csvParam := pickScopedOrFlat(
		func(id string) string { return "/v1/clusters/" + url.PathEscape(id) + "/workloads" },
		"/v1/workloads",
		f.ClusterIDs,
	)
	params := url.Values{}
	if csvParam != "" {
		params.Set("cluster_id", csvParam)
	}
	setCostMode(params, f.CostMode)
	setCSV(params, "namespace", f.Namespaces)
	setCSV(params, "kind", f.Kinds)
	setCSV(params, "team", f.Teams)
	setCSV(params, "department", f.Departments)
	setTriBool(params, "has_hpa", f.HasHPA)
	if f.MinCostHourly != "" {
		params.Set("min_cost_hourly", f.MinCostHourly)
	}
	params = appendCursorParams(params, f.PagedOpts)
	return DoEnvelopeGet[[]types.Workload](ctx, c, path, params)
}

func (c *Client) GetWorkload(
	ctx context.Context, workloadUID string, opts WorkloadGetOpts,
) (*types.Workload, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, opts.CostMode)
	if opts.ClusterID != "" {
		params.Set("cluster_id", opts.ClusterID)
	}
	return DoEnvelopeGet[*types.Workload](ctx, c, "/v1/workloads/"+url.PathEscape(workloadUID), params)
}

func (c *Client) ListWorkloadPods(
	ctx context.Context, workloadUID string, f PodFilter,
) ([]types.Pod, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, f.CostMode)
	if f.ClusterID != "" {
		params.Set("cluster_id", f.ClusterID)
	}
	setCSV(params, "namespace", f.Namespaces)
	setCSV(params, "node_uid", f.NodeUIDs)
	if f.Phase != "" {
		params.Set("phase", f.Phase)
	}
	if f.QoSClass != "" {
		params.Set("qos_class", f.QoSClass)
	}
	setTriBool(params, "has_hostpath", f.HasHostPath)
	setTriBool(params, "has_emptydir", f.HasEmptyDir)
	setTriBool(params, "host_network", f.HostNetwork)
	params = appendCursorParams(params, f.PagedOpts)
	return DoEnvelopeGet[[]types.Pod](ctx, c, "/v1/workloads/"+url.PathEscape(workloadUID)+"/pods", params)
}
