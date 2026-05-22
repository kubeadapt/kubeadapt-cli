package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

// getWorkloadCmd fetches a single workload by its Kubernetes metadata.uid.
var getWorkloadCmd = &cobra.Command{
	Use:   "workload <workload-uid>",
	Short: "Show a workload",
	Long:  `Show details for a single workload by k8s metadata.uid.`,
	Args:  cobra.ExactArgs(1),
	Example: `  kubeadapt get workload 11111111-2222-3333-4444-555555555555
  kubeadapt get workload 11111111-... --cost-mode workload_only -o yaml`,
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

		opts := api.WorkloadGetOpts{
			CostModeOpt: api.CostModeOpt{CostMode: paged.CostMode},
		}
		w, _, err := c.GetWorkload(ctx, args[0], opts)
		if err != nil {
			return fmt.Errorf("get workload %s: %w", args[0], err)
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), w)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), w)
		default:
			return output.RenderWorkload(cmd.OutOrStdout(), *w)
		}
	},
}

func init() {
	getCmd.AddCommand(getWorkloadCmd)
}
