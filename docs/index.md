# kubeadapt-cli

Command-line interface for the [KubeAdapt](https://kubeadapt.com) Kubernetes cost optimization platform.

## Installation

### Homebrew (recommended)

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

## Usage

```bash
kubeadapt-cli [command] [flags]
```

### Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `table` | Output format: `table`, `json`, `yaml` |
| `--verbose` | `-v` | `false` | Enable debug logging |
| `--api-key` | | | API key (overrides stored config) |
| `--no-color` | | `false` | Disable colored output |
| `--config` | | | Config file path |

### Commands

| Command | Description |
|---------|-------------|
| `auth login` | Store API key |
| `auth status` | Show current auth status |
| `auth logout` | Remove stored credentials |
| `get overview` | Organization dashboard |
| `get clusters` | List all clusters |
| `get workloads` | List workloads |
| `get recommendations` | Cost optimization suggestions |
| `get costs teams` | Cost breakdown by team |
| `get costs departments` | Cost breakdown by department |
| `version` | Print version information |
| `completion` | Generate shell completion scripts |

## Configuration

Config file is stored at `~/.kubeadapt/config.yaml`.

Environment variable `KUBEADAPT_API_KEY` overrides the stored config.

## Development

```bash
task test          # Run tests with race detector
task lint          # Run linter
task build         # Build for current platform
task build-all     # Build for all platforms
```
