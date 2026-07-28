# Kubeadapt CLI

[![Test Suite](https://github.com/kubeadapt/kubeadapt-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/kubeadapt/kubeadapt-cli/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/kubeadapt/kubeadapt-cli)](https://github.com/kubeadapt/kubeadapt-cli/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubeadapt/kubeadapt-cli)](https://goreportcard.com/report/github.com/kubeadapt/kubeadapt-cli)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

CLI for the [Kubeadapt](https://kubeadapt.io) Kubernetes cost optimization platform.
Browse clusters, workloads, nodes, and cost-saving recommendations from the
terminal, with JSON and YAML output for scripting.

```console
$ kubeadapt get clusters
╭──────────────────────────────────────┬─────────────────┬──────────┬───────────┬─────────────┬────────┬───────┬───────┬──────────╮
│ ID                                   │ Name            │ Provider │ Region    │ Environment │ Status │ CPU%  │ Mem%  │ $/hr     │
├──────────────────────────────────────┼─────────────────┼──────────┼───────────┼─────────────┼────────┼───────┼───────┼──────────┤
│ 3f8a1c02-5d41-4e7b-9a13-2c6e80f4b7d9 │ prod-cluster    │ aws      │ us-east-1 │ production  │ active │ 41.2% │ 63.8% │ $12.4700 │
│ 9b2e47d1-8c05-4f36-b1ae-7d40e9a2c518 │ staging-cluster │ aws      │ us-east-1 │ staging     │ active │ 11.5% │ 29.1% │ $4.2100  │
╰──────────────────────────────────────┴─────────────────┴──────────┴───────────┴─────────────┴────────┴───────┴───────┴──────────╯

Showing 2 (limit 100). End of results.

$ kubeadapt get recommendations --priority high -o json | jq -r '.data[0].savings.estimated_hourly.amount'
3.6584
```

## Installation

```bash
# Homebrew
brew install kubeadapt/tap/kubeadapt

# From source
go install github.com/kubeadapt/kubeadapt-cli/cmd/kubeadapt@latest
```

The Homebrew formula ships a prebuilt binary and has no other requirements. The
`go install` path needs Go 1.26 or newer.

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

Config file location. The first match wins:

1. `$XDG_CONFIG_HOME/kubeadapt/config.yaml`, if `XDG_CONFIG_HOME` is set
2. `~/.kubeadapt/config.yaml` (legacy), only if that file already exists
3. `os.UserConfigDir()/kubeadapt/config.yaml`, which is `~/.config/kubeadapt/config.yaml` on Linux and `~/Library/Application Support/kubeadapt/config.yaml` on macOS

The legacy path is probed before the platform config dir, so an existing `~/.kubeadapt/config.yaml` keeps being used after an upgrade. New installs never create it.

```yaml
api_url: https://public-api.kubeadapt.io
api_key: ka_your_api_key_here
```

**Environment variables** override the config file:

| Variable | Description |
|---|---|
| `KUBEADAPT_API_KEY` | API key |
| `KUBEADAPT_API_URL` | API endpoint |
| `NO_COLOR` | Any non-empty value disables colored output ([no-color.org](https://no-color.org)) |
| `KUBEADAPT_NO_UPDATE_CHECK` | Any non-empty value disables the update check (see [Update Check](#update-check)) |

`kubeadapt auth login` writes the config atomically (a sibling temp file, `fsync`, then `rename`) with `0600` permissions. An interrupted write cannot truncate the existing config. Re-running it overwrites only the API key and URL.

The key is verified before anything is written. If the API rejects it, the config is left untouched and existing credentials keep working. If the key cannot be reached at all (offline, DNS failure), it is saved and verified on first use.

## Commands

### Authentication

| Command | Description |
|---|---|
| `auth login` | Save API key and URL to config |
| `auth status` | Show current auth state; exits 0 only if the CLI can authenticate right now |
| `auth logout` | Remove stored credentials |

`auth status` is a scriptable gate. It exits 0 if and only if the API accepts the
current credentials. Both "unauthorized" and "no key configured" exit 1, because
neither can call the API. The full report is still printed first, in every output
format, so a failing run stays diagnosable.

```bash
kubeadapt auth status >/dev/null || exec kubeadapt auth login
```

`auth logout` distinguishes a missing config from an unreadable one. No config
file means there is nothing to remove, which is success (exit 0). A config that
exists but cannot be parsed or read exits 1 and removes nothing, because the API
key is still on disk.

`auth login --quiet` still emits the unverified-key warning on stderr. If the key
could not be checked (offline, DNS failure) it is saved and verified on first use,
and `--quiet` does not hide that.

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

`health` and `version` both honor `-o json` and `-o yaml`, so they can be consumed by scripts:

```bash
kubeadapt health -o json | jq .status
kubeadapt version -o json | jq -r .version
```

`health -o json` emits `status`, `url`, and `version` (omitted when the server does not report one). Under `--quiet`, the table form prints only the `Status:` line.

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

`get node-groups` rejects `--limit`, `--cursor`, `--paginate`, and `--include-total`. That endpoint returns every node group in one response, so a `--limit 2` would return all rows rather than two.

### Rate limits and `--max-wait`

`--paginate` reads the server's `X-RateLimit-*` headers and pauses before it would trip the limit. It also retries any 429 it does hit, honoring `Retry-After`. Pauses are announced on stderr and are interruptible with Ctrl-C.

`--max-wait` caps the total time spent waiting out rate limits across the whole run:

```bash
# Default: give up after 15m of cumulative waiting
kubeadapt get workloads --paginate -o json

# Wait as long as it takes
kubeadapt get workloads --paginate --max-wait 0 -o json

# Bail out early
kubeadapt get workloads --paginate --max-wait 90s -o json
```

A larger `--limit` (up to 500) cuts the request count and avoids most pauses. The CLI says so on stderr when `--paginate` runs with a small `--limit`.

### Resuming a partial run

A paginated run can stop early: the `--max-wait` budget runs out, the connection drops, you hit Ctrl-C. The pages already fetched are kept. The CLI writes what it collected to stdout, writes a resume cursor to **both stdout and stderr**, and **exits non-zero**.

The resume cursor reaches stdout in every output format: `-o json` and `-o yaml` carry it in the document, and table mode appends a `PARTIAL RESULTS:` line. A script that redirects stderr away still sees that the answer is truncated.

```bash
kubeadapt get workloads --paginate --max-wait 90s -o json > page1.json
# exit status 1
# stderr:
#   kubeadapt: partial results - 1400 items fetched before the run stopped.
#     reason: --max-wait budget exhausted: ...
#     resume with: --cursor=eyJpZCI6IjEyMyJ9

kubeadapt get workloads --paginate --max-wait 90s > page1.txt
# exit status 1
# stdout (after the table):
#   PARTIAL RESULTS: 1400 items fetched before the run stopped; resume with --cursor=eyJpZCI6IjEyMyJ9
# stderr: the same three-line report as above
```

Resume from that cursor:

```bash
kubeadapt get workloads --paginate --cursor eyJpZCI6IjEyMyJ9 -o json > page2.json
```

Under `-o json` and `-o yaml` the truncated document carries the cursor as structured fields, so a script never has to parse stderr - and table mode carries it on stdout too, via the `PARTIAL RESULTS:` line above:

```json
{
  "data": [],
  "meta": {},
  "partial": true,
  "resume_cursor": "eyJpZCI6IjEyMyJ9"
}
```

Check `partial` before treating the output as a complete answer. A cost total computed from a truncated page set is wrong, not merely incomplete.

Cursors expire (24h TTL) and are bound to both the endpoint and the query that minted them. If the resume cursor is rejected, the CLI says so and you need to re-run from the start for a consistent snapshot. Changing filters or sort between pages invalidates the cursor.

## Output Formats

```bash
kubeadapt get clusters -o table   # default
kubeadapt get clusters -o json
kubeadapt get clusters -o yaml
```

`-o` accepts only those three values. Anything else is an error; there is no fallback to `table`.

**YAML keys match JSON keys.** Both formats derive their key names from the same source, so `-o yaml` emits `is_stale`, `k8s_version`, `created_at`, `availability_zones`, `last_seen_at`, and `total_cores`. Keys are ordered alphabetically, not in struct-declaration order. Rely on key names, not position.

**Empty result sets** render as `[]`, never `null`, so `jq '.data[]'` and `jq '.data | length'` work against a no-match response without a guard.

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

Endpoints that don't support it (cluster, node, node-group, recommendation, team-assignments, organization overview) reject the flag with a clear error before sending any request. Team assignments record attribution, not cost.

## Unrecognized Filters

The API reports query parameters it did not recognize. When that happens on a `get` list command, the CLI warns on stderr:

```
kubeadapt: warning: the API ignored unknown parameter "namespaces" - results are NOT filtered by it.
```

The result set is unfiltered, not empty: the response looks like a real answer while being far too broad. The warning fires once per run rather than once per page, and `--quiet` does not suppress it.

## Global Flags

| Flag | Short | Description |
|---|---|---|
| `--api-key` | | API key (overrides config) |
| `--api-url` | | API endpoint (overrides config) |
| `--output` | `-o` | Output format: table, json, yaml |
| `--no-color` | | Disable colored output (see also `NO_COLOR`) |
| `--verbose` | `-v` | Debug logging, including the server `request_id` |
| `--quiet` | `-q` | Suppress non-essential output. Never suppresses a next-page cursor or a partial-run marker |
| `--config` | | Config file path |

`get` subcommands additionally accept `--cursor`, `--limit`, `--paginate`, `--include-total`, and `--max-wait`. See [Pagination](#pagination).

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `2` | Usage or pre-flight error. The invocation was rejected before any API call was made |
| `1` | Every other failure: API errors, network failures, authentication failures, partial results |

Exit code `2` covers an unknown command, subcommand, or flag; an invalid
`--output`, `--limit`, `--cost-mode`, `--max-wait`, or `--top-clusters-limit`; a
missing required flag; and a wrong positional argument count. Nothing was sent to
the API, so retrying without changing the command line cannot help.

An API-layer rejection such as HTTP 400 or 422 is **not** a usage error. The
request was well-formed enough to send, so it exits `1`.

```bash
kubeadapt get clusters -o xml        # exit 2 - invalid --output
kubeadapt get clusters --offset 1    # exit 2 - unknown flag
kubeadapt frobnicate                 # exit 2 - unknown command
kubeadapt get clusters               # exit 1 - no API key configured
```

Distinguish the two in a script:

```bash
kubeadapt get clusters -o json > out.json
case $? in
  0) ;;                                  # success
  2) echo "bad invocation; check the command line" >&2; exit 2 ;;
  *) echo "call failed; safe to retry" >&2; exit 1 ;;
esac
```

**Exit codes 3-125 are unassigned and callers must not depend on any specific
nonzero value beyond "failed".** Treat any nonzero code other than `2` as a
generic failure. Reserving that range keeps a finer-grained taxonomy possible
later without breaking existing scripts.

## Update Check

Release builds check GitHub at most once every 24 hours for a newer version and
print a one-line upgrade notice on stderr, after any error the command reported.
The check is skipped entirely when any of the following is true:

- `KUBEADAPT_NO_UPDATE_CHECK` is set to any non-empty value
- `CI` is set to any non-empty value
- stdout is not a terminal, such as when output is piped or redirected

That keeps the CLI off the network in pipelines and air-gapped environments,
where the 3-second lookup is pure latency and GitHub's unauthenticated per-IP
rate limit is shared across every job behind the same egress address. Builds
from source report version `dev` and never check.

## Shell Completions

```bash
source <(kubeadapt completion bash)
source <(kubeadapt completion zsh)
kubeadapt completion fish | source
```

Once loaded, tab completion resolves resource IDs from the API, so you don't
have to copy UUIDs out of a list command:

```bash
kubeadapt get cluster <TAB>
# 3f8a1c02-5d41-4e7b-9a13-2c6e80f4b7d9  prod-cluster
# 9b2e47d1-8c05-4f36-b1ae-7d40e9a2c518  staging-cluster

kubeadapt get workloads --cluster-id <TAB>   # same, for the flag
kubeadapt get recommendations --status <TAB> # pending applied dismissed archived
```

ID completion calls the API and is capped at 100 results. Enum flags complete
offline. If you aren't authenticated or the API is slow, completion returns
nothing rather than an error, so it never blocks your shell.

## Development

```bash
task build    # Build binary
task test     # Tests with race detector
task lint     # Linter
task fmt      # Format
task vuln     # Vulnerability check
```

## Contributing

Issues and pull requests are welcome.

Git hooks are managed with [lefthook](https://github.com/evilmartians/lefthook).
Run `lefthook install` once after cloning to enable them:

- **pre-commit** runs `gofmt` and `goimports` on staged Go files, `golangci-lint`
  on changes new since `HEAD~`, and a `trufflehog` scan for verified secrets.
- **commit-msg** runs `cz check`, so commit messages must follow
  [Conventional Commits](https://www.conventionalcommits.org). Allowed types and
  scopes are listed in `.cz.yaml`.
- **pre-push** runs the test suite with the race detector.

Run `task test` and `task lint` before opening a pull request.

## Security

Report vulnerabilities to **security@kubeadapt.io** rather than opening a public
issue. See [SECURITY.md](SECURITY.md) for the disclosure process, supported
versions, and how the CLI stores your API key.

## Support

Bug reports and feature requests belong in
[GitHub Issues](https://github.com/kubeadapt/kubeadapt-cli/issues).

## License

Apache License 2.0
