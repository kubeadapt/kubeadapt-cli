package cmd

import (
	"context"
	"fmt"
	"time"

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
		if cmd.Flags().Changed("cost-mode") {
			return fmt.Errorf("--cost-mode is not accepted by the recommendations endpoint")
		}
		rctx := getRunContext(cmd)
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

		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		var allItems []types.Recommendation
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.RecommendationFilter{
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
			}
			items, meta, err := c.ListRecommendations(ctx, f)
			if err != nil {
				return fmt.Errorf("list recommendations: %w", err)
			}
			allItems = append(allItems, items...)
			lastMeta = meta
			if !paged.Paginate || meta == nil || meta.Pagination == nil || !meta.Pagination.HasMore {
				break
			}
			cursor = meta.Pagination.NextCursor
			if cursor == "" {
				break
			}
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSONWithMeta(cmd.OutOrStdout(), allItems, lastMeta)
		case formatYAML:
			return output.RenderYAMLWithMeta(cmd.OutOrStdout(), allItems, lastMeta)
		default:
			return output.RenderRecommendations(cmd.OutOrStdout(), allItems, lastMeta)
		}
	},
}

func init() {
	f := getRecommendationsCmd.Flags()
	f.StringSlice("cluster-id", nil, "Filter by cluster ID (repeatable)")
	f.StringSlice("namespace", nil, "Filter by namespace (repeatable)")
	f.String("recommendation-type", "", "workload_rightsizing")
	f.String("status", "", "pending|applied|dismissed|archived")
	f.String("risk-level", "", "low|medium|high")
	f.String("priority", "", "high|medium|low")
	f.String("resource-type", "", "Deployment|StatefulSet|DaemonSet|Pod|Node")
	f.StringSlice("workload-uid", nil, "Filter by workload UID (repeatable)")
	f.String("min-savings-hourly", "", "Minimum hourly savings (decimal)")
	getCmd.AddCommand(getRecommendationsCmd)
}
