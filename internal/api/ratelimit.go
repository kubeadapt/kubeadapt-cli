package api

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Updated atomically after every successful or rate-limited response; reads
// return a value copy.
type RateLimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// Goroutine-safe holder for the latest value observed on the wire.
type rateLimitSnapshot struct {
	mu  sync.RWMutex
	val RateLimit
}

func (r *rateLimitSnapshot) load() RateLimit {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.val
}

// Reset is Unix epoch seconds by convention, but an HTTP-date is accepted as an
// absolute time. Malformed values are silently zeroed - this telemetry is
// best-effort and must never break a request.
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

// Pausing at the floor beats tripping the limiter: the server is fail-closed and
// answers every 429 with a flat Retry-After: 60, so overrunning the window by
// one request costs a full minute no matter how close the reset was.
func (r RateLimit) PaceUntil(now time.Time, floor int) time.Duration {
	if r.Limit <= 0 || r.Remaining > floor || r.Reset.IsZero() {
		return 0
	}
	if d := r.Reset.Sub(now); d > 0 {
		return d
	}
	return 0
}

// Rate-limit waits run for whole minutes, so a bare time.Sleep would swallow
// Ctrl-C for that long.
func SleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
