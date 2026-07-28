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

## Documentation

Full command reference, flags, and usage guides live at
**<https://kubeadapt.io/docs/v1/cli/overview>**. This page covers install and first run only.

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
kubeadapt auth login                          # authenticate
kubeadapt health                              # connectivity probe, no auth required
kubeadapt get clusters
kubeadapt get workloads --cluster-id <id>
kubeadapt get recommendations --priority high
```

The CLI covers clusters, workloads, nodes, namespaces, recommendations, teams,
and departments. Run `kubeadapt get --help` for the list, or read the full
reference at <https://kubeadapt.io/docs/v1/cli/overview>.

## Configuration

By default the config file is `config.yaml` under your platform config
directory, such as `~/.config/kubeadapt/` on Linux. `XDG_CONFIG_HOME` and an
existing legacy `~/.kubeadapt/config.yaml` take precedence; see the docs for the
full lookup order.

```yaml
api_url: https://public-api.kubeadapt.io
api_key: ka_your_api_key_here
```

Environment variables override the config file:

| Variable | Description |
|---|---|
| `KUBEADAPT_API_KEY` | API key |
| `KUBEADAPT_API_URL` | API endpoint |
| `NO_COLOR` | Any non-empty value disables colored output ([no-color.org](https://no-color.org)) |
| `KUBEADAPT_NO_UPDATE_CHECK` | Any non-empty value disables the update check |

`kubeadapt auth login` writes the config atomically with `0600` permissions, and
verifies the key before writing anything.

## More

- **Output formats:** `-o table|json|yaml`. Money fields in JSON and YAML are
  decimal strings, not raw floats.
- **Pagination:** cursor-based. `--paginate` fetches all pages; a run that stops
  early exits non-zero and prints a resume cursor.
- **Cost mode:** `--cost-mode fully_loaded` (default) or `workload_only` on the
  endpoints that support cost attribution.
- **Unrecognized filters:** the CLI warns on stderr when the API ignores a query
  parameter, because the result set is unfiltered rather than empty.
- **Shell completions:** `kubeadapt completion bash|zsh|fish`.
- **Global flags:** `--api-key`, `--api-url`, `-o/--output`, `--no-color`,
  `-v/--verbose`, `-q/--quiet`, `--config`.

Details for all of these are at <https://kubeadapt.io/docs/v1/cli/overview>.

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `2` | Usage or pre-flight error. The invocation was rejected before any API call was made |
| `1` | Every other failure: API errors, network failures, authentication failures, partial results |

An API-layer rejection such as HTTP 400 or 422 is **not** a usage error. The
request was well-formed enough to send, so it exits `1`.

**Exit codes 3-125 are unassigned and callers must not depend on any specific
nonzero value beyond "failed".** Treat any nonzero code other than `2` as a
generic failure. Reserving that range keeps a finer-grained taxonomy possible
later without breaking existing scripts.

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

Git hooks are managed with [lefthook](https://github.com/evilmartians/lefthook);
run `lefthook install` once after cloning. Commit messages follow
[Conventional Commits](https://www.conventionalcommits.org), with allowed types
and scopes listed in `.cz.yaml`. Run `task test` and `task lint` before opening
a pull request.

## Security

Report vulnerabilities to **security@kubeadapt.io** rather than opening a public
issue. See [SECURITY.md](SECURITY.md) for the disclosure process, supported
versions, and how the CLI stores your API key.

## Support

Bug reports and feature requests belong in
[GitHub Issues](https://github.com/kubeadapt/kubeadapt-cli/issues).

## License

Apache License 2.0
