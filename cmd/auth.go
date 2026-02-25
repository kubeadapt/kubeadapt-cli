package cmd

import "github.com/spf13/cobra"

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Manage authentication to the KubeAdapt API. Use subcommands to login, check status, or logout.`,
}

func init() {
	rootCmd.AddCommand(authCmd)
}
