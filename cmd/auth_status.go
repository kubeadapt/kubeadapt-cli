package cmd

import (
	"fmt"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load(cfgFile)
		if err != nil {
			fmt.Println("Not authenticated. Run 'kubeadapt auth login' to authenticate.")
			return nil
		}

		fmt.Printf("API URL:  %s\n", c.APIURL)
		if c.APIKey != "" {
			fmt.Printf("API Key:  %s\n", config.MaskAPIKey(c.APIKey))
		} else {
			fmt.Println("API Key:  (not set)")
		}

		return nil
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}
