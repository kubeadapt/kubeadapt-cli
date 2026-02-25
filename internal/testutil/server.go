package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

func validAuth(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	return token != "" && token != auth // has Bearer prefix and non-empty token
}

// MockHandler creates a handler that returns the given response.
func MockHandler(statusCode int, response interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validAuth(r) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"detail": "Missing authorization"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if response != nil {
			_ = json.NewEncoder(w).Encode(response)
		}
	}
}

// NewMockServer creates a test server with routes matching the KubeAdapt API.
func NewMockServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/overview", MockHandler(http.StatusOK, SampleOverview()))

	mux.HandleFunc("/v1/clusters", MockHandler(http.StatusOK, map[string]interface{}{
		"clusters": SampleClusters(),
		"total":    len(SampleClusters()),
	}))

	clustersByID := map[string]interface{}{}
	for _, c := range SampleClusters() {
		clustersByID[c.ID] = c
	}
	mux.HandleFunc("/v1/clusters/", func(w http.ResponseWriter, r *http.Request) {
		if !validAuth(r) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"detail": "Missing authorization"})
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/v1/clusters/")
		if strings.HasSuffix(id, "/dashboard") {
			clusterID := strings.TrimSuffix(id, "/dashboard")
			if _, ok := clustersByID[clusterID]; ok {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(SampleClusterDashboard())
				return
			}
		}
		if strings.HasSuffix(id, "/capacity-planning") {
			clusterID := strings.TrimSuffix(id, "/capacity-planning")
			if _, ok := clustersByID[clusterID]; ok {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(SampleCapacityPlanning())
				return
			}
		}
		if cluster, ok := clustersByID[id]; ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cluster)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"detail": "Cluster not found"})
	})

	mux.HandleFunc("/v1/workloads", MockHandler(http.StatusOK, map[string]interface{}{
		"workloads": SampleWorkloads(),
		"total":     len(SampleWorkloads()),
	}))

	mux.HandleFunc("/v1/nodes", MockHandler(http.StatusOK, map[string]interface{}{
		"nodes": SampleNodes(),
		"total": len(SampleNodes()),
	}))

	mux.HandleFunc("/v1/recommendations", MockHandler(http.StatusOK, map[string]interface{}{
		"recommendations":                 SampleRecommendations(),
		"total":                           len(SampleRecommendations()),
		"total_potential_savings_monthly": 26.64,
	}))

	mux.HandleFunc("/v1/costs/teams", MockHandler(http.StatusOK, map[string]interface{}{
		"teams": []interface{}{},
		"total": 0,
	}))

	mux.HandleFunc("/v1/costs/departments", MockHandler(http.StatusOK, map[string]interface{}{
		"departments": []interface{}{},
		"total":       0,
	}))

	mux.HandleFunc("/v1/node-groups", MockHandler(http.StatusOK, map[string]interface{}{
		"node_groups": []interface{}{},
		"total":       0,
	}))

	mux.HandleFunc("/v1/namespaces", MockHandler(http.StatusOK, map[string]interface{}{
		"namespaces": SampleNamespaces(),
		"total":      len(SampleNamespaces()),
		"summary": map[string]interface{}{
			"total_hourly_cost": 0.77,
			"total_pods":        18,
			"total_workloads":   8,
		},
	}))

	mux.HandleFunc("/v1/persistent-volumes", MockHandler(http.StatusOK, map[string]interface{}{
		"persistent_volumes": []interface{}{},
		"total":              0,
	}))

	mux.HandleFunc("/v1/dashboard", MockHandler(http.StatusOK, SampleDashboard()))

	return httptest.NewServer(mux)
}
