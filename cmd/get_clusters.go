package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "List clusters",
	Long: `List clusters visible to the current API key. Filter by provider, region,
environment, or status. Pagination is cursor-based via --cursor + --limit.`,
	Args: cobra.NoArgs,
	Example: `  kubeadapt get clusters
  kubeadapt get clusters --provider aws --region eu-west-1
  kubeadapt get clusters --status active --limit 50
  kubeadapt get clusters --paginate -o json`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if err := rejectCostMode(cmd, api.EndpointClusters); err != nil {
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

		provider, _ := cmd.Flags().GetString("provider")
		region, _ := cmd.Flags().GetString("region")
		env, _ := cmd.Flags().GetString("environment")
		status, _ := cmd.Flags().GetString("status")

		fetch := func(ctx context.Context, cursor string) ([]types.Cluster, *types.Meta, error) {
			return c.ListClusters(ctx, api.ClusterFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				Provider:    provider,
				Region:      region,
				Environment: env,
				Status:      status,
			})
		}
		return runPaginatedList(cmd, c, paged, fetch, output.RenderClusters)
	},
}

func init() {
	getClustersCmd.Flags().String("provider", "", "Filter by cloud provider (aws|gcp|azure|on-prem)")
	getClustersCmd.Flags().String("region", "", "Filter by region (e.g. us-east-1)")
	getClustersCmd.Flags().String("environment", "", "Filter by environment (production|non-production|staging|dev)")
	getClustersCmd.Flags().String("status", "", "Filter by status (pending|active|disconnected|error|discovered)")
	registerEnumFlag(getClustersCmd, "provider", "aws", "gcp", "azure", "on-prem")
	registerEnumFlag(getClustersCmd, "environment", "production", "non-production", "staging", "dev")
	registerEnumFlag(getClustersCmd, "status", "pending", "active", "disconnected", "error", "discovered")
	getCmd.AddCommand(getClustersCmd)
}
