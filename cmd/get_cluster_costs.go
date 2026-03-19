package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getClusterCostsCmd = &cobra.Command{
	Use:     "cluster-costs [cluster-id]",
	Short:   "Show cost distribution over time for a cluster",
	Example: "  kubeadapt get cluster-costs abc123 --timeframe 7d",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		timeframe, _ := cmd.Flags().GetString("timeframe")

		resp, err := client.GetClusterCostDistribution(cmd.Context(), args[0], timeframe)
		if err != nil {
			return err
		}

		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderCostDistribution(resp, isNoColor(cmd))
		})
	},
}

func init() {
	addTimeframeFlag(getClusterCostsCmd)
	getCmd.AddCommand(getClusterCostsCmd)
}
