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

var getTeamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "List teams with cost attribution",
	Args:  cobra.NoArgs,
	Example: `  kubeadapt get teams
  kubeadapt get teams --cost-mode workload_only
  kubeadapt get teams --department-id dept-123 --paginate`,
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

		deptIDs, _ := cmd.Flags().GetStringSlice("department-id")
		origins, _ := cmd.Flags().GetStringSlice("origin")

		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		var allItems []types.Team
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.TeamFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				CostModeOpt:   api.CostModeOpt{CostMode: paged.CostMode},
				DepartmentIDs: deptIDs,
				Origins:       origins,
			}
			items, meta, err := c.ListTeams(ctx, f)
			if err != nil {
				return fmt.Errorf("list teams: %w", err)
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
			return output.RenderTeams(cmd.OutOrStdout(), allItems, lastMeta)
		}
	},
}

func init() {
	getTeamsCmd.Flags().StringSlice("department-id", nil, "Filter by department ID (repeatable)")
	getTeamsCmd.Flags().StringSlice("origin", nil, "Filter by origin: k8s (auto-discovered from K8s labels) or kubeadapt (created in dashboard) (repeatable)")
	getCmd.AddCommand(getTeamsCmd)
}
