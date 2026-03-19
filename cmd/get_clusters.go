package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "List all clusters",
	Example: `  kubeadapt get clusters
  kubeadapt get clusters -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching clusters...", client.GetClusters)
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderClusters(resp.Clusters, resp.Total, isNoColor(cmd))
		})
	},
}

func init() {
	getCmd.AddCommand(getClustersCmd)
}
