package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	verbose      bool
	configFile   string
)

var rootCmd = &cobra.Command{
	Use:   "replace-me",
	Short: "Replace with a short description of your CLI",
	Long:  `Replace with a longer description of your CLI tool.`,
}

// Execute is the entry point called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json|yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file path (default: $HOME/.replace-me/config.yaml)")
}
