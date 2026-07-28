package api

import (
	"net/url"
	"strconv"
	"strings"
)

// PagedOpts is embedded into per-resource filter structs.
type PagedOpts struct {
	Limit        int    // 0 means server default (100)
	Cursor       string // empty means first page
	IncludeTotal bool   // adds ?include_total=true (expensive)
}

// CostModeOpt is embedded only by endpoints accepting cost_mode. Validation
// happens in the flag parser; the transport just forwards.
type CostModeOpt struct {
	CostMode string // "" | "fully_loaded" | "workload_only"
}

// Endpoints not embedding CostModeOpt must not call this - the absent field is
// how they opt out.
func setCostMode(params url.Values, cm string) {
	if cm != "" {
		params.Set("cost_mode", cm)
	}
}

// Empty strings are dropped to keep the query stable when callers pre-allocate
// slices.
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

// A nil pointer skips the param entirely - the tri-state encoding used by all
// filter structs.
func setTriBool(params url.Values, key string, v *bool) {
	if v != nil {
		params.Set(key, strconv.FormatBool(*v))
	}
}
