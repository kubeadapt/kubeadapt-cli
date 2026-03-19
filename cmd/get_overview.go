package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Show organization dashboard overview",
	Example: `  kubeadapt get overview
  kubeadapt get overview -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching overview...", client.GetOverview)
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderOverview(resp, isNoColor(cmd))
		})
	},
}

func init() {
	getCmd.AddCommand(getOverviewCmd)
}
