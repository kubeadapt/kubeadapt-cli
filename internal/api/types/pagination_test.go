package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int { return &v }

func TestMeta_Unmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  Meta
	}{
		{
			name: "all fields populated",
			input: `{
				"request_id":"req-1",
				"applied_at":"2026-05-20T12:00:00Z",
				"cost_mode":"fully_loaded",
				"pagination":{"next_cursor":"cur","has_more":true,"limit":50,"total_count":120}
			}`,
			want: Meta{
				RequestID: "req-1",
				AppliedAt: "2026-05-20T12:00:00Z",
				CostMode:  "fully_loaded",
				Pagination: &Pagination{
					NextCursor: "cur",
					HasMore:    true,
					Limit:      50,
					TotalCount: intPtr(120),
				},
			},
		},
		{
			name: "no pagination block",
			input: `{
				"request_id":"req-2",
				"applied_at":"2026-05-20T12:01:00Z",
				"cost_mode":"workload_only"
			}`,
			want: Meta{
				RequestID: "req-2",
				AppliedAt: "2026-05-20T12:01:00Z",
				CostMode:  "workload_only",
			},
		},
		{
			name: "cost_mode omitted (cluster-style endpoint)",
			input: `{
				"request_id":"req-3",
				"applied_at":"2026-05-20T12:02:00Z"
			}`,
			want: Meta{
				RequestID: "req-3",
				AppliedAt: "2026-05-20T12:02:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Meta
			require.NoError(t, json.Unmarshal([]byte(tt.input), &got))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPagination_Unmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  Pagination
	}{
		{
			name:  "end of list (no next_cursor, has_more=false)",
			input: `{"has_more":false,"limit":50}`,
			want:  Pagination{HasMore: false, Limit: 50},
		},
		{
			name:  "mid-list with cursor",
			input: `{"next_cursor":"abc==","has_more":true,"limit":25}`,
			want:  Pagination{NextCursor: "abc==", HasMore: true, Limit: 25},
		},
		{
			name:  "total_count populated",
			input: `{"next_cursor":"abc","has_more":true,"limit":25,"total_count":42}`,
			want:  Pagination{NextCursor: "abc", HasMore: true, Limit: 25, TotalCount: intPtr(42)},
		},
		{
			name:  "total_count omitted is nil pointer",
			input: `{"has_more":false,"limit":10}`,
			want:  Pagination{HasMore: false, Limit: 10, TotalCount: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Pagination
			require.NoError(t, json.Unmarshal([]byte(tt.input), &got))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPagination_TotalCountPointer(t *testing.T) {
	t.Parallel()

	// Explicit check that TotalCount is a pointer that distinguishes "absent" from "zero".
	var p Pagination
	require.NoError(t, json.Unmarshal([]byte(`{"has_more":false,"limit":10}`), &p))
	assert.Nil(t, p.TotalCount, "TotalCount should be nil when absent")

	var p2 Pagination
	require.NoError(t, json.Unmarshal([]byte(`{"has_more":false,"limit":10,"total_count":0}`), &p2))
	require.NotNil(t, p2.TotalCount, "TotalCount should be *int=0 when present and zero")
	assert.Equal(t, 0, *p2.TotalCount)
}
