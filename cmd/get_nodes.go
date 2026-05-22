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

var getNodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List nodes",
	Long: `List nodes across one or more clusters. The nodes endpoint rejects
--cost-mode (a node has a single physical bill). Pagination is cursor-based
via --cursor + --limit.`,
	Args: cobra.NoArgs,
	Example: `  kubeadapt get nodes
  kubeadapt get nodes --cluster-id abc123
  kubeadapt get nodes --is-spot --architecture arm64
  kubeadapt get nodes --capacity-type spot --paginate -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if cmd.Flags().Changed("cost-mode") {
			return fmt.Errorf("--cost-mode is not accepted by the nodes endpoint")
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
		nodeGroups, _ := cmd.Flags().GetStringSlice("node-group")
		instanceType, _ := cmd.Flags().GetString("instance-type")
		zone, _ := cmd.Flags().GetString("zone")
		architecture, _ := cmd.Flags().GetString("architecture")
		capacityType, _ := cmd.Flags().GetString("capacity-type")

		var isSpot *bool
		if cmd.Flags().Changed("is-spot") {
			v, _ := cmd.Flags().GetBool("is-spot")
			isSpot = &v
		}
		var isReady *bool
		if cmd.Flags().Changed("is-ready") {
			v, _ := cmd.Flags().GetBool("is-ready")
			isReady = &v
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		var allItems []types.Node
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.NodeFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				ClusterIDs:   clusterIDs,
				NodeGroups:   nodeGroups,
				InstanceType: instanceType,
				Zone:         zone,
				IsSpot:       isSpot,
				IsReady:      isReady,
				Architecture: architecture,
				CapacityType: capacityType,
			}
			items, meta, err := c.ListNodes(ctx, f)
			if err != nil {
				return fmt.Errorf("list nodes: %w", err)
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
			return output.RenderNodes(cmd.OutOrStdout(), allItems, lastMeta)
		}
	},
}

func init() {
	getNodesCmd.Flags().StringSlice("cluster-id", nil, "Filter by cluster ID (repeatable or comma-separated)")
	getNodesCmd.Flags().StringSlice("node-group", nil, "Filter by node group name (repeatable or comma-separated)")
	getNodesCmd.Flags().String("instance-type", "", "Filter by instance type (e.g. m5.large)")
	getNodesCmd.Flags().String("zone", "", "Filter by availability zone")
	getNodesCmd.Flags().Bool("is-spot", false, "Filter by spot instance (tri-state: only applied if set)")
	getNodesCmd.Flags().Bool("is-ready", false, "Filter by node readiness (tri-state: only applied if set)")
	getNodesCmd.Flags().String("architecture", "", "Filter by CPU architecture (amd64|arm64)")
	getNodesCmd.Flags().String("capacity-type", "", "Filter by capacity type (on-demand|spot)")
	getCmd.AddCommand(getNodesCmd)
}
