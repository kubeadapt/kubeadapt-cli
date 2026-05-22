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
		if cmd.Flags().Changed("cost-mode") {
			return fmt.Errorf("--cost-mode is not accepted by the clusters endpoint")
		}
		rctx := getRunContext(cmd)
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

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		var allItems []types.Cluster
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.ClusterFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				Provider:    provider,
				Region:      region,
				Environment: env,
				Status:      status,
			}
			items, meta, err := c.ListClusters(ctx, f)
			if err != nil {
				return fmt.Errorf("list clusters: %w", err)
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
			return output.RenderClusters(cmd.OutOrStdout(), allItems, lastMeta)
		}
	},
}

func init() {
	getClustersCmd.Flags().String("provider", "", "Filter by cloud provider (aws|gcp|azure|on-prem)")
	getClustersCmd.Flags().String("region", "", "Filter by region (e.g. us-east-1)")
	getClustersCmd.Flags().String("environment", "", "Filter by environment (production|non-production|staging|dev)")
	getClustersCmd.Flags().String("status", "", "Filter by status (pending|active|disconnected|error|discovered)")
	getCmd.AddCommand(getClustersCmd)
}
