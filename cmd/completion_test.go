package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The deprecation is a compile-time fact, not a runtime one: ExactValidArgs
// and MatchAll(ExactArgs(1), OnlyValidArgs) behave identically, so only the
// source can witness the difference.
func TestCompletion_ArgsValidatorNotDeprecated(t *testing.T) {
	require.NotNil(t, completionCmd.Args, "completion must validate its positional argument")

	assert.Error(t, completionCmd.Args(completionCmd, []string{"tcsh"}),
		"an unsupported shell must be rejected")
	assert.NoError(t, completionCmd.Args(completionCmd, []string{"bash"}),
		"a supported shell must be accepted")
	assert.Error(t, completionCmd.Args(completionCmd, []string{}),
		"a missing shell argument must be rejected")
	assert.Error(t, completionCmd.Args(completionCmd, []string{"bash", "zsh"}),
		"more than one shell must be rejected")

	src, err := os.ReadFile("completion.go")
	require.NoError(t, err)
	assert.NotContains(t, string(src), "ExactValidArgs",
		"cobra.ExactValidArgs is deprecated since cobra 1.8; use MatchAll(ExactArgs(1), OnlyValidArgs)")
}

func TestCompletion_ValidArgsUnchanged(t *testing.T) {
	assert.Equal(t, []string{"bash", "zsh", "fish", "powershell"}, completionCmd.ValidArgs)
}
