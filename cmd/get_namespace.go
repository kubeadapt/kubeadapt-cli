package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

var getNamespaceCmd = &cobra.Command{
	Use:   "namespace <name>",
	Short: "Show a namespace",
	Long:  `Show details for a single namespace within a cluster. Requires --cluster-id.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rctx := getRunContext(cmd)
		c, err := newAPIClientFromCmd(cmd)
		if err != nil {
			return err
		}
		paged, err := parsePagedFlags(cmd)
		if err != nil {
			return err
		}
		clusterID, err := cmd.Flags().GetString("cluster-id")
		if err != nil {
			return err
		}
		if clusterID == "" {
			return fmt.Errorf("--cluster-id is required")
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		ns, _, err := c.GetNamespace(ctx, clusterID, args[0], api.NamespaceGetOpts{
			CostModeOpt: api.CostModeOpt{CostMode: paged.CostMode},
		})
		if err != nil {
			return fmt.Errorf("get namespace %s: %w", args[0], err)
		}

		outFmt := formatTable
		if rctx != nil {
			outFmt = rctx.OutputFmt
		}
		switch outFmt {
		case formatJSON:
			return output.RenderJSON(cmd.OutOrStdout(), ns)
		case formatYAML:
			return output.RenderYAML(cmd.OutOrStdout(), ns)
		default:
			return output.RenderNamespace(cmd.OutOrStdout(), *ns)
		}
	},
}

func init() {
	getNamespaceCmd.Flags().String("cluster-id", "", "Cluster ID that owns the namespace (required)")
	_ = getNamespaceCmd.MarkFlagRequired("cluster-id")
	getCmd.AddCommand(getNamespaceCmd)
}
