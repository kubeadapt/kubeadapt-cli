package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kubeadapt/kubeadapt-cli/internal/version"
)

// The server returns its correlation id in the envelope body, not as a header,
// so reading only X-Request-ID left every --verbose log line blank and made
// support requests unanswerable.
func TestResolveRequestID_FallsBackToEnvelopeBody(t *testing.T) {
	t.Parallel()

	body := []byte(`{"data":[],"meta":{"request_id":"req-from-body","applied_at":"2026-07-26T00:00:00Z"}}`)

	assert.Equal(t, "req-from-body", resolveRequestID(nil, body))
}

func TestResolveRequestID_PrefersHeaderWhenPresent(t *testing.T) {
	t.Parallel()

	hdr := map[string][]string{"X-Request-Id": {"req-from-header"}}
	body := []byte(`{"meta":{"request_id":"req-from-body"}}`)

	assert.Equal(t, "req-from-header", resolveRequestID(hdr, body))
}

func TestResolveRequestID_ToleratesGarbageBody(t *testing.T) {
	t.Parallel()

	assert.Empty(t, resolveRequestID(nil, []byte("not json at all")))
	assert.Empty(t, resolveRequestID(nil, nil))
}

// A released binary reporting itself as "dev" makes server-side version
// analytics useless and hides which CLI build produced a bad request.
//
// Deliberately not parallel: this mutates the package-level version.Version,
// and every request in this package reads it through buildUserAgent.
func TestUserAgent_CarriesBuildVersion(t *testing.T) {
	assert.Contains(t, buildUserAgent(), version.Version)

	// Proves derivation rather than a coincidental literal: a local build's
	// version genuinely is "dev", so equality alone can't distinguish the two.
	orig := version.Version
	t.Cleanup(func() { version.Version = orig })
	version.Version = "9.9.9"
	assert.Equal(t, "kubeadapt-cli/9.9.9", buildUserAgent())
}
