package cmd

import (
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/spf13/cobra"
)

func newAPIClientFromCmd(cmd *cobra.Command) (*api.Client, error) {
	rc := getRunContext(cmd)
	if rc == nil || rc.Config == nil {
		return nil, fmt.Errorf("not authenticated. Run 'kubeadapt auth login' first")
	}
	if rc.Config.APIKey == "" {
		return nil, fmt.Errorf("no API key configured. Run 'kubeadapt auth login' first")
	}
	opts := []api.Option{api.WithLogger(rc.Logger)}
	// Retry only under --paginate. A single request that 429s should still fail
	// fast: the user is waiting on one answer and can decide for themselves,
	// whereas abandoning a 700-request run over one 429 throws away real work.
	if paginateRequested(cmd) {
		opts = append(opts,
			api.WithRetryOnRateLimit(true),
			api.WithMaxRetries(paginateRetries),
			api.WithSleeper(paginateSleep),
		)
	}
	return api.NewClient(rc.Config.APIURL, rc.Config.APIKey, opts...), nil
}

func paginateRequested(cmd *cobra.Command) bool {
	v, err := cmd.Flags().GetBool(flagPaginate)
	return err == nil && v
}

const (
	flagCostMode     = "cost-mode"
	flagCursor       = "cursor"
	flagLimit        = "limit"
	flagPaginate     = "paginate"
	flagIncludeTotal = "include-total"
)

const (
	formatTable = "table"
	formatJSON  = "json"
	formatYAML  = "yaml"
)

// PagedFlags mirrors api.PagedOpts + api.CostModeOpt.
type PagedFlags struct {
	CostMode     string
	Cursor       string
	Limit        int
	Paginate     bool
	IncludeTotal bool
	MaxWait      time.Duration
}

func isValidCostMode(s string) bool {
	switch s {
	case "fully_loaded", "workload_only":
		return true
	}
	return false
}

// Validates the cost-mode enum and limit range, returning a user-printable
// error. Must be called from every `get *` list subcommand's RunE.
func parsePagedFlags(cmd *cobra.Command) (PagedFlags, error) {
	var f PagedFlags
	var err error

	if f.CostMode, err = cmd.Flags().GetString(flagCostMode); err != nil {
		return f, fmt.Errorf("read %s: %w", flagCostMode, err)
	}
	if !isValidCostMode(f.CostMode) {
		return f, usagef("invalid --cost-mode %q (must be one of: fully_loaded, workload_only)", f.CostMode)
	}
	if f.Cursor, err = cmd.Flags().GetString(flagCursor); err != nil {
		return f, fmt.Errorf("read %s: %w", flagCursor, err)
	}
	if f.Limit, err = cmd.Flags().GetInt(flagLimit); err != nil {
		return f, fmt.Errorf("read %s: %w", flagLimit, err)
	}
	if f.Limit < 1 || f.Limit > 500 {
		return f, usagef("invalid --limit %d (must be 1..500)", f.Limit)
	}
	if f.Paginate, err = cmd.Flags().GetBool(flagPaginate); err != nil {
		return f, fmt.Errorf("read %s: %w", flagPaginate, err)
	}
	if f.IncludeTotal, err = cmd.Flags().GetBool(flagIncludeTotal); err != nil {
		return f, fmt.Errorf("read %s: %w", flagIncludeTotal, err)
	}
	// Absent flag rather than absent value: parsePagedFlags is also called
	// against trimmed-down command trees that register only the flags they
	// exercise, so a missing --max-wait means "take the default", not an error.
	f.MaxWait = defaultMaxWait
	if cmd.Flags().Lookup(flagMaxWait) != nil {
		if f.MaxWait, err = cmd.Flags().GetDuration(flagMaxWait); err != nil {
			return f, fmt.Errorf("read %s: %w", flagMaxWait, err)
		}
	}
	if f.MaxWait < 0 {
		return f, usagef("invalid --%s %s (must be >= 0; 0 means unbounded)", flagMaxWait, f.MaxWait)
	}
	return f, nil
}
