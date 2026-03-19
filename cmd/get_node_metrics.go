package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNodeMetricsCmd = &cobra.Command{
	Use:     "node-metrics [node-uid]",
	Short:   "Show time-series metrics for a node",
	Example: "  kubeadapt get node-metrics node-uid-123 --cluster-id abc123 --timeframe 24h",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		timeframe, _ := cmd.Flags().GetString("timeframe")

		resp, err := client.GetNodeMetrics(cmd.Context(), args[0], clusterID, timeframe)
		if err != nil {
			return err
		}

		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderNodeMetrics(resp, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getNodeMetricsCmd)
	_ = getNodeMetricsCmd.MarkFlagRequired("cluster-id")
	addTimeframeFlag(getNodeMetricsCmd)
	getCmd.AddCommand(getNodeMetricsCmd)
}
