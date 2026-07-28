package update

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The cached and freshly-fetched code paths build the same sentence from
// different sources - one v-stripped, one the raw git tag - so the same
// upgrade could be announced as "0.1.3 -> 0.2.1" or "0.1.3 -> v0.2.1"
// depending only on whether the 24h cache happened to be warm.
func TestUpgradeMessage_IdenticalForCachedAndFreshTag(t *testing.T) {
	t.Parallel()

	fromCache := upgradeMessage("0.1.3", "0.2.1")
	fromTag := upgradeMessage("0.1.3", "v0.2.1")

	assert.Equal(t, fromCache, fromTag)
}

func TestUpgradeMessage_RendersBothVersions(t *testing.T) {
	t.Parallel()

	msg := upgradeMessage("0.1.3", "v0.2.1")

	assert.Contains(t, msg, "0.1.3")
	assert.Contains(t, msg, "0.2.1")
	assert.Contains(t, msg, "brew upgrade kubeadapt")
	assert.False(t, strings.Contains(msg, "vv"), "tag prefix must not be doubled")
}
