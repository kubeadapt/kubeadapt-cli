package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getClusterCmd = &cobra.Command{
	Use:   "cluster [id]",
	Short: "Show details of a specific cluster",
	Args:  cobra.ExactArgs(1),
	Example: `  kubeadapt get cluster abc123
  kubeadapt get cluster abc123 -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching cluster details...", func(ctx context.Context) (*types.ClusterResponse, error) {
			return client.GetCluster(ctx, args[0])
		})
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderCluster(resp, isNoColor(cmd))
		})
	},
}

func init() {
	getCmd.AddCommand(getClusterCmd)
}
