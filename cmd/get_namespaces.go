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

var getNamespacesCmd = &cobra.Command{
	Use:   "namespaces",
	Short: "List namespaces",
	Long: `List namespaces. Use --cluster-id to scope (one cluster) or filter (multiple).

NOTE: --team and --department filters are no longer accepted by the new public
API. Filter via team-assignments or department-assignments queries instead.`,
	Args: cobra.NoArgs,
	Example: `  kubeadapt get namespaces
  kubeadapt get namespaces --cluster-id <id>
  kubeadapt get namespaces --min-cost-hourly 0.05 --paginate`,
	RunE: func(cmd *cobra.Command, _ []string) error {
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
		minCost, _ := cmd.Flags().GetString("min-cost-hourly")

		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		var all []types.Namespace
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.NamespaceFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				CostModeOpt:   api.CostModeOpt{CostMode: paged.CostMode},
				ClusterIDs:    clusterIDs,
				MinCostHourly: minCost,
			}
			items, meta, err := c.ListNamespaces(ctx, f)
			if err != nil {
				return fmt.Errorf("list namespaces: %w", err)
			}
			all = append(all, items...)
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
			return output.RenderJSONWithMeta(cmd.OutOrStdout(), all, lastMeta)
		case formatYAML:
			return output.RenderYAMLWithMeta(cmd.OutOrStdout(), all, lastMeta)
		default:
			return output.RenderNamespaces(cmd.OutOrStdout(), all, lastMeta)
		}
	},
}

func init() {
	getNamespacesCmd.Flags().StringSlice("cluster-id", nil, "Filter by cluster ID (repeatable; one becomes a scoped path)")
	getNamespacesCmd.Flags().String("min-cost-hourly", "", "Minimum hourly cost (decimal)")
	getCmd.AddCommand(getNamespacesCmd)
}
