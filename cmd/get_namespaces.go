package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNamespacesCmd = &cobra.Command{
	Use:   "namespaces",
	Short: "List namespaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		team, _ := cmd.Flags().GetString("team")
		department, _ := cmd.Flags().GetString("department")

		resp, err := client.GetNamespaces(context.Background(), clusterID, team, department)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderNamespaces(resp.Namespaces, noColor)
		})
	},
}

func init() {
	getNamespacesCmd.Flags().String("cluster-id", "", "Filter by cluster ID")
	getNamespacesCmd.Flags().String("team", "", "Filter by team")
	getNamespacesCmd.Flags().String("department", "", "Filter by department")
	getCmd.AddCommand(getNamespacesCmd)
}
