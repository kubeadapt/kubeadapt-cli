package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNodeCmd = &cobra.Command{
	Use:               "node <node-uid>",
	Short:             "Show a node",
	Long:              `Show details for a single node by k8s metadata.uid.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: firstArgOnly(completeNodeUIDs),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := rejectCostMode(cmd, api.EndpointNode); err != nil {
			return err
		}
		rctx := getRunContext(cmd)
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		n, _, err := c.GetNode(ctx, args[0])
		if err != nil {
			return fmt.Errorf("get node %s: %w", args[0], err)
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), n)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), n)
		default:
			return output.RenderNode(cmd.OutOrStdout(), *n)
		}
	},
}

func init() {
	getCmd.AddCommand(getNodeCmd)
}
