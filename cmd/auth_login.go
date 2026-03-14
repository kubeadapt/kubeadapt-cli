package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Kubeadapt API",
	Long:  `Authenticate with the Kubeadapt API by providing your API key. The key is stored securely in ~/.kubeadapt/config.yaml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("api-key")

		if key == "" {
			// Prompt for API key
			fmt.Print("Enter your Kubeadapt API key: ")
			if term.IsTerminal(int(os.Stdin.Fd())) {
				keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("reading API key: %w", err)
				}
				key = string(keyBytes)
				fmt.Println() // newline after hidden input
			} else {
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					key = scanner.Text()
				}
			}
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		// Load or create config
		c, err := config.Load(cfgFile)
		if err != nil {
			c = config.Default()
		}

		// Set the API key and optional URL override
		c.APIKey = key
		if apiURL != "" {
			c.APIURL = apiURL
		}

		if err := config.Save(c, cfgFile); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Authenticated successfully. API key saved to %s\n", config.DefaultPath())
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
}
