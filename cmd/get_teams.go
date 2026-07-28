package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getTeamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "List teams with cost attribution",
	Args:  cobra.NoArgs,
	Example: `  kubeadapt get teams
  kubeadapt get teams --cost-mode workload_only
  kubeadapt get teams --department-id dept-123 --paginate`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		paged, err := parsePagedFlags(cmd)
		if err != nil {
			return err
		}

		deptIDs, _ := cmd.Flags().GetStringSlice("department-id")
		origins, _ := cmd.Flags().GetStringSlice("origin")

		fetch := func(ctx context.Context, cursor string) ([]types.Team, *types.Meta, error) {
			return c.ListTeams(ctx, api.TeamFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				CostModeOpt:   api.CostModeOpt{CostMode: paged.CostMode},
				DepartmentIDs: deptIDs,
				Origins:       origins,
			})
		}
		return runPaginatedList(cmd, c, paged, fetch, output.RenderTeams)
	},
}

func init() {
	getTeamsCmd.Flags().StringSlice("department-id", nil, "Filter by department ID (repeatable)")
	getTeamsCmd.Flags().StringSlice("origin", nil, "Filter by origin: k8s (auto-discovered from K8s labels) or kubeadapt (created in dashboard) (repeatable)")
	registerEnumFlag(getTeamsCmd, "origin", "k8s", "kubeadapt")
	getCmd.AddCommand(getTeamsCmd)
}
