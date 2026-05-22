package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// ClusterFilter narrows the result set of ListClusters. The endpoint REJECTS
// cost_mode (cluster cost is a single physical number) so the struct has no
// CostModeOpt — there is no way for a caller to send the param via this type.
type ClusterFilter struct {
	PagedOpts
	Provider    string // "" | aws | gcp | azure | on-prem
	Region      string
	Environment string // "" | production | non-production | staging | dev
	Status      string // "" | pending | active | disconnected | error | discovered
}

// ListClusters lists clusters visible to the current API key. Pagination is
// cursor-based via f.PagedOpts; pass an empty Cursor for the first page and
// use Meta.Pagination.NextCursor on subsequent calls.
func (c *Client) ListClusters(
	ctx context.Context, f ClusterFilter,
) ([]types.Cluster, *types.Meta, error) {
	params := url.Values{}
	if f.Provider != "" {
		params.Set("provider", f.Provider)
	}
	if f.Region != "" {
		params.Set("region", f.Region)
	}
	if f.Environment != "" {
		params.Set("environment", f.Environment)
	}
	if f.Status != "" {
		params.Set("status", f.Status)
	}
	params = appendCursorParams(params, f.Cursor, f.Limit, f.IncludeTotal)
	return DoEnvelopeGet[[]types.Cluster](ctx, c, "/v1/clusters", params)
}

// GetCluster fetches a single cluster by ID via GET /v1/clusters/{id}. The
// endpoint rejects cost_mode.
func (c *Client) GetCluster(
	ctx context.Context, clusterID string,
) (*types.Cluster, *types.Meta, error) {
	return DoEnvelopeGet[*types.Cluster](ctx, c, "/v1/clusters/"+clusterID, nil)
}
