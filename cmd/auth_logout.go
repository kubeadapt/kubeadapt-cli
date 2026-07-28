package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/spf13/cobra"
)

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored authentication credentials",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Deliberately the on-disk config, not the resolved RunContext one:
		// writing that back would persist a KUBEADAPT_API_KEY env value into
		// the file that logout is supposed to be clearing.
		c, err := config.Load(cfgFile)
		if err != nil {
			// Only a genuinely absent file proves there is nothing to remove.
			// A parse or permission failure leaves the key on disk.
			if errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintln(cmd.OutOrStdout(), "No stored credentials found.")
				return nil
			}
			return fmt.Errorf("cannot remove credentials: %w", err)
		}

		c.APIKey = ""
		if err := config.Save(c, cfgFile); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Logged out successfully. API key removed.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLogoutCmd)
}
