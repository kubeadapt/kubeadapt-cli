# Kubeadapt CLI

CLI for the [Kubeadapt](https://kubeadapt.io) Kubernetes cost optimization platform.

## Installation

```bash
# Homebrew
brew install kubeadapt/tap/kubeadapt

# From source
go install github.com/kubeadapt/kubeadapt-cli@latest
```

## Quick Start

```bash
kubeadapt auth login
kubeadapt get overview
kubeadapt get clusters
kubeadapt get recommendations
```

## Commands

```bash
# Auth
kubeadapt auth login / status / logout

# Resources
kubeadapt get overview
kubeadapt get dashboard
kubeadapt get clusters
kubeadapt get cluster <id>
kubeadapt get workloads
kubeadapt get nodes
kubeadapt get recommendations
kubeadapt get costs teams
kubeadapt get costs departments
kubeadapt get node-groups
kubeadapt get namespaces
kubeadapt get persistent-volumes
```

Most commands support filters:

```bash
kubeadapt get workloads --cluster-id cls-001 --namespace default --kind Deployment
kubeadapt get recommendations --type rightsizing --status active
```

## Output Formats

```bash
kubeadapt get clusters -o table   # default
kubeadapt get clusters -o json
kubeadapt get clusters -o yaml
```

## Configuration

Config at `~/.kubeadapt/config.yaml`:

```yaml
api_url: https://public-api.kubeadapt.io
api_key: ka_your_api_key_here
```

`KUBEADAPT_API_KEY` env var overrides the stored config.

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--api-key` | | API key (overrides config) |
| `--output` | `-o` | Output format: table, json, yaml |
| `--no-color` | | Disable colored output |
| `--verbose` | `-v` | Debug logging |
| `--config` | | Config file path |

## Shell Completions

```bash
source <(kubeadapt completion bash)
source <(kubeadapt completion zsh)
kubeadapt completion fish | source
```

## Development

```bash
task build    # Build binary
task test     # Tests with race detector
task lint     # Linter
task fmt      # Format
task vuln     # Vulnerability check
```

## License

Apache License 2.0
