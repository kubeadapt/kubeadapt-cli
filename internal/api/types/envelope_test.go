package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fixtureItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestEnvelope_UnmarshalSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantData fixtureItem
		wantMeta Meta
	}{
		{
			name: "single object payload",
			input: `{
				"data": {"id":"c-1","name":"prod"},
				"meta": {"request_id":"req-abc","applied_at":"2026-05-20T12:00:00Z"}
			}`,
			wantData: fixtureItem{ID: "c-1", Name: "prod"},
			wantMeta: Meta{RequestID: "req-abc", AppliedAt: "2026-05-20T12:00:00Z"},
		},
		{
			name: "with cost_mode and pagination",
			input: `{
				"data": {"id":"c-2","name":"dev"},
				"meta": {
					"request_id":"req-def",
					"applied_at":"2026-05-20T12:01:00Z",
					"cost_mode":"fully_loaded",
					"pagination":{"next_cursor":"abc","has_more":true,"limit":50}
				}
			}`,
			wantData: fixtureItem{ID: "c-2", Name: "dev"},
			wantMeta: Meta{
				RequestID: "req-def",
				AppliedAt: "2026-05-20T12:01:00Z",
				CostMode:  "fully_loaded",
				Pagination: &Pagination{
					NextCursor: "abc",
					HasMore:    true,
					Limit:      50,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var env Envelope[fixtureItem]
			require.NoError(t, json.Unmarshal([]byte(tt.input), &env))
			assert.Equal(t, tt.wantData, env.Data)
			assert.Equal(t, tt.wantMeta, env.Meta)
			assert.Nil(t, env.Error)
		})
	}
}

func TestEnvelope_UnmarshalSliceData(t *testing.T) {
	t.Parallel()

	const input = `{
		"data": [{"id":"a","name":"alpha"},{"id":"b","name":"beta"}],
		"meta": {"request_id":"r","applied_at":"2026-05-20T12:00:00Z"}
	}`

	var env Envelope[[]fixtureItem]
	require.NoError(t, json.Unmarshal([]byte(input), &env))
	require.Len(t, env.Data, 2)
	assert.Equal(t, "a", env.Data[0].ID)
	assert.Equal(t, "b", env.Data[1].ID)
}

func TestEnvelope_UnmarshalError(t *testing.T) {
	t.Parallel()

	const input = `{
		"data": null,
		"meta": {"request_id":"req-err","applied_at":"2026-05-20T12:00:00Z"},
		"error": {
			"code":"not_found",
			"message":"cluster not found",
			"details":[{"field":"cluster_id"}]
		}
	}`

	var env Envelope[*fixtureItem]
	require.NoError(t, json.Unmarshal([]byte(input), &env))
	assert.Nil(t, env.Data)
	require.NotNil(t, env.Error)
	assert.Equal(t, "not_found", env.Error.Code)
	assert.Equal(t, "cluster not found", env.Error.Message)
	require.Len(t, env.Error.Details, 1)
	assert.Equal(t, "cluster_id", env.Error.Details[0]["field"])
}

func TestEnvelope_MarshalRoundTrip(t *testing.T) {
	t.Parallel()

	original := Envelope[fixtureItem]{
		Data: fixtureItem{ID: "c-1", Name: "prod"},
		Meta: Meta{
			RequestID: "req-1",
			AppliedAt: "2026-05-20T12:00:00Z",
			CostMode:  "workload_only",
			Pagination: &Pagination{
				NextCursor: "cursor-xyz",
				HasMore:    false,
				Limit:      25,
			},
		},
	}

	raw, err := json.Marshal(original)
	require.NoError(t, err)

	var round Envelope[fixtureItem]
	require.NoError(t, json.Unmarshal(raw, &round))

	assert.Equal(t, original, round)
}
