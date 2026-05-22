package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/spf13/cobra"
)

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load(cfgFile)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Not authenticated. Run 'kubeadapt auth login' to authenticate.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "API URL:    %s\n", c.APIURL)
		if c.APIKey != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "API Key:    %s\n", config.MaskAPIKey(c.APIKey))
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "API Key:    (not set)")
			return nil
		}

		client, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()
		org, meta, err := client.GetOrganization(ctx)
		switch {
		case err == nil:
			fmt.Fprintf(cmd.OutOrStdout(), "Status:     Connected\n")
			if org != nil && org.Metadata.Name != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Org:        %s (%s)\n", org.Metadata.Name, org.ID)
			} else if org != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Org ID:     %s\n", org.ID)
			}
			if meta != nil && meta.RequestID != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Request ID: %s\n", meta.RequestID)
			}
		case api.IsUnauthorized(err):
			fmt.Fprintf(cmd.OutOrStdout(), "Status:     Unauthorized (run `kubeadapt auth login`)\n")
		default:
			fmt.Fprintf(cmd.OutOrStdout(), "Status:     Error: %v\n", err)
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}
