package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getWorkloadsCmd = &cobra.Command{
	Use:   "workloads",
	Short: "List workloads",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		namespace, _ := cmd.Flags().GetString("namespace")
		kind, _ := cmd.Flags().GetString("kind")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		resp, err := client.GetWorkloads(cmd.Context(), clusterID, namespace, kind, limit, offset)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderWorkloads(resp.Workloads, noColor)
		})
	},
}

func init() {
	getWorkloadsCmd.Flags().String("cluster-id", "", "Filter by cluster ID")
	getWorkloadsCmd.Flags().String("namespace", "", "Filter by namespace")
	getWorkloadsCmd.Flags().String("kind", "", "Filter by workload kind")
	getWorkloadsCmd.Flags().Int("limit", 0, "Maximum number of results")
	getWorkloadsCmd.Flags().Int("offset", 0, "Number of results to skip")
	getCmd.AddCommand(getWorkloadsCmd)
}
