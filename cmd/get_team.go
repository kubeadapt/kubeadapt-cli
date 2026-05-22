package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getTeamCmd = &cobra.Command{
	Use:   "team <team-id>",
	Short: "Show a team",
	Args:  cobra.ExactArgs(1),
	Example: `  kubeadapt get team team-abc-123
  kubeadapt get team team-abc-123 --cost-mode workload_only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rctx := getRunContext(cmd)
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		paged, err := parsePagedFlags(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		t, _, err := c.GetTeam(ctx, args[0], api.TeamGetOpts{
			CostModeOpt: api.CostModeOpt{CostMode: paged.CostMode},
		})
		if err != nil {
			return fmt.Errorf("get team %s: %w", args[0], err)
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), t)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), t)
		default:
			return output.RenderTeam(cmd.OutOrStdout(), *t)
		}
	},
}

func init() {
	getCmd.AddCommand(getTeamCmd)
}
