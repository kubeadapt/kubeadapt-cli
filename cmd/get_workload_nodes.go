package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getWorkloadNodesCmd = &cobra.Command{
	Use:     "workload-nodes [workload-uid]",
	Short:   "Show node distribution for a workload",
	Example: "  kubeadapt get workload-nodes wl-uid-123 --cluster-id abc123",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		resp, err := client.GetWorkloadNodes(cmd.Context(), args[0], clusterID)
		if err != nil {
			return err
		}

		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderWorkloadNodes(resp, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getWorkloadNodesCmd)
	_ = getWorkloadNodesCmd.MarkFlagRequired("cluster-id")
	getCmd.AddCommand(getWorkloadNodesCmd)
}
