# KubeAdapt CLI

Command-line interface for the [KubeAdapt](https://kubeadapt.com) Kubernetes cost optimization platform.

## Installation

### Homebrew

```bash
brew install kubeadapt/tap/kubeadapt-cli

# Upgrade to latest version
brew upgrade kubeadapt/tap/kubeadapt-cli
```

### Install Script

```bash
curl -sSL https://raw.githubusercontent.com/kubeadapt/kubeadapt-cli/main/scripts/install.sh | bash
```

### From Source

```bash
go install github.com/kubeadapt/kubeadapt-cli@latest
```

### Build from Source

```bash
git clone https://github.com/kubeadapt/kubeadapt-cli.git
cd kubeadapt-cli
task build
```

## Quick Start

```bash
# Authenticate
kubeadapt auth login

# View organization overview
kubeadapt get overview

# List clusters
kubeadapt get clusters

# View cost optimization recommendations
kubeadapt get recommendations
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
kubeadapt get dashboard            # Organization dashboard with cost trends
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
```

Most `get` commands support optional filters:

```bash
kubeadapt get workloads --cluster-id cls-001 --namespace default --kind Deployment
kubeadapt get nodes --cluster-id cls-001 --limit 10
kubeadapt get recommendations --type rightsizing --status active
kubeadapt get dashboard --days 7
```

## Output Formats

All `get` commands support the `--output` flag:

```bash
kubeadapt get clusters -o table   # Default styled table
kubeadapt get clusters -o json    # JSON
kubeadapt get clusters -o yaml    # YAML
```

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--api-key` | | API key (overrides stored config) |
| `--output` | `-o` | Output format: table, json, yaml |
| `--no-color` | | Disable colored output |
| `--verbose` | `-v` | Enable debug logging |
| `--config` | | Config file path |

## Configuration

Config is stored at `~/.kubeadapt/config.yaml`:

```yaml
api_url: https://api.kubeadapt.com
api_key: ka_your_api_key_here
```

Environment variable `KUBEADAPT_API_KEY` overrides the stored config.

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
task build    # Build binary
task test     # Run tests with race detector
task lint     # Run linter
task fmt      # Format code
task vuln     # Run vulnerability check
```

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.
