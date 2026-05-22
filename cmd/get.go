package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:     "get",
	Short:   "Display resources",
	Long:    `Display Kubeadapt resources including clusters, workloads, nodes, recommendations, costs, and more.`,
	GroupID: groupData,
	Example: `  kubeadapt get clusters
  kubeadapt get workloads --cluster-id abc123
  kubeadapt get recommendations --status open`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return fmt.Errorf("unknown subcommand %q for %q", args[0], cmd.CommandPath())
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Cobra does NOT chain PersistentPreRunE through the command tree —
		// defining one on getCmd hides root's PreRun (which builds the
		// RunContext). Invoke it explicitly first so every `get *` subcommand
		// receives a populated RunContext via getRunContext.
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

func init() {
	getCmd.PersistentFlags().String(flagCostMode, "fully_loaded", "Cost attribution mode for namespace/workload/pod/team/department endpoints (fully_loaded|workload_only)")
	getCmd.PersistentFlags().String(flagCursor, "", "Pagination cursor (opaque token from previous response)")
	getCmd.PersistentFlags().Int(flagLimit, 100, "Page size (1-500)")
	getCmd.PersistentFlags().Bool(flagPaginate, false, "Automatically fetch all pages")
	getCmd.PersistentFlags().Bool(flagIncludeTotal, false, "Include total_count in pagination metadata (expensive)")

	rootCmd.AddCommand(getCmd)
}
