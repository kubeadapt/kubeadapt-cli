package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show organization dashboard",
	Example: `  kubeadapt get dashboard
  kubeadapt get dashboard -o yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching dashboard...", client.GetDashboard)
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderDashboard(resp, isNoColor(cmd))
		})
	},
}

func init() {
	getCmd.AddCommand(getDashboardCmd)
}
