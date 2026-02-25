package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		nodeGroup, _ := cmd.Flags().GetString("node-group")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		resp, err := client.GetNodes(context.Background(), clusterID, nodeGroup, limit, offset)
		if err != nil {
			return err
		}

		return renderOutput(outputFmt, resp, func() {
			output.RenderNodes(resp.Nodes, noColor)
		})
	},
}

func init() {
	getNodesCmd.Flags().String("cluster-id", "", "Filter by cluster ID")
	getNodesCmd.Flags().String("node-group", "", "Filter by node group")
	getNodesCmd.Flags().Int("limit", 0, "Maximum number of results")
	getNodesCmd.Flags().Int("offset", 0, "Number of results to skip")
	getCmd.AddCommand(getNodesCmd)
}
