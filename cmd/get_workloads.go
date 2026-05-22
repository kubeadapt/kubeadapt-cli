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

// getWorkloadsCmd lists workloads visible to the current API key with
// cursor-based pagination and a rich set of filters. A single --cluster-id
// triggers the scoped path (handled by the API client); multiple values are
// forwarded as a query filter.
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
		rctx := getRunContext(cmd)
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

		var hasHPAPtr *bool
		if cmd.Flags().Changed("has-hpa") {
			v, _ := cmd.Flags().GetBool("has-hpa")
			hasHPAPtr = &v
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		var allItems []types.Workload
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.WorkloadFilter{
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
				HasHPA:        hasHPAPtr,
				MinCostHourly: minCost,
			}
			items, meta, err := c.ListWorkloads(ctx, f)
			if err != nil {
				return fmt.Errorf("list workloads: %w", err)
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
			return output.RenderWorkloads(cmd.OutOrStdout(), allItems, lastMeta)
		}
	},
}

func init() {
	getWorkloadsCmd.Flags().StringSlice("cluster-id", nil, "Filter by cluster ID (repeatable; one becomes a scoped path)")
	getWorkloadsCmd.Flags().StringSlice("namespace", nil, "Filter by namespace (repeatable)")
	getWorkloadsCmd.Flags().StringSlice("kind", nil, "Filter by kind (Deployment|StatefulSet|DaemonSet) (repeatable)")
	getWorkloadsCmd.Flags().StringSlice("team", nil, "Filter by team (repeatable)")
	getWorkloadsCmd.Flags().StringSlice("department", nil, "Filter by department (repeatable)")
	getWorkloadsCmd.Flags().Bool("has-hpa", false, "Filter workloads with/without Horizontal Pod Autoscaler")
	getWorkloadsCmd.Flags().String("min-cost-hourly", "", "Minimum hourly cost (decimal, e.g. 0.50)")
	getCmd.AddCommand(getWorkloadsCmd)
}
