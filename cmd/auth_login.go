package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Kubeadapt API",
	Long:  `Authenticate with the Kubeadapt API by providing your API key. The key is stored securely in ~/.kubeadapt/config.yaml.`,
	Example: `  kubeadapt auth login
  kubeadapt auth login --api-key your-key-here`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("api-key")
		if key == "" {
			var err error
			if key, err = promptAPIKey(cmd); err != nil {
				return err
			}
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		c, err := config.Load(cfgFile)
		if err != nil {
			c = config.Default()
		}
		if apiURL != "" {
			c.APIURL = apiURL
		}

		// Verify before touching disk. Writing first and rolling back on 401
		// destroyed a previously working key whenever the user fat-fingered a
		// new one, so the config must stay untouched until the key is proven.
		client := api.NewClient(c.APIURL, key)
		_, _, verifyErr := client.GetOrganization(cmd.Context())
		if verifyErr != nil && api.IsUnauthorized(verifyErr) {
			return fmt.Errorf("invalid API key: rejected by %s. Nothing was written; your existing credentials are unchanged", c.APIURL)
		}

		c.APIKey = key
		path := resolvedConfigPath()
		if err := config.Save(c, cfgFile); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		// A transport failure is not a rejection - the key may well be valid and
		// the user may be offline, so store it and defer verification. --quiet
		// cannot hide this: it is a caveat about state just written to disk.
		if verifyErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Could not verify API key (%v). Key saved to %s - it will be verified on first use.\n", verifyErr, path)
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Authenticated. API key saved to %s.\n", path)
		return nil
	},
}

// promptAPIKey reads a key from stdin. The prompt goes to stderr so that
// redirecting stdout to a file or pipe yields only command output.
func promptAPIKey(cmd *cobra.Command) (string, error) {
	fmt.Fprint(cmd.ErrOrStderr(), "Enter your Kubeadapt API key: ")

	if term.IsTerminal(int(os.Stdin.Fd())) {
		keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", fmt.Errorf("reading API key: %w", err)
		}
		fmt.Fprintln(cmd.ErrOrStderr())
		return string(keyBytes), nil
	}

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}

// resolvedConfigPath mirrors the path selection inside config.Save so success
// messages name the file that was actually written rather than the default.
func resolvedConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}
	return config.DefaultPath()
}

func init() {
	authCmd.AddCommand(authLoginCmd)
}
