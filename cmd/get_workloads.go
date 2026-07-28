package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

// A single --cluster-id triggers the scoped path in the API client; multiple
// values are forwarded as a query filter.
var getWorkloadsCmd = &cobra.Command{
	Use:   "workloads",
	Short: "List workloads",
	Long: `List workloads visible to the current API key. Filter by cluster, namespace,
kind, team, department, HPA presence, or minimum hourly cost. Pagination is
cursor-based via --cursor + --limit.`,
	Args: cobra.NoArgs,
	Example: `  kubeadapt get workloads
  kubeadapt get workloads --cluster-id abc123 --namespace default
  kubeadapt get workloads --kind Deployment --kind StatefulSet --limit 50
  kubeadapt get workloads --has-hpa=true --min-cost-hourly 0.50
  kubeadapt get workloads --paginate -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		paged, err := parsePagedFlags(cmd)
		if err != nil {
			return err
		}

		clusters, _ := cmd.Flags().GetStringSlice("cluster-id")
		namespaces, _ := cmd.Flags().GetStringSlice("namespace")
		kinds, _ := cmd.Flags().GetStringSlice("kind")
		teams, _ := cmd.Flags().GetStringSlice("team")
		departments, _ := cmd.Flags().GetStringSlice("department")
		minCost, _ := cmd.Flags().GetString("min-cost-hourly")
		hasHPA := triStateBool(cmd, "has-hpa")

		fetch := func(ctx context.Context, cursor string) ([]types.Workload, *types.Meta, error) {
			return c.ListWorkloads(ctx, api.WorkloadFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				CostModeOpt:   api.CostModeOpt{CostMode: paged.CostMode},
				ClusterIDs:    clusters,
				Namespaces:    namespaces,
				Kinds:         kinds,
				Teams:         teams,
				Departments:   departments,
				HasHPA:        hasHPA,
				MinCostHourly: minCost,
			})
		}
		return runPaginatedList(cmd, c, paged, fetch, output.RenderWorkloads)
	},
}

func init() {
	getWorkloadsCmd.Flags().StringSlice("cluster-id", nil, clusterIDFilterUsage)
	getWorkloadsCmd.Flags().StringSlice("namespace", nil, "Filter by namespace (repeatable)")
	getWorkloadsCmd.Flags().StringSlice("kind", nil, "Filter by kind (Deployment|StatefulSet|DaemonSet) (repeatable)")
	getWorkloadsCmd.Flags().StringSlice("team", nil, "Filter by team (repeatable)")
	getWorkloadsCmd.Flags().StringSlice("department", nil, "Filter by department (repeatable)")
	getWorkloadsCmd.Flags().Bool("has-hpa", false, "Filter workloads with/without Horizontal Pod Autoscaler "+triStateSuffix)
	getWorkloadsCmd.Flags().String("min-cost-hourly", "", "Minimum hourly cost (decimal, e.g. 0.50)")
	registerClusterIDFlag(getWorkloadsCmd)
	registerEnumFlag(getWorkloadsCmd, "kind", "Deployment", "StatefulSet", "DaemonSet")
	getCmd.AddCommand(getWorkloadsCmd)
}
