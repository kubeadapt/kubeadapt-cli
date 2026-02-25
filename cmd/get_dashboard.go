package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show organization dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		days, _ := cmd.Flags().GetInt("days")

		resp, err := client.GetDashboard(context.Background(), days)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderDashboard(resp, noColor)
		})
	},
}

func init() {
	getDashboardCmd.Flags().Int("days", 30, "Number of days for cost trends")
	getCmd.AddCommand(getDashboardCmd)
}
