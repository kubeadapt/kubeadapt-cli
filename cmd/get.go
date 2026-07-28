package cmd

import (
	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:     "get",
	Short:   "Display resources",
	Long:    `Display Kubeadapt resources including clusters, workloads, nodes, recommendations, costs, and more.`,
	GroupID: groupData,
	Example: `  kubeadapt get clusters
  kubeadapt get workloads --cluster-id abc123
  kubeadapt get recommendations --status pending`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return usagef("unknown subcommand %q for %q", args[0], cmd.CommandPath())
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Cobra does not chain PersistentPreRunE: defining one here hides root's,
		// which builds the RunContext. Invoke it explicitly so every `get *`
		// subcommand still receives a populated RunContext.
		if rootCmd.PersistentPreRunE != nil {
			if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
				return err
			}
		}
		// Validate get-group flags up front so subcommands never see invalid input.
		if _, err := parsePagedFlags(cmd); err != nil {
			return err
		}
		return nil
	},
}

// Shared flag vocabulary for the `get` subcommands. One spelling per concept
// keeps `--help` from implying that identically-named flags behave differently.
const (
	clusterIDFilterUsage = "Filter by cluster ID (repeatable or comma-separated)"

	// A bool flag read through Changed() has three states, and --flag=false
	// meaning something other than omission is not guessable from the name.
	triStateSuffix = "(tri-state: only applied if set)"
)

// --cost-mode is persistent on `get`, so every subcommand advertises it in
// --help whether or not the route honors it. Silently ignoring it would report
// a fully-loaded number as if it were workload-only.
func rejectCostMode(cmd *cobra.Command, endpoint string) error {
	if !cmd.Flags().Changed(flagCostMode) {
		return nil
	}
	mode, _ := cmd.Flags().GetString(flagCostMode)
	return api.RejectCostMode(endpoint, mode)
}

// singleClusterID enforces the one-cluster contract for the detail commands.
// They take --cluster-id as a StringSlice purely so the flag parses the same
// way everywhere; a namespace or node group still lives in exactly one cluster.
func singleClusterID(cmd *cobra.Command) (string, error) {
	ids, err := cmd.Flags().GetStringSlice("cluster-id")
	if err != nil {
		return "", err
	}
	if len(ids) > 1 {
		return "", usagef("--cluster-id accepts exactly one cluster ID here, got %d", len(ids))
	}
	if len(ids) == 0 || ids[0] == "" {
		return "", usagef("--cluster-id is required")
	}
	return ids[0], nil
}

// triStateBool distinguishes an explicitly-set bool flag from an omitted one,
// returning nil for the latter so the filter is left off the request entirely.
func triStateBool(cmd *cobra.Command, name string) *bool {
	if !cmd.Flags().Changed(name) {
		return nil
	}
	v, _ := cmd.Flags().GetBool(name)
	return &v
}

func init() {
	getCmd.PersistentFlags().String(flagCostMode, "fully_loaded", "Cost attribution mode for namespace/workload/pod/team/department endpoints (fully_loaded|workload_only)")
	getCmd.PersistentFlags().String(flagCursor, "", "Pagination cursor (opaque token from previous response)")
	getCmd.PersistentFlags().Int(flagLimit, 100, "Page size (1-500)")
	getCmd.PersistentFlags().Bool(flagPaginate, false, "Automatically fetch all pages")
	getCmd.PersistentFlags().Bool(flagIncludeTotal, false, "Include total_count in pagination metadata (expensive)")

	registerEnumFlag(getCmd, flagCostMode, "fully_loaded", "workload_only")
	rootCmd.AddCommand(getCmd)
}
