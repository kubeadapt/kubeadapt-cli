package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getClusterCmd = &cobra.Command{
	Use:   "cluster <cluster-id>",
	Short: "Show a cluster",
	Long:  `Show details for a single cluster by UUID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("cost-mode") {
			return fmt.Errorf("--cost-mode is not accepted by the cluster endpoint")
		}
		rctx := getRunContext(cmd)
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		cluster, _, err := c.GetCluster(ctx, args[0])
		if err != nil {
			return fmt.Errorf("get cluster %s: %w", args[0], err)
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), cluster)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), cluster)
		default:
			return output.RenderCluster(cmd.OutOrStdout(), *cluster)
		}
	},
}

func init() {
	getCmd.AddCommand(getClusterCmd)
}
