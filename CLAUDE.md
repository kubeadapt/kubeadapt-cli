# kubeadapt-cli

## Commands
- Build: `task build` (GoReleaser snapshot, single-target)
- Build All Platforms: `task build-all`
- Test: `task test` (race detector + coverage)
- Lint: `task lint`
- Format: `task fmt`
- Install: `task install`
- Vuln: `task vuln`

## Architecture
Go 1.26 | Flat | Module: `github.com/kubeadapt/kubeadapt-cli`
Entry: `main.go`

```
cmd/                → Cobra commands (root, get_*, auth_*, tui, version, completion)
internal/
  api/              → HTTP client + types (client.go, errors.go, types/)
  config/           → Config file load/save (~/.kubeadapt/config.yaml)
  logger/           → Zap logger factory (New(debug bool))
  output/           → Table/JSON/YAML renderers
  tui/              → Bubbletea TUI (app, views, components)
  version/          → Version ldflags injection
```

## Debug Logging
Pass `--verbose` / `-v` to enable zap debug output to stderr:
```
kubeadapt get clusters --verbose
```
Logs API request method, path, response status, and round-trip duration.
Logger is initialized in `PersistentPreRunE` (cmd/root.go) and injected into
the API client via `api.WithLogger(log)` in `newAPIClient()` (cmd/helpers.go).

## Testing
Uses stdlib `testing` + `internal/testutil` mock HTTP server.
Race detection on by default via `task test`. No coverage threshold configured.

## Key Dependencies
- github.com/spf13/cobra          — CLI framework
- github.com/charmbracelet/bubbletea + bubbles + lipgloss — TUI
- go.uber.org/zap                 — Structured debug logging
- gopkg.in/yaml.v3                — Config file parsing
- github.com/guptarohit/asciigraph — ASCII charts in TUI

## Domain Context
CLI tool for the Kubeadapt platform. Authenticates via API key, queries kubeadapt-api
REST endpoints, renders data as tables/JSON/YAML or launches an interactive TUI dashboard.
Distributed as a standalone binary via GoReleaser (GitHub Releases + Homebrew tap + krew).
