package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every character here changes the shape of a URL when appended raw: '/' adds a
// segment, '?' and '#' end the path, '%' starts an escape, and a space is not
// legal in a request line at all. ':' is legal in a path segment and must
// survive untouched.
const awkwardSegment = "pool/a?b#c d:e%f"

// The server routes on ServeMux wildcards, which match one escaped segment
// each, so a test passes only if every user-supplied value arrives as exactly
// the segment it was meant to be.
func TestClient_EscapesUserSuppliedPathSegments(t *testing.T) {
	const s = awkwardSegment
	tests := []struct {
		name    string
		pattern string
		call    func(context.Context, *api.Client) error
		want    map[string]string
	}{
		{
			name:    "cluster",
			pattern: "GET /v1/clusters/{cluster}",
			call:    func(ctx context.Context, c *api.Client) error { _, _, err := c.GetCluster(ctx, s); return err },
			want:    map[string]string{"cluster": s},
		},
		{
			name:    "cluster namespaces",
			pattern: "GET /v1/clusters/{cluster}/namespaces",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.ListNamespaces(ctx, api.NamespaceFilter{ClusterIDs: []string{s}})
				return err
			},
			want: map[string]string{"cluster": s},
		},
		{
			name:    "namespace",
			pattern: "GET /v1/clusters/{cluster}/namespaces/{namespace}",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.GetNamespace(ctx, s, s+"-ns", api.NamespaceGetOpts{})
				return err
			},
			want: map[string]string{"cluster": s, "namespace": s + "-ns"},
		},
		{
			name:    "cluster node groups",
			pattern: "GET /v1/clusters/{cluster}/node-groups",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.ListNodeGroups(ctx, api.NodeGroupFilter{ClusterIDs: []string{s}})
				return err
			},
			want: map[string]string{"cluster": s},
		},
		{
			name:    "node group",
			pattern: "GET /v1/clusters/{cluster}/node-groups/{name}",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.GetNodeGroup(ctx, s, s+"-group")
				return err
			},
			want: map[string]string{"cluster": s, "name": s + "-group"},
		},
		{
			name:    "cluster nodes",
			pattern: "GET /v1/clusters/{cluster}/nodes",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.ListNodes(ctx, api.NodeFilter{ClusterIDs: []string{s}})
				return err
			},
			want: map[string]string{"cluster": s},
		},
		{
			name:    "node",
			pattern: "GET /v1/nodes/{node}",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.GetNode(ctx, s, api.NodeGetOpts{})
				return err
			},
			want: map[string]string{"node": s},
		},
		{
			name:    "cluster workloads",
			pattern: "GET /v1/clusters/{cluster}/workloads",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.ListWorkloads(ctx, api.WorkloadFilter{ClusterIDs: []string{s}})
				return err
			},
			want: map[string]string{"cluster": s},
		},
		{
			name:    "workload",
			pattern: "GET /v1/workloads/{workload}",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.GetWorkload(ctx, s, api.WorkloadGetOpts{})
				return err
			},
			want: map[string]string{"workload": s},
		},
		{
			name:    "workload pods",
			pattern: "GET /v1/workloads/{workload}/pods",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.ListWorkloadPods(ctx, s, api.PodFilter{})
				return err
			},
			want: map[string]string{"workload": s},
		},
		{
			name:    "recommendation",
			pattern: "GET /v1/recommendations/{rec}",
			call:    func(ctx context.Context, c *api.Client) error { _, _, err := c.GetRecommendation(ctx, s); return err },
			want:    map[string]string{"rec": s},
		},
		{
			name:    "team",
			pattern: "GET /v1/teams/{team}",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.GetTeam(ctx, s, api.TeamGetOpts{})
				return err
			},
			want: map[string]string{"team": s},
		},
		{
			name:    "team assignments",
			pattern: "GET /v1/teams/{team}/assignments",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.ListTeamAssignments(ctx, s, api.AssignmentFilter{})
				return err
			},
			want: map[string]string{"team": s},
		},
		{
			name:    "department",
			pattern: "GET /v1/departments/{dept}",
			call: func(ctx context.Context, c *api.Client) error {
				_, _, err := c.GetDepartment(ctx, s, api.DepartmentGetOpts{})
				return err
			},
			want: map[string]string{"dept": s},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := map[string]string{}
			var query string
			mux := http.NewServeMux()
			mux.HandleFunc(tt.pattern, func(w http.ResponseWriter, r *http.Request) {
				for key := range tt.want {
					got[key] = r.PathValue(key)
				}
				query = r.URL.RawQuery
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":null,"meta":{}}`))
			})
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("request did not match %q: escaped path %q", tt.pattern, r.URL.EscapedPath())
				http.NotFound(w, r)
			})
			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)

			require.NoError(t, tt.call(t.Context(), api.NewClient(srv.URL, "k")))
			assert.Equal(t, tt.want, got, "each value must arrive as a single, intact path segment")
			assert.NotContains(t, query, "b#c", "a '?' in a name must not start the query string")
		})
	}
}
