package cmd

import (
	"context"

	"go.uber.org/zap"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/spf13/cobra"
)

type runContextKey struct{}

// Created once in PersistentPreRunE and stored in cobra's context.
type RunContext struct {
	Config    *config.Config
	Logger    *zap.Logger
	OutputFmt string
	NoColor   bool
	Verbose   bool
	Quiet     bool
}

func withRunContext(cmd *cobra.Command, rc *RunContext) {
	ctx := context.WithValue(cmd.Context(), runContextKey{}, rc)
	cmd.SetContext(ctx)
}

// Walks up the parent chain to find the context set by PersistentPreRunE.
func getRunContext(cmd *cobra.Command) *RunContext {
	ctx := cmd.Context()
	if ctx == nil {
		return nil
	}
	rc, _ := ctx.Value(runContextKey{}).(*RunContext)
	return rc
}
