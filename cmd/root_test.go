package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeadapt/kubeadapt-cli/internal/api/types"
	"github.com/kubeadapt/kubeadapt-cli/internal/config"
	"github.com/kubeadapt/kubeadapt-cli/internal/output"
)

const ansiEscape = "\x1b"

func resetGlobals(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		apiURL = ""
		apiKey = ""
		cfgFile = ""
		outputFmt = "table"
		noColor = false
		verbose = false
		quiet = false
		// PersistentPreRunE now mutates package state inside internal/output.
		// Leaving it set would leak into whichever test runs next.
		output.SetNoColor(false)
		output.SetQuiet(false)
	})
}

// forceColorProfile makes lipgloss emit ANSI even though `go test` runs
// without a TTY, where the default renderer negotiates the Ascii profile and
// silently strips every style. Without it the colored and --no-color paths
// render byte-identical output and the assertions prove nothing.
func forceColorProfile(t *testing.T) {
	t.Helper()
	prev := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(prev) })
}

func dummyCmd() *cobra.Command {
	c := &cobra.Command{Use: "status"}
	c.SetContext(context.Background())
	return c
}

func namedCmd(name string) *cobra.Command {
	c := &cobra.Command{Use: name}
	c.SetContext(context.Background())
	return c
}

func colorSampleClusters() []types.Cluster {
	return []types.Cluster{{
		ID:   "cluster-prod-east",
		Kind: "cluster",
		Metadata: types.ClusterMetadata{
			Name:        "prod-east",
			Provider:    "aws",
			Region:      "us-east-1",
			Environment: "production",
			Status:      "connected",
		},
		Cost: types.ClusterCost{
			CurrentRunRateHourly: types.Money{Amount: "12.5000", Currency: "USD"},
		},
	}}
}

// Points cfgFile at a path that cannot exist so PersistentPreRunE falls back to
// config.Default() rather than reading the developer's real config.
func isolatedConfig(t *testing.T) {
	t.Helper()
	cfgFile = filepath.Join(t.TempDir(), "nonexistent.yaml")
}

func TestNoColorFlag_DisablesANSI(t *testing.T) {
	resetGlobals(t)
	forceColorProfile(t)
	isolatedConfig(t)
	items := colorSampleClusters()

	noColor = false
	require.NoError(t, rootCmd.PersistentPreRunE(dummyCmd(), nil))
	var colored bytes.Buffer
	require.NoError(t, output.RenderClusters(&colored, items, nil))
	require.Contains(t, colored.String(), ansiEscape,
		"precondition: without --no-color the table must be styled")

	noColor = true
	require.NoError(t, rootCmd.PersistentPreRunE(dummyCmd(), nil))
	var plain bytes.Buffer
	require.NoError(t, output.RenderClusters(&plain, items, nil))
	assert.NotContains(t, plain.String(), ansiEscape,
		"--no-color must suppress every ANSI sequence")
}

func TestNoColorEnv_DisablesANSI(t *testing.T) {
	resetGlobals(t)
	forceColorProfile(t)
	isolatedConfig(t)
	items := colorSampleClusters()

	noColor = false
	t.Setenv("NO_COLOR", "1")
	require.NoError(t, rootCmd.PersistentPreRunE(dummyCmd(), nil))
	var plain bytes.Buffer
	require.NoError(t, output.RenderClusters(&plain, items, nil))
	assert.NotContains(t, plain.String(), ansiEscape,
		"NO_COLOR must suppress every ANSI sequence")
}

func TestQuiet_SuppressesPaginationFooter(t *testing.T) {
	resetGlobals(t)
	isolatedConfig(t)
	items := colorSampleClusters()
	meta := &types.Meta{Pagination: &types.Pagination{Limit: 50, HasMore: false}}

	quiet = false
	require.NoError(t, rootCmd.PersistentPreRunE(dummyCmd(), nil))
	var loud bytes.Buffer
	require.NoError(t, output.RenderClusters(&loud, items, meta))
	require.Contains(t, loud.String(), "Showing 1",
		"precondition: the footer renders when --quiet is unset")

	quiet = true
	require.NoError(t, rootCmd.PersistentPreRunE(dummyCmd(), nil))
	var hushed bytes.Buffer
	require.NoError(t, output.RenderClusters(&hushed, items, meta))
	out := hushed.String()
	assert.NotContains(t, out, "Showing", "--quiet must drop the pagination footer")
	assert.Contains(t, out, "prod-east", "--quiet must not drop the data itself")
}

func TestOutputFormat_RejectsUnknown(t *testing.T) {
	resetGlobals(t)
	isolatedConfig(t)

	outputFmt = "xml"
	err := rootCmd.PersistentPreRunE(dummyCmd(), nil)
	require.Error(t, err, "-o xml must be rejected, not silently rendered as a table")
	for _, want := range []string{`"xml"`, "table", "json", "yaml"} {
		assert.Contains(t, err.Error(), want)
	}
}

func TestOutputFormat_AcceptsKnown(t *testing.T) {
	resetGlobals(t)
	isolatedConfig(t)

	for _, format := range []string{formatTable, formatJSON, formatYAML} {
		t.Run(format, func(t *testing.T) {
			outputFmt = format
			assert.NoError(t, rootCmd.PersistentPreRunE(dummyCmd(), nil))
		})
	}
}

// The config/logger setup in PersistentPreRunE is skipped for these three
// commands, but `-o` is a presentation flag they still accept, so validation
// has to run ahead of that early return.
func TestOutputFormat_RejectsUnknown_OnEarlyReturnCommands(t *testing.T) {
	resetGlobals(t)
	isolatedConfig(t)

	for _, name := range []string{"login", "version", "completion"} {
		t.Run(name, func(t *testing.T) {
			outputFmt = "xml"
			err := rootCmd.PersistentPreRunE(namedCmd(name), nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"xml"`)
		})
	}
}

func TestRootLongText_MatchesConfigDefault(t *testing.T) {
	assert.Contains(t, rootCmd.Long, config.Default().APIURL,
		"help text must quote the API endpoint the CLI actually defaults to")
	assert.NotContains(t, rootCmd.Long, "https://api.kubeadapt.io",
		"the bare api.kubeadapt.io host is dead and must not be advertised")
}

// DefaultPath only returns the legacy ~/.kubeadapt path when that file already
// exists, so advertising it unconditionally misdescribes every fresh install.
func TestRootLongText_QuotesResolvedConfigPath(t *testing.T) {
	assert.Contains(t, rootCmd.Long, config.DefaultPath(),
		"help text must name the path this machine actually resolves to")
	assert.NotContains(t, rootCmd.Long, "Configuration is stored in ~/.kubeadapt/config.yaml",
		"the unconditional legacy-path claim is wrong for a fresh install")
}

func TestRootLongText_DocumentsUpdateOptOut(t *testing.T) {
	assert.Contains(t, rootCmd.Long, "KUBEADAPT_NO_UPDATE_CHECK",
		"the update-check opt-out must be discoverable from --help")
}

// The registered completion values are the same list the server enforces, so
// deriving from them keeps this test honest if the enum ever changes.
func statusEnumValues(t *testing.T) []string {
	t.Helper()
	fn, ok := getRecommendationsCmd.GetFlagCompletionFunc("status")
	require.True(t, ok, "--status must have registered completions")
	values, _ := fn(getRecommendationsCmd, nil, "")
	return values
}

func TestGetExample_UsesRegisteredStatusValue(t *testing.T) {
	const marker = "--status "
	idx := strings.Index(getCmd.Example, marker)
	require.GreaterOrEqual(t, idx, 0, "the get example must still demonstrate --status")

	used := strings.Fields(getCmd.Example[idx+len(marker):])[0]
	assert.Contains(t, statusEnumValues(t), used,
		"copy-pasting the --help example must not produce a server rejection")
}

func TestReport_UpgradeBannerFollowsError(t *testing.T) {
	resetGlobals(t)
	const banner = "A new version of kubeadapt is available: 0.1.0 -> 9.9.9"

	var buf bytes.Buffer
	code := report(&buf, errors.New("boom"), func() string { return banner })

	out := buf.String()
	require.Contains(t, out, "Error: boom")
	require.Contains(t, out, banner)
	assert.Less(t, strings.Index(out, "Error: boom"), strings.Index(out, banner),
		"a failed run must end on its own error, not on the upgrade nag")
	assert.Equal(t, exitError, code)
}

func TestReport_QuietSuppressesBanner(t *testing.T) {
	resetGlobals(t)
	quiet = true

	var buf bytes.Buffer
	report(&buf, nil, func() string { return "banner" })

	assert.NotContains(t, buf.String(), "banner", "--quiet must drop the upgrade nag")
}

func TestReport_SuccessPrintsNothingButBanner(t *testing.T) {
	resetGlobals(t)

	var buf bytes.Buffer
	code := report(&buf, nil, func() string { return "" })

	assert.Empty(t, buf.String())
	assert.Equal(t, exitOK, code)
}

func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}

func TestConfigOverrideChain_FlagOverridesEnv(t *testing.T) {
	resetGlobals(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	t.Setenv("KUBEADAPT_API_URL", "http://from-env.com")
	t.Setenv("KUBEADAPT_API_KEY", "env-key")

	cfgFile = cfgPath
	apiURL = "http://from-flag.com"
	apiKey = "flag-key"

	cmd := dummyCmd()
	require.NoError(t, rootCmd.PersistentPreRunE(cmd, nil))

	rc := getRunContext(cmd)
	require.NotNil(t, rc, "RunContext is nil after PersistentPreRunE")
	assert.Equal(t, "http://from-flag.com", rc.Config.APIURL, "flag wins")
	assert.Equal(t, "flag-key", rc.Config.APIKey, "flag wins")
}

func TestConfigOverrideChain_EnvOverridesConfig(t *testing.T) {
	resetGlobals(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	t.Setenv("KUBEADAPT_API_URL", "http://from-env.com")
	t.Setenv("KUBEADAPT_API_KEY", "env-key")

	cfgFile = cfgPath

	cmd := dummyCmd()
	require.NoError(t, rootCmd.PersistentPreRunE(cmd, nil))

	rc := getRunContext(cmd)
	require.NotNil(t, rc, "RunContext is nil after PersistentPreRunE")
	assert.Equal(t, "http://from-env.com", rc.Config.APIURL, "env wins")
	assert.Equal(t, "env-key", rc.Config.APIKey, "env wins")
}

func TestConfigOverrideChain_ConfigFallback(t *testing.T) {
	resetGlobals(t)

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeConfigFile(t, cfgPath, "api_url: http://from-config.com\napi_key: config-key\n")

	cfgFile = cfgPath

	cmd := dummyCmd()
	require.NoError(t, rootCmd.PersistentPreRunE(cmd, nil))

	rc := getRunContext(cmd)
	require.NotNil(t, rc, "RunContext is nil after PersistentPreRunE")
	assert.Equal(t, "http://from-config.com", rc.Config.APIURL, "config wins")
	assert.Equal(t, "config-key", rc.Config.APIKey, "config wins")
}

func TestConfigMissing_UsesDefault(t *testing.T) {
	resetGlobals(t)

	cfgFile = filepath.Join(t.TempDir(), "nonexistent.yaml")

	cmd := dummyCmd()
	require.NoError(t, rootCmd.PersistentPreRunE(cmd, nil))

	rc := getRunContext(cmd)
	require.NotNil(t, rc, "RunContext is nil after PersistentPreRunE")
	const wantURL = "https://public-api.kubeadapt.io"
	assert.Equal(t, wantURL, rc.Config.APIURL)
}
