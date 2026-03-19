package cmd

import (
	"context"
	"fmt"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

// fetchWithSpinner shows a spinner while fn executes, then hides it.
func fetchWithSpinner[T any](ctx context.Context, msg string, fn func(context.Context) (T, error)) (T, error) {
	sp := newSpinner(msg)
	sp.start()
	result, err := fn(ctx)
	sp.stop()
	return result, err
}

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

func renderOutputFromCmd(cmd *cobra.Command, data any, tableFunc func()) error {
	rc := getRunContext(cmd)
	format := "table"
	if rc != nil {
		format = rc.OutputFmt
	}
	switch format {
	case "json":
		return output.JSON(data)
	case "yaml":
		return output.YAML(data)
	default:
		tableFunc()
		return nil
	}
}

func addClusterIDFlag(cmd *cobra.Command) {
	cmd.Flags().String("cluster-id", "", "Filter by cluster ID")
}

func addLimitOffsetFlags(cmd *cobra.Command) {
	cmd.Flags().Int("limit", 0, "Maximum number of results")
	cmd.Flags().Int("offset", 0, "Number of results to skip")
}

func addNamespaceFlag(cmd *cobra.Command) {
	cmd.Flags().String("namespace", "", "Filter by namespace")
}

func addTimeframeFlag(cmd *cobra.Command) {
	cmd.Flags().String("timeframe", "", "Time range (e.g. 24h, 7d, 30d)")
}

func isNoColor(cmd *cobra.Command) bool {
	rc := getRunContext(cmd)
	if rc != nil {
		return rc.NoColor
	}
	return false
}
