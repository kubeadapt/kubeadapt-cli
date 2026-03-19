package cmd

import (
	"context"

	"go.uber.org/zap"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/spf13/cobra"
)

// runContextKey is the context key for RunContext.
type runContextKey struct{}

// RunContext holds all runtime state that was previously stored as package globals.
// It is created once in PersistentPreRunE and stored in cobra's context.
type RunContext struct {
	Config    *config.Config
	Logger    *zap.Logger
	OutputFmt string
	NoColor   bool
	Verbose   bool
	Quiet     bool
}

// withRunContext stores the RunContext in the command's context.
func withRunContext(cmd *cobra.Command, rc *RunContext) {
	ctx := context.WithValue(cmd.Context(), runContextKey{}, rc)
	cmd.SetContext(ctx)
}

// getRunContext retrieves the RunContext from the command's context.
// It walks up the parent chain to find the context set by PersistentPreRunE.
func getRunContext(cmd *cobra.Command) *RunContext {
	ctx := cmd.Context()
	if ctx == nil {
		return nil
	}
	rc, _ := ctx.Value(runContextKey{}).(*RunContext)
	return rc
}
