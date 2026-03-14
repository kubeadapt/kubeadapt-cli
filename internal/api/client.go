package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

const defaultTimeout = 30 * time.Second

// Client is the Kubeadapt API client.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

// Option configures the Client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithLogger sets the debug logger.
func WithLogger(l *zap.Logger) Option {
	return func(c *Client) {
		c.logger = l
	}
}

// NewClient creates a new API client.
func NewClient(baseURL, apiKey string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger: zap.NewNop(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	start := time.Now()
	c.logger.Debug("API request", zap.String("method", method), zap.String("path", path))

	u := strings.TrimRight(c.baseURL, "/") + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if err := json.Unmarshal(respBody, apiErr); err != nil {
			apiErr.Message = string(respBody)
		}
		c.logger.Debug("API error",
			zap.String("path", path),
			zap.Int("status", resp.StatusCode),
			zap.Duration("duration", time.Since(start)),
		)
		return apiErr
	}

	c.logger.Debug("API response",
		zap.String("path", path),
		zap.Int("status", resp.StatusCode),
		zap.Duration("duration", time.Since(start)),
	)

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

func (c *Client) get(ctx context.Context, path string, params url.Values, result any) error {
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// GetOverview fetches the organization overview.
func (c *Client) GetOverview(ctx context.Context) (*types.OverviewResponse, error) {
	var resp types.OverviewResponse
	err := c.get(ctx, "/v1/overview", nil, &resp)
	return &resp, err
}

// GetDashboard fetches the organization dashboard.
func (c *Client) GetDashboard(ctx context.Context, days int) (*types.DashboardResponse, error) {
	params := url.Values{}
	if days > 0 {
		params.Set("days", strconv.Itoa(days))
	}
	var resp types.DashboardResponse
	err := c.get(ctx, "/v1/dashboard", params, &resp)
	return &resp, err
}

// GetClusters fetches all clusters.
func (c *Client) GetClusters(ctx context.Context) (*types.ClusterListResponse, error) {
	var resp types.ClusterListResponse
	err := c.get(ctx, "/v1/clusters", nil, &resp)
	return &resp, err
}

// GetCluster fetches a single cluster by ID.
func (c *Client) GetCluster(ctx context.Context, id string) (*types.ClusterResponse, error) {
	var resp types.ClusterResponse
	err := c.get(ctx, "/v1/clusters/"+id, nil, &resp)
	return &resp, err
}

// GetWorkloads fetches workloads with optional filters.
func (c *Client) GetWorkloads(ctx context.Context, clusterID, namespace, kind string, limit, offset int) (*types.WorkloadListResponse, error) {
	params := url.Values{}
	if clusterID != "" {
		params.Set("cluster_id", clusterID)
	}
	if namespace != "" {
		params.Set("namespace", namespace)
	}
	if kind != "" {
		params.Set("kind", kind)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	var resp types.WorkloadListResponse
	err := c.get(ctx, "/v1/workloads", params, &resp)
	return &resp, err
}

// GetNodes fetches nodes with optional filters.
func (c *Client) GetNodes(ctx context.Context, clusterID, nodeGroup string, limit, offset int) (*types.NodeListResponse, error) {
	params := url.Values{}
	if clusterID != "" {
		params.Set("cluster_id", clusterID)
	}
	if nodeGroup != "" {
		params.Set("node_group", nodeGroup)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	var resp types.NodeListResponse
	err := c.get(ctx, "/v1/nodes", params, &resp)
	return &resp, err
}

// GetRecommendations fetches recommendations with optional filters.
func (c *Client) GetRecommendations(ctx context.Context, clusterID, recType, status string, limit, offset int) (*types.RecommendationListResponse, error) {
	params := url.Values{}
	if clusterID != "" {
		params.Set("cluster_id", clusterID)
	}
	if recType != "" {
		params.Set("recommendation_type", recType)
	}
	if status != "" {
		params.Set("status", status)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	var resp types.RecommendationListResponse
	err := c.get(ctx, "/v1/recommendations", params, &resp)
	return &resp, err
}

// GetCostsTeams fetches cost breakdown by team.
func (c *Client) GetCostsTeams(ctx context.Context, clusterID string) (*types.TeamCostListResponse, error) {
	params := url.Values{}
	if clusterID != "" {
		params.Set("cluster_id", clusterID)
	}
	var resp types.TeamCostListResponse
	err := c.get(ctx, "/v1/costs/teams", params, &resp)
	return &resp, err
}

// GetCostsDepartments fetches cost breakdown by department.
func (c *Client) GetCostsDepartments(ctx context.Context, clusterID string) (*types.DepartmentCostListResponse, error) {
	params := url.Values{}
	if clusterID != "" {
		params.Set("cluster_id", clusterID)
	}
	var resp types.DepartmentCostListResponse
	err := c.get(ctx, "/v1/costs/departments", params, &resp)
	return &resp, err
}

// GetNodeGroups fetches node groups.
func (c *Client) GetNodeGroups(ctx context.Context, clusterID string) (*types.NodeGroupListResponse, error) {
	params := url.Values{}
	if clusterID != "" {
		params.Set("cluster_id", clusterID)
	}
	var resp types.NodeGroupListResponse
	err := c.get(ctx, "/v1/node-groups", params, &resp)
	return &resp, err
}

// GetNamespaces fetches namespaces.
func (c *Client) GetNamespaces(ctx context.Context, clusterID, team, department string) (*types.NamespaceListResponse, error) {
	params := url.Values{}
	if clusterID != "" {
		params.Set("cluster_id", clusterID)
	}
	if team != "" {
		params.Set("team", team)
	}
	if department != "" {
		params.Set("department", department)
	}
	var resp types.NamespaceListResponse
	err := c.get(ctx, "/v1/namespaces", params, &resp)
	return &resp, err
}

// GetPersistentVolumes fetches persistent volumes.
func (c *Client) GetPersistentVolumes(ctx context.Context, clusterID, namespace, storageClass string) (*types.PersistentVolumeListResponse, error) {
	params := url.Values{}
	if clusterID != "" {
		params.Set("cluster_id", clusterID)
	}
	if namespace != "" {
		params.Set("namespace", namespace)
	}
	if storageClass != "" {
		params.Set("storage_class", storageClass)
	}
	var resp types.PersistentVolumeListResponse
	err := c.get(ctx, "/v1/persistent-volumes", params, &resp)
	return &resp, err
}

// --- Metrics / Detail endpoints ---

// GetClusterDashboard fetches dashboard summary metrics for a cluster.
func (c *Client) GetClusterDashboard(ctx context.Context, clusterID string) (*types.ClusterDashboardResponse, error) {
	var resp types.ClusterDashboardResponse
	err := c.get(ctx, "/v1/clusters/"+clusterID+"/dashboard", nil, &resp)
	return &resp, err
}

// GetClusterCostDistribution fetches time-series cost/utilization data.
func (c *Client) GetClusterCostDistribution(ctx context.Context, clusterID, timeframe string) (*types.CostDistributionResponse, error) {
	params := url.Values{}
	if timeframe != "" {
		params.Set("timeframe", timeframe)
	}
	var resp types.CostDistributionResponse
	err := c.get(ctx, "/v1/clusters/"+clusterID+"/cost-distribution", params, &resp)
	return &resp, err
}

// GetNodeMetrics fetches time-series CPU/memory history for a node.
func (c *Client) GetNodeMetrics(ctx context.Context, nodeUID, clusterID, timeframe string) (*types.NodeMetricsResponse, error) {
	params := url.Values{}
	params.Set("cluster_id", clusterID)
	if timeframe != "" {
		params.Set("timeframe", timeframe)
	}
	var resp types.NodeMetricsResponse
	err := c.get(ctx, "/v1/nodes/"+nodeUID+"/metrics", params, &resp)
	return &resp, err
}

// GetNodeGroupDetails fetches detailed node group information.
func (c *Client) GetNodeGroupDetails(ctx context.Context, groupName, clusterID string) (*types.NodeGroupDetailResponse, error) {
	params := url.Values{}
	params.Set("cluster_id", clusterID)
	var resp types.NodeGroupDetailResponse
	err := c.get(ctx, "/v1/node-groups/"+groupName+"/details", params, &resp)
	return &resp, err
}

// GetWorkloadMetrics fetches time-series CPU/memory/cost trends for a workload.
func (c *Client) GetWorkloadMetrics(ctx context.Context, workloadUID, clusterID, timeframe string) (*types.WorkloadMetricsResponse, error) {
	params := url.Values{}
	params.Set("cluster_id", clusterID)
	if timeframe != "" {
		params.Set("timeframe", timeframe)
	}
	var resp types.WorkloadMetricsResponse
	err := c.get(ctx, "/v1/workloads/"+workloadUID+"/metrics", params, &resp)
	return &resp, err
}

// GetWorkloadNodes fetches node distribution for a workload.
func (c *Client) GetWorkloadNodes(ctx context.Context, workloadUID, clusterID string) (*types.WorkloadNodesResponse, error) {
	params := url.Values{}
	params.Set("cluster_id", clusterID)
	var resp types.WorkloadNodesResponse
	err := c.get(ctx, "/v1/workloads/"+workloadUID+"/nodes", params, &resp)
	return &resp, err
}

// GetNamespaceDetails fetches detailed namespace information.
func (c *Client) GetNamespaceDetails(ctx context.Context, name, clusterID string) (*types.NamespaceDetailResponse, error) {
	params := url.Values{}
	params.Set("cluster_id", clusterID)
	var resp types.NamespaceDetailResponse
	err := c.get(ctx, "/v1/namespaces/"+name+"/details", params, &resp)
	return &resp, err
}

// GetNamespaceTrends fetches time-series trends for a namespace.
func (c *Client) GetNamespaceTrends(ctx context.Context, name, clusterID, timeframe string) (*types.NamespaceTrendsResponse, error) {
	params := url.Values{}
	params.Set("cluster_id", clusterID)
	if timeframe != "" {
		params.Set("timeframe", timeframe)
	}
	var resp types.NamespaceTrendsResponse
	err := c.get(ctx, "/v1/namespaces/"+name+"/trends", params, &resp)
	return &resp, err
}
