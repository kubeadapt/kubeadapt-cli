package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/kubeadapt/kubeadapt-cli/internal/version"
)

// versionInfo is the machine-readable shape of the version block. The keys are
// the contract consumed by `kubeadapt version -o json | jq`, so they are
// snake_case rather than a transliteration of the table labels.
type versionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Built     string `json:"built"`
	GoVersion string `json:"go_version"`
	OSArch    string `json:"os_arch"`
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Print version information",
	GroupID: groupUtility,
	Args:    cobra.NoArgs,
	Example: `  kubeadapt version`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		info := versionInfo{
			Version:   version.Version,
			Commit:    version.Commit,
			Built:     version.Date,
			GoVersion: runtime.Version(),
			OSArch:    runtime.GOOS + "/" + runtime.GOARCH,
		}

		// root's PersistentPreRunE returns early for `version` without building
		// a RunContext, so the flag variable is the only source here.
		outFmt := outputFmt
		if rctx := getRunContext(cmd); rctx != nil && rctx.OutputFmt != "" {
			outFmt = rctx.OutputFmt
		}

		w := cmd.OutOrStdout()
		switch outFmt {
		case formatJSON:
			return output.RenderJSON(w, info)
		case formatYAML:
			return output.RenderYAML(w, info)
		}

		fmt.Fprintf(w, "kubeadapt %s\n", info.Version)
		fmt.Fprintf(w, "  Commit:     %s\n", info.Commit)
		fmt.Fprintf(w, "  Built:      %s\n", info.Built)
		fmt.Fprintf(w, "  Go version: %s\n", info.GoVersion)
		fmt.Fprintf(w, "  OS/Arch:    %s\n", info.OSArch)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
