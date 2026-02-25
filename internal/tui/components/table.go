package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kubeadapt/kubeadapt-cli/internal/tui"
)

// Table is a simple styled table component for TUI views.
type Table struct {
	Headers  []string
	Rows     [][]string
	RowIDs   []string // Optional IDs for each row (for drill-down navigation)
	Selected int
	Width    int

	ViewportHeight int        // Number of visible rows (0 = show all, for backward compat)
	ScrollOffset   int        // Index of first visible row
	FilterQuery    string     // Current filter text (empty = no filter)
	filteredRows   [][]string // Filtered subset of Rows (private)
	filteredRowIDs []string   // Filtered subset of RowIDs (private)
}

// NewTable creates a new Table.
func NewTable(headers []string, rows [][]string) *Table {
	return &Table{
		Headers:  headers,
		Rows:     rows,
		Selected: 0,
	}
}

func (t *Table) activeRows() ([][]string, []string) {
	if t.FilterQuery != "" && t.filteredRows != nil {
		return t.filteredRows, t.filteredRowIDs
	}
	return t.Rows, t.RowIDs
}

func (t *Table) TotalRows() int {
	rows, _ := t.activeRows()
	return len(rows)
}

func (t *Table) MoveUp() {
	if t.Selected > 0 {
		t.Selected--
	}
	if t.Selected < t.ScrollOffset {
		t.ScrollOffset = t.Selected
	}
}

func (t *Table) MoveDown() {
	rows, _ := t.activeRows()
	if t.Selected < len(rows)-1 {
		t.Selected++
	}
	if t.ViewportHeight > 0 && t.Selected >= t.ScrollOffset+t.ViewportHeight {
		t.ScrollOffset = t.Selected - t.ViewportHeight + 1
	}
}

func (t *Table) MovePageDown() {
	rows, _ := t.activeRows()
	pageSize := t.ViewportHeight
	if pageSize <= 0 {
		pageSize = 10
	}
	t.Selected += pageSize
	if t.Selected >= len(rows) {
		t.Selected = len(rows) - 1
	}
	if t.Selected < 0 {
		t.Selected = 0
	}
	if t.ViewportHeight > 0 && t.Selected >= t.ScrollOffset+t.ViewportHeight {
		t.ScrollOffset = t.Selected - t.ViewportHeight + 1
	}
}

func (t *Table) MovePageUp() {
	pageSize := t.ViewportHeight
	if pageSize <= 0 {
		pageSize = 10
	}
	t.Selected -= pageSize
	if t.Selected < 0 {
		t.Selected = 0
	}
	if t.Selected < t.ScrollOffset {
		t.ScrollOffset = t.Selected
	}
}

func (t *Table) ScrollTo(index int) {
	rows, _ := t.activeRows()
	if index < 0 {
		index = 0
	}
	if index >= len(rows) {
		index = len(rows) - 1
	}
	if index < 0 {
		return
	}
	t.Selected = index
	if t.Selected < t.ScrollOffset {
		t.ScrollOffset = t.Selected
	}
	if t.ViewportHeight > 0 && t.Selected >= t.ScrollOffset+t.ViewportHeight {
		t.ScrollOffset = t.Selected - t.ViewportHeight + 1
	}
}

func (t *Table) ApplyFilter(query string) {
	t.FilterQuery = query
	if query == "" {
		t.filteredRows = nil
		t.filteredRowIDs = nil
		return
	}
	lowerQuery := strings.ToLower(query)
	t.filteredRows = nil
	t.filteredRowIDs = nil
	for i, row := range t.Rows {
		for _, cell := range row {
			if strings.Contains(strings.ToLower(cell), lowerQuery) {
				t.filteredRows = append(t.filteredRows, row)
				if i < len(t.RowIDs) {
					t.filteredRowIDs = append(t.filteredRowIDs, t.RowIDs[i])
				}
				break
			}
		}
	}
	t.Selected = 0
	t.ScrollOffset = 0
}

func (t *Table) ClearFilter() {
	t.FilterQuery = ""
	t.filteredRows = nil
	t.filteredRowIDs = nil
	if t.Selected >= len(t.Rows) {
		t.Selected = 0
	}
	t.ScrollOffset = 0
}

func (t *Table) SelectedRow() []string {
	rows, _ := t.activeRows()
	if len(rows) == 0 || t.Selected < 0 || t.Selected >= len(rows) {
		return nil
	}
	return rows[t.Selected]
}

func (t *Table) SelectedID() string {
	_, ids := t.activeRows()
	if len(ids) == 0 || t.Selected < 0 || t.Selected >= len(ids) {
		return ""
	}
	return ids[t.Selected]
}

func (t *Table) SelectedIndex() int {
	rows, _ := t.activeRows()
	if len(rows) == 0 || t.Selected < 0 || t.Selected >= len(rows) {
		return -1
	}
	return t.Selected
}

// Render renders the table as a string.
func (t *Table) Render() string {
	if len(t.Headers) == 0 {
		return ""
	}

	const cellPadding = 2 // Padding(0,1) = 1 char each side

	rows, _ := t.activeRows()

	colWidths := make([]int, len(t.Headers))
	for i, h := range t.Headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Cap column widths to fit terminal
	if t.Width > 0 {
		totalWidth := 0
		for _, w := range colWidths {
			totalWidth += w + cellPadding
		}
		if totalWidth > t.Width {
			// Minimum content width per column: header length (at least 4)
			minWidths := make([]int, len(colWidths))
			totalMin := 0
			for i, h := range t.Headers {
				minWidths[i] = len(h)
				if minWidths[i] < 4 {
					minWidths[i] = 4
				}
				totalMin += minWidths[i] + cellPadding
			}

			// Distribute remaining space proportionally to columns that need more
			extraSpace := t.Width - totalMin
			if extraSpace < 0 {
				extraSpace = 0
			}

			totalExtra := 0
			extras := make([]int, len(colWidths))
			for i := range colWidths {
				extras[i] = colWidths[i] - minWidths[i]
				if extras[i] < 0 {
					extras[i] = 0
				}
				totalExtra += extras[i]
			}

			for i := range colWidths {
				if totalExtra > 0 && extraSpace > 0 {
					extra := int(float64(extras[i]) / float64(totalExtra) * float64(extraSpace))
					colWidths[i] = minWidths[i] + extra
				} else {
					colWidths[i] = minWidths[i]
				}
			}
		}
	}

	var sb strings.Builder

	// Header
	var headerCells []string
	for i, h := range t.Headers {
		w := colWidths[i] + cellPadding
		cell := tui.TableHeaderStyle.Width(w).MaxWidth(w).Render(truncate(h, colWidths[i]))
		headerCells = append(headerCells, cell)
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	sb.WriteString("\n")

	// Separator
	for i, cw := range colWidths {
		sb.WriteString(strings.Repeat("─", cw+cellPadding))
		if i < len(colWidths)-1 {
			sb.WriteString("┼")
		}
	}
	sb.WriteString("\n")

	startIdx := 0
	endIdx := len(rows)
	if t.ViewportHeight > 0 {
		startIdx = t.ScrollOffset
		endIdx = t.ScrollOffset + t.ViewportHeight
		if endIdx > len(rows) {
			endIdx = len(rows)
		}
		if startIdx > endIdx {
			startIdx = endIdx
		}
	}

	for rowIdx := startIdx; rowIdx < endIdx; rowIdx++ {
		row := rows[rowIdx]
		var cells []string
		for i := range t.Headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			w := colWidths[i] + cellPadding
			style := tui.TableCellStyle.Width(w).MaxWidth(w)
			if rowIdx == t.Selected {
				style = tui.TableSelectedStyle.Width(w).MaxWidth(w)
			}
			cells = append(cells, style.Render(truncate(cell, colWidths[i])))
		}
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells...))
		sb.WriteString("\n")
	}

	if t.ViewportHeight > 0 && len(rows) > 0 {
		indicator := fmt.Sprintf("▼ %d/%d", t.Selected+1, len(rows))
		sb.WriteString(tui.ScrollIndicatorStyle.Width(t.Width).Render(indicator))
		sb.WriteString("\n")
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
