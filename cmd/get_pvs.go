package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getPVsCmd = &cobra.Command{
	Use:     "persistent-volumes",
	Short:   "List persistent volumes",
	Aliases: []string{"pvs"},
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		namespace, _ := cmd.Flags().GetString("namespace")
		storageClass, _ := cmd.Flags().GetString("storage-class")

		resp, err := client.GetPersistentVolumes(context.Background(), clusterID, namespace, storageClass)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderPersistentVolumes(resp.PersistentVolumes, noColor)
		})
	},
}

func init() {
	getPVsCmd.Flags().String("cluster-id", "", "Filter by cluster ID")
	getPVsCmd.Flags().String("namespace", "", "Filter by namespace")
	getPVsCmd.Flags().String("storage-class", "", "Filter by storage class")
	getCmd.AddCommand(getPVsCmd)
}
