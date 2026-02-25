package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getClusterCmd = &cobra.Command{
	Use:   "cluster [id]",
	Short: "Show details of a specific cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		resp, err := client.GetCluster(context.Background(), args[0])
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderCluster(resp, noColor)
		})
	},
}

func init() {
	getCmd.AddCommand(getClusterCmd)
}
