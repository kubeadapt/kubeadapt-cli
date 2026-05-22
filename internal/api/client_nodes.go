package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// NodeFilter narrows the result set of ListNodes. The endpoint REJECTS
// cost_mode (node has a single physical bill), so this struct has no
// CostModeOpt. ClusterIDs picks scoped vs flat path.
type NodeFilter struct {
	PagedOpts
	ClusterIDs   []string // single → /v1/clusters/{cid}/nodes; multi/empty → /v1/nodes
	NodeGroups   []string // csv
	InstanceType string
	Zone         string
	IsSpot       *bool
	IsReady      *bool
	Architecture string
	CapacityType string // on-demand|spot
}

// ListNodes lists nodes. With a single ClusterID it calls the scoped path;
// otherwise it calls the flat path and forwards the cluster_id list as CSV.
func (c *Client) ListNodes(
	ctx context.Context, f NodeFilter,
) ([]types.Node, *types.Meta, error) {
	path, csvParam := pickScopedOrFlat(
		func(id string) string { return "/v1/clusters/" + id + "/nodes" },
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
	params = appendCursorParams(params, f.Cursor, f.Limit, f.IncludeTotal)
	return DoEnvelopeGet[[]types.Node](ctx, c, path, params)
}

// GetNode fetches a node detail via GET /v1/nodes/{node_uid}. The endpoint
// rejects cost_mode.
func (c *Client) GetNode(
	ctx context.Context, nodeUID string,
) (*types.Node, *types.Meta, error) {
	return DoEnvelopeGet[*types.Node](ctx, c, "/v1/nodes/"+nodeUID, nil)
}
