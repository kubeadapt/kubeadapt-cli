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

var getNodeGroupsCmd = &cobra.Command{
	Use:   "node-groups",
	Short: "List node groups",
	Long: `List node groups across one or more clusters. The node-groups endpoint
rejects --cost-mode. Pagination is cursor-based via --cursor + --limit.`,
	Args: cobra.NoArgs,
	Example: `  kubeadapt get node-groups
  kubeadapt get node-groups --cluster-id abc123
  kubeadapt get node-groups --paginate -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if cmd.Flags().Changed("cost-mode") {
			return fmt.Errorf("--cost-mode is not accepted by the node-groups endpoint")
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

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		var allItems []types.NodeGroup
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.NodeGroupFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				ClusterIDs: clusterIDs,
			}
			items, meta, err := c.ListNodeGroups(ctx, f)
			if err != nil {
				return fmt.Errorf("list node groups: %w", err)
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
			return output.RenderNodeGroups(cmd.OutOrStdout(), allItems, lastMeta)
		}
	},
}

func init() {
	getNodeGroupsCmd.Flags().StringSlice("cluster-id", nil, "Filter by cluster ID (repeatable or comma-separated)")
	getCmd.AddCommand(getNodeGroupsCmd)
}
