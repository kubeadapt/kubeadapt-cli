package cmd

import (
	"cmp"
	"fmt"
	"io"
	"os"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/logger"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/kubeadapt/kubeadapt-cli/internal/update"
	"github.com/spf13/cobra"
)

const (
	groupData    = "data"
	groupAuth    = "auth"
	groupUtility = "utility"
)

// Flag variables - still needed for cobra binding, but only used in PersistentPreRunE
// to populate the RunContext. Commands access state via getRunContext(cmd).
var (
	cfgFile   string
	apiURL    string
	apiKey    string
	outputFmt string
	noColor   bool
	verbose   bool
	quiet     bool
)

// The resolved path is printed rather than the precedence rules, because
// DefaultPath picks between XDG, a legacy ~/.kubeadapt, and os.UserConfigDir
// depending on what already exists, and only one of those is true per machine.
func configPathForHelp() string {
	if p := config.DefaultPath(); p != "" {
		return p
	}
	return "the platform config directory"
}

// Derived from config.Default() rather than restated, so the advertised
// endpoint cannot drift away from the one the CLI actually dials.
var rootLong = fmt.Sprintf(`Kubeadapt CLI provides command-line access to the Kubeadapt platform
for Kubernetes cost optimization, resource management, and recommendations.

Environment variables:
  KUBEADAPT_API_URL   Override the API endpoint (default: %s)
  KUBEADAPT_API_KEY   Provide the API key (overrides config file)
  NO_COLOR            Any non-empty value disables colored output
  KUBEADAPT_NO_UPDATE_CHECK  Any non-empty value disables the update check

Configuration lives at %s on this machine; the full lookup order is in the README.
Use 'kubeadapt auth login' to authenticate.`, config.Default().APIURL, configPathForHelp())

func validateOutputFmt(format string) error {
	switch format {
	case formatTable, formatJSON, formatYAML:
		return nil
	}
	return usagef("invalid --output %q (must be one of: %s, %s, %s)",
		format, formatTable, formatJSON, formatYAML)
}

var rootCmd = &cobra.Command{
	Use:           "kubeadapt",
	Short:         "Kubeadapt CLI - Kubernetes cost optimization",
	Long:          rootLong,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Presentation flags are resolved ahead of the early return below,
		// because login/version/completion still accept and honor them.
		if err := validateOutputFmt(outputFmt); err != nil {
			return err
		}
		// https://no-color.org - any non-empty value disables color.
		colorDisabled := noColor || os.Getenv("NO_COLOR") != ""
		output.SetNoColor(colorDisabled)
		output.SetQuiet(quiet)

		if cmd.Name() == "login" || cmd.Name() == "version" || cmd.Name() == "completion" {
			return nil
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			cfg = config.Default()
		}

		// Resolution order (first non-empty wins): CLI flag > env var > config file.
		cfg.APIURL = cmp.Or(apiURL, os.Getenv("KUBEADAPT_API_URL"), cfg.APIURL)
		cfg.APIKey = cmp.Or(apiKey, os.Getenv("KUBEADAPT_API_KEY"), cfg.APIKey)

		log, logErr := logger.New(verbose)
		if logErr != nil {
			return fmt.Errorf("initializing logger: %w", logErr)
		}

		rc := &RunContext{
			Config:    cfg,
			Logger:    log,
			OutputFmt: outputFmt,
			NoColor:   colorDisabled,
			Verbose:   verbose,
			Quiet:     quiet,
		}
		withRunContext(cmd, rc)

		return nil
	},
}

func Execute() {
	if code := run(os.Stderr); code != exitOK {
		os.Exit(code)
	}
}

// Everything except the os.Exit, so tests exercise the same wiring the binary
// does. markUsageErrors runs here rather than in an init: sibling init
// functions are what register the subcommands, and their order is not fixed.
func run(w io.Writer) int {
	markUsageErrors(rootCmd)
	err := rootCmd.Execute()

	if rc := getRunContext(rootCmd); rc != nil && rc.Logger != nil {
		_ = rc.Logger.Sync()
	}

	return report(w, err, update.CheckForUpdate)
}

// The upgrade banner is emitted after the failure, so a failed run ends on its
// own error rather than on an unrelated nag. banner is injected so the ordering
// stays assertable without a terminal.
func report(w io.Writer, err error, banner func() string) int {
	if err != nil {
		fmt.Fprintf(w, "Error: %s\n", friendlyError(err))
		if !verbose {
			fmt.Fprintln(w, "  Use --verbose for more details.")
		}
	}

	if !quiet {
		if msg := banner(); msg != "" {
			fmt.Fprintln(w, "")
			fmt.Fprintln(w, msg)
		}
	}

	return exitCodeFor(err)
}

func init() {
	rootCmd.AddGroup(
		&cobra.Group{ID: groupData, Title: "Data Commands:"},
		&cobra.Group{ID: groupAuth, Title: "Authentication:"},
		&cobra.Group{ID: groupUtility, Title: "Utility:"},
	)

	rootCmd.AddCommand(newHealthCmd())

	// Inherited by every subcommand via Command.FlagErrorFunc, so an unknown
	// flag or an unparseable flag value lands on the same exit code as the
	// validators below instead of being reported as a runtime failure.
	rootCmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return asUsageError(err)
	})

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.kubeadapt/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "Kubeadapt API URL")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Kubeadapt API key")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "Output format (table|json|yaml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
}
