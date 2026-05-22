package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getRecommendationCmd = &cobra.Command{
	Use:   "recommendation <id>",
	Short: "Show a recommendation",
	Long:  `Show details for a single recommendation by UUID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("cost-mode") {
			return fmt.Errorf("--cost-mode is not accepted by the recommendation endpoint")
		}
		rctx := getRunContext(cmd)
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		r, _, err := c.GetRecommendation(ctx, args[0])
		if err != nil {
			return fmt.Errorf("get recommendation %s: %w", args[0], err)
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), r)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), r)
		default:
			return output.RenderRecommendation(cmd.OutOrStdout(), *r)
		}
	},
}

func init() {
	getCmd.AddCommand(getRecommendationCmd)
}
