package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getCapacityPlanningCmd = &cobra.Command{
	Use:   "capacity-planning",
	Short: "Show cluster capacity planning summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		resp, err := client.GetCapacityPlanning(cmd.Context(), clusterID)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderCapacityPlanning(resp, noColor)
		})
	},
}

func init() {
	getCapacityPlanningCmd.Flags().String("cluster-id", "", "Cluster ID (required)")
	if err := getCapacityPlanningCmd.MarkFlagRequired("cluster-id"); err != nil {
		panic(err)
	}
	getCmd.AddCommand(getCapacityPlanningCmd)
}
