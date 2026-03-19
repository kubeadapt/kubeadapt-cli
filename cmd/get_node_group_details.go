package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNodeGroupDetailsCmd = &cobra.Command{
	Use:     "node-group-details [group-name]",
	Short:   "Show details of a specific node group",
	Example: "  kubeadapt get node-group-details my-group --cluster-id abc123",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		resp, err := client.GetNodeGroupDetails(cmd.Context(), args[0], clusterID)
		if err != nil {
			return err
		}

		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderNodeGroupDetails(resp, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getNodeGroupDetailsCmd)
	_ = getNodeGroupDetailsCmd.MarkFlagRequired("cluster-id")
	getCmd.AddCommand(getNodeGroupDetailsCmd)
}
