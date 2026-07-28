package cmd

import (
	"context"

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

		fetch := func(ctx context.Context, cursor string) ([]types.Namespace, *types.Meta, error) {
			return c.ListNamespaces(ctx, api.NamespaceFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				CostModeOpt:   api.CostModeOpt{CostMode: paged.CostMode},
				ClusterIDs:    clusterIDs,
				MinCostHourly: minCost,
			})
		}
		return runPaginatedList(cmd, c, paged, fetch, output.RenderNamespaces)
	},
}

func init() {
	getNamespacesCmd.Flags().StringSlice("cluster-id", nil, clusterIDFilterUsage)
	getNamespacesCmd.Flags().String("min-cost-hourly", "", "Minimum hourly cost (decimal)")
	registerClusterIDFlag(getNamespacesCmd)
	getCmd.AddCommand(getNamespacesCmd)
}
