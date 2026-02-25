# replace-me

Replace with a description of your CLI tool.

## Installation

```bash
curl -sSL https://raw.githubusercontent.com/kubeadapt/replace-me/main/scripts/install.sh | bash
```

Or via `go install`:

```bash
go install github.com/kubeadapt/replace-me@latest
```

## Usage

```bash
replace-me [command] [flags]
```

### Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `table` | Output format: `table`, `json`, `yaml` |
| `--verbose` | `-v` | `false` | Enable verbose output |
| `--config` | | | Config file path |

### Commands

| Command | Description |
|---------|-------------|
| `version` | Print version information |
| `completion` | Generate shell completion scripts |
| `help` | Help about any command |

## Configuration

Config file is loaded from `$HOME/.replace-me/config.yaml` by default.

## Development

```bash
task test          # Run tests
task lint          # Run linter
task build         # Build for current platform
task build-all     # Build for all platforms
task release-local # Build full release snapshot locally
```
