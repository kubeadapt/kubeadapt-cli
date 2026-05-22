package cmd

import (
	"fmt"

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
	return api.NewClient(rc.Config.APIURL, rc.Config.APIKey, api.WithLogger(rc.Logger)), nil
}

// Flag-name constants for the persistent pagination + cost-mode flags
// registered on getCmd. Used by parsePagedFlags and any subcommand that
// needs to read them by name.
const (
	flagCostMode     = "cost-mode"
	flagCursor       = "cursor"
	flagLimit        = "limit"
	flagPaginate     = "paginate"
	flagIncludeTotal = "include-total"
)

// Output format constants for the --output / -o flag. Used in every list and
// detail subcommand to dispatch between table, JSON, and YAML rendering.
const (
	formatTable = "table"
	formatJSON  = "json"
	formatYAML  = "yaml"
)

// PagedFlags is the resolved set of pagination + cost-mode flags shared by
// every `kubeadapt get` list subcommand. It mirrors api.PagedOpts +
// api.CostModeOpt and is produced by parsePagedFlags(cmd).
type PagedFlags struct {
	CostMode     string
	Cursor       string
	Limit        int
	Paginate     bool
	IncludeTotal bool
}

func isValidCostMode(s string) bool {
	switch s {
	case "fully_loaded", "workload_only":
		return true
	}
	return false
}

// parsePagedFlags reads the PersistentFlags from the get command tree and
// returns them as a PagedFlags. It validates the cost-mode enum and the
// limit range; on error it returns an error suitable for printing to the
// user. It MUST be called from every `get *` list subcommand's RunE.
func parsePagedFlags(cmd *cobra.Command) (PagedFlags, error) {
	var f PagedFlags
	var err error

	if f.CostMode, err = cmd.Flags().GetString(flagCostMode); err != nil {
		return f, fmt.Errorf("read %s: %w", flagCostMode, err)
	}
	if !isValidCostMode(f.CostMode) {
		return f, fmt.Errorf("invalid --cost-mode %q (must be one of: fully_loaded, workload_only)", f.CostMode)
	}
	if f.Cursor, err = cmd.Flags().GetString(flagCursor); err != nil {
		return f, fmt.Errorf("read %s: %w", flagCursor, err)
	}
	if f.Limit, err = cmd.Flags().GetInt(flagLimit); err != nil {
		return f, fmt.Errorf("read %s: %w", flagLimit, err)
	}
	if f.Limit < 1 || f.Limit > 500 {
		return f, fmt.Errorf("invalid --limit %d (must be 1..500)", f.Limit)
	}
	if f.Paginate, err = cmd.Flags().GetBool(flagPaginate); err != nil {
		return f, fmt.Errorf("read %s: %w", flagPaginate, err)
	}
	if f.IncludeTotal, err = cmd.Flags().GetBool(flagIncludeTotal); err != nil {
		return f, fmt.Errorf("read %s: %w", flagIncludeTotal, err)
	}
	return f, nil
}
