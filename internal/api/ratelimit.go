package api

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimit is a point-in-time snapshot of the API's rate-limit headers.
// It is updated atomically by the client after every successful or rate-limited
// response. Reads via Client.RateLimit() return a value copy.
type RateLimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// rateLimitSnapshot is the goroutine-safe holder for the latest RateLimit
// value observed on the wire. It is internal to the api package.
type rateLimitSnapshot struct {
	mu  sync.RWMutex
	val RateLimit
}

// load returns a value copy of the current snapshot.
func (r *rateLimitSnapshot) load() RateLimit {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.val
}

// captureFromHeaders parses X-RateLimit-Limit, X-RateLimit-Remaining, and
// X-RateLimit-Reset from the response headers. Reset is a Unix epoch seconds
// integer per backend conventions; if it parses as an HTTP-date it is used as
// the absolute reset time. Malformed values are silently zeroed — rate-limit
// telemetry is best-effort and must never break a request.
func (r *rateLimitSnapshot) captureFromHeaders(h http.Header) {
	if h == nil {
		return
	}
	limit, _ := strconv.Atoi(h.Get("X-RateLimit-Limit"))
	remaining, _ := strconv.Atoi(h.Get("X-RateLimit-Remaining"))

	var reset time.Time
	if v := h.Get("X-RateLimit-Reset"); v != "" {
		if secs, err := strconv.ParseInt(v, 10, 64); err == nil {
			reset = time.Unix(secs, 0).UTC()
		} else if t, err := http.ParseTime(v); err == nil {
			reset = t.UTC()
		}
	}

	r.mu.Lock()
	r.val = RateLimit{Limit: limit, Remaining: remaining, Reset: reset}
	r.mu.Unlock()
}
