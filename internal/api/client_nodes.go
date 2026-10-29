package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// The endpoint rejects cost_mode, so this struct has no CostModeOpt.
// ClusterIDs picks the scoped vs flat path.
type NodeFilter struct {
	PagedOpts
	ClusterIDs   []string // single -> /v1/clusters/{cid}/nodes; multi/empty -> /v1/nodes
	NodeGroups   []string // csv
	InstanceType string
	Zone         string
	IsSpot       *bool
	IsReady      *bool
	Architecture string
	CapacityType string // on-demand|spot
}

// One ClusterID uses the scoped path; otherwise the flat path with the
// cluster_id list forwarded as CSV.
func (c *Client) ListNodes(
	ctx context.Context, f NodeFilter,
) ([]types.Node, *types.Meta, error) {
	path, csvParam := pickScopedOrFlat(
		func(id string) string { return "/v1/clusters/" + url.PathEscape(id) + "/nodes" },
		"/v1/nodes",
		f.ClusterIDs,
	)
	params := url.Values{}
	if csvParam != "" {
		params.Set("cluster_id", csvParam)
	}
	setCSV(params, "node_group", f.NodeGroups)
	if f.InstanceType != "" {
		params.Set("instance_type", f.InstanceType)
	}
	if f.Zone != "" {
		params.Set("zone", f.Zone)
	}
	setTriBool(params, "is_spot", f.IsSpot)
	setTriBool(params, "is_ready", f.IsReady)
	if f.Architecture != "" {
		params.Set("architecture", f.Architecture)
	}
	if f.CapacityType != "" {
		params.Set("capacity_type", f.CapacityType)
	}
	params = appendCursorParams(params, f.PagedOpts)
	return DoEnvelopeGet[[]types.Node](ctx, c, path, params)
}

// ClusterID picks one cluster when the node UID exists in more than one.
type NodeGetOpts struct {
	ClusterID string
}

func (c *Client) GetNode(
	ctx context.Context, nodeUID string, opts NodeGetOpts,
) (*types.Node, *types.Meta, error) {
	var params url.Values
	if opts.ClusterID != "" {
		params = url.Values{"cluster_id": {opts.ClusterID}}
	}
	return DoEnvelopeGet[*types.Node](ctx, c, "/v1/nodes/"+url.PathEscape(nodeUID), params)
}
