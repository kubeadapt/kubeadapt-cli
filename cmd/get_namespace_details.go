package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNamespaceDetailsCmd = &cobra.Command{
	Use:     "namespace-details [name]",
	Short:   "Show details of a specific namespace",
	Example: "  kubeadapt get namespace-details kube-system --cluster-id abc123",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		resp, err := client.GetNamespaceDetails(cmd.Context(), args[0], clusterID)
		if err != nil {
			return err
		}

		return renderOutputFromCmd(cmd, resp, func() {
			output.RenderNamespaceDetails(resp, isNoColor(cmd))
		})
	},
}

func init() {
	addClusterIDFlag(getNamespaceDetailsCmd)
	_ = getNamespaceDetailsCmd.MarkFlagRequired("cluster-id")
	getCmd.AddCommand(getNamespaceDetailsCmd)
}
