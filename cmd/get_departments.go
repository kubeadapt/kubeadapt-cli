package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getDepartmentsCmd = &cobra.Command{
	Use:   "departments",
	Short: "List departments with cost attribution",
	Args:  cobra.NoArgs,
	Example: `  kubeadapt get departments
  kubeadapt get departments --cost-mode workload_only
  kubeadapt get departments --origin kubeadapt --paginate`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		paged, err := parsePagedFlags(cmd)
		if err != nil {
			return err
		}

		origins, _ := cmd.Flags().GetStringSlice("origin")

		fetch := func(ctx context.Context, cursor string) ([]types.Department, *types.Meta, error) {
			return c.ListDepartments(ctx, api.DepartmentFilter{
				PagedOpts:   paged.pageOpts(cursor),
				CostModeOpt: api.CostModeOpt{CostMode: paged.CostMode},
				Origins:     origins,
			})
		}
		order := costDesc(func(d types.Department) (types.Money, string, string) {
			return d.Cost.CurrentRunRateHourly, d.ID, ""
		})
		return runPaginatedList(cmd, c, paged, fetch, output.RenderDepartments, order)
	},
}

func init() {
	getDepartmentsCmd.Flags().StringSlice("origin", nil, "Filter by origin: k8s (auto-discovered from K8s labels) or kubeadapt (created in dashboard) (repeatable)")
	registerEnumFlag(getDepartmentsCmd, "origin", originValues...)
	getCmd.AddCommand(getDepartmentsCmd)
}
