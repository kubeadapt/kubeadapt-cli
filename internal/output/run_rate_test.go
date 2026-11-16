package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A null run rate used to decode into an empty Money and re-encode as
// {"amount":"","currency":""}, and cost_state was dropped, so -o json broke
// any script summing the amounts.
func TestNullRunRateSurvivesEveryFormat(t *testing.T) {
	payload := `{"id":"c-1","kind":"Cluster","metadata":{"name":"idle"},"cost":{
		"current_run_rate_hourly":null,"cost_state":"not_reporting",
		"cost_as_of":"2026-11-15T10:00:00Z","basis":"all","partial":false}}`
	var c types.Cluster
	require.NoError(t, json.Unmarshal([]byte(payload), &c))

	var js bytes.Buffer
	require.NoError(t, RenderJSON(&js, c))
	var got struct {
		Cost map[string]json.RawMessage `json:"cost"`
	}
	require.NoError(t, json.Unmarshal(js.Bytes(), &got))
	assert.JSONEq(t, `null`, string(got.Cost["current_run_rate_hourly"]))
	assert.JSONEq(t, `"not_reporting"`, string(got.Cost["cost_state"]))
	assert.JSONEq(t, `"2026-11-15T10:00:00Z"`, string(got.Cost["cost_as_of"]))

	var yml bytes.Buffer
	require.NoError(t, RenderYAML(&yml, c))
	assert.Contains(t, yml.String(), "current_run_rate_hourly: null")

	var table bytes.Buffer
	require.NoError(t, RenderClusters(&table, []types.Cluster{c}, nil))
	assert.Contains(t, table.String(), "not reporting")
}
