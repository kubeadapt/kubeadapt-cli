package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getRecommendationsCmd = &cobra.Command{
	Use:   "recommendations",
	Short: "List recommendations",
	Example: `  kubeadapt get recommendations
  kubeadapt get recommendations --status open --type rightsize
  kubeadapt get recommendations --cluster-id abc123 --limit 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		recType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching recommendations...", func(ctx context.Context) (*types.RecommendationListResponse, error) {
			return client.GetRecommendations(ctx, clusterID, recType, status, limit, offset)
		})
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderRecommendations(resp.Recommendations, resp.Total, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getRecommendationsCmd)
	getRecommendationsCmd.Flags().String("type", "", "Filter by recommendation type")
	getRecommendationsCmd.Flags().String("status", "", "Filter by status")
	addLimitOffsetFlags(getRecommendationsCmd)
	getCmd.AddCommand(getRecommendationsCmd)
}
