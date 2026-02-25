package api

import (
	"context"
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
)

func TestGetOverview(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetOverview(context.Background())
	if err != nil {
		t.Fatalf("GetOverview() error: %v", err)
	}
	if resp.ClusterCount != 3 {
		t.Errorf("expected 3 clusters, got %d", resp.ClusterCount)
	}
	if resp.TotalNodes != 15 {
		t.Errorf("expected 15 nodes, got %d", resp.TotalNodes)
	}
	if resp.MTDActualCost == nil {
		t.Error("expected MTDActualCost to be non-nil")
	}
	if resp.RunRate == nil {
		t.Error("expected RunRate to be non-nil")
	}
	if resp.EfficiencyScore == nil {
		t.Error("expected EfficiencyScore to be non-nil")
	}
}

func TestGetClusters(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetClusters(context.Background())
	if err != nil {
		t.Fatalf("GetClusters() error: %v", err)
	}
	if len(resp.Clusters) != 2 {
		t.Errorf("expected 2 clusters, got %d", len(resp.Clusters))
	}
	if resp.Clusters[0].Name != "production-us" {
		t.Errorf("expected first cluster name 'production-us', got %q", resp.Clusters[0].Name)
	}
	if resp.Clusters[0].EfficiencyScore == nil {
		t.Error("expected EfficiencyScore to be non-nil")
	}
	if resp.Clusters[0].MonthlyCost == nil {
		t.Error("expected MonthlyCost to be non-nil")
	}
}

func TestGetCluster(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetCluster(context.Background(), "cls-001")
	if err != nil {
		t.Fatalf("GetCluster() error: %v", err)
	}
	if resp.ID != "cls-001" {
		t.Errorf("expected ID 'cls-001', got %q", resp.ID)
	}
}

func TestUnauthorized(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "") // no key
	_, err := client.GetOverview(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsAuthError() {
		t.Errorf("expected auth error, got status %d", apiErr.StatusCode)
	}
}

func TestGetNodes(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetNodes(context.Background(), "", "", 0, 0)
	if err != nil {
		t.Fatalf("GetNodes() error: %v", err)
	}
	if len(resp.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(resp.Nodes))
	}
	if resp.Nodes[0].NodeName != "ip-10-0-1-42.ec2.internal" {
		t.Errorf("expected first node name 'ip-10-0-1-42.ec2.internal', got %q", resp.Nodes[0].NodeName)
	}
	if resp.Nodes[0].PodCount == nil {
		t.Error("expected PodCount to be non-nil")
	}
	if resp.Nodes[0].MonthlyCost == nil {
		t.Error("expected MonthlyCost to be non-nil")
	}
}

func TestGetWorkloads(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetWorkloads(context.Background(), "", "", "", 0, 0)
	if err != nil {
		t.Fatalf("GetWorkloads() error: %v", err)
	}
	if len(resp.Workloads) != 2 {
		t.Errorf("expected 2 workloads, got %d", len(resp.Workloads))
	}
	if resp.Workloads[0].WorkloadName != "api-gateway" {
		t.Errorf("expected first workload name 'api-gateway', got %q", resp.Workloads[0].WorkloadName)
	}
	if resp.Workloads[0].EfficiencyScore == nil {
		t.Error("expected EfficiencyScore to be non-nil")
	}
	if resp.Workloads[0].MonthlyCost == nil {
		t.Error("expected MonthlyCost to be non-nil")
	}
}

func TestGetNamespaces(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetNamespaces(context.Background(), "", "", "")
	if err != nil {
		t.Fatalf("GetNamespaces() error: %v", err)
	}
	if len(resp.Namespaces) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(resp.Namespaces))
	}
	if resp.Namespaces[0].Name != "default" {
		t.Errorf("expected first namespace name 'default', got %q", resp.Namespaces[0].Name)
	}
	if resp.Namespaces[0].EfficiencyScore == nil {
		t.Error("expected EfficiencyScore to be non-nil")
	}
	if resp.Namespaces[0].MonthlyCost == nil {
		t.Error("expected MonthlyCost to be non-nil")
	}
	if resp.Namespaces[0].ContainerCount == nil {
		t.Error("expected ContainerCount to be non-nil")
	}
}

func TestGetRecommendations(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetRecommendations(context.Background(), "", "", "", 0, 0)
	if err != nil {
		t.Fatalf("GetRecommendations() error: %v", err)
	}
	if len(resp.Recommendations) != 2 {
		t.Errorf("expected 2 recommendations, got %d", len(resp.Recommendations))
	}
	if resp.Recommendations[0].Priority == nil {
		t.Error("expected Priority to be non-nil")
	}
}

func TestGetDashboard(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetDashboard(context.Background(), 30)
	if err != nil {
		t.Fatalf("GetDashboard() error: %v", err)
	}
	if resp.ClusterCount != 3 {
		t.Errorf("expected 3 clusters, got %d", resp.ClusterCount)
	}
	if resp.TotalMonthlyCost != 9125.00 {
		t.Errorf("expected monthly cost 9125.00, got %f", resp.TotalMonthlyCost)
	}
	if len(resp.TopClusters) != 2 {
		t.Errorf("expected 2 top clusters, got %d", len(resp.TopClusters))
	}
	if resp.EfficiencyScore == nil {
		t.Error("expected EfficiencyScore to be non-nil")
	}
}

func TestGetClusterDashboard(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetClusterDashboard(context.Background(), "cls-001")
	if err != nil {
		t.Fatalf("GetClusterDashboard() error: %v", err)
	}
	if resp.ClusterID != "cls-001" {
		t.Errorf("expected cluster ID 'cls-001', got %q", resp.ClusterID)
	}
	if resp.CostBreakdown == nil {
		t.Error("expected CostBreakdown to be non-nil")
	}
	if resp.MTDActualCost == nil {
		t.Error("expected MTDActualCost to be non-nil")
	}
	if len(resp.RecommendationSummary) == 0 {
		t.Error("expected RecommendationSummary to be non-empty")
	}
}

func TestGetCapacityPlanning(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	resp, err := client.GetCapacityPlanning(context.Background(), "cls-001")
	if err != nil {
		t.Fatalf("GetCapacityPlanning() error: %v", err)
	}
	if resp.ClusterID != "cls-001" {
		t.Errorf("expected cluster ID 'cls-001', got %q", resp.ClusterID)
	}
	if resp.SpotVsOnDemand.Total != 10 {
		t.Errorf("expected 10 total nodes, got %d", resp.SpotVsOnDemand.Total)
	}
	if resp.PodDensity.TotalPods != 85 {
		t.Errorf("expected 85 total pods, got %d", resp.PodDensity.TotalPods)
	}
	if len(resp.NodeGroups) != 2 {
		t.Errorf("expected 2 node groups, got %d", len(resp.NodeGroups))
	}
}

func TestGetCluster404(t *testing.T) {
	server := testutil.NewMockServer()
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	_, err := client.GetCluster(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent cluster")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsNotFound() {
		t.Errorf("expected not found error, got status %d", apiErr.StatusCode)
	}
}
