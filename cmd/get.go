package cmd

import "github.com/spf13/cobra"

var getCmd = &cobra.Command{
	Use:     "get",
	Short:   "Display resources",
	Long:    `Display Kubeadapt resources including clusters, workloads, nodes, recommendations, costs, and more.`,
	GroupID: groupData,
	Example: `  kubeadapt get clusters
  kubeadapt get workloads --cluster-id abc123
  kubeadapt get recommendations --status open`,
}

func init() {
	rootCmd.AddCommand(getCmd)
}
