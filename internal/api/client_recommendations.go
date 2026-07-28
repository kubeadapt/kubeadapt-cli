package api

import (
	"context"
	"net/url"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// Savings are mode-agnostic, so the endpoint rejects cost_mode and this struct
// has no CostModeOpt.
type RecommendationFilter struct {
	PagedOpts
	ClusterIDs         []string // csv
	Namespaces         []string // csv
	RecommendationType string   // workload_rightsizing
	Status             string   // pending|applied|dismissed|archived
	RiskLevel          string   // low|medium|high
	Priority           string   // high|medium|low
	ResourceType       string   // Deployment|StatefulSet|DaemonSet|Pod|Node
	WorkloadUIDs       []string // csv
	MinSavingsHourly   string   // decimal string
}

func (c *Client) ListRecommendations(
	ctx context.Context, f RecommendationFilter,
) ([]types.Recommendation, *types.Meta, error) {
	params := url.Values{}
	setCSV(params, "cluster_id", f.ClusterIDs)
	setCSV(params, "namespace", f.Namespaces)
	if f.RecommendationType != "" {
		params.Set("recommendation_type", f.RecommendationType)
	}
	if f.Status != "" {
		params.Set("status", f.Status)
	}
	if f.RiskLevel != "" {
		params.Set("risk_level", f.RiskLevel)
	}
	if f.Priority != "" {
		params.Set("priority", f.Priority)
	}
	if f.ResourceType != "" {
		params.Set("resource_type", f.ResourceType)
	}
	setCSV(params, "workload_uid", f.WorkloadUIDs)
	if f.MinSavingsHourly != "" {
		params.Set("min_savings_hourly", f.MinSavingsHourly)
	}
	params = appendCursorParams(params, f.Cursor, f.Limit, f.IncludeTotal)
	return DoEnvelopeGet[[]types.Recommendation](ctx, c, "/v1/recommendations", params)
}

func (c *Client) GetRecommendation(
	ctx context.Context, recID string,
) (*types.Recommendation, *types.Meta, error) {
	return DoEnvelopeGet[*types.Recommendation](ctx, c, "/v1/recommendations/"+recID, nil)
}
