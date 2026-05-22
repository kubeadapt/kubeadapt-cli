package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

// flagTopClustersLimit is the local flag name for the dashboard subcommand's
// top-N cluster slice size. 0 means "use server default" (currently 5).
const flagTopClustersLimit = "top-clusters-limit"

// getDashboardCmd registers the `get dashboard` subcommand which calls
// GET /v1/organization/dashboard on the Kubeadapt public API. It honors
// the persistent --cost-mode flag from getCmd (via parsePagedFlags) and a
// local --top-clusters-limit flag that bounds the top-clusters slice.
var getDashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show organization dashboard",
	Long: `Display the organization dashboard: MTD spend, projected monthly, total
recommendations, top clusters, and savings summary. Accepts --cost-mode and
--top-clusters-limit to tune the response.`,
	Example: `  kubeadapt get dashboard
  kubeadapt get dashboard --cost-mode workload_only
  kubeadapt get dashboard --top-clusters-limit 10 -o yaml`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		rctx := getRunContext(cmd)
		if rctx == nil {
			return fmt.Errorf("failed to get run context")
		}
		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}

		paged, err := parsePagedFlags(cmd)
		if err != nil {
			return err
		}
		topLimit, err := cmd.Flags().GetInt(flagTopClustersLimit)
		if err != nil {
			return fmt.Errorf("read %s: %w", flagTopClustersLimit, err)
		}
		if topLimit < 0 || topLimit > 20 {
			return fmt.Errorf("invalid --%s %d (must be 0..20; 0 means server default)", flagTopClustersLimit, topLimit)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		opts := api.OrganizationDashboardOpts{
			CostModeOpt:      api.CostModeOpt{CostMode: paged.CostMode},
			TopClustersLimit: topLimit,
		}
		dash, _, err := client.GetOrganizationDashboard(ctx, opts)
		if err != nil {
			return fmt.Errorf("get dashboard: %w", err)
		}
		if dash == nil {
			return fmt.Errorf("get dashboard: empty response")
		}

		switch rctx.OutputFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), dash)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), dash)
		default:
			return output.RenderOrganizationDashboard(cmd.OutOrStdout(), *dash)
		}
	},
}

func init() {
	getDashboardCmd.Flags().Int(flagTopClustersLimit, 0, "Number of top clusters to include (1-20; 0 means server default)")
	getCmd.AddCommand(getDashboardCmd)
}
