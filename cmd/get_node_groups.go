package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNodeGroupsCmd = &cobra.Command{
	Use:   "node-groups",
	Short: "List node groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		resp, err := client.GetNodeGroups(context.Background(), clusterID)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderNodeGroups(resp.NodeGroups, noColor)
		})
	},
}

func init() {
	getNodeGroupsCmd.Flags().String("cluster-id", "", "Filter by cluster ID")
	getCmd.AddCommand(getNodeGroupsCmd)
}
