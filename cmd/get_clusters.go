package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "List all clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		resp, err := client.GetClusters(cmd.Context())
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderClusters(resp.Clusters, noColor)
		})
	},
}

func init() {
	getCmd.AddCommand(getClustersCmd)
}
