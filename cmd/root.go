package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/logger"
	"github.com/kubeadapt/kubeadapt-cli/internal/update"
	"github.com/spf13/cobra"
)

const (
	groupData    = "data"
	groupAuth    = "auth"
	groupUtility = "utility"
)

// Flag variables — still needed for cobra binding, but only used in PersistentPreRunE
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

var rootCmd = &cobra.Command{
	Use:   "kubeadapt",
	Short: "Kubeadapt CLI - Kubernetes cost optimization",
	Long: `Kubeadapt CLI provides command-line access to the Kubeadapt platform
for Kubernetes cost optimization, resource management, and recommendations.

Environment variables:
  KUBEADAPT_API_URL   Override the API endpoint (default: https://public-api.kubeadapt.io)
  KUBEADAPT_API_KEY   Provide the API key (overrides config file)

Configuration is stored in ~/.kubeadapt/config.yaml (or $XDG_CONFIG_HOME/kubeadapt/config.yaml).
Use 'kubeadapt auth login' to authenticate.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "login" || cmd.Name() == "version" || cmd.Name() == "completion" {
			return nil
		}

		var cfg *config.Config
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			cfg = config.Default()
		}

		if envURL := os.Getenv("KUBEADAPT_API_URL"); envURL != "" {
			cfg.APIURL = envURL
		}
		if envKey := os.Getenv("KUBEADAPT_API_KEY"); envKey != "" {
			cfg.APIKey = envKey
		}

		if apiURL != "" {
			cfg.APIURL = apiURL
		}
		if apiKey != "" {
			cfg.APIKey = apiKey
		}

		log, logErr := logger.New(verbose)
		if logErr != nil {
			return fmt.Errorf("initializing logger: %w", logErr)
		}

		rc := &RunContext{
			Config:    cfg,
			Logger:    log,
			OutputFmt: outputFmt,
			NoColor:   noColor,
			Verbose:   verbose,
			Quiet:     quiet,
		}
		withRunContext(cmd, rc)

		return nil
	},
}

// Execute runs the root command with proper error handling.
func Execute() {
	err := rootCmd.Execute()

	// Sync logger if it was initialized
	if rc := getRunContext(rootCmd); rc != nil && rc.Logger != nil {
		_ = rc.Logger.Sync()
	}

	if updateMsg := update.CheckForUpdate(); updateMsg != "" {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, updateMsg)
	}

	if err == nil {
		return
	}

	// FlagError → show usage + exit 2
	var flagErr *FlagError
	if errors.As(err, &flagErr) {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", flagErr.Err)
		// Don't print usage for the root command (too verbose), only for subcommands
		os.Exit(2)
	}

	// All other errors → friendly message + exit 1
	fmt.Fprintf(os.Stderr, "Error: %s\n", friendlyError(err))
	if !verbose {
		fmt.Fprintln(os.Stderr, "  Use --verbose for more details.")
	}
	os.Exit(1)
}

func init() {
	rootCmd.AddGroup(
		&cobra.Group{ID: groupData, Title: "Data Commands:"},
		&cobra.Group{ID: groupAuth, Title: "Authentication:"},
		&cobra.Group{ID: groupUtility, Title: "Utility:"},
	)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.kubeadapt/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "Kubeadapt API URL")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Kubeadapt API key")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "Output format (table|json|yaml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
}
