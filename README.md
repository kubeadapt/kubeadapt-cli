# Kubeadapt CLI

CLI for the [Kubeadapt](https://kubeadapt.io) Kubernetes cost optimization platform.

## Installation

```bash
# Homebrew
brew install kubeadapt/tap/kubeadapt

# From source
go install github.com/kubeadapt/kubeadapt-cli@latest
```

## Quickstart

```bash
# Authenticate
kubeadapt auth login

# Check connectivity (no auth required)
kubeadapt health

# Browse your data
kubeadapt get clusters
kubeadapt get workloads --cluster-id <id>
kubeadapt get recommendations --priority high
```

## Configuration

Config file location (in order of precedence):

1. `$XDG_CONFIG_HOME/kubeadapt/config.yaml`
2. `~/.config/kubeadapt/config.yaml`
3. `~/.kubeadapt/config.yaml` (legacy fallback, still supported)

```yaml
api_url: https://api.kubeadapt.io
api_key: ka_your_api_key_here
```

**Environment variables** override the config file:

| Variable | Description |
|---|---|
| `KUBEADAPT_API_KEY` | API key |
| `KUBEADAPT_API_URL` | API endpoint |

`kubeadapt auth login` writes the config atomically with `0600` permissions. Running it again overwrites only the API key and URL — it's safe to re-run.

## Commands

### Authentication

| Command | Description |
|---|---|
| `auth login` | Save API key and URL to config |
| `auth status` | Show current auth state |
| `auth logout` | Remove stored credentials |

### Organization

| Command | Description |
|---|---|
| `get overview` | Organization-level cost summary |
| `get dashboard` | Month-to-date billed cost, savings potential, and top clusters |

### Clusters

| Command | Description |
|---|---|
| `get clusters` | List clusters |
| `get cluster <id>` | Show a single cluster |

### Workloads

| Command | Description |
|---|---|
| `get workloads` | List workloads |
| `get workload <uid>` | Show a single workload |
| `get pods <workload-uid>` | List pods for a workload |

### Nodes

| Command | Description |
|---|---|
| `get nodes` | List nodes |
| `get node <uid>` | Show a single node |
| `get node-groups` | List node groups |
| `get node-group <name>` | Show a single node group (requires `--cluster-id`) |

### Namespaces

| Command | Description |
|---|---|
| `get namespaces` | List namespaces |
| `get namespace <name>` | Show a single namespace (requires `--cluster-id`) |

### Recommendations

| Command | Description |
|---|---|
| `get recommendations` | List cost-saving recommendations |
| `get recommendation <id>` | Show a single recommendation |

### Teams & Departments

| Command | Description |
|---|---|
| `get teams` | List teams with cost attribution |
| `get team <id>` | Show a single team |
| `get team-assignments <team-id>` | List entity assignments for a team |
| `get departments` | List departments with cost attribution |
| `get department <id>` | Show a single department |

### Utility

| Command | Description |
|---|---|
| `health` | Unauthenticated connectivity probe to `GET /health` |
| `version` | Print CLI version |
| `completion <shell>` | Generate shell completion script |

## Pagination

All `get *` list commands use cursor-based pagination.

```bash
# Fetch the first page (default limit: 100)
kubeadapt get workloads --limit 50

# Fetch the next page using the cursor from the previous response
kubeadapt get workloads --cursor eyJpZCI6IjEyMyJ9 --limit 50

# Fetch all pages automatically
kubeadapt get workloads --paginate -o json

# Include total count in metadata (costs an extra DB query)
kubeadapt get workloads --include-total
```

`--offset` is not supported. Cobra rejects it as an unknown flag.

## Output Formats

```bash
kubeadapt get clusters -o table   # default
kubeadapt get clusters -o json
kubeadapt get clusters -o yaml
```

**Money fields** in JSON/YAML output are decimal strings, not raw floats:

```json
{
  "cost": {
    "current_run_rate_hourly": {
      "amount": "12.4700",
      "currency": "USD"
    }
  }
}
```

Extract the numeric value with `jq`:

```bash
kubeadapt get workloads -o json | jq '.data[].cost.current_run_rate_hourly.amount | tonumber'
```

## Cost Mode

The `--cost-mode` flag controls cost attribution for namespace, workload, pod, team, and department endpoints:

```bash
kubeadapt get workloads --cost-mode fully_loaded    # default: includes node overhead
kubeadapt get workloads --cost-mode workload_only   # workload resource cost only
```

Endpoints that don't support it (cluster, node, node-group, recommendation, organization overview) reject the flag with a clear error before sending any request.

## Global Flags

| Flag | Short | Description |
|---|---|---|
| `--api-key` | | API key (overrides config) |
| `--api-url` | | API endpoint (overrides config) |
| `--output` | `-o` | Output format: table, json, yaml |
| `--no-color` | | Disable colored output |
| `--verbose` | `-v` | Debug logging |
| `--quiet` | `-q` | Suppress non-essential output |
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
