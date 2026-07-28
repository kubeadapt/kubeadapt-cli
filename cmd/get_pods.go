package cmd

import (
	"context"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getPodsCmd = &cobra.Command{
	Use:   "pods <workload-uid>",
	Short: "List pods for a workload",
	Long: `List pods that belong to the workload identified by k8s metadata.uid.
Accepts pod-specific filters (phase, qos-class, host-path, empty-dir, host-network)
and the standard pagination flags.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: firstArgOnly(completeWorkloadUIDs),
	Example: `  kubeadapt get pods 11111111-...-aaaaaa
  kubeadapt get pods 11111111-...-aaaaaa --phase Running --qos-class Burstable
  kubeadapt get pods 11111111-...-aaaaaa --has-hostpath --paginate`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		hasHostPath := triStateBool(cmd, "has-hostpath")
		hasEmptyDir := triStateBool(cmd, "has-emptydir")
		hostNetwork := triStateBool(cmd, "host-network")

		fetch := func(ctx context.Context, cursor string) ([]types.Pod, *types.Meta, error) {
			return c.ListWorkloadPods(ctx, args[0], api.PodFilter{
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
				HasHostPath: hasHostPath,
				HasEmptyDir: hasEmptyDir,
				HostNetwork: hostNetwork,
			})
		}
		return runPaginatedList(cmd, c, paged, fetch, output.RenderPods)
	},
}

func init() {
	getPodsCmd.Flags().StringSlice("namespace", nil, "Filter by namespace (repeatable)")
	getPodsCmd.Flags().StringSlice("node-uid", nil, "Filter by node UID (repeatable)")
	getPodsCmd.Flags().String("phase", "", "Filter by pod phase (Pending|Running|Succeeded|Failed|Unknown)")
	getPodsCmd.Flags().String("qos-class", "", "Filter by QoS class (Guaranteed|Burstable|BestEffort)")
	getPodsCmd.Flags().Bool("has-hostpath", false, "Filter pods with/without hostPath volumes "+triStateSuffix)
	getPodsCmd.Flags().Bool("has-emptydir", false, "Filter pods with/without emptyDir volumes "+triStateSuffix)
	getPodsCmd.Flags().Bool("host-network", false, "Filter pods with/without hostNetwork "+triStateSuffix)
	registerEnumFlag(getPodsCmd, "phase", "Pending", "Running", "Succeeded", "Failed", "Unknown")
	registerEnumFlag(getPodsCmd, "qos-class", "Guaranteed", "Burstable", "BestEffort")
	getCmd.AddCommand(getPodsCmd)
}
