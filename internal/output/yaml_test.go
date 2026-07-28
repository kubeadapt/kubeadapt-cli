package output

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
)

// testYAMLData mirrors how the real /v1 resource types are tagged: `json:`
// only. RenderYAML derives YAML keys from those tags, so a fixture tagged
// `yaml:` would exercise a tag set no production type actually carries.
type testYAMLData struct {
	Name string   `json:"name"`
	Cost *float64 `json:"cost"`
}

func TestRenderYAML_ValidStruct(t *testing.T) {
	cost := 42.5
	data := testYAMLData{Name: "test", Cost: &cost}
	var buf bytes.Buffer
	require.NoError(t, RenderYAML(&buf, data))
	got := buf.String()
	for _, want := range []string{"name:", "test", "cost:"} {
		assert.Contains(t, got, want)
	}
}

func TestRenderYAML_NilFields(t *testing.T) {
	data := testYAMLData{Name: "test", Cost: nil}
	var buf bytes.Buffer
	require.NoError(t, RenderYAML(&buf, data))
	assert.Contains(t, buf.String(), "name:")
}

// yamlSampleCluster is a fully-populated Cluster used by the key-fidelity tests.
// Every field that carries a multi-word json tag is set, because those are
// exactly the fields that silently degraded when the encoder fell back to
// lowercasing Go field names.
func yamlSampleCluster() types.Cluster {
	total := 7
	return types.Cluster{
		ID:   "clu_123",
		Kind: "cluster",
		Metadata: types.ClusterMetadata{
			Name:              "prod-eu",
			Provider:          "aws",
			Service:           "eks",
			Region:            "eu-west-1",
			AvailabilityZones: []string{"eu-west-1a", "eu-west-1b"},
			Environment:       "production",
			Status:            "active",
			IsStale:           true,
			K8sVersion:        "1.31.4",
			AgentVersion:      "0.9.2",
			DiscoverySource:   "agent",
			CreatedAt:         "2025-01-02T03:04:05Z",
			LastSeenAt:        "2025-06-07T08:09:10Z",
		},
		Utilization: types.ClusterUtilization{
			Counts: types.ClusterCounts{
				Nodes:             3,
				RunningPods:       12,
				RunningContainers: 30,
				PersistentVolumes: total,
			},
		},
		Cost: types.ClusterCost{
			CurrentRunRateHourly: types.Money{Amount: "12.4700", Currency: "USD"},
			LastUpdatedAt:        "2025-06-07T08:00:00Z",
		},
	}
}

func TestRenderYAML_PreservesSnakeCaseKeys(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, RenderYAML(&buf, yamlSampleCluster()))
	got := buf.String()

	for _, want := range []string{"is_stale:", "k8s_version:", "created_at:", "availability_zones:", "last_seen_at:", "running_pods:"} {
		assert.Contains(t, got, want, "snake_case key from the json tag must survive YAML rendering")
	}
	for _, mangled := range []string{"isstale:", "k8sversion:", "createdat:", "availabilityzones:", "lastseenat:", "runningpods:"} {
		assert.NotContains(t, got, mangled, "field name must not be lowercased into a mangled key")
	}
}

func TestRenderYAML_KeysMatchJSONKeys(t *testing.T) {
	value := yamlSampleCluster()

	var jsonBuf, yamlBuf bytes.Buffer
	require.NoError(t, RenderJSON(&jsonBuf, value))
	require.NoError(t, RenderYAML(&yamlBuf, value))

	var fromJSON, fromYAML any
	require.NoError(t, json.Unmarshal(jsonBuf.Bytes(), &fromJSON))
	require.NoError(t, yamlv3.Unmarshal(yamlBuf.Bytes(), &fromYAML))

	assert.Equal(t, keyPaths(fromJSON), keyPaths(fromYAML),
		"-o json and -o yaml must describe the same document")
}

func TestRenderYAMLWithMeta_EnvelopeKeys(t *testing.T) {
	meta := &types.Meta{
		RequestID: "req_1",
		AppliedAt: "2025-06-07T08:09:10Z",
		Pagination: &types.Pagination{
			NextCursor: "cur_2",
			HasMore:    true,
			Limit:      50,
		},
	}

	var buf bytes.Buffer
	require.NoError(t, RenderYAMLWithMeta(&buf, []types.Cluster{yamlSampleCluster()}, meta))

	// Decoded into a bare map rather than the typed envelope: types.Meta has no
	// yaml tags, so a typed decode would look for `requestid` and mask exactly
	// the key mangling this test exists to catch.
	var envelope map[string]any
	require.NoError(t, yamlv3.Unmarshal(buf.Bytes(), &envelope))

	require.Len(t, envelope["data"], 1)
	metaBlock, ok := envelope["meta"].(map[string]any)
	require.True(t, ok, "meta must decode as a mapping")
	assert.Equal(t, "req_1", metaBlock["request_id"])

	got := buf.String()
	assert.Contains(t, got, "data:")
	assert.Contains(t, got, "meta:")
	assert.Contains(t, got, "request_id:")
	assert.Contains(t, got, "next_cursor:")
	assert.Contains(t, got, "has_more:")
}

func TestRenderYAML_NilSliceEmitsEmptyList(t *testing.T) {
	var clusters []types.Cluster

	var buf bytes.Buffer
	require.NoError(t, RenderYAML(&buf, clusters))
	assert.Contains(t, buf.String(), "[]")
	assert.NotContains(t, buf.String(), "null")
}

func TestRenderYAMLWithMeta_NilSliceEmitsEmptyList(t *testing.T) {
	var clusters []types.Cluster

	var buf bytes.Buffer
	require.NoError(t, RenderYAMLWithMeta(&buf, clusters, &types.Meta{RequestID: "req_1"}))
	assert.Contains(t, buf.String(), "data: []")
}

// keyPaths flattens a decoded document into a sorted list of dotted key paths
// so two encodings can be compared on key identity alone, independent of the
// order the encoder chose to emit them in.
func keyPaths(v any) []string {
	var out []string
	var walk func(prefix string, v any)
	walk = func(prefix string, v any) {
		switch t := v.(type) {
		case map[string]any:
			for k, child := range t {
				path := k
				if prefix != "" {
					path = prefix + "." + k
				}
				out = append(out, path)
				walk(path, child)
			}
		case []any:
			for _, item := range t {
				walk(prefix+"[]", item)
			}
		}
	}
	walk("", v)
	sort.Strings(out)
	return out
}
