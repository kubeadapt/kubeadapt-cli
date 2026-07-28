package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
)

// These keys are a scripting contract (`health -o json | jq .status`); renaming
// one breaks callers even though nothing in-tree references them.
type healthInfo struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
	URL     string `json:"url"`
}

// Uses the unauthenticated /health endpoint, so it diagnoses networking or DNS
// without consuming an API key.
func newHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "health",
		Short:   "Check API server health (no authentication required)",
		GroupID: groupUtility,
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

			info := healthInfo{Status: body.Status, Version: body.Version, URL: url}

			outFmt := outputFmt
			if rctx.OutputFmt != "" {
				outFmt = rctx.OutputFmt
			}

			w := cmd.OutOrStdout()
			switch outFmt {
			case formatJSON:
				return output.RenderJSON(w, info)
			case formatYAML:
				return output.RenderYAML(w, info)
			}

			// The status is the answer the command exists to give, so --quiet
			// trims only the surrounding context lines.
			_, _ = fmt.Fprintf(w, "Status:  %s\n", info.Status)
			if rctx.Quiet {
				return nil
			}
			if info.Version != "" {
				_, _ = fmt.Fprintf(w, "Version: %s\n", info.Version)
			}
			_, _ = fmt.Fprintf(w, "URL:     %s\n", info.URL)
			return nil
		},
	}
}
