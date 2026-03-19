package cmd

import "github.com/spf13/cobra"

var authCmd = &cobra.Command{
	Use:     "auth",
	Short:   "Manage authentication",
	Long:    `Manage authentication to the Kubeadapt API. Use subcommands to login, check status, or logout.`,
	GroupID: groupAuth,
	Example: `  kubeadapt auth login
  kubeadapt auth status
  kubeadapt auth logout`,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
