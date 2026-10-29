# Changelog

All notable changes to the Kubeadapt CLI are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-10-29

### Breaking

- **`-o yaml` key names changed.** YAML keys are now derived from the same
  source as JSON keys, so they are snake_case instead of a lowercased Go field
  name. The renames:

  | Old key | New key |
  |---|---|
  | `isstale` | `is_stale` |
  | `k8sversion` | `k8s_version` |
  | `createdat` | `created_at` |
  | `availabilityzones` | `availability_zones` |
  | `lastseenat` | `last_seen_at` |
  | `totalcores` | `total_cores` |

  YAML keys are also now ordered alphabetically rather than in
  struct-declaration order.

  **Migration:** update any `yq` expression that targets an old key, for example
  `yq '.data[].isstale'` becomes `yq '.data[].is_stale'`. Anything reading YAML
  by key position must switch to reading by key name. `-o json` consumers are
  unaffected; the JSON keys did not change, and YAML now matches them.

- **`--paginate` exits non-zero when it returns partial results.** A run cut
  short by an exhausted `--max-wait` budget, a dropped connection, or Ctrl-C now
  prints the pages it collected to stdout and exits `1`. Previously it returned
  an error and no data at all.

  **Migration:** a script that treats any non-zero exit from `--paginate` as
  "no output" will now discard real data. Check the `partial` field before using
  the payload, and resume from `resume_cursor`:

  ```bash
  kubeadapt get workloads --paginate -o json > out.json || true
  if [ "$(jq -r '.partial // false' out.json)" = "true" ]; then
    cursor=$(jq -r .resume_cursor out.json)
    kubeadapt get workloads --paginate --cursor "$cursor" -o json > out2.json
  fi
  ```

  Under `-o table` the collected rows are still rendered, followed by a
  `PARTIAL RESULTS:` line on stdout carrying the resume cursor. The same
  details also go to stderr.

- **`get node-groups` rejects `--limit`, `--cursor`, `--paginate`, and
  `--include-total`.** That endpoint returns every node group in a single
  response, so those flags never did anything. `--limit 2` previously returned
  all rows while appearing to work.

  **Migration:** drop the flag. `kubeadapt get node-groups --paginate -o json`
  becomes `kubeadapt get node-groups -o json`. `--cost-mode` was already
  rejected and still is.

- **`-o` rejects unknown formats.** Only `table`, `json`, and `yaml` are
  accepted. Anything else is an error instead of a silent fall back to `table`.

  **Migration:** fix the typo. A pipeline that ran `-o josn` and parsed the
  result was already parsing a table, not JSON; it now fails loudly at the
  point of the mistake.

- **Source installs use `cmd/kubeadapt`.** `go install
  github.com/kubeadapt/kubeadapt-cli@latest` built a binary named
  `kubeadapt-cli`. Install with
  `go install github.com/kubeadapt/kubeadapt-cli/cmd/kubeadapt@latest` and
  delete the old binary. Homebrew is unaffected.

- **Cleartext API URLs are refused.** An `http://` URL that is not loopback,
  from `--api-url`, `KUBEADAPT_API_URL` or the config file, exits `2` before
  any request is sent. A redirect to one is refused too. Use `https://`.

- **A broken or missing config file is an error.** A file that does not parse
  exits `1` and names the file. A `--config` path that does not exist exits
  `2`. A missing file at the default location still means defaults.

- **`auth status` exits `1` unless the CLI can authenticate**: no key, a
  rejected key or an unreachable API. The report is still printed.

- **Invalid flag values exit `2` before any request.** Every flag with a fixed
  set of values (`--status`, `--kind`, `--entity-type`, `--cost-mode` and the
  rest) is checked locally, and the error lists the accepted values.
  `--cost-mode` on a command that does not accept it also exits `2`.

- **`get team-assignments --entity-type` takes `workload`, `namespace` or
  `pv`.** `cluster` was never accepted by the API.

### Added

- Dynamic shell completion for resource IDs. `kubeadapt get cluster <TAB>` now
  resolves real IDs from the API and shows each one's name alongside it, on all
  eight detail commands and on `--cluster-id`. Enum flags (`--status`,
  `--priority`, `--provider`, and 12 others) complete offline. Completion is
  capped at 100 results and fails silently - an expired key or a slow API
  yields no candidates rather than an error, and never blocks the shell.
- `--max-wait` on `get` subcommands. Caps the total time `--paginate` spends
  waiting out API rate limits before it returns partial results. Default `15m`,
  `0` means unbounded. Negative values are rejected.
- `--paginate` now paces itself against the server's `X-RateLimit-*` headers,
  pausing before it would trip the limit instead of absorbing a 429 and its flat
  60-second `Retry-After`. It also retries a 429 it does hit, honoring
  `Retry-After`. Pauses are announced on stderr and are interruptible with
  Ctrl-C. Single (non-paginated) requests still fail fast on 429 rather than
  retrying.
- Resume cursors for truncated `--paginate` runs. Under `-o json` and `-o yaml`
  the document carries `partial: true` and `resume_cursor`; under any format the
  cursor and the reason are written to stderr.
- `--paginate` with a small `--limit` now prints a one-time stderr hint that a
  larger `--limit` (up to 500) cuts the request count and avoids most
  rate-limit pauses.
- `NO_COLOR` environment variable is honored per [no-color.org](https://no-color.org).
  Any non-empty value disables colored output, in addition to `--no-color`.
- `version` and `health` support `-o json` and `-o yaml`, making
  `kubeadapt health -o json | jq .status` and
  `kubeadapt version -o json | jq -r .version` possible.
- `get team-assignments` rejects `--cost-mode`, matching the nine other
  commands whose endpoints do not accept it. An assignment records attribution,
  not cost.
- `--cluster-id` on `get workload`, `get node` and `get pods`, for a UID that
  exists in more than one cluster. Without it, `get workload` and `get node`
  fail and print the cluster IDs as ready-to-paste flags; `get pods` lists
  pods from every such cluster.
- `-o json` and `-o yaml` keep each cost block's breakdown (`components`,
  `unallocated`, `basis`, `partial`) and `meta.basis`.
- A warning on stderr, with the `chmod 600` command, when the config file
  holds an API key that other users can read. Not shown on Windows.
- Specific messages for `UPSTREAM_UNAVAILABLE`, `UPSTREAM_TIMEOUT`,
  `RATE_LIMIT_UNAVAILABLE`, and for a key with no cluster in scope.
- `get nodes --capacity-type` accepts `reserved` and `karpenter`.
- `KUBEADAPT_NO_UPDATE_CHECK` disables the update check. It is also skipped
  when `CI` is set or stdout is not a terminal, so CI jobs no longer pay a
  GitHub API round trip per invocation.

### Changed

- Default API URL is now `https://public-api.kubeadapt.io`. The previously
  documented `https://api.kubeadapt.io` is not a live API host.
  **Migration:** a config file with an explicit `api_url: https://api.kubeadapt.io`
  is still read as written and will keep failing. Update it, or delete the line
  to pick up the new default.
- `--quiet` and `--no-color` are now applied. They are resolved before the
  command runs, so `login`, `version`, and `completion` honor them too.
  `--quiet` additionally suppresses the update-available notice.
- Config is written atomically: a sibling temp file, `fsync`, then `rename`. A
  crash or a full disk can no longer truncate the file holding the only copy of
  your API key.
- `auth login` prompt moved from stdout to stderr, so redirecting stdout
  captures only command output.
- `--verbose` request logging now resolves the server `request_id` from the
  response envelope rather than only from the `X-Request-ID` header.
- `get pods` tables start with a `Cluster ID` column.
- An error response that is not from the API, such as a proxy's HTML 404, is
  reported as `unexpected response from <host> (HTTP 404, text/html)` instead
  of printing the body.
- Response bodies are capped at 32 MiB, and error bodies at 64 KiB.
- Exit codes follow a documented contract: `0` success, `2` for a usage or
  pre-flight error (unknown command or flag, invalid flag value, wrong number
  of arguments, missing required flag), `1` for everything else. Some of these
  exited `1` before. See the README.
- The upgrade notice is printed after an error message rather than before it.
- Em dashes and arrows in output are replaced with ASCII, which legacy Windows
  console code pages rendered as mojibake.
- `auth login --help` no longer suggests `--api-key <key>`; it shows the
  hidden prompt, a piped key and `KUBEADAPT_API_KEY`.
- `--config` help names the default path this machine resolves to.
- `get team-assignments --source` help describes `backfill_v1` as "imported
  from earlier team labels". The value itself is unchanged.

### Fixed

- `auth login` no longer erases a working API key when the new key is rejected.
  The key is verified before anything is written, and on any API error (a
  `401`, but also a `429` or a `5xx`) the config is left untouched. A transport
  failure is not treated as a rejection: the key is saved and verified on first
  use. A config file that exists but cannot be read or parsed is no longer
  overwritten with defaults.
- Empty result sets emit `"data": []` instead of `"data": null`, so
  `jq '.data[]'` and `jq '.data | length'` work without a guard.
- API error responses now render their `details`, including the `allowed` value
  list returned with a rejected enum.
- Rate-limit errors state the actual retry delay from `Retry-After` instead of
  "wait a moment".
- "Resource not found" hints name the resource that was actually missing.
  A missing node suggests `kubeadapt get nodes` rather than always
  `kubeadapt get clusters`.
- `--verbose` shows a real `request_id`. It was always blank.
- The `User-Agent` reports the real build version instead of a hardcoded
  `kubeadapt-cli/dev`.
- IDs and names are escaped when placed in a request path. A node group name
  containing `/`, `?`, `#`, `%` or a space was sent as a different URL and
  failed or read the wrong resource.
- `auth status` reports the credential the CLI would actually send. It read the
  config file directly, ignoring `KUBEADAPT_API_KEY` and `--api-key`. It also
  prints its report when the API client cannot be built.
- `auth logout` no longer reports "No stored credentials found." when the config
  file is unreadable or corrupt while the key is still on disk.
- `auth login` warns that a key could not be verified even under `--quiet`.
- `--quiet` no longer drops the next-page cursor from the table footer.
- `--paginate` reports a page that claims more results but carries no cursor as
  a partial run instead of a complete one.
- `get dashboard --top-clusters-limit` is validated before any credentials are
  needed.
- The `get --help` example used a `--status` value the API rejects.
- `--paginate` no longer repeats or skips rows on lists sorted by cost. A cost
  changes between requests, so the walk now pages in a stable order and the
  result is sorted by cost once all pages are in.

### Removed

- `scripts/install.sh`. It downloaded a binary name the releases do not
  publish, so it failed for everyone. Install with Homebrew or `go install`.
