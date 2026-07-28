package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	authStatusNotAuthenticated = "not_authenticated"
	authStatusNoAPIKey         = "no_api_key"
	authStatusConnected        = "connected"
	authStatusUnauthorized     = "unauthorized"
	authStatusError            = "error"
)

// authStatusReport is the machine-readable shape of `auth status`. The API key
// is always masked: this output is meant to be safe to paste into a bug report.
type authStatusReport struct {
	APIURL       string `json:"api_url" yaml:"api_url"`
	APIKeyMasked string `json:"api_key_masked" yaml:"api_key_masked"`
	Status       string `json:"status" yaml:"status"`
	OrgName      string `json:"org_name,omitempty" yaml:"org_name,omitempty"`
	OrgID        string `json:"org_id,omitempty" yaml:"org_id,omitempty"`
	RequestID    string `json:"request_id,omitempty" yaml:"request_id,omitempty"`
	Error        string `json:"error,omitempty" yaml:"error,omitempty"`
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// The report is printed before the exit status is decided so that a
		// failure still yields the full diagnostic, in every output format.
		// buildAuthStatusReport returns no error for exactly that reason: every
		// failure has to arrive as a status on the report, not instead of one.
		report := buildAuthStatusReport(cmd)
		if err := renderAuthStatusReport(cmd.OutOrStdout(), authOutputFormat(cmd), report); err != nil {
			return err
		}
		return authStatusExitError(report)
	},
}

// effectiveConfig returns the credentials the CLI would actually send: the
// RunContext copy has --api-key/--api-url and the KUBEADAPT_* env vars already
// applied. Loading the file directly reports a key that may never be used.
func effectiveConfig(cmd *cobra.Command) (*config.Config, error) {
	if rc := getRunContext(cmd); rc != nil && rc.Config != nil {
		return rc.Config, nil
	}
	return config.Load(cfgFile)
}

func buildAuthStatusReport(cmd *cobra.Command) authStatusReport {
	c, err := effectiveConfig(cmd)
	if err != nil {
		return authStatusReport{Status: authStatusNotAuthenticated}
	}

	report := authStatusReport{APIURL: c.APIURL, Status: authStatusNoAPIKey}
	if c.APIKey == "" {
		return report
	}
	report.APIKeyMasked = config.MaskAPIKey(c.APIKey)

	client, err := newAPIClientFromCmd(cmd)
	if err != nil {
		report.Status = authStatusError
		report.Error = err.Error()
		return report
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()

	org, meta, err := client.GetOrganization(ctx)
	switch {
	case err == nil:
		report.Status = authStatusConnected
		if org != nil {
			report.OrgID = org.ID
			report.OrgName = org.Metadata.Name
		}
		if meta != nil {
			report.RequestID = meta.RequestID
		}
	case api.IsUnauthorized(err):
		report.Status = authStatusUnauthorized
	default:
		report.Status = authStatusError
		report.Error = err.Error()
	}
	return report
}

// authStatusExitError makes `kubeadapt auth status` exit 0 if and only if the
// CLI can authenticate right now, so it is usable as a gate in a script. A
// missing key is as fatal as a rejected one: neither can call the API.
func authStatusExitError(r authStatusReport) error {
	switch r.Status {
	case authStatusConnected:
		return nil
	case authStatusUnauthorized:
		return fmt.Errorf("API key rejected by %s. Run 'kubeadapt auth login' to re-authenticate", r.APIURL)
	case authStatusError:
		return fmt.Errorf("could not verify credentials against %s: %s", r.APIURL, r.Error)
	default:
		return fmt.Errorf("no API key configured. Run 'kubeadapt auth login' first")
	}
}

func renderAuthStatusReport(out io.Writer, format string, r authStatusReport) error {
	if format != formatTable {
		return renderAuthStatus(out, format, r)
	}
	if r.Status == authStatusNotAuthenticated {
		fmt.Fprintln(out, "Not authenticated. Run 'kubeadapt auth login' to authenticate.")
		return nil
	}

	fmt.Fprintf(out, "API URL:    %s\n", r.APIURL)
	if r.Status == authStatusNoAPIKey {
		fmt.Fprintln(out, "API Key:    (not set)")
		return nil
	}
	fmt.Fprintf(out, "API Key:    %s\n", r.APIKeyMasked)

	switch r.Status {
	case authStatusConnected:
		fmt.Fprintf(out, "Status:     Connected\n")
		switch {
		case r.OrgName != "":
			fmt.Fprintf(out, "Org:        %s (%s)\n", r.OrgName, r.OrgID)
		case r.OrgID != "":
			fmt.Fprintf(out, "Org ID:     %s\n", r.OrgID)
		}
		if r.RequestID != "" {
			fmt.Fprintf(out, "Request ID: %s\n", r.RequestID)
		}
	case authStatusUnauthorized:
		fmt.Fprintf(out, "Status:     Unauthorized (run `kubeadapt auth login`)\n")
	default:
		fmt.Fprintf(out, "Status:     Error: %s\n", r.Error)
	}
	return nil
}

func renderAuthStatus(w io.Writer, format string, r authStatusReport) error {
	if format == formatYAML {
		return output.RenderYAML(w, r)
	}
	return output.RenderJSON(w, r)
}

// authOutputFormat prefers the RunContext, which resolves -o for every command,
// but falls back to the flag var so direct RunE calls in tests still work.
func authOutputFormat(cmd *cobra.Command) string {
	if rc := getRunContext(cmd); rc != nil && rc.OutputFmt != "" {
		return rc.OutputFmt
	}
	return outputFmt
}

func init() {
	authCmd.AddCommand(authStatusCmd)
}
