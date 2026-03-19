package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getWorkloadsCmd = &cobra.Command{
	Use:   "workloads",
	Short: "List workloads",
	Example: `  kubeadapt get workloads
  kubeadapt get workloads --cluster-id abc123 --namespace default
  kubeadapt get workloads --kind Deployment --limit 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		namespace, _ := cmd.Flags().GetString("namespace")
		kind, _ := cmd.Flags().GetString("kind")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching workloads...", func(ctx context.Context) (*types.WorkloadListResponse, error) {
			return client.GetWorkloads(ctx, clusterID, namespace, kind, limit, offset)
		})
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderWorkloads(resp.Workloads, resp.Total, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getWorkloadsCmd)
	addNamespaceFlag(getWorkloadsCmd)
	getWorkloadsCmd.Flags().String("kind", "", "Filter by workload kind")
	addLimitOffsetFlags(getWorkloadsCmd)
	getCmd.AddCommand(getWorkloadsCmd)
}
