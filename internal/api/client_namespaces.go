package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// ClusterIDs selects the scoped vs flat endpoint via pickScopedOrFlat.
type NamespaceFilter struct {
	PagedOpts
	CostModeOpt
	ClusterIDs    []string // single -> scoped; multi/empty -> flat
	MinCostHourly string   // decimal string for `min_cost_hourly`
}

type NamespaceGetOpts struct {
	CostModeOpt
}

// One ClusterID uses the scoped path; zero or many use the flat path with the
// cluster_id list forwarded as CSV.
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

func (c *Client) GetNamespace(
	ctx context.Context, clusterID, namespace string, opts NamespaceGetOpts,
) (*types.Namespace, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, opts.CostMode)
	path := "/v1/clusters/" + clusterID + "/namespaces/" + namespace
	return DoEnvelopeGet[*types.Namespace](ctx, c, path, params)
}
