// Package output renders Kubeadapt /v1 resource types as human-friendly
// tables, JSON, and YAML. Renderers always accept an io.Writer — no helper
// here writes to os.Stdout. cmd/* is responsible for choosing the writer.
package output

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
)

const noValue = "-"

// FormatMoney formats a value-type Money for human display. Returns "-" if
// the Money is the zero value. For known currencies (USD) it emits
// "$1234.5678"; for others it emits "<CCY> 1234.5678". Always four decimals.
func FormatMoney(m types.Money) string {
	return m.String()
}

// FormatMoneyPtr is like FormatMoney but accepts a pointer; nil → "-".
func FormatMoneyPtr(m *types.Money) string {
	if m == nil {
		return noValue
	}
	return m.String()
}

// FormatPercentage formats a value in [0, 100] as "73.5%". NaN, infinity,
// or negative values render as "-".
func FormatPercentage(p float64) string {
	if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 {
		return noValue
	}
	return fmt.Sprintf("%.1f%%", p)
}

// FormatBytes formats a byte count using binary IEC prefixes (KiB, MiB,
// GiB, TiB, PiB). Negative values render as "-". Zero renders as "0 B".
func FormatBytes(b int64) string {
	if b < 0 {
		return noValue
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffixes := []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	if exp >= len(suffixes) {
		exp = len(suffixes) - 1
	}
	return fmt.Sprintf("%.1f %s", float64(b)/float64(div), suffixes[exp])
}

// FormatCores formats a CPU-core count as "12.50". Negative, NaN, or
// infinite values render as "-".
func FormatCores(c float64) string {
	if math.IsNaN(c) || math.IsInf(c, 0) || c < 0 {
		return noValue
	}
	return strconv.FormatFloat(c, 'f', 2, 64)
}

// FormatCursor returns "(none)" for an empty cursor and the cursor verbatim
// otherwise. The full string is intentionally returned (not truncated) so that
// PaginationFooter can show a copy-pasteable value to the user.
func FormatCursor(c string) string {
	if c == "" {
		return "(none)"
	}
	return c
}

// PaginationFooter renders a multi-line cursor-aware footer beneath a list
// table. Returns "" when meta is nil or carries no Pagination block.
//
// When more pages exist, the full next_cursor is printed on its own line so
// users can copy-paste it into --cursor=, alongside a hint to use --paginate
// to auto-fetch every page.
func PaginationFooter(itemsShown int, meta *types.Meta) string {
	if meta == nil || meta.Pagination == nil {
		return ""
	}
	p := meta.Pagination

	var b strings.Builder
	b.WriteString("Showing ")
	b.WriteString(strconv.Itoa(itemsShown))
	if p.TotalCount != nil {
		b.WriteString(" of ")
		b.WriteString(strconv.Itoa(*p.TotalCount))
	}
	if p.Limit > 0 {
		b.WriteString(" (limit ")
		b.WriteString(strconv.Itoa(p.Limit))
		b.WriteString(")")
	}
	b.WriteString(".")

	if p.HasMore {
		b.WriteString(" More results available — use --paginate to auto-fetch, or copy this cursor:\n")
		b.WriteString("  --cursor=")
		b.WriteString(p.NextCursor)
	} else {
		b.WriteString(" End of results.")
	}
	return b.String()
}

func formatBool(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

func formatIntPlain(v int) string {
	return strconv.Itoa(v)
}

func formatStr(s string) string {
	if s == "" {
		return noValue
	}
	return s
}

func formatRefName(ref *types.NestedRef) string {
	if ref == nil {
		return noValue
	}
	return formatStr(ref.Name)
}
