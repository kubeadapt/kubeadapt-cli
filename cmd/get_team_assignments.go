package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getTeamAssignmentsCmd = &cobra.Command{
	Use:   "team-assignments <team-id>",
	Short: "List a team's assignments",
	Long: `List workload/namespace/cluster assignments for a team. Use filters to narrow
by entity type, cluster, or assignment source. The assignments endpoint rejects
--cost-mode; an assignment records attribution, not cost.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: firstArgOnly(completeTeamIDs),
	Example: `  kubeadapt get team-assignments team-abc-123
  kubeadapt get team-assignments team-abc-123 --entity-type workload
  kubeadapt get team-assignments team-abc-123 --cluster-id c1 --cluster-id c2 --paginate`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := rejectCostMode(cmd, api.EndpointTeamAssignments); err != nil {
			return err
		}
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

		fetch := func(ctx context.Context, cursor string) ([]types.TeamAssignment, *types.Meta, error) {
			return c.ListTeamAssignments(ctx, args[0], api.AssignmentFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				EntityType: entityType,
				ClusterIDs: clusterIDs,
				Source:     source,
			})
		}
		return runPaginatedList(cmd, c, paged, fetch, output.RenderTeamAssignments)
	},
}

func init() {
	getTeamAssignmentsCmd.Flags().String("entity-type", "", "Filter by entity type (namespace|workload|cluster)")
	getTeamAssignmentsCmd.Flags().StringSlice("cluster-id", nil, clusterIDFilterUsage)
	getTeamAssignmentsCmd.Flags().String("source", "", "Filter by source: k8s_label (auto from K8s label) | user_manual (UI) | namespace_auto (namespace default) | backfill_v1 (migration backfill)")
	registerClusterIDFlag(getTeamAssignmentsCmd)
	registerEnumFlag(getTeamAssignmentsCmd, "entity-type", "namespace", "workload", "cluster")
	getCmd.AddCommand(getTeamAssignmentsCmd)
}
