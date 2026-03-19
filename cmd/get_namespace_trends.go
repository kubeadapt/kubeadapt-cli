package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNamespaceTrendsCmd = &cobra.Command{
	Use:     "namespace-trends [name]",
	Short:   "Show time-series trends for a namespace",
	Example: "  kubeadapt get namespace-trends kube-system --cluster-id abc123 --timeframe 7d",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		timeframe, _ := cmd.Flags().GetString("timeframe")

		resp, err := client.GetNamespaceTrends(cmd.Context(), args[0], clusterID, timeframe)
		if err != nil {
			return err
		}

		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderNamespaceTrends(resp, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getNamespaceTrendsCmd)
	_ = getNamespaceTrendsCmd.MarkFlagRequired("cluster-id")
	addTimeframeFlag(getNamespaceTrendsCmd)
	getCmd.AddCommand(getNamespaceTrendsCmd)
}
