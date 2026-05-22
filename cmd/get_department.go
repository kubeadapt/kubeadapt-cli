package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getDepartmentCmd = &cobra.Command{
	Use:   "department <dept-id>",
	Short: "Show a department",
	Args:  cobra.ExactArgs(1),
	Example: `  kubeadapt get department dept-abc-123
  kubeadapt get department dept-abc-123 --cost-mode workload_only`,
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

		d, _, err := c.GetDepartment(ctx, args[0], api.DepartmentGetOpts{
			CostModeOpt: api.CostModeOpt{CostMode: paged.CostMode},
		})
		if err != nil {
			return fmt.Errorf("get department %s: %w", args[0], err)
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), d)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), d)
		default:
			return output.RenderDepartment(cmd.OutOrStdout(), *d)
		}
	},
}

func init() {
	getCmd.AddCommand(getDepartmentCmd)
}
