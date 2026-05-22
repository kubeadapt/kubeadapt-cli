package types

// Pagination is the cursor-based pagination block returned in Meta.Pagination
// on list endpoints. The NextCursor is opaque (base64url-encoded JSON internally
// to the server) and must be echoed back verbatim on the follow-up request.
//
// HasMore is the authoritative "is there another page?" signal; do NOT rely on
// NextCursor being empty to mean "end of results" because a cursor might be
// emitted even at the final page in some endpoints.
//
// TotalCount is only populated when the caller passes ?include_total=true.
type Pagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
	TotalCount *int   `json:"total_count,omitempty"`
}

// Meta is the response metadata block. RequestID is a UUID assigned by the
// server and useful for support; AppliedAt is RFC3339; CostMode is one of
// "fully_loaded" or "workload_only" and is only present on endpoints that
// accept the ?cost_mode= query parameter (cluster/node/org-root/recommendation
// endpoints omit it).
type Meta struct {
	RequestID  string      `json:"request_id"`
	AppliedAt  string      `json:"applied_at"`
	CostMode   string      `json:"cost_mode,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}
