package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getRecommendationsCmd = &cobra.Command{
	Use:   "recommendations",
	Short: "List recommendations",
	Long: `List cost-saving recommendations. Filter by type, status, risk level,
priority, resource type, cluster, namespace, workload, or minimum savings.`,
	Args: cobra.NoArgs,
	Example: `  kubeadapt get recommendations
  kubeadapt get recommendations --priority high --status pending
  kubeadapt get recommendations --recommendation-type workload_rightsizing --risk-level low
  kubeadapt get recommendations --min-savings-hourly 0.10 --paginate`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := rejectCostMode(cmd, api.EndpointRecommendations); err != nil {
			return err
		}
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		paged, err := parsePagedFlags(cmd)
		if err != nil {
			return err
		}

		clusterIDs, _ := cmd.Flags().GetStringSlice("cluster-id")
		namespaces, _ := cmd.Flags().GetStringSlice("namespace")
		recType, _ := cmd.Flags().GetString("recommendation-type")
		status, _ := cmd.Flags().GetString("status")
		risk, _ := cmd.Flags().GetString("risk-level")
		priority, _ := cmd.Flags().GetString("priority")
		resType, _ := cmd.Flags().GetString("resource-type")
		workloadUIDs, _ := cmd.Flags().GetStringSlice("workload-uid")
		minSavings, _ := cmd.Flags().GetString("min-savings-hourly")

		fetch := func(ctx context.Context, cursor string) ([]types.Recommendation, *types.Meta, error) {
			return c.ListRecommendations(ctx, api.RecommendationFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				ClusterIDs:         clusterIDs,
				Namespaces:         namespaces,
				RecommendationType: recType,
				Status:             status,
				RiskLevel:          risk,
				Priority:           priority,
				ResourceType:       resType,
				WorkloadUIDs:       workloadUIDs,
				MinSavingsHourly:   minSavings,
			})
		}
		return runPaginatedList(cmd, c, paged, fetch, output.RenderRecommendations)
	},
}

func init() {
	f := getRecommendationsCmd.Flags()
	f.StringSlice("cluster-id", nil, clusterIDFilterUsage)
	f.StringSlice("namespace", nil, "Filter by namespace (repeatable)")
	f.String("recommendation-type", "", "Filter by recommendation type (workload_rightsizing)")
	f.String("status", "", "Filter by status (pending|applied|dismissed|archived)")
	f.String("risk-level", "", "Filter by risk level (low|medium|high)")
	f.String("priority", "", "Filter by priority (low|medium|high)")
	f.String("resource-type", "", "Filter by resource type (Deployment|StatefulSet|DaemonSet|Pod|Node)")
	f.StringSlice("workload-uid", nil, "Filter by workload UID (repeatable)")
	f.String("min-savings-hourly", "", "Minimum hourly savings (decimal)")
	registerClusterIDFlag(getRecommendationsCmd)
	registerEnumFlag(getRecommendationsCmd, "recommendation-type", "workload_rightsizing")
	registerEnumFlag(getRecommendationsCmd, "status", "pending", "applied", "dismissed", "archived")
	registerEnumFlag(getRecommendationsCmd, "risk-level", "low", "medium", "high")
	registerEnumFlag(getRecommendationsCmd, "priority", "low", "medium", "high")
	registerEnumFlag(getRecommendationsCmd, "resource-type", "Deployment", "StatefulSet", "DaemonSet", "Pod", "Node")
	getCmd.AddCommand(getRecommendationsCmd)
}
