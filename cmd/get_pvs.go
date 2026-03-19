package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getPVsCmd = &cobra.Command{
	Use:     "persistent-volumes",
	Short:   "List persistent volumes",
	Aliases: []string{"pvs"},
	Example: `  kubeadapt get persistent-volumes
  kubeadapt get pvs --cluster-id abc123 --namespace monitoring`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		namespace, _ := cmd.Flags().GetString("namespace")
		storageClass, _ := cmd.Flags().GetString("storage-class")
		resp, err := fetchWithSpinner(cmd.Context(), "Fetching persistent volumes...", func(ctx context.Context) (*types.PersistentVolumeListResponse, error) {
			return client.GetPersistentVolumes(ctx, clusterID, namespace, storageClass)
		})
		if err != nil {
			return err
		}
		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderPersistentVolumes(resp.PersistentVolumes, resp.Total, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getPVsCmd)
	addNamespaceFlag(getPVsCmd)
	getPVsCmd.Flags().String("storage-class", "", "Filter by storage class")
	getCmd.AddCommand(getPVsCmd)
}
