package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNodeGroupsCmd = &cobra.Command{
	Use:   "node-groups",
	Short: "List node groups",
	Example: `  kubeadapt get node-groups
  kubeadapt get node-groups --cluster-id abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching node groups...", func(ctx context.Context) (*types.NodeGroupListResponse, error) {
			return client.GetNodeGroups(ctx, clusterID)
		})
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderNodeGroups(resp.NodeGroups, resp.Total, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getNodeGroupsCmd)
	getCmd.AddCommand(getNodeGroupsCmd)
}
