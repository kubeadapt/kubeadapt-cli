package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List nodes",
	Example: `  kubeadapt get nodes
  kubeadapt get nodes --cluster-id abc123
  kubeadapt get nodes --node-group general-purpose --limit 20`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		nodeGroup, _ := cmd.Flags().GetString("node-group")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching nodes...", func(ctx context.Context) (*types.NodeListResponse, error) {
			return client.GetNodes(ctx, clusterID, nodeGroup, limit, offset)
		})
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderNodes(resp.Nodes, resp.Total, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getNodesCmd)
	getNodesCmd.Flags().String("node-group", "", "Filter by node group")
	addLimitOffsetFlags(getNodesCmd)
	getCmd.AddCommand(getNodesCmd)
}
