package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// The endpoint rejects cost_mode, so this struct deliberately has no
// CostModeOpt - callers cannot send the param at all.
type ClusterFilter struct {
	PagedOpts
	Provider    string // "" | aws | gcp | azure | on-prem
	Region      string
	Environment string // "" | production | non-production | staging | dev
	Status      string // "" | pending | active | disconnected | error | discovered
}

// ListClusters returns only clusters visible to the current API key.
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

func (c *Client) GetCluster(
	ctx context.Context, clusterID string,
) (*types.Cluster, *types.Meta, error) {
	return DoEnvelopeGet[*types.Cluster](ctx, c, "/v1/clusters/"+clusterID, nil)
}
