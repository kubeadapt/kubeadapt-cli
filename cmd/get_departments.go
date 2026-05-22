package cmd

import (
	"context"
	"fmt"
	"time"

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
		rctx := getRunContext(cmd)
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		paged, err := parsePagedFlags(cmd)
		if err != nil {
			return err
		}

		origins, _ := cmd.Flags().GetStringSlice("origin")

		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		var allItems []types.Department
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.DepartmentFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				CostModeOpt: api.CostModeOpt{CostMode: paged.CostMode},
				Origins:     origins,
			}
			items, meta, err := c.ListDepartments(ctx, f)
			if err != nil {
				return fmt.Errorf("list departments: %w", err)
			}
			allItems = append(allItems, items...)
			lastMeta = meta
			if !paged.Paginate || meta == nil || meta.Pagination == nil || !meta.Pagination.HasMore {
				break
			}
			cursor = meta.Pagination.NextCursor
			if cursor == "" {
				break
			}
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSONWithMeta(cmd.OutOrStdout(), allItems, lastMeta)
		case formatYAML:
			return output.RenderYAMLWithMeta(cmd.OutOrStdout(), allItems, lastMeta)
		default:
			return output.RenderDepartments(cmd.OutOrStdout(), allItems, lastMeta)
		}
	},
}

func init() {
	getDepartmentsCmd.Flags().StringSlice("origin", nil, "Filter by origin: k8s (auto-discovered from K8s labels) or kubeadapt (created in dashboard) (repeatable)")
	getCmd.AddCommand(getDepartmentsCmd)
}
