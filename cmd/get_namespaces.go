package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNamespacesCmd = &cobra.Command{
	Use:   "namespaces",
	Short: "List namespaces",
	Example: `  kubeadapt get namespaces
  kubeadapt get namespaces --cluster-id abc123 --team platform`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		team, _ := cmd.Flags().GetString("team")
		department, _ := cmd.Flags().GetString("department")
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching namespaces...", func(ctx context.Context) (*types.NamespaceListResponse, error) {
			return client.GetNamespaces(ctx, clusterID, team, department)
		})
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderNamespaces(resp.Namespaces, resp.Total, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getNamespacesCmd)
	getNamespacesCmd.Flags().String("team", "", "Filter by team")
	getNamespacesCmd.Flags().String("department", "", "Filter by department")
	getCmd.AddCommand(getNamespacesCmd)
}
