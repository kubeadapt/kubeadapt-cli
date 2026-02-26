package cmd

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/logger"
	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	apiURL    string
	apiKey    string
	outputFmt string
	noColor   bool
	verbose   bool
	cfg       *config.Config
	log       *zap.Logger
)

var rootCmd = &cobra.Command{
	Use:   "kubeadapt",
	Short: "KubeAdapt CLI - Kubernetes cost optimization",
	Long:  `KubeAdapt CLI provides command-line access to the KubeAdapt platform for Kubernetes cost optimization, resource management, and recommendations.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for auth login and version commands
		if cmd.Name() == "login" || cmd.Name() == "version" || cmd.Name() == "completion" {
			return nil
		}

		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			// Non-fatal: config might not exist yet
			cfg = config.Default()
		}

		// Environment variables override config values
		if envURL := os.Getenv("KUBEADAPT_API_URL"); envURL != "" {
			cfg.APIURL = envURL
		}
		if envKey := os.Getenv("KUBEADAPT_API_KEY"); envKey != "" {
			cfg.APIKey = envKey
		}

		// CLI flags override everything
		if apiURL != "" {
			cfg.APIURL = apiURL
		}
		if apiKey != "" {
			cfg.APIKey = apiKey
		}

		var logErr error
		log, logErr = logger.New(verbose)
		if logErr != nil {
			return fmt.Errorf("initializing logger: %w", logErr)
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if log != nil {
		_ = log.Sync()
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.kubeadapt/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "KubeAdapt API URL")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "KubeAdapt API key")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "Output format (table|json|yaml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}
