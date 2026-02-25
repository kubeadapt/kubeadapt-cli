package components

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

func TestSidebarRenderAllItems(t *testing.T) {
	s := NewSidebar()
	s.ActiveView = tui.ViewOverview

	output := s.Render()

	expectedLabels := []string{
		"Overview",
		"Clusters",
		"Workloads",
		"Nodes",
		"Recomm.",
		"Costs",
		"Namespc.",
		"NodeGrps",
		"PVs",
		"Help",
	}

	for _, label := range expectedLabels {
		if !strings.Contains(output, label) {
			t.Errorf("Render() output missing label %q", label)
		}
	}
}

func TestSidebarActiveIndicator(t *testing.T) {
	s := NewSidebar()
	s.ActiveView = tui.ViewClusters

	output := s.Render()

	// Active view should have filled circle indicator
	if !strings.Contains(output, "●") {
		t.Error("Render() output missing active indicator '●'")
	}

	// Active view should contain "Clusters"
	if !strings.Contains(output, "Clusters") {
		t.Error("Render() output missing 'Clusters' label")
	}

	// Inactive views should have empty circle indicator
	if !strings.Contains(output, "○") {
		t.Error("Render() output missing inactive indicator '○'")
	}

	// Count indicators - should have 1 active (●) and 9 inactive (○)
	activeCount := strings.Count(output, "●")
	inactiveCount := strings.Count(output, "○")

	if activeCount != 1 {
		t.Errorf("Expected 1 active indicator, got %d", activeCount)
	}
	if inactiveCount != 9 {
		t.Errorf("Expected 9 inactive indicators, got %d", inactiveCount)
	}
}

func TestSidebarWidth(t *testing.T) {
	s := NewSidebar()
	s.Width = 20
	s.ActiveView = tui.ViewOverview

	output := s.Render()

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		// Skip empty lines
		if line == "" {
			continue
		}

		// Visual width check - each line should be reasonable for width 20
		// We allow some flexibility for ANSI codes and border characters
		if len(line) > 200 { // Generous upper bound accounting for ANSI codes
			t.Errorf("Line %d appears too long (len=%d): %q", i, len(line), line)
		}
	}
}

func TestSidebarDifferentActiveViews(t *testing.T) {
	tests := []struct {
		viewID tui.ViewID
		label  string
	}{
		{tui.ViewOverview, "Overview"},
		{tui.ViewClusters, "Clusters"},
		{tui.ViewWorkloads, "Workloads"},
		{tui.ViewNodes, "Nodes"},
		{tui.ViewRecommendations, "Recomm."},
		{tui.ViewCosts, "Costs"},
		{tui.ViewNamespaces, "Namespc."},
		{tui.ViewNodeGroups, "NodeGrps"},
		{tui.ViewPVs, "PVs"},
		{tui.ViewHelp, "Help"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("view_%d", tt.viewID), func(t *testing.T) {
			s := NewSidebar()
			s.ActiveView = tt.viewID

			output := s.Render()

			// Should contain the label
			if !strings.Contains(output, tt.label) {
				t.Errorf("Render() output missing label %q for view %d", tt.label, tt.viewID)
			}

			// Should have exactly 1 active indicator
			activeCount := strings.Count(output, "●")
			if activeCount != 1 {
				t.Errorf("Expected 1 active indicator for view %d, got %d", tt.viewID, activeCount)
			}

			// Should have exactly 9 inactive indicators
			inactiveCount := strings.Count(output, "○")
			if inactiveCount != 9 {
				t.Errorf("Expected 9 inactive indicators for view %d, got %d", tt.viewID, inactiveCount)
			}
		})
	}
}
