package api

import (
	"net/url"
	"strconv"
	"strings"
)

// PagedOpts is the common pagination payload accepted by every list endpoint.
// It is embedded into per-resource filter structs.
//
// Limit==0 means "use the server default" (currently 100). Cursor is the
// opaque next_cursor value returned by Meta.Pagination on the previous
// page; pass an empty string for the first page. IncludeTotal adds
// ?include_total=true, which forces the server to compute the total
// count (expensive — opt-in only).
type PagedOpts struct {
	Limit        int    // 0 means server default (100)
	Cursor       string // empty means first page
	IncludeTotal bool   // adds ?include_total=true (expensive)
}

// CostModeOpt is embedded into filter structs for endpoints that accept the
// cost_mode query parameter. The empty string means "do not send the param"
// (the server picks its own default, currently "fully_loaded"). Valid
// non-empty values are "fully_loaded" and "workload_only" — the per-flag
// parser enforces that; the transport just forwards.
type CostModeOpt struct {
	CostMode string // "" | "fully_loaded" | "workload_only"
}

// setCostMode forwards the cost_mode value when non-empty. Endpoints whose
// filter struct does NOT embed CostModeOpt must not call this — the
// absence of the field is how callers opt out.
func setCostMode(params url.Values, cm string) {
	if cm != "" {
		params.Set("cost_mode", cm)
	}
}

// setCSV writes a comma-separated list under key when the input slice has
// at least one non-empty element. Empty strings inside values are dropped
// to keep the query stable when callers pre-allocate slices.
func setCSV(params url.Values, key string, values []string) {
	nonEmpty := make([]string, 0, len(values))
	for _, v := range values {
		if v != "" {
			nonEmpty = append(nonEmpty, v)
		}
	}
	if len(nonEmpty) > 0 {
		params.Set(key, strings.Join(nonEmpty, ","))
	}
}

// setTriBool writes the boolean as "true"/"false" when the pointer is non-nil
// and skips the param entirely when it is nil. This is the standard
// tri-state encoding used by the filter structs.
func setTriBool(params url.Values, key string, v *bool) {
	if v != nil {
		params.Set(key, strconv.FormatBool(*v))
	}
}
