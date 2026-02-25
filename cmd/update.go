package cmd

import "github.com/spf13/cobra"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update resources",
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
