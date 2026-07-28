package cmd

import (
	"context"

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
		if err := rejectCostMode(cmd, api.EndpointNodes); err != nil {
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
		nodeGroups, _ := cmd.Flags().GetStringSlice("node-group")
		instanceType, _ := cmd.Flags().GetString("instance-type")
		zone, _ := cmd.Flags().GetString("zone")
		architecture, _ := cmd.Flags().GetString("architecture")
		capacityType, _ := cmd.Flags().GetString("capacity-type")
		isSpot := triStateBool(cmd, "is-spot")
		isReady := triStateBool(cmd, "is-ready")

		fetch := func(ctx context.Context, cursor string) ([]types.Node, *types.Meta, error) {
			return c.ListNodes(ctx, api.NodeFilter{
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
			})
		}
		return runPaginatedList(cmd, c, paged, fetch, output.RenderNodes)
	},
}

func init() {
	getNodesCmd.Flags().StringSlice("cluster-id", nil, clusterIDFilterUsage)
	getNodesCmd.Flags().StringSlice("node-group", nil, "Filter by node group name (repeatable or comma-separated)")
	getNodesCmd.Flags().String("instance-type", "", "Filter by instance type (e.g. m5.large)")
	getNodesCmd.Flags().String("zone", "", "Filter by availability zone")
	getNodesCmd.Flags().Bool("is-spot", false, "Filter by spot instance "+triStateSuffix)
	getNodesCmd.Flags().Bool("is-ready", false, "Filter by node readiness "+triStateSuffix)
	getNodesCmd.Flags().String("architecture", "", "Filter by CPU architecture (amd64|arm64)")
	getNodesCmd.Flags().String("capacity-type", "", "Filter by capacity type (on-demand|spot)")
	registerClusterIDFlag(getNodesCmd)
	registerEnumFlag(getNodesCmd, "architecture", "amd64", "arm64")
	registerEnumFlag(getNodesCmd, "capacity-type", "on-demand", "spot")
	getCmd.AddCommand(getNodesCmd)
}
