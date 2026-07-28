package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// The endpoint rejects cost_mode, so this struct has no CostModeOpt.
type NodeGroupFilter struct {
	PagedOpts
	ClusterIDs []string // single -> /v1/clusters/{cid}/node-groups; multi/empty -> /v1/node-groups
}

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

func (c *Client) GetNodeGroup(
	ctx context.Context, clusterID, name string,
) (*types.NodeGroup, *types.Meta, error) {
	path := "/v1/clusters/" + clusterID + "/node-groups/" + name
	return DoEnvelopeGet[*types.NodeGroup](ctx, c, path, nil)
}
