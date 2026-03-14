package cmd

import "github.com/spf13/cobra"

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Display resources",
	Long:  `Display Kubeadapt resources including clusters, workloads, nodes, recommendations, costs, and more.`,
}

func init() {
	rootCmd.AddCommand(getCmd)
}
