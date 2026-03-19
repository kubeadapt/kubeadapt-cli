package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getClusterDashboardCmd = &cobra.Command{
	Use:     "cluster-dashboard [cluster-id]",
	Short:   "Show dashboard summary for a cluster",
	Example: "  kubeadapt get cluster-dashboard abc123",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		resp, err := client.GetClusterDashboard(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderClusterDashboard(resp, isNoColor(cmd))
		})
	},
}

func init() {
	getCmd.AddCommand(getClusterDashboardCmd)
}
