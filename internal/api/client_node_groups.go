package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// NodeGroupFilter narrows the result set of ListNodeGroups. The endpoint
// REJECTS cost_mode, so this struct has no CostModeOpt.
type NodeGroupFilter struct {
	PagedOpts
	ClusterIDs []string // single → /v1/clusters/{cid}/node-groups; multi/empty → /v1/node-groups
}

// ListNodeGroups lists node groups across one or more clusters.
func (c *Client) ListNodeGroups(
	ctx context.Context, f NodeGroupFilter,
) ([]types.NodeGroup, *types.Meta, error) {
	path, csvParam := pickScopedOrFlat(
		func(id string) string { return "/v1/clusters/" + id + "/node-groups" },
		"/v1/node-groups",
		f.ClusterIDs,
	)
	params := url.Values{}
	if csvParam != "" {
		params.Set("cluster_id", csvParam)
	}
	params = appendCursorParams(params, f.Cursor, f.Limit, f.IncludeTotal)
	return DoEnvelopeGet[[]types.NodeGroup](ctx, c, path, params)
}

// GetNodeGroup fetches a node group detail via
// GET /v1/clusters/{cluster_id}/node-groups/{name}. The endpoint rejects
// cost_mode.
func (c *Client) GetNodeGroup(
	ctx context.Context, clusterID, name string,
) (*types.NodeGroup, *types.Meta, error) {
	path := "/v1/clusters/" + clusterID + "/node-groups/" + name
	return DoEnvelopeGet[*types.NodeGroup](ctx, c, path, nil)
}
