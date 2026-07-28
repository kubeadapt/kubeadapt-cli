package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	flagMaxWait       = "max-wait"
	defaultMaxWait    = 15 * time.Minute
	singlePageTimeout = 60 * time.Second

	// Stop this many requests short of the window. The server is fail-closed
	// with a flat Retry-After: 60, so the reserve costs seconds while skipping
	// it costs a minute.
	rateLimitFloor = 5

	// Retries are the safety net behind pacing, not the mechanism. Two bounds
	// what a retry can add on top of the pacing budget.
	paginateRetries = 2

	// Below the default page size, --paginate needs enough requests that the
	// rate limit becomes the dominant cost, so say so once.
	smallLimitHint = 100

	// Room for the in-flight request that follows the final permitted wait.
	waitSlack = 2 * time.Minute
)

var errMaxWaitExceeded = errors.New("--max-wait budget exhausted")

var errMissingCursor = errors.New(
	"the API reported more results but returned no cursor to fetch them; results may be incomplete")

// paginateSleep is the single wait used by both the pacer and the client's
// Retry-After retry, so one budget covers every kind of rate-limit wait and
// tests can make all of them observable through one seam.
var paginateSleep = api.SleepContext

// PageFetcher retrieves one page starting at cursor. Each `get` list
// subcommand supplies one closing over its own filter struct.
type PageFetcher[T any] func(ctx context.Context, cursor string) ([]T, *types.Meta, error)

// pageRun is the outcome of a pagination run: everything successfully
// collected, plus whether the run was cut short and where to resume from.
type pageRun[T any] struct {
	Items        []T
	Meta         *types.Meta
	Partial      bool
	ResumeCursor string
	Err          error
}

// partialDoc is the emitted shape when a run is truncated. The partial flag is
// the whole point: a consumer piping cost data into a budget calculation has to
// be able to tell a short answer from a complete one.
type partialDoc struct {
	Data         any         `json:"data"`
	Meta         *types.Meta `json:"meta"`
	Partial      bool        `json:"partial"`
	ResumeCursor string      `json:"resume_cursor,omitempty"`
}

type paginator struct {
	paged     PagedFlags
	stderr    io.Writer
	now       func() time.Time
	sleep     func(context.Context, time.Duration) error
	rateLimit func() api.RateLimit
	spent     time.Duration
	announced bool
	warned    bool

	ignoredWarned bool
}

// A run cut short still renders what it collected before returning the error.
// client must be the same one fetch issues through: the pacer steers on that
// client's rate-limit snapshot, and a second client's is never populated.
func runPaginatedList[T any](
	cmd *cobra.Command,
	client *api.Client,
	paged PagedFlags,
	fetch PageFetcher[T],
	renderTable func(io.Writer, []T, *types.Meta) error,
) error {
	ctx, cancel := paginateContext(cmd, paged)
	defer cancel()

	p := &paginator{
		paged:     paged,
		stderr:    cmd.ErrOrStderr(),
		now:       time.Now,
		sleep:     paginateSleep,
		rateLimit: client.RateLimit,
	}
	return renderPaginated(cmd, collectPages(ctx, p, fetch), renderTable)
}

// paginateContext bounds the whole run. A single page keeps the original 60s
// ceiling; a paginated run needs room for its entire wait budget, which is
// minutes, so reusing 60s there would abort every long run.
func paginateContext(cmd *cobra.Command, paged PagedFlags) (context.Context, context.CancelFunc) {
	parent := cmd.Context()
	if parent == nil {
		parent = context.Background()
	}
	switch {
	case !paged.Paginate:
		return context.WithTimeout(parent, singlePageTimeout)
	case paged.MaxWait <= 0:
		return context.WithCancel(parent)
	default:
		return context.WithTimeout(parent, paged.MaxWait+waitSlack)
	}
}

func collectPages[T any](ctx context.Context, p *paginator, fetch PageFetcher[T]) pageRun[T] {
	var run pageRun[T]
	cursor := p.paged.Cursor
	pages := 0

	for {
		items, meta, err := fetch(ctx, cursor)
		if err != nil {
			run.Partial = pages > 0
			run.ResumeCursor = cursor
			run.Err = err
			return run
		}
		pages++
		run.Items = append(run.Items, items...)
		run.Meta = meta
		p.warnIgnoredParams(meta)
		p.warnSmallLimit(meta)

		if !p.paged.Paginate || meta == nil || meta.Pagination == nil || !meta.Pagination.HasMore {
			return run
		}
		next := meta.Pagination.NextCursor
		if next == "" {
			// HasMore is authoritative, so this page claims more results and
			// gives no way to reach them. Reporting a complete run here would
			// pass a silently truncated answer off as the whole set.
			run.Partial = true
			run.ResumeCursor = cursor
			run.Err = errMissingCursor
			return run
		}
		if err := p.pace(ctx, pages, len(run.Items), meta); err != nil {
			run.Partial = true
			run.ResumeCursor = next
			run.Err = err
			return run
		}
		cursor = next
	}
}

// pace waits out the rate-limit window before the next request when the budget
// is nearly spent. The budget check precedes the wait so an impossible wait is
// reported instantly instead of after sleeping through it.
func (p *paginator) pace(ctx context.Context, pages, itemsSoFar int, meta *types.Meta) error {
	rl := p.rateLimit()
	d := rl.PaceUntil(p.now(), rateLimitFloor)
	if d <= 0 {
		return nil
	}
	if p.paged.MaxWait > 0 && p.spent+d > p.paged.MaxWait {
		return fmt.Errorf("%w: pausing %s more would exceed --%s %s (%s already spent)",
			errMaxWaitExceeded, d.Round(time.Second), flagMaxWait,
			p.paged.MaxWait, p.spent.Round(time.Second))
	}
	p.announce(d, pages, itemsSoFar, meta, rl)
	if err := p.sleep(ctx, d); err != nil {
		return err
	}
	p.spent += d
	return nil
}

func (p *paginator) announce(d time.Duration, pages, itemsSoFar int, meta *types.Meta, rl api.RateLimit) {
	if p.announced {
		return
	}
	p.announced = true
	budget := "unbounded"
	if p.paged.MaxWait > 0 {
		budget = p.paged.MaxWait.String()
	}
	fmt.Fprintf(p.stderr,
		"kubeadapt: rate limit reached after %d pages; pausing %s for the window to reset.%s --%s is %s. Press Ctrl-C to stop and keep what has been fetched.\n",
		pages, d.Round(time.Second), p.eta(meta, itemsSoFar, d, rl), flagMaxWait, budget)
}

func (p *paginator) eta(meta *types.Meta, itemsSoFar int, pause time.Duration, rl api.RateLimit) string {
	reqs, ok := remainingRequests(meta, itemsSoFar, p.paged.Limit)
	perWindow := rl.Limit - rateLimitFloor
	if !ok || perWindow <= 0 {
		return ""
	}
	total := time.Duration(ceilDiv(reqs, perWindow)) * pause
	return fmt.Sprintf(" Estimated %s of further waiting.", total.Round(time.Second))
}

// Reports filters the server took with 200 but never applied, so a typo like
// --namespaces silently widening results is not read as a real number. Not gated
// on --quiet, and emitted once per run since every page repeats the list.
func (p *paginator) warnIgnoredParams(meta *types.Meta) {
	if p.ignoredWarned || meta == nil || len(meta.IgnoredParams) == 0 {
		return
	}
	p.ignoredWarned = true
	noun, pronoun := "parameter", "it"
	if len(meta.IgnoredParams) > 1 {
		noun, pronoun = "parameters", "them"
	}
	fmt.Fprintf(p.stderr,
		"kubeadapt: warning: the API ignored unknown %s %s - results are NOT filtered by %s.\n",
		noun, quoteJoin(meta.IgnoredParams), pronoun)
}

func quoteJoin(vals []string) string {
	quoted := make([]string, 0, len(vals))
	for _, v := range vals {
		quoted = append(quoted, strconv.Quote(v))
	}
	return strings.Join(quoted, ", ")
}

func (p *paginator) warnSmallLimit(meta *types.Meta) {
	if p.warned || !p.paged.Paginate || p.paged.Limit >= smallLimitHint {
		return
	}
	p.warned = true
	if reqs, ok := remainingRequests(meta, 0, p.paged.Limit); ok {
		fmt.Fprintf(p.stderr,
			"kubeadapt: --%s with --%s %d needs ~%d requests; --%s 500 would need ~%d and avoid most rate-limit pauses.\n",
			flagPaginate, flagLimit, p.paged.Limit, reqs, flagLimit, ceilDiv(reqs*p.paged.Limit, 500))
		return
	}
	fmt.Fprintf(p.stderr,
		"kubeadapt: --%s with --%s %d issues one request per %d items; a larger --%s (up to 500) cuts the request count and avoids most rate-limit pauses.\n",
		flagPaginate, flagLimit, p.paged.Limit, p.paged.Limit, flagLimit)
}

func remainingRequests(meta *types.Meta, itemsSoFar, pageSize int) (int, bool) {
	if meta == nil || meta.Pagination == nil || meta.Pagination.TotalCount == nil || pageSize <= 0 {
		return 0, false
	}
	remaining := *meta.Pagination.TotalCount - itemsSoFar
	if remaining <= 0 {
		return 0, false
	}
	return ceilDiv(remaining, pageSize), true
}

func ceilDiv(a, b int) int {
	if b <= 0 {
		return 0
	}
	return (a + b - 1) / b
}

func renderPaginated[T any](
	cmd *cobra.Command,
	run pageRun[T],
	renderTable func(io.Writer, []T, *types.Meta) error,
) error {
	if run.Err != nil && !run.Partial {
		return run.Err
	}

	outFmt := formatTable
	if rc := getRunContext(cmd); rc != nil && rc.OutputFmt != "" {
		outFmt = rc.OutputFmt
	}
	w := cmd.OutOrStdout()

	var renderErr error
	switch {
	case run.Partial && outFmt == formatJSON:
		renderErr = output.RenderJSON(w, partialDocFor(run))
	case run.Partial && outFmt == formatYAML:
		renderErr = output.RenderYAML(w, partialDocFor(run))
	case outFmt == formatJSON:
		renderErr = output.RenderJSONWithMeta(w, run.Items, run.Meta)
	case outFmt == formatYAML:
		renderErr = output.RenderYAMLWithMeta(w, run.Items, run.Meta)
	default:
		if renderErr = renderTable(w, run.Items, run.Meta); renderErr == nil && run.Partial {
			renderErr = writePartialNotice(w, run)
		}
	}
	if renderErr != nil {
		return renderErr
	}
	if run.Err == nil {
		return nil
	}

	reportPartial(cmd.ErrOrStderr(), run)
	return fmt.Errorf("pagination stopped after %d items: %w", len(run.Items), run.Err)
}

func partialDocFor[T any](run pageRun[T]) partialDoc {
	items := run.Items
	if items == nil {
		items = []T{}
	}
	return partialDoc{Data: items, Meta: run.Meta, Partial: true, ResumeCursor: run.ResumeCursor}
}

// The json/yaml documents carry `partial`, but a table had no stdout marker at
// all: stderr is routinely redirected away, and a truncated table is otherwise
// indistinguishable from a complete one. Not gated on --quiet.
func writePartialNotice[T any](w io.Writer, run pageRun[T]) error {
	_, err := fmt.Fprintf(w, "\nPARTIAL RESULTS: %s fetched before the run stopped; resume with --%s=%s\n",
		pluralize(len(run.Items), "item"), flagCursor, output.FormatCursor(run.ResumeCursor))
	if err != nil {
		return fmt.Errorf("writing partial notice: %w", err)
	}
	return nil
}

func reportPartial[T any](w io.Writer, run pageRun[T]) {
	fmt.Fprintf(w, "kubeadapt: partial results - %s fetched before the run stopped.\n",
		pluralize(len(run.Items), "item"))
	// friendlyError returns its own already-indented continuation lines; nest
	// them one level deeper so they read as part of "reason:".
	fmt.Fprintf(w, "  reason: %s\n", strings.ReplaceAll(friendlyError(run.Err), "\n", "\n  "))
	fmt.Fprintf(w, "  resume with: --%s=%s\n", flagCursor, output.FormatCursor(run.ResumeCursor))
	if api.IsCursorExpired(run.Err) {
		fmt.Fprintln(w, "  note: that cursor is no longer valid; re-run from the start for a consistent snapshot.")
	}
}

func init() {
	getCmd.PersistentFlags().Duration(flagMaxWait, defaultMaxWait,
		"Maximum total time --paginate may spend waiting out API rate limits before returning partial results (0 = unbounded)")
}
