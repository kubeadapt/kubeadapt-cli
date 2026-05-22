package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// NamespaceFilter narrows the result set of ListNamespaces. ClusterIDs picks
// between the scoped (/v1/clusters/{cid}/namespaces) and the flat
// (/v1/namespaces) endpoint via pickScopedOrFlat.
type NamespaceFilter struct {
	PagedOpts
	CostModeOpt
	ClusterIDs    []string // single → scoped; multi/empty → flat
	MinCostHourly string   // decimal string for `min_cost_hourly`
}

// NamespaceGetOpts captures optional query parameters accepted by
// GET /v1/clusters/{cid}/namespaces/{name}.
type NamespaceGetOpts struct {
	CostModeOpt
}

// ListNamespaces lists namespaces. With a single ClusterID it calls the
// scoped path; with zero or multiple it calls the flat path and forwards
// the cluster_id list as a CSV query param.
func (c *Client) ListNamespaces(
	ctx context.Context, f NamespaceFilter,
) ([]types.Namespace, *types.Meta, error) {
	path, csvParam := pickScopedOrFlat(
		func(id string) string { return "/v1/clusters/" + id + "/namespaces" },
		"/v1/namespaces",
		f.ClusterIDs,
	)
	params := url.Values{}
	if csvParam != "" {
		params.Set("cluster_id", csvParam)
	}
	setCostMode(params, f.CostMode)
	if f.MinCostHourly != "" {
		params.Set("min_cost_hourly", f.MinCostHourly)
	}
	params = appendCursorParams(params, f.Cursor, f.Limit, f.IncludeTotal)
	return DoEnvelopeGet[[]types.Namespace](ctx, c, path, params)
}

// GetNamespace fetches a namespace detail via
// GET /v1/clusters/{cluster_id}/namespaces/{namespace}.
func (c *Client) GetNamespace(
	ctx context.Context, clusterID, namespace string, opts NamespaceGetOpts,
) (*types.Namespace, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, opts.CostMode)
	path := "/v1/clusters/" + clusterID + "/namespaces/" + namespace
	return DoEnvelopeGet[*types.Namespace](ctx, c, path, params)
}
