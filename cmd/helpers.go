package cmd

import (
	"fmt"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
)

func newAPIClient() (*api.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("not authenticated. Run 'kubeadapt auth login' first")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("no API key configured. Run 'kubeadapt auth login' first")
	}
	return api.NewClient(cfg.APIURL, cfg.APIKey), nil
}

func renderOutput(format string, data interface{}, tableFunc func()) error {
	switch format {
	case "json":
		return output.JSON(data)
	case "yaml":
		return output.YAML(data)
	default:
		tableFunc()
		return nil
	}
}
