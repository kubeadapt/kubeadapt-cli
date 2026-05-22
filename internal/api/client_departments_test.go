package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListDepartments(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.ListDepartments(t.Context(), api.DepartmentFilter{})
	require.NoError(t, err)
	require.NotEmpty(t, got, "expected at least one department")
}

func TestGetDepartment(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, _, err := c.GetDepartment(t.Context(), "dept-1", api.DepartmentGetOpts{
		CostModeOpt: api.CostModeOpt{CostMode: "workload_only"},
	})
	require.NoError(t, err)
	require.NotNil(t, got, "expected department payload")
	assert.NotEmpty(t, got.ID, "expected department payload")
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Equal(t, "workload_only", last.Query.Get("cost_mode"))
}
