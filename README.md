# KubeAdapt CLI

Command-line interface and interactive TUI for the [KubeAdapt](https://kubeadapt.com) Kubernetes cost optimization platform.

## Installation

### From Source

```bash
go install github.com/kubeadapt/kubeadapt-cli@latest
```

### From Release

```bash
curl -sSL https://raw.githubusercontent.com/kubeadapt/kubeadapt-cli/main/scripts/install.sh | bash
```

### Build from Source

```bash
git clone https://github.com/kubeadapt/kubeadapt-cli.git
cd kubeadapt-cli
make build
```

## Quick Start

```bash
# Authenticate
kubeadapt auth login

# View organization overview
kubeadapt get overview

# List clusters
kubeadapt get clusters

# Launch interactive TUI
kubeadapt tui
```

## Commands

### Authentication

```bash
kubeadapt auth login          # Store API key
kubeadapt auth status         # Show current auth status
kubeadapt auth logout         # Remove stored credentials
```

### Resources

```bash
kubeadapt get overview             # Organization dashboard
kubeadapt get clusters             # List all clusters
kubeadapt get cluster <id>         # Single cluster details
kubeadapt get workloads            # List workloads
kubeadapt get nodes                # List nodes
kubeadapt get recommendations      # Cost optimization suggestions
kubeadapt get costs teams          # Cost breakdown by team
kubeadapt get costs departments    # Cost breakdown by department
kubeadapt get node-groups          # List node groups
kubeadapt get namespaces           # List namespaces
kubeadapt get persistent-volumes   # List persistent volumes
kubeadapt get integrations         # List integrations
kubeadapt get integration <id>     # Single integration details
```

### Integration Management

```bash
kubeadapt create integration --name "Slack Alerts" --type slack --config '{"channel":"#alerts"}'
kubeadapt update integration <id> --enabled
kubeadapt delete integration <id> --yes
```

### Interactive TUI

```bash
kubeadapt tui
```

Navigate with number keys (1-9), j/k for rows, Tab for sub-tabs, r to refresh, ? for help, q to quit.

## Output Formats

All `get` commands support `--output` flag:

```bash
kubeadapt get clusters --output table   # Default styled table
kubeadapt get clusters --output json    # JSON
kubeadapt get clusters --output yaml    # YAML
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--api-url` | KubeAdapt API URL |
| `--api-key` | API key (overrides stored config) |
| `--output` | Output format: table, json, yaml |
| `--no-color` | Disable colored output |
| `--config` | Config file path |
| `--verbose` | Verbose output |

## Configuration

Config is stored at `~/.kubeadapt/config.yaml`:

```yaml
api_url: https://api.kubeadapt.com
api_key: ka_your_api_key_here
```

## Shell Completions

```bash
# Bash
source <(kubeadapt completion bash)

# Zsh
source <(kubeadapt completion zsh)

# Fish
kubeadapt completion fish | source
```

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linter
make all      # Format, vet, lint, test, build
```

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
