package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

// getOverviewCmd registers the `get overview` subcommand which calls
// GET /v1/organization on the Kubeadapt public API and renders the
// tenant-level snapshot (capacity, utilization, costs, run-rate).
var getOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Show organization overview",
	Long:  `Display a snapshot of the organization: capacity, utilization, costs, and run-rate.`,
	Example: `  kubeadapt get overview
  kubeadapt get overview -o json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if cmd.Flags().Changed(flagCostMode) {
			return fmt.Errorf("--cost-mode is not accepted by the organization endpoint")
		}
		rctx := getRunContext(cmd)
		if rctx == nil {
			return fmt.Errorf("failed to get run context")
		}
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		org, _, err := client.GetOrganization(ctx)
		if err != nil {
			return fmt.Errorf("get organization: %w", err)
		}
		if org == nil {
			return fmt.Errorf("get organization: empty response")
		}

		switch rctx.OutputFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), org)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), org)
		default:
			return output.RenderOrganization(cmd.OutOrStdout(), *org)
		}
	},
}

func init() {
	getCmd.AddCommand(getOverviewCmd)
}
