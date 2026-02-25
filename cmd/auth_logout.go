package cmd

import (
	"fmt"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/spf13/cobra"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load(cfgFile)
		if err != nil {
			fmt.Println("No stored credentials found.")
			return nil
		}

		c.APIKey = ""
		if err := config.Save(c, cfgFile); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Println("Logged out successfully. API key removed.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLogoutCmd)
}
