package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getRecommendationsCmd = &cobra.Command{
	Use:   "recommendations",
	Short: "List recommendations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		recType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		resp, err := client.GetRecommendations(context.Background(), clusterID, recType, status, limit, offset)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderRecommendations(resp.Recommendations, noColor)
		})
	},
}

func init() {
	getRecommendationsCmd.Flags().String("cluster-id", "", "Filter by cluster ID")
	getRecommendationsCmd.Flags().String("type", "", "Filter by recommendation type")
	getRecommendationsCmd.Flags().String("status", "", "Filter by status")
	getRecommendationsCmd.Flags().Int("limit", 0, "Maximum number of results")
	getRecommendationsCmd.Flags().Int("offset", 0, "Number of results to skip")
	getCmd.AddCommand(getRecommendationsCmd)
}
