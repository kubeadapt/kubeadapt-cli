package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for kubeadapt.

To load completions:

Bash:
  $ source <(kubeadapt completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ kubeadapt completion bash > /etc/bash_completion.d/kubeadapt
  # macOS:
  $ kubeadapt completion bash > $(brew --prefix)/etc/bash_completion.d/kubeadapt

Zsh:
  $ source <(kubeadapt completion zsh)
  # To load completions for each session, execute once:
  $ kubeadapt completion zsh > "${fpath[1]}/_kubeadapt"

Fish:
  $ kubeadapt completion fish | source
  # To load completions for each session, execute once:
  $ kubeadapt completion fish > ~/.config/fish/completions/kubeadapt.fish

PowerShell:
  PS> kubeadapt completion powershell | Out-String | Invoke-Expression
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
