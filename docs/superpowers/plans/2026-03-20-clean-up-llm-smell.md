# CLI Code Quality Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove copy-paste patterns, dead code, and unnecessary comments from kubeadapt-cli without changing any observable behavior.

**Architecture:** Four targeted refactors: (1) delete dead watch.go, (2) standardize flag helpers in cmd, (3) add `fetchWithSpinner` generic to collapse list-command boilerplate, (4) add `renderList` generic to collapse table.go list-renderer boilerplate. No new abstractions beyond what's needed. Tests pass before and after each task.

**Tech Stack:** Go 1.26, cobra, lipgloss/table, testify

---

## File Map

| File | Change |
|---|---|
| `cmd/watch.go` | **Delete** — dead code, watch feature never implemented |
| `cmd/helpers.go` | Remove dead `addPaginationFlags`; add `addLimitOffsetFlags`, `addNamespaceFlag`; add `fetchWithSpinner[T]`; remove restatement comments |
| `cmd/root.go` | Remove 3 restatement comments |
| `cmd/get_workloads.go` | Use flag helpers; use `fetchWithSpinner` |
| `cmd/get_nodes.go` | Use flag helpers; use `fetchWithSpinner` |
| `cmd/get_namespaces.go` | Use flag helpers; use `fetchWithSpinner` |
| `cmd/get_recommendations.go` | Use flag helpers; use `fetchWithSpinner` |
| `cmd/get_pvs.go` | Use flag helpers; use `fetchWithSpinner` |
| `cmd/get_node_groups.go` | Use flag helpers; use `fetchWithSpinner` |
| `cmd/get_costs_teams.go` | Use flag helpers; use `fetchWithSpinner` |
| `cmd/get_costs_departments.go` | Use flag helpers; use `fetchWithSpinner` |
| `cmd/get_clusters.go` | Use `fetchWithSpinner` |
| `cmd/get_cluster.go` | Use `fetchWithSpinner` |
| `cmd/get_overview.go` | Use `fetchWithSpinner` |
| `cmd/get_dashboard.go` | Use `fetchWithSpinner` |
| `internal/output/table.go` | Add `renderList[T]`; simplify all list `RenderX` functions |

---

## Task 1: Delete dead code

**Files:**
- Delete: `cmd/watch.go`
- Modify: `cmd/helpers.go` (remove `addPaginationFlags`)

- [ ] **Step 1: Delete watch.go**

  ```bash
  cd /path/to/kubeadapt-cli && rm cmd/watch.go
  ```

- [ ] **Step 2: Remove addPaginationFlags from helpers.go**

  In `cmd/helpers.go`, delete lines 48–52 (the `addPaginationFlags` function and its comment). The function is defined but never called anywhere.

  Final `helpers.go` after removal — the `addPaginationFlags` block is gone:
  ```go
  // addClusterIDFlag adds the --cluster-id flag to the given command.
  func addClusterIDFlag(cmd *cobra.Command) {
  	cmd.Flags().String("cluster-id", "", "Filter by cluster ID")
  }

  // addTimeframeFlag adds the --timeframe flag to the given command.
  func addTimeframeFlag(cmd *cobra.Command) {
  	cmd.Flags().String("timeframe", "", "Time range (e.g. 24h, 7d, 30d)")
  }
  ```

- [ ] **Step 3: Verify build and tests pass**

  ```bash
  cd /path/to/kubeadapt-cli && go build ./... && go test ./...
  ```

  Expected: all tests pass, no compile errors.

- [ ] **Step 4: Commit**

  ```bash
  git add -A && git commit -m "chore: remove dead watch.go and unused addPaginationFlags"
  ```

---

## Task 2: Standardize flag helpers

**Files:**
- Modify: `cmd/helpers.go` — add `addLimitOffsetFlags`, `addNamespaceFlag`
- Modify: `cmd/get_workloads.go`, `cmd/get_nodes.go`, `cmd/get_namespaces.go`, `cmd/get_recommendations.go`, `cmd/get_pvs.go`, `cmd/get_node_groups.go`

Currently some `init()` functions inline flag registration while others use helpers. Standardize everything through helpers.

- [ ] **Step 1: Add missing flag helpers to helpers.go**

  Add after `addClusterIDFlag`:
  ```go
  func addLimitOffsetFlags(cmd *cobra.Command) {
  	cmd.Flags().Int("limit", 0, "Maximum number of results")
  	cmd.Flags().Int("offset", 0, "Number of results to skip")
  }

  func addNamespaceFlag(cmd *cobra.Command) {
  	cmd.Flags().String("namespace", "", "Filter by namespace")
  }
  ```

- [ ] **Step 2: Update get_workloads.go init()**

  Replace the inlined flag registration:
  ```go
  func init() {
  	addClusterIDFlag(getWorkloadsCmd)
  	getWorkloadsCmd.Flags().String("namespace", "", "Filter by namespace")
  	getWorkloadsCmd.Flags().String("kind", "", "Filter by workload kind")
  	addLimitOffsetFlags(getWorkloadsCmd)
  	getCmd.AddCommand(getWorkloadsCmd)
  }
  ```

- [ ] **Step 3: Update get_nodes.go init()**

  ```go
  func init() {
  	addClusterIDFlag(getNodesCmd)
  	getNodesCmd.Flags().String("node-group", "", "Filter by node group")
  	addLimitOffsetFlags(getNodesCmd)
  	getCmd.AddCommand(getNodesCmd)
  }
  ```

- [ ] **Step 4: Update get_namespaces.go init()**

  ```go
  func init() {
  	addClusterIDFlag(getNamespacesCmd)
  	getNamespacesCmd.Flags().String("team", "", "Filter by team")
  	getNamespacesCmd.Flags().String("department", "", "Filter by department")
  	getCmd.AddCommand(getNamespacesCmd)
  }
  ```

- [ ] **Step 5: Update get_recommendations.go init()**

  ```go
  func init() {
  	addClusterIDFlag(getRecommendationsCmd)
  	getRecommendationsCmd.Flags().String("type", "", "Filter by recommendation type")
  	getRecommendationsCmd.Flags().String("status", "", "Filter by status")
  	addLimitOffsetFlags(getRecommendationsCmd)
  	getCmd.AddCommand(getRecommendationsCmd)
  }
  ```

- [ ] **Step 6: Update get_pvs.go init()**

  ```go
  func init() {
  	addClusterIDFlag(getPVsCmd)
  	addNamespaceFlag(getPVsCmd)
  	getPVsCmd.Flags().String("storage-class", "", "Filter by storage class")
  	getCmd.AddCommand(getPVsCmd)
  }
  ```

- [ ] **Step 7: Update get_node_groups.go init()**

  `get_node_groups.go` only has `cluster-id` (no limit/offset). Expected result:
  ```go
  func init() {
  	addClusterIDFlag(getNodeGroupsCmd)
  	getCmd.AddCommand(getNodeGroupsCmd)
  }
  ```

- [ ] **Step 8: Verify build and tests pass**

  ```bash
  go build ./... && go test ./...
  ```

  Expected: all tests pass. Flag behavior is identical — only the registration call site changes.

- [ ] **Step 9: Commit**

  ```bash
  git add cmd/helpers.go cmd/get_workloads.go cmd/get_nodes.go cmd/get_namespaces.go \
    cmd/get_recommendations.go cmd/get_pvs.go cmd/get_node_groups.go
  git commit -m "chore: standardize flag registration through helpers"
  ```

---

## Task 3: Add fetchWithSpinner and apply to list commands

**Files:**
- Modify: `cmd/helpers.go` — add `fetchWithSpinner[T any]`
- Modify: all list commands that currently have the 4-line spinner block

The 4-line spinner pattern:
```go
sp := newSpinner("Fetching X...")
sp.start()
resp, err := client.GetX(cmd.Context(), ...)
sp.stop()
```
...becomes 1 line inside a closure.

- [ ] **Step 1: Add fetchWithSpinner to helpers.go**

  Add at the top of helpers.go, before the flag helpers:
  ```go
  // fetchWithSpinner shows a spinner while fn executes, then hides it.
  func fetchWithSpinner[T any](msg string, fn func() (T, error)) (T, error) {
  	sp := newSpinner(msg)
  	sp.start()
  	result, err := fn()
  	sp.stop()
  	return result, err
  }
  ```

- [ ] **Step 2: Update get_clusters.go**

  Replace RunE body with:
  ```go
  RunE: func(cmd *cobra.Command, args []string) error {
  	client, err := newAPIClientFromCmd(cmd)
  	if err != nil {
  		return err
  	}
  	resp, err := fetchWithSpinner("Fetching clusters...", func() (*types.ClusterListResponse, error) {
  		return client.GetClusters(cmd.Context())
  	})
  	if err != nil {
  		return err
  	}
  	return renderOutputFromCmd(cmd, resp, func() {
  		output.RenderClusters(resp.Clusters, resp.Total, isNoColor(cmd))
  	})
  },
  ```

  > Note: The explicit return type in the closure is valid Go but requires importing `internal/api/types`, which the cmd files do not currently do. If you omit the explicit return type annotation entirely, Go infers `T` from the `client.GetX(...)` call and no extra import is needed. Either approach compiles; prefer the no-annotation form to keep the cmd files free of the types import.

- [ ] **Step 3: Update get_cluster.go, get_overview.go, get_dashboard.go the same way**

  Same pattern: `fetchWithSpinner("Fetching ...", func() (ReturnType, error) { return client.GetX(...) })`

- [ ] **Step 4: Update get_workloads.go, get_nodes.go, get_namespaces.go, get_recommendations.go, get_pvs.go**

  For commands with filter flags, the flags are read before the fetch call:
  ```go
  RunE: func(cmd *cobra.Command, args []string) error {
  	client, err := newAPIClientFromCmd(cmd)
  	if err != nil {
  		return err
  	}
  	clusterID, _ := cmd.Flags().GetString("cluster-id")
  	namespace, _ := cmd.Flags().GetString("namespace")
  	kind, _ := cmd.Flags().GetString("kind")
  	limit, _ := cmd.Flags().GetInt("limit")
  	offset, _ := cmd.Flags().GetInt("offset")
  	resp, err := fetchWithSpinner("Fetching workloads...", func() (*types.WorkloadListResponse, error) {
  		return client.GetWorkloads(cmd.Context(), clusterID, namespace, kind, limit, offset)
  	})
  	if err != nil {
  		return err
  	}
  	return renderOutputFromCmd(cmd, resp, func() {
  		output.RenderWorkloads(resp.Workloads, resp.Total, isNoColor(cmd))
  	})
  },
  ```

- [ ] **Step 5: Read and update get_node_groups.go, get_costs_teams.go, get_costs_departments.go the same way**

  Read each file, apply the same pattern.

- [ ] **Step 6: Verify build and tests pass**

  ```bash
  go build ./... && go test ./...
  ```

  Expected: all tests pass. The spinner behavior is unchanged — only the call site is cleaner.

- [ ] **Step 7: Commit**

  ```bash
  git add cmd/helpers.go cmd/get_clusters.go cmd/get_cluster.go cmd/get_overview.go \
    cmd/get_dashboard.go cmd/get_workloads.go cmd/get_nodes.go cmd/get_namespaces.go \
    cmd/get_recommendations.go cmd/get_pvs.go cmd/get_node_groups.go \
    cmd/get_costs_teams.go cmd/get_costs_departments.go
  git commit -m "refactor: extract fetchWithSpinner to remove 4-line spinner boilerplate"
  ```

---

## Task 4: Refactor output/table.go with renderList helper

**Files:**
- Modify: `internal/output/table.go` — add `renderList[T any]`; update all list renderers

The repeating pattern in every list renderer:
```go
if len(items) == 0 {
    fmt.Fprintln(os.Stdout, StyleMuted.Render("No X found."))
    return
}
rows := make([][]string, 0, len(items))
for _, item := range items {
    rows = append(rows, []string{ ... })
}
renderTable(headers, rows, noColor)
renderPaginationFooter(len(items), total)
```

- [ ] **Step 1: Add renderList generic helper to table.go**

  Add after `renderPaginationFooter`:
  ```go
  func renderList[T any](items []T, emptyMsg string, headers []string, rowFn func(T) []string, noColor bool, total int) {
  	if len(items) == 0 {
  		fmt.Fprintln(os.Stdout, StyleMuted.Render(emptyMsg))
  		return
  	}
  	rows := make([][]string, 0, len(items))
  	for _, item := range items {
  		rows = append(rows, rowFn(item))
  	}
  	renderTable(headers, rows, noColor)
  	renderPaginationFooter(len(items), total)
  }
  ```

- [ ] **Step 2: Rewrite RenderWorkloads using renderList**

  Before (18 lines):
  ```go
  func RenderWorkloads(workloads []types.WorkloadResponse, total int, noColor bool) {
  	if len(workloads) == 0 {
  		fmt.Fprintln(os.Stdout, StyleMuted.Render("No workloads found. Try adjusting your filters or check 'kubeadapt get clusters' first."))
  		return
  	}
  	headers := []string{"Name", "Kind", "Namespace", "Cluster", "Replicas", "Efficiency", "Monthly $", "$/hr"}
  	rows := make([][]string, 0, len(workloads))
  	for _, w := range workloads {
  		rows = append(rows, []string{
  			w.WorkloadName, w.WorkloadKind, w.Namespace, w.ClusterName,
  			fmt.Sprintf("%d/%d", w.AvailableReplicas, w.Replicas),
  			FormatPercentPtr(w.EfficiencyScore), FormatCostPtr(w.MonthlyCost), FormatCost(w.HourlyCost),
  		})
  	}
  	renderTable(headers, rows, noColor)
  	renderPaginationFooter(len(workloads), total)
  }
  ```

  After (10 lines):
  ```go
  func RenderWorkloads(workloads []types.WorkloadResponse, total int, noColor bool) {
  	renderList(workloads, "No workloads found. Try adjusting your filters or check 'kubeadapt get clusters' first.",
  		[]string{"Name", "Kind", "Namespace", "Cluster", "Replicas", "Efficiency", "Monthly $", "$/hr"},
  		func(w types.WorkloadResponse) []string {
  			return []string{
  				w.WorkloadName, w.WorkloadKind, w.Namespace, w.ClusterName,
  				fmt.Sprintf("%d/%d", w.AvailableReplicas, w.Replicas),
  				FormatPercentPtr(w.EfficiencyScore), FormatCostPtr(w.MonthlyCost), FormatCost(w.HourlyCost),
  			}
  		}, noColor, total)
  }
  ```

- [ ] **Step 3: Rewrite RenderNodes using renderList**

  `RenderNodes` has conditional coloring based on `n.IsReady`. The `rowFn` closure captures `noColor`, so this works cleanly:
  ```go
  func RenderNodes(nodes []types.NodeResponse, total int, noColor bool) {
  	renderList(nodes, "No nodes found.",
  		[]string{"Name", "Cluster", "Instance", "Ready", "CPU", "Memory", "Pods", "Spot", "$/hr"},
  		func(n types.NodeResponse) []string {
  			ready := FormatBool(n.IsReady)
  			if !noColor {
  				if n.IsReady {
  					ready = StyleSuccess.Render("Yes")
  				} else {
  					ready = StyleError.Render("No")
  				}
  			}
  			return []string{
  				n.NodeName, n.ClusterName, FormatOptionalString(n.InstanceType), ready,
  				FormatFloat(n.CPUAllocatable, 1), FormatMemoryGB(n.MemoryAllocatableGB),
  				FormatIntPtr(n.PodCount), FormatBool(n.SpotInstance), FormatCost(n.HourlyCost),
  			}
  		}, noColor, total)
  }
  ```

- [ ] **Step 4: Rewrite RenderRecommendations using renderList**

  Has conditional coloring for status and priority — same closure approach:
  ```go
  func RenderRecommendations(recs []types.RecommendationResponse, total int, noColor bool) {
  	renderList(recs, "No recommendations found. Your resources may already be optimized!",
  		[]string{"ID", "Type", "Cluster", "Resource", "Priority", "Status", "Monthly Savings"},
  		func(r types.RecommendationResponse) []string {
  			resource := FormatOptionalString(r.ResourceName)
  			if ns := FormatOptionalString(r.Namespace); ns != "-" && ns != "" {
  				resource = ns + "/" + resource
  			}
  			status := r.Status
  			priority := FormatOptionalString(r.Priority)
  			if !noColor {
  				switch r.Status {
  				case "applied":
  					status = StyleSuccess.Render(r.Status)
  				case "dismissed":
  					status = StyleMuted.Render(r.Status)
  				case "open":
  					status = StyleWarning.Render(r.Status)
  				}
  				if r.Priority != nil {
  					switch *r.Priority {
  					case "critical", "high":
  						priority = StyleError.Render(*r.Priority)
  					case "medium":
  						priority = StyleWarning.Render(*r.Priority)
  					case "low":
  						priority = StyleSuccess.Render(*r.Priority)
  					}
  				}
  			}
  			return []string{ShortID(r.ID), r.RecommendationType, r.ClusterName, resource, priority, status, FormatCost(r.EstimatedMonthlySavings)}
  		}, noColor, total)
  }
  ```

- [ ] **Step 5: Rewrite RenderTeamCosts and RenderDepartmentCosts using renderList**

  These two are almost identical. After refactoring, each is ~10 lines (was ~22):
  ```go
  func RenderTeamCosts(costs []types.TeamCostResponse, total int, noColor bool) {
  	renderList(costs, "No team cost data available.",
  		[]string{"Team", "Namespaces", "Workloads", "Pods", "CPU", "Memory", "$/hr", "$/mo"},
  		func(c types.TeamCostResponse) []string {
  			return []string{
  				c.Team, FormatInt(c.NamespaceCount), FormatInt(c.WorkloadCount), FormatInt(c.PodCount),
  				FormatFloat(c.TotalCPUCores, 1), FormatMemoryGB(c.TotalMemoryGB),
  				FormatCost(c.HourlyCost), FormatCost(c.MonthlyCost),
  			}
  		}, noColor, total)
  }

  func RenderDepartmentCosts(costs []types.DepartmentCostResponse, total int, noColor bool) {
  	renderList(costs, "No department cost data available.",
  		[]string{"Department", "Namespaces", "Workloads", "Pods", "CPU", "Memory", "$/hr", "$/mo"},
  		func(c types.DepartmentCostResponse) []string {
  			return []string{
  				c.Department, FormatInt(c.NamespaceCount), FormatInt(c.WorkloadCount), FormatInt(c.PodCount),
  				FormatFloat(c.TotalCPUCores, 1), FormatMemoryGB(c.TotalMemoryGB),
  				FormatCost(c.HourlyCost), FormatCost(c.MonthlyCost),
  			}
  		}, noColor, total)
  }
  ```

- [ ] **Step 6: Rewrite RenderNodeGroups, RenderNamespaces, RenderPersistentVolumes using renderList**

  Same pattern for each. Read each function, apply `renderList` shape.

- [ ] **Step 7: Skip time-series renderers**

  `RenderCostDistribution`, `RenderNodeMetrics`, `RenderWorkloadMetrics`, `RenderWorkloadNodes`, `RenderNamespaceTrends` have no empty-state message and no pagination footer. Forcing `renderList` onto them would require fake empty strings and `total=len(items)` hacks that make the code worse. Leave these functions as-is.

- [ ] **Step 8: Leave RenderClusters as-is**

  `RenderClusters` outputs: table → "Potential Savings..." note → pagination footer. This ordering cannot be preserved with `renderList` (which outputs table → pagination footer internally). Do not apply `renderList` to `RenderClusters` — the function stays exactly as it is. Same for `RenderOverview` and `RenderDashboard` which also have trailing notes and are not list renderers.

- [ ] **Step 9: Run tests**

  ```bash
  go test ./internal/output/...
  ```

  Expected: all 8 table tests pass, output is byte-for-byte identical.

- [ ] **Step 10: Run full test suite**

  ```bash
  go build ./... && go test ./...
  ```

- [ ] **Step 11: Commit**

  ```bash
  git add internal/output/table.go
  git commit -m "refactor: add renderList generic to remove list-rendering boilerplate in table.go"
  ```

---

## Task 5: Remove restatement comments

**Files:**
- Modify: `cmd/root.go`
- Modify: `cmd/helpers.go`

- [ ] **Step 1: Remove restatement comments from root.go**

  Remove these three comments (lines 47, 59, 67–68):
  - `// Skip config loading for commands that don't need it` — the condition is self-evident
  - `// Environment variables override config values` — the variable names say this
  - `// CLI flags override everything` — same

  Keep:
  - `// Flag variables — still needed for cobra binding...` — this explains a non-obvious design decision
  - `// FlagError → show usage + exit 2` — this is meaningful
  - `// All other errors → friendly message + exit 1` — meaningful

- [ ] **Step 2: Remove restatement comments from helpers.go**

  Remove:
  - `// newAPIClientFromCmd creates an API client using the RunContext from the command.` — restates the function name
  - `// renderOutputFromCmd renders data in the format specified by the RunContext.` — same
  - `// addClusterIDFlag adds the --cluster-id flag to the given command.` — same
  - `// addTimeframeFlag adds the --timeframe flag to the given command.` — same
  - `// isNoColor returns whether color is disabled for this command.` — same
  - `// --- Flag helpers for DRY registration ---` — section header noise

  Keep:
  - `// fetchWithSpinner shows a spinner while fn executes, then hides it.` — new function, keep for clarity
  - `// addLimitOffsetFlags...` — remove, same as others

- [ ] **Step 3: Verify build**

  ```bash
  go build ./...
  ```

- [ ] **Step 4: Commit**

  ```bash
  git add cmd/root.go cmd/helpers.go
  git commit -m "chore: remove comments that restate the code"
  ```

---

## Verification

After all tasks, run:

```bash
go build ./... && go test ./... -count=1
```

Expected output:
```
ok  github.com/kubeadapt/kubeadapt-cli/cmd
ok  github.com/kubeadapt/kubeadapt-cli/internal/api
ok  github.com/kubeadapt/kubeadapt-cli/internal/config
ok  github.com/kubeadapt/kubeadapt-cli/internal/output
```

Smoke test the CLI binary:
```bash
go run . --help
go run . get --help
go run . get clusters --help
go run . get workloads --help
```

Expected: all flags present, help text identical to pre-refactor.
