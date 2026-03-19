package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getWorkloadMetricsCmd = &cobra.Command{
	Use:     "workload-metrics [workload-uid]",
	Short:   "Show time-series metrics for a workload",
	Example: "  kubeadapt get workload-metrics wl-uid-123 --cluster-id abc123 --timeframe 7d",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		timeframe, _ := cmd.Flags().GetString("timeframe")

		resp, err := client.GetWorkloadMetrics(cmd.Context(), args[0], clusterID, timeframe)
		if err != nil {
			return err
		}

		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderWorkloadMetrics(resp, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getWorkloadMetricsCmd)
	_ = getWorkloadMetricsCmd.MarkFlagRequired("cluster-id")
	addTimeframeFlag(getWorkloadMetricsCmd)
	getCmd.AddCommand(getWorkloadMetricsCmd)
}
