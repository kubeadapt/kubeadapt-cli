package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
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

// The key is sent as a bearer token on every request, so a cleartext endpoint
// exposes it to anything on the path. Loopback is exempt because a locally-run
// API never puts the key on a network, and blocking it would break local dev.
func requireSecureAPIURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return usagef("invalid --api-url %q: %v", raw, err)
	}
	switch {
	case u.Scheme == "https":
		return nil
	case u.Scheme == "http" && isLoopbackHost(u.Hostname()):
		return nil
	case u.Scheme == "":
		return usagef("invalid --api-url %q: missing scheme (expected https://...)", raw)
	}
	return usagef("refusing to send the API key in cleartext to %q. "+
		"Use https://, or http:// with a loopback host (localhost, 127.0.0.1, ::1) for local development", raw)
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
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
