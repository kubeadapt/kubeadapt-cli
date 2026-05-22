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

// getPodsCmd lists pods that belong to a given workload (k8s metadata.uid)
// with cursor-based pagination and pod-specific filters.
var getPodsCmd = &cobra.Command{
	Use:   "pods <workload-uid>",
	Short: "List pods for a workload",
	Long: `List pods that belong to the workload identified by k8s metadata.uid.
Accepts pod-specific filters (phase, qos-class, host-path, empty-dir, host-network)
and the standard pagination flags.`,
	Args: cobra.ExactArgs(1),
	Example: `  kubeadapt get pods 11111111-...-aaaaaa
  kubeadapt get pods 11111111-...-aaaaaa --phase Running --qos-class Burstable
  kubeadapt get pods 11111111-...-aaaaaa --has-hostpath --paginate`,
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

		namespaces, _ := cmd.Flags().GetStringSlice("namespace")
		nodeUIDs, _ := cmd.Flags().GetStringSlice("node-uid")
		phase, _ := cmd.Flags().GetString("phase")
		qos, _ := cmd.Flags().GetString("qos-class")

		var triHostPath, triEmptyDir, triHostNet *bool
		if cmd.Flags().Changed("has-hostpath") {
			v, _ := cmd.Flags().GetBool("has-hostpath")
			triHostPath = &v
		}
		if cmd.Flags().Changed("has-emptydir") {
			v, _ := cmd.Flags().GetBool("has-emptydir")
			triEmptyDir = &v
		}
		if cmd.Flags().Changed("host-network") {
			v, _ := cmd.Flags().GetBool("host-network")
			triHostNet = &v
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		var allItems []types.Pod
		var lastMeta *types.Meta
		cursor := paged.Cursor
		for {
			f := api.PodFilter{
				PagedOpts: api.PagedOpts{
					Limit:        paged.Limit,
					Cursor:       cursor,
					IncludeTotal: paged.IncludeTotal,
				},
				CostModeOpt: api.CostModeOpt{CostMode: paged.CostMode},
				Namespaces:  namespaces,
				NodeUIDs:    nodeUIDs,
				Phase:       phase,
				QoSClass:    qos,
				HasHostPath: triHostPath,
				HasEmptyDir: triEmptyDir,
				HostNetwork: triHostNet,
			}
			items, meta, err := c.ListWorkloadPods(ctx, args[0], f)
			if err != nil {
				return fmt.Errorf("list pods for workload %s: %w", args[0], err)
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
			return output.RenderPods(cmd.OutOrStdout(), allItems, lastMeta)
		}
	},
}

func init() {
	getPodsCmd.Flags().StringSlice("namespace", nil, "Filter by namespace (repeatable)")
	getPodsCmd.Flags().StringSlice("node-uid", nil, "Filter by node UID (repeatable)")
	getPodsCmd.Flags().String("phase", "", "Filter by pod phase (Pending|Running|Succeeded|Failed|Unknown)")
	getPodsCmd.Flags().String("qos-class", "", "Filter by QoS class (Guaranteed|Burstable|BestEffort)")
	getPodsCmd.Flags().Bool("has-hostpath", false, "Filter pods with/without hostPath volumes")
	getPodsCmd.Flags().Bool("has-emptydir", false, "Filter pods with/without emptyDir volumes")
	getPodsCmd.Flags().Bool("host-network", false, "Filter pods with/without hostNetwork")
	getCmd.AddCommand(getPodsCmd)
}
