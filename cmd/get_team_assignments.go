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

var getTeamAssignmentsCmd = &cobra.Command{
	Use:   "team-assignments <team-id>",
	Short: "List a team's assignments",
	Long: `List workload/namespace/cluster assignments for a team. Use filters to narrow
by entity type, cluster, or assignment source.`,
	Args: cobra.ExactArgs(1),
	Example: `  kubeadapt get team-assignments team-abc-123
  kubeadapt get team-assignments team-abc-123 --entity-type workload
  kubeadapt get team-assignments team-abc-123 --cluster-id c1 --cluster-id c2 --paginate`,
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

		entityType, _ := cmd.Flags().GetString("entity-type")
		clusterIDs, _ := cmd.Flags().GetStringSlice("cluster-id")
		source, _ := cmd.Flags().GetString("source")

		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		var allItems []types.TeamAssignment
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.AssignmentFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				EntityType: entityType,
				ClusterIDs: clusterIDs,
				Source:     source,
			}
			items, meta, err := c.ListTeamAssignments(ctx, args[0], f)
			if err != nil {
				return fmt.Errorf("list team assignments for %s: %w", args[0], err)
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
			return output.RenderTeamAssignments(cmd.OutOrStdout(), allItems, lastMeta)
		}
	},
}

func init() {
	getTeamAssignmentsCmd.Flags().String("entity-type", "", "Filter by entity type (namespace|workload|cluster)")
	getTeamAssignmentsCmd.Flags().StringSlice("cluster-id", nil, "Filter by cluster ID (repeatable)")
	getTeamAssignmentsCmd.Flags().String("source", "", "Filter by source: k8s_label (auto from K8s label) | user_manual (UI) | namespace_auto (namespace default) | backfill_v1 (migration backfill)")
	getCmd.AddCommand(getTeamAssignmentsCmd)
}
