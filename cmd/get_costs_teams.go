package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var costsCmd = &cobra.Command{
	Use:   "costs",
	Short: "Show cost breakdowns",
}

var getCostsTeamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Show cost breakdown by team",
	Example: `  kubeadapt get costs teams
  kubeadapt get costs teams --cluster-id abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching team costs...", func(ctx context.Context) (*types.TeamCostListResponse, error) {
			return client.GetCostsTeams(ctx, clusterID)
		})
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderTeamCosts(resp.Teams, resp.Total, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getCostsTeamsCmd)
	costsCmd.AddCommand(getCostsTeamsCmd)
	getCmd.AddCommand(costsCmd)
}
