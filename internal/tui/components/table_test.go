package components

import (
	"strings"
	"testing"
)

func TestTableVirtualScrollViewport(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := make([][]string, 20)
	for i := 0; i < 20; i++ {
		rows[i] = []string{"row" + string(rune('A'+i)), "active"}
	}

	table := NewTable(headers, rows)
	table.ViewportHeight = 5
	table.Width = 50

	output := table.Render()

	lines := strings.Split(output, "\n")
	dataRowCount := 0
	for _, line := range lines {
		if line != "" && !strings.Contains(line, "Name") && !strings.Contains(line, "─") && !strings.Contains(line, "▼") {
			dataRowCount++
		}
	}

	if dataRowCount > 5 {
		t.Errorf("Expected at most 5 visible data rows, got %d", dataRowCount)
	}

	if !strings.Contains(output, "▼") {
		t.Error("Expected scroll indicator '▼' in output")
	}
}

func TestTableMoveDownScrolls(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := make([][]string, 20)
	for i := 0; i < 20; i++ {
		rows[i] = []string{"row" + string(rune('A'+i)), "active"}
	}

	table := NewTable(headers, rows)
	table.ViewportHeight = 5
	table.Width = 50

	for i := 0; i < 6; i++ {
		table.MoveDown()
	}

	if table.ScrollOffset <= 0 {
		t.Errorf("Expected ScrollOffset > 0 after moving down 6 times, got %d", table.ScrollOffset)
	}

	if table.Selected != 6 {
		t.Errorf("Expected Selected == 6, got %d", table.Selected)
	}
}

func TestTableMoveUpScrolls(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := make([][]string, 20)
	for i := 0; i < 20; i++ {
		rows[i] = []string{"row" + string(rune('A'+i)), "active"}
	}

	table := NewTable(headers, rows)
	table.ViewportHeight = 5
	table.Width = 50

	for i := 0; i < 10; i++ {
		table.MoveDown()
	}

	initialOffset := table.ScrollOffset
	if initialOffset <= 0 {
		t.Fatal("Expected ScrollOffset > 0 after moving down")
	}

	for i := 0; i < 10; i++ {
		table.MoveUp()
	}

	if table.ScrollOffset != 0 {
		t.Errorf("Expected ScrollOffset == 0 after moving back to top, got %d", table.ScrollOffset)
	}

	if table.Selected != 0 {
		t.Errorf("Expected Selected == 0, got %d", table.Selected)
	}
}

func TestTableApplyFilter(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"auth-service", "running"},
		{"billing-svc", "running"},
		{"cache-redis", "stopped"},
	}

	table := NewTable(headers, rows)
	table.ApplyFilter("auth")

	if table.TotalRows() != 1 {
		t.Errorf("Expected 1 row after filtering for 'auth', got %d", table.TotalRows())
	}

	selectedRow := table.SelectedRow()
	if selectedRow == nil {
		t.Fatal("Expected non-nil selected row")
	}

	if !strings.Contains(selectedRow[0], "auth") {
		t.Errorf("Expected selected row to contain 'auth', got %v", selectedRow)
	}
}

func TestTableClearFilter(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"auth-service", "running"},
		{"billing-svc", "running"},
		{"cache-redis", "stopped"},
	}

	table := NewTable(headers, rows)
	table.ApplyFilter("auth")

	if table.TotalRows() != 1 {
		t.Fatalf("Expected 1 row after filtering, got %d", table.TotalRows())
	}

	table.ClearFilter()

	if table.TotalRows() != 3 {
		t.Errorf("Expected 3 rows after clearing filter, got %d", table.TotalRows())
	}

	if table.FilterQuery != "" {
		t.Errorf("Expected empty FilterQuery after clearing, got %q", table.FilterQuery)
	}
}

func TestTableBackwardCompat(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"service-1", "running"},
		{"service-2", "running"},
		{"service-3", "stopped"},
		{"service-4", "running"},
		{"service-5", "stopped"},
	}

	table := NewTable(headers, rows)
	table.ViewportHeight = 0
	table.Width = 50

	output := table.Render()

	for _, row := range rows {
		if !strings.Contains(output, row[0]) {
			t.Errorf("Expected output to contain %q, but it was missing", row[0])
		}
	}

	if strings.Contains(output, "▼") {
		t.Error("Expected no scroll indicator when ViewportHeight == 0")
	}
}

func TestTableSelectedID(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"service-1", "running"},
		{"service-2", "running"},
		{"service-3", "stopped"},
	}
	rowIDs := []string{"id-1", "id-2", "id-3"}

	table := NewTable(headers, rows)
	table.RowIDs = rowIDs

	table.MoveDown()
	table.MoveDown()

	selectedID := table.SelectedID()
	if selectedID != "id-3" {
		t.Errorf("Expected SelectedID() == 'id-3', got %q", selectedID)
	}

	if table.Selected != 2 {
		t.Errorf("Expected Selected == 2, got %d", table.Selected)
	}
}

func TestTableMovePageDown(t *testing.T) {
	headers := []string{"Name"}
	rows := make([][]string, 30)
	for i := 0; i < 30; i++ {
		rows[i] = []string{"row" + string(rune('A'+i))}
	}

	table := NewTable(headers, rows)
	table.ViewportHeight = 10
	table.Width = 50

	table.MovePageDown()

	if table.Selected != 10 {
		t.Errorf("Expected Selected == 10 after page down, got %d", table.Selected)
	}

	if table.ScrollOffset <= 0 {
		t.Errorf("Expected ScrollOffset > 0 after page down, got %d", table.ScrollOffset)
	}
}

func TestTableMovePageUp(t *testing.T) {
	headers := []string{"Name"}
	rows := make([][]string, 30)
	for i := 0; i < 30; i++ {
		rows[i] = []string{"row" + string(rune('A'+i))}
	}

	table := NewTable(headers, rows)
	table.ViewportHeight = 10
	table.Width = 50

	table.MovePageDown()
	table.MovePageDown()

	initialSelected := table.Selected
	if initialSelected < 10 {
		t.Fatal("Expected Selected >= 10 after two page downs")
	}

	table.MovePageUp()

	if table.Selected >= initialSelected {
		t.Errorf("Expected Selected to decrease after page up, was %d, now %d", initialSelected, table.Selected)
	}
}

func TestTableScrollTo(t *testing.T) {
	headers := []string{"Name"}
	rows := make([][]string, 20)
	for i := 0; i < 20; i++ {
		rows[i] = []string{"row" + string(rune('A'+i))}
	}

	table := NewTable(headers, rows)
	table.ViewportHeight = 5
	table.Width = 50

	table.ScrollTo(15)

	if table.Selected != 15 {
		t.Errorf("Expected Selected == 15, got %d", table.Selected)
	}

	if table.ScrollOffset <= 0 {
		t.Errorf("Expected ScrollOffset > 0 when scrolled to row 15, got %d", table.ScrollOffset)
	}
}

func TestTableFilterWithRowIDs(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"auth-service", "running"},
		{"billing-svc", "running"},
		{"cache-redis", "stopped"},
		{"auth-worker", "running"},
	}
	rowIDs := []string{"id-1", "id-2", "id-3", "id-4"}

	table := NewTable(headers, rows)
	table.RowIDs = rowIDs
	table.ApplyFilter("auth")

	if table.TotalRows() != 2 {
		t.Errorf("Expected 2 rows after filtering for 'auth', got %d", table.TotalRows())
	}

	selectedID := table.SelectedID()
	if selectedID != "id-1" {
		t.Errorf("Expected first filtered row to have ID 'id-1', got %q", selectedID)
	}

	table.MoveDown()
	selectedID = table.SelectedID()
	if selectedID != "id-4" {
		t.Errorf("Expected second filtered row to have ID 'id-4', got %q", selectedID)
	}
}
