package cmd

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/spf13/cobra"
)

// Exit codes are a permanent public contract. 3-125 are deliberately left
// unassigned so a richer taxonomy stays possible without breaking callers.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// usageError marks a failure the user fixes by retyping the command. It is
// reserved for pre-flight rejections: if a request reached the API, a 400 or
// 422 is still an ordinary failure, because the invocation itself was well-formed.
type usageError struct{ err error }

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return e.err }

func usagef(format string, a ...any) error {
	return &usageError{err: fmt.Errorf(format, a...)}
}

func asUsageError(err error) error {
	if err == nil || isUsageError(err) {
		return err
	}
	return &usageError{err: err}
}

func isUsageError(err error) bool {
	_, ok := errors.AsType[*usageError](err)
	return ok
}

// Cobra reports an unknown command (from Find) and a missing required flag as
// plain fmt.Errorf values with no exported sentinel and no interception hook,
// so those two are matched on shape; TestExitCode_* pin the wording.
func isCobraUsageError(err error) bool {
	msg := err.Error()
	return strings.HasPrefix(msg, "unknown command ") ||
		strings.HasPrefix(msg, "required flag(s) ")
}

func exitCodeFor(err error) int {
	switch {
	case err == nil:
		return exitOK
	case isUsageError(err), isCobraUsageError(err):
		return exitUsage
	default:
		return exitError
	}
}

var markUsageErrorsOnce sync.Once

// Positional-argument validators are wrapped once across the whole tree rather
// than at each of the ~27 definition sites, so cobra.NoArgs and cobra.ExactArgs
// failures carry the usage class without every command having to remember to.
func markUsageErrors(root *cobra.Command) {
	markUsageErrorsOnce.Do(func() { wrapArgsValidator(root) })
}

func wrapArgsValidator(c *cobra.Command) {
	for _, sub := range c.Commands() {
		wrapArgsValidator(sub)
	}
	inner := c.Args
	if inner == nil {
		return
	}
	c.Args = func(cmd *cobra.Command, args []string) error {
		return asUsageError(inner(cmd, args))
	}
}

// notFoundHints maps each *_NOT_FOUND code to the subcommand that lists that
// resource. A generic "get clusters" hint is actively misleading when the
// missing thing was a workload or a node.
var notFoundHints = map[api.ErrorCode]string{
	api.CodeClusterNotFound:        "clusters",
	api.CodeNamespaceNotFound:      "namespaces",
	api.CodeWorkloadNotFound:       "workloads",
	api.CodeNodeNotFound:           "nodes",
	api.CodeNodeGroupNotFound:      "node-groups",
	api.CodeTeamNotFound:           "teams",
	api.CodeDepartmentNotFound:     "departments",
	api.CodeRecommendationNotFound: "recommendations",
}

func friendlyError(err error) string {
	if err == nil {
		return ""
	}

	if apiErr, ok := errors.AsType[*api.APIError](err); ok {
		return apiErrorMessage(apiErr)
	}

	if dnsErr, ok := errors.AsType[*net.DNSError](err); ok {
		return fmt.Sprintf("Cannot resolve %s: check your network connection and --api-url flag.", dnsErr.Name)
	}

	if opErr, ok := errors.AsType[*net.OpError](err); ok {
		if opErr.Op == "dial" {
			return "Cannot connect to Kubeadapt API: connection refused.\n  Check your network connection or use --api-url to set a different endpoint."
		}
		return fmt.Sprintf("Network error: %s\n  Check your network connection.", opErr.Err)
	}

	if isTimeout(err) {
		return "Request timed out: the API did not respond in time.\n  Try again or use --verbose to see request details."
	}

	msg := err.Error()
	if strings.Contains(msg, "not authenticated") || strings.Contains(msg, "no API key") {
		return msg // already user-friendly
	}

	return msg
}

func apiErrorMessage(e *api.APIError) string {
	var msg string
	switch {
	case e.IsAuthError():
		msg = "Authentication failed: API key is invalid or expired.\n  Run 'kubeadapt auth login' to re-authenticate."
	case e.IsForbidden():
		msg = "Access denied: you don't have permission for this resource.\n  Check your API key permissions at https://app.kubeadapt.io/settings/api-keys"
	case e.IsNotFound():
		msg = fmt.Sprintf("Resource not found: %s\n  Use 'kubeadapt get %s' to list available resources.",
			e.Message, notFoundHint(e.Code))
	case e.IsRateLimited():
		msg = rateLimitMessage(e.RetryAfter)
	case e.IsServerError():
		msg = "Kubeadapt API error: the service is experiencing issues.\n  Check https://status.kubeadapt.io or try again later."
	default:
		msg = fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return msg + formatDetails(e.Details)
}

func notFoundHint(code api.ErrorCode) string {
	if resource, ok := notFoundHints[code]; ok {
		return resource
	}
	return "clusters"
}

func rateLimitMessage(retryAfter time.Duration) string {
	if retryAfter <= 0 {
		return "Rate limited: too many requests.\n  Wait a moment and try again."
	}
	return fmt.Sprintf("Rate limited: too many requests.\n  Retry after %s.", humanizeDuration(retryAfter))
}

func humanizeDuration(d time.Duration) string {
	if d < time.Minute {
		return pluralize(int(math.Ceil(d.Seconds())), "second")
	}
	return pluralize(int(math.Ceil(d.Minutes())), "minute")
}

func pluralize(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

// The API inlines free-form keys per error code (VALIDATION_ERROR sends
// field/allowed, FORBIDDEN sends granted/required), so unknown keys are printed
// verbatim rather than dropped.
func formatDetails(details []map[string]any) string {
	var b strings.Builder
	for _, d := range details {
		if line := formatDetail(d); line != "" {
			b.WriteString("\n  " + line)
		}
	}
	return b.String()
}

func formatDetail(d map[string]any) string {
	rest := make(map[string]any, len(d))
	var prefix string
	for k, v := range d {
		if s, ok := v.(string); ok && k == "field" && s != "" {
			prefix = s + ": "
			continue
		}
		rest[k] = v
	}

	pairs := joinPairs(rest)
	if pairs == "" {
		return strings.TrimSuffix(prefix, ": ")
	}
	return prefix + pairs
}

func joinPairs(m map[string]any) string {
	parts := make([]string, 0, len(m))
	for _, k := range slices.Sorted(maps.Keys(m)) {
		parts = append(parts, k+"="+detailValue(m[k]))
	}
	return strings.Join(parts, ", ")
}

func detailValue(v any) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		return t
	case []any:
		items := make([]string, len(t))
		for i, item := range t {
			items[i] = detailValue(item)
		}
		return "[" + strings.Join(items, ", ") + "]"
	case map[string]any:
		return "{" + joinPairs(t) + "}"
	default:
		return fmt.Sprintf("%v", t)
	}
}

func isTimeout(err error) bool {
	type timeouter interface {
		Timeout() bool
	}
	var t timeouter
	return errors.As(err, &t) && t.Timeout()
}
