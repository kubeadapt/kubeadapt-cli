package types

// Pagination is cursor-based. NextCursor is opaque and must be echoed back
// verbatim. HasMore is the authoritative end-of-results signal - a cursor may
// be emitted even on the final page. TotalCount needs ?include_total=true.
type Pagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
	TotalCount *int   `json:"total_count,omitempty"`
}

// Meta is response metadata. CostMode is "fully_loaded" or "workload_only" and
// is present only on endpoints accepting ?cost_mode=. Basis echoes the cost
// basis on cost-bearing endpoints; each block's own basis is on its CostVector.
type Meta struct {
	RequestID  string      `json:"request_id"`
	AppliedAt  string      `json:"applied_at"`
	CostMode   string      `json:"cost_mode,omitempty"`
	Basis      string      `json:"basis,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}
