package api

import (
	"context"
	"net/url"
	"strconv"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

type OrganizationDashboardOpts struct {
	CostModeOpt
	TopClustersLimit int // 0 -> server default (5); 1-20 valid
}

// The endpoint rejects cost_mode, so no opts are exposed.
func (c *Client) GetOrganization(ctx context.Context) (*types.Organization, *types.Meta, error) {
	return DoEnvelopeGet[*types.Organization](ctx, c, "/v1/organization", nil)
}

// TopClustersLimit is forwarded as ?top_clusters_limit=N only when non-zero.
func (c *Client) GetOrganizationDashboard(
	ctx context.Context, opts OrganizationDashboardOpts,
) (*types.OrganizationDashboard, *types.Meta, error) {
	params := url.Values{}
	setCostMode(params, opts.CostMode)
	if opts.TopClustersLimit > 0 {
		params.Set("top_clusters_limit", strconv.Itoa(opts.TopClustersLimit))
	}
	return DoEnvelopeGet[*types.OrganizationDashboard](ctx, c, "/v1/organization/dashboard", params)
}
