package tui

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

// View identifiers.
type ViewID int

const (
	ViewOverview ViewID = iota
	ViewClusters
	ViewWorkloads
	ViewNodes
	ViewRecommendations
	ViewCosts
	ViewNamespaces
	ViewNodeGroups
	ViewPVs
	ViewHelp
)

// SwitchViewMsg tells the app to switch to a different view.
type SwitchViewMsg struct {
	View ViewID
}

// DataLoadedMsg carries loaded API data.
type DataLoadedMsg struct {
	View ViewID
	Data interface{}
	Err  error
}

// OverviewData holds overview data.
type OverviewData struct {
	Overview *types.OverviewResponse
}

// ClustersData holds clusters data.
type ClustersData struct {
	Clusters []types.ClusterResponse
}

// WorkloadsData holds workloads data.
type WorkloadsData struct {
	Workloads []types.WorkloadResponse
	Total     int
}

// NodesData holds nodes data.
type NodesData struct {
	Nodes []types.NodeResponse
	Total int
}

// RecommendationsData holds recommendations data.
type RecommendationsData struct {
	Recommendations []types.RecommendationResponse
	Total           int
}

// CostsData holds costs data.
type CostsData struct {
	Teams       []types.TeamCostResponse
	Departments []types.DepartmentCostResponse
}

// NamespacesData holds namespaces data.
type NamespacesData struct {
	Namespaces []types.NamespaceResponse
}

// WindowSizeMsg is sent when the terminal is resized.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// PushDetailMsg tells the app to push a detail view onto the navigation stack.
type PushDetailMsg struct {
	View       ViewInterface
	Breadcrumb string // e.g. "prod-us-east"
}

// PopDetailMsg tells the app to pop the top detail view from the navigation stack.
type PopDetailMsg struct{}

// DetailDataLoadedMsg carries data for a detail view.
type DetailDataLoadedMsg struct {
	EntityType string
	EntityID   string
	Data       interface{}
	Err        error
}

// FilterChangedMsg is sent when the filter query changes.
type FilterChangedMsg struct {
	Query string
}

// FilterClearedMsg is sent when the filter is cleared.
type FilterClearedMsg struct{}
