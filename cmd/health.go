package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// newHealthCmd returns the `kubeadapt health` subcommand. It calls the
// public API's unauthenticated /health endpoint to verify connectivity
// and print the server status + version. It is most useful when
// diagnosing networking or DNS issues without consuming an API key.
func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check API server health (no authentication required)",
		Long: `Performs an unauthenticated GET /health request against the configured
API URL (--api-url / KUBEADAPT_API_URL / config file) and prints the result.

Exits with a non-zero status if the server is unreachable or returns a
non-2xx status.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			rctx := getRunContext(cmd)
			if rctx == nil {
				return fmt.Errorf("failed to get run context")
			}

			baseURL := strings.TrimRight(rctx.Config.APIURL, "/")
			if baseURL == "" {
				return fmt.Errorf("API URL is not configured (set --api-url, KUBEADAPT_API_URL, or run `kubeadapt auth login`)")
			}
			url := baseURL + "/health"

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("build request: %w", err)
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("User-Agent", "kubeadapt-cli/health")

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("GET %s: %w", url, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "API health probe failed: HTTP %d\n", resp.StatusCode)
				return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
			}

			var body struct {
				Status  string `json:"status"`
				Version string `json:"version,omitempty"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				return fmt.Errorf("decode health response: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:  %s\n", body.Status)
			if body.Version != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Version: %s\n", body.Version)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "URL:     %s\n", url)
			return nil
		},
	}
}
