package cmd

import (
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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		resp, err := client.GetCostsTeams(cmd.Context(), clusterID)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderTeamCosts(resp.Teams, noColor)
		})
	},
}

func init() {
	getCostsTeamsCmd.Flags().String("cluster-id", "", "Filter by cluster ID")
	costsCmd.AddCommand(getCostsTeamsCmd)
	getCmd.AddCommand(costsCmd)
}
