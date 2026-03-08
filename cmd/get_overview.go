package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Show organization dashboard overview",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		resp, err := client.GetOverview(cmd.Context())
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderOverview(resp, noColor)
		})
	},
}

func init() {
	getCmd.AddCommand(getOverviewCmd)
}
