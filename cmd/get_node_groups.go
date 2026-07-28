package cmd

import (
	"context"
	"fmt"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNodeGroupsCmd = &cobra.Command{
	Use:   "node-groups",
	Short: "List node groups",
	Long: `List node groups across one or more clusters. The node-groups endpoint returns
every group in a single response, so it rejects --cost-mode and the pagination
flags (--limit, --cursor, --paginate, --include-total).`,
	Args: cobra.NoArgs,
	Example: `  kubeadapt get node-groups
  kubeadapt get node-groups --cluster-id abc123
  kubeadapt get node-groups -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := rejectCostMode(cmd, api.EndpointNodeGroups); err != nil {
			return err
		}
		// The flags are persistent on `get`, so they show up in --help here even
		// though the route has no pagination at all. Silently ignoring them made
		// --limit 2 return all 40 rows; reject instead of pretending.
		for _, name := range []string{flagLimit, flagCursor, flagPaginate, flagIncludeTotal} {
			if cmd.Flags().Changed(name) {
				return fmt.Errorf("--%s is not accepted by the node-groups endpoint, which returns every node group in one response", name)
			}
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

		fetch := func(ctx context.Context, cursor string) ([]types.NodeGroup, *types.Meta, error) {
			return c.ListNodeGroups(ctx, api.NodeGroupFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				ClusterIDs: clusterIDs,
			})
		}
		return runPaginatedList(cmd, c, paged, fetch, output.RenderNodeGroups)
	},
}

func init() {
	getNodeGroupsCmd.Flags().StringSlice("cluster-id", nil, clusterIDFilterUsage)
	registerClusterIDFlag(getNodeGroupsCmd)
	getCmd.AddCommand(getNodeGroupsCmd)
}
