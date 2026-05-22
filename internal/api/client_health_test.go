package api_test

import (
	"testing"

	"github.com/kubeadapt/kubeadapt-cli/internal/api"
	"github.com/kubeadapt/kubeadapt-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	ms := testutil.NewMockServer(t)
	c := api.NewClient(ms.URL, "test-key")
	got, err := c.Health(t.Context())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "ok", got.Status)
	last := ms.Requests()[len(ms.Requests())-1]
	assert.Empty(t, last.Authorization, "Health must not send Authorization header")
}
