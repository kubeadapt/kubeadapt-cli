package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaceUntil(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		rl   RateLimit
		want time.Duration
	}{
		{
			name: "absent headers never pace",
			rl:   RateLimit{},
		},
		{
			name: "budget above floor proceeds immediately",
			rl:   RateLimit{Limit: 100, Remaining: 40, Reset: now.Add(30 * time.Second)},
		},
		{
			name: "at floor waits for the window to reset",
			rl:   RateLimit{Limit: 100, Remaining: 5, Reset: now.Add(30 * time.Second)},
			want: 30 * time.Second,
		},
		{
			name: "exhausted waits for the window to reset",
			rl:   RateLimit{Limit: 100, Remaining: 0, Reset: now.Add(47 * time.Second)},
			want: 47 * time.Second,
		},
		{
			name: "reset already elapsed proceeds immediately",
			rl:   RateLimit{Limit: 100, Remaining: 0, Reset: now.Add(-time.Second)},
		},
		{
			name: "missing reset cannot be paced against",
			rl:   RateLimit{Limit: 100, Remaining: 0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.rl.PaceUntil(now, 5))
		})
	}
}

func TestSleepContext_RespectsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	start := time.Now()
	err := SleepContext(ctx, time.Hour)

	require.ErrorIs(t, err, context.Canceled)
	assert.Less(t, time.Since(start), time.Second,
		"a cancelled context must abort the wait instead of running the timer down")
}

func TestSleepContext_WaitsOutTheDuration(t *testing.T) {
	start := time.Now()
	require.NoError(t, SleepContext(t.Context(), 20*time.Millisecond))
	assert.GreaterOrEqual(t, time.Since(start), 20*time.Millisecond)
}

func TestSleepContext_NonPositiveReturnsImmediately(t *testing.T) {
	require.NoError(t, SleepContext(t.Context(), 0))
	require.NoError(t, SleepContext(t.Context(), -time.Second))
}
