package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNodeGroupCmd = &cobra.Command{
	Use:   "node-group <name>",
	Short: "Show a node group",
	Long:  `Show details for a single node group by name within a cluster. Requires --cluster-id.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("cost-mode") {
			return fmt.Errorf("--cost-mode is not accepted by the node-group endpoint")
		}
		rctx := getRunContext(cmd)
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		clusterID, err := cmd.Flags().GetString("cluster-id")
		if err != nil {
			return err
		}
		if clusterID == "" {
			return fmt.Errorf("--cluster-id is required")
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		ng, _, err := c.GetNodeGroup(ctx, clusterID, args[0])
		if err != nil {
			return fmt.Errorf("get node-group %s: %w", args[0], err)
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), ng)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), ng)
		default:
			return output.RenderNodeGroup(cmd.OutOrStdout(), *ng)
		}
	},
}

func init() {
	getNodeGroupCmd.Flags().String("cluster-id", "", "Cluster ID that owns the node group (required)")
	_ = getNodeGroupCmd.MarkFlagRequired("cluster-id")
	getCmd.AddCommand(getNodeGroupCmd)
}
