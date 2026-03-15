package cmd

import (
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

		resp, err := client.GetDashboard(cmd.Context())
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderDashboard(resp, noColor)
		})
	},
}

func init() {
	getCmd.AddCommand(getDashboardCmd)
}
