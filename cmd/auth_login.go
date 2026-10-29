package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
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
	Long: `Authenticate with the Kubeadapt API by providing your API key. The key is
verified, then saved to the config file (see --config) readable only by you.`,
	Example: `  # Prompt for the key; input is hidden
  kubeadapt auth login

  # Pipe it from a secret store instead of typing it
  pass show kubeadapt/api-key | kubeadapt auth login

  # Or skip the config file and read the key from the environment
  export KUBEADAPT_API_KEY="$(pass show kubeadapt/api-key)"`,
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
			// Only a genuinely absent file means there is nothing to preserve.
			// A parse or permission failure has a real config behind it, and
			// starting from Default() would overwrite a working key.
			if !errors.Is(err, fs.ErrNotExist) {
				return fmt.Errorf("cannot save credentials: %w", err)
			}
			c = config.Default()
		}
		if apiURL != "" {
			c.APIURL = apiURL
		}
		if err := requireSecureAPIURL(c.APIURL); err != nil {
			return err
		}

		// Verify before touching disk. Writing first and rolling back on 401
		// destroyed a previously working key whenever the user fat-fingered a
		// new one, so the config must stay untouched until the key is proven.
		client := api.NewClient(c.APIURL, key)
		_, _, verifyErr := client.GetOrganization(cmd.Context())
		if apiErr, ok := errors.AsType[*api.APIError](verifyErr); ok {
			return rejectedKeyError(c.APIURL, apiErr)
		}

		c.APIKey = key
		path := resolvedConfigPath()
		if err := config.Save(c, cfgFile); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		// Only a transport failure reaches here: the key may well be valid and
		// the user may be offline or behind a proxy, so store it and defer
		// verification. --quiet cannot hide this: it is a caveat about state
		// just written to disk.
		if verifyErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Could not verify API key (%v). Key saved to %s - it will be verified on first use.\n", verifyErr, path)
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Authenticated. API key saved to %s.\n", path)
		return nil
	},
}

// A response of any kind proves the server was reachable and still did not
// accept the key. Persisting it would replace a working credential with one
// already known to be unusable, so nothing is written.
func rejectedKeyError(url string, e *api.APIError) error {
	const unchanged = "Nothing was written; your existing credentials are unchanged"
	if e.IsAuthError() {
		return fmt.Errorf("invalid API key: rejected by %s. %s", url, unchanged)
	}
	return fmt.Errorf("could not verify API key against %s: %s. %s", url, e, unchanged)
}

// promptAPIKey reads a key from the command's input. The prompt goes to stderr
// so that redirecting stdout to a file or pipe yields only command output.
func promptAPIKey(cmd *cobra.Command) (string, error) {
	fmt.Fprint(cmd.ErrOrStderr(), "Enter your Kubeadapt API key: ")

	in := cmd.InOrStdin()
	// Echo suppression needs a real descriptor, so a non-file reader (a test's
	// buffer, an embedded caller's pipe) must fall through to the plain read.
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		keyBytes, err := term.ReadPassword(int(f.Fd()))
		if err != nil {
			return "", fmt.Errorf("reading API key: %w", err)
		}
		fmt.Fprintln(cmd.ErrOrStderr())
		return string(keyBytes), nil
	}

	scanner := bufio.NewScanner(in)
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
