package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetAuthFlags restores rootCmd's persistent flag set and the package-level
// flag vars to their declared defaults. Cobra retains parsed flag values on the
// shared rootCmd singleton, so without this a `--api-key` set by one test leaks
// into the next one and silently skips the interactive prompt path.
func resetAuthFlags(t *testing.T) {
	t.Helper()
	reset := func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		// Cobra only pushes the root context down when the subcommand's own ctx
		// is nil, so a context left behind by an earlier Execute would shadow
		// the one this test passes in.
		for _, c := range []*cobra.Command{rootCmd, authCmd, authLoginCmd, authLogoutCmd, authStatusCmd} {
			c.SetContext(nil)
		}
		rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		})
		cfgFile, apiURL, apiKey = "", "", ""
		outputFmt = formatTable
		noColor, verbose, quiet = false, false, false
	}
	reset()
	t.Cleanup(reset)
	// PersistentPreRunE lets KUBEADAPT_* env vars outrank the config file, so a
	// developer's shell would otherwise redirect these tests at a real endpoint.
	t.Setenv("KUBEADAPT_API_URL", "")
	t.Setenv("KUBEADAPT_API_KEY", "")
}

// runAuth executes the real command tree so flag parsing, Args validators, and
// PersistentPreRunE all run exactly as they do for a user.
func runAuth(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	rootCmd.SetArgs(args)
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	err = rootCmd.Execute()
	return out.String(), errOut.String(), err
}

func unauthorizedServer(t *testing.T) *httptest.Server {
	t.Helper()
	// api.IsUnauthorized dispatches on the envelope error *code*, not the HTTP
	// status, so the body must carry UNAUTHORIZED for the 401 branch to fire.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"invalid api key"}}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func orgServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":"org-42","metadata":{"name":"Acme"}},"meta":{"request_id":"req-abc"}}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func seedConfig(t *testing.T, apiURLValue, key string) (path, contents string) {
	t.Helper()
	path = filepath.Join(t.TempDir(), "config.yaml")
	contents = "version: 1\napi_url: " + apiURLValue + "\napi_key: " + key + "\n"
	require.NoError(t, os.WriteFile(path, []byte(contents), 0600))
	return path, contents
}

// captureOSStdout swaps the process-level os.Stdout so a test can prove that a
// command wrote nothing to the real stdout (as opposed to cmd.OutOrStdout()).
func captureOSStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	require.NoError(t, w.Close())
	os.Stdout = orig
	captured := <-done
	require.NoError(t, r.Close())
	return captured
}

// Defect 1: a rejected key must not clobber the key already on disk.
// Defect 2: the 401 message must name the key that was rejected and state that
// the on-disk config is untouched.
func TestAuthLogin_RejectedKeyPreservesConfigAndExplains(t *testing.T) {
	resetAuthFlags(t)
	srv := unauthorizedServer(t)
	cfgPath, before := seedConfig(t, srv.URL, "good-existing-key")

	_, _, err := runAuth(t, "auth", "login", "--config", cfgPath, "--api-key", "typod-key")
	require.Error(t, err, "a 401 must surface as an error")

	after, readErr := os.ReadFile(cfgPath)
	require.NoError(t, readErr)
	assert.Equal(t, before, string(after), "failed login must leave the config byte-identical")

	assert.Contains(t, err.Error(), "invalid API key")
	assert.Contains(t, err.Error(), "unchanged", "message must tell the user their existing config survived")
}

// Defect 3: success must print the path actually written, not DefaultPath().
func TestAuthLogin_PrintsResolvedConfigPath(t *testing.T) {
	resetAuthFlags(t)
	srv := orgServer(t)
	cfgPath, _ := seedConfig(t, srv.URL, "")

	stdout, _, err := runAuth(t, "auth", "login", "--config", cfgPath, "--api-key", "fresh-key")
	require.NoError(t, err)
	assert.Contains(t, stdout, cfgPath, "success message must name the file that was written")
}

// Defect 5: the prompt is UI, not data - it belongs on stderr so `kubeadapt
// auth login > key.txt` style pipes stay clean.
func TestAuthLogin_PromptGoesToStderrNotStdout(t *testing.T) {
	resetAuthFlags(t)
	srv := orgServer(t)
	cfgPath, _ := seedConfig(t, srv.URL, "")

	stdinR, stdinW, err := os.Pipe()
	require.NoError(t, err)
	origStdin := os.Stdin
	os.Stdin = stdinR
	t.Cleanup(func() {
		os.Stdin = origStdin
		_ = stdinR.Close()
	})
	_, _ = stdinW.WriteString("piped-key\n")
	require.NoError(t, stdinW.Close())

	var stderr string
	osStdout := captureOSStdout(t, func() {
		_, stderr, err = runAuth(t, "auth", "login", "--config", cfgPath)
	})
	require.NoError(t, err)

	assert.Contains(t, stderr, "Enter your Kubeadapt API key", "prompt must go to stderr")
	assert.NotContains(t, osStdout, "Enter your Kubeadapt API key", "prompt must not reach os.Stdout")
}

// deadServerURL returns the URL of a listener that has already been closed, so
// a request to it fails at the transport layer: neither success nor a 401.
func deadServerURL(t *testing.T) string {
	t.Helper()
	dead := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	u := dead.URL
	dead.Close()
	return u
}

// Defect 9: --quiet hid the fact that an unverified key had been written to
// disk. Quiet suppresses chatter; it must never suppress a caveat about state
// the command just persisted.
func TestAuthLogin_QuietStillWarnsAboutUnverifiedKey(t *testing.T) {
	resetAuthFlags(t)
	cfgPath, _ := seedConfig(t, deadServerURL(t), "")

	_, stderr, err := runAuth(t, "auth", "login", "--config", cfgPath, "--api-key", "unverifiable", "--quiet")
	require.NoError(t, err, "a transport failure must still persist the key")
	assert.Contains(t, stderr, "Warning",
		"--quiet must not hide that an unverified key was written to disk")
	assert.Contains(t, stderr, cfgPath, "the warning must name the file that was written")
}

// A config that exists but cannot be parsed or read is not an absent config.
// Falling back to Default() there overwrote a working key with a fresh file.
func TestAuthLogin_UnreadableConfigIsNotOverwritten(t *testing.T) {
	tests := []struct {
		name     string
		write    func(*testing.T, string)
		skipRoot bool
	}{
		{
			name: "corrupt yaml",
			write: func(t *testing.T, path string) {
				require.NoError(t, os.WriteFile(path, []byte("api_key: [unterminated\n"), 0600))
			},
		},
		{
			name:     "unreadable file",
			skipRoot: true,
			write: func(t *testing.T, path string) {
				require.NoError(t, os.WriteFile(path, []byte("api_key: good-existing-key\n"), 0600))
				require.NoError(t, os.Chmod(path, 0000))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipRoot && os.Geteuid() == 0 {
				t.Skip("root ignores file permissions")
			}
			resetAuthFlags(t)
			srv := orgServer(t)
			cfgPath := filepath.Join(t.TempDir(), "config.yaml")
			tt.write(t, cfgPath)
			before, readErr := os.ReadFile(cfgPath)

			_, _, err := runAuth(t, "auth", "login", "--config", cfgPath,
				"--api-url", srv.URL, "--api-key", "new-key")

			require.Error(t, err, "an unreadable config must not be silently replaced")
			if readErr == nil {
				after, err := os.ReadFile(cfgPath)
				require.NoError(t, err)
				assert.Equal(t, string(before), string(after), "the existing file must be byte-identical")
			}
		})
	}
}

// The key travels as a bearer token, so a cleartext endpoint leaks it on every
// request. Loopback is exempt because a locally-run API never leaves the host.
func TestAuthLogin_RejectsCleartextNonLoopbackURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "remote http", url: "http://api.example.com", wantErr: true},
		{name: "remote http with port", url: "http://10.0.0.5:8080", wantErr: true},
		{name: "no scheme", url: "api.example.com", wantErr: true},
		{name: "localhost", url: "http://localhost:8080"},
		{name: "ipv4 loopback", url: "http://127.0.0.1:8080"},
		{name: "ipv6 loopback", url: "http://[::1]:8080"},
		{name: "https remote", url: "https://api.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAuthFlags(t)
			cfgPath, before := seedConfig(t, "https://old.example.com", "good-existing-key")

			_, _, err := runAuth(t, "auth", "login", "--config", cfgPath,
				"--api-url", tt.url, "--api-key", "new-key")

			after, readErr := os.ReadFile(cfgPath)
			require.NoError(t, readErr)
			if !tt.wantErr {
				// Every allowed URL here is unreachable, so the run lands on the
				// transport path: saved with a warning, which is the point.
				assert.NotEqual(t, before, string(after), "an allowed URL must still reach the save path")
				return
			}
			require.Error(t, err, "%s must be rejected before any request is made", tt.url)
			assert.Equal(t, exitUsage, exitCodeFor(err), "a bad --api-url is a usage error")
			assert.Equal(t, before, string(after), "a rejected URL must leave the config untouched")
		})
	}
}

// A server that answered but did not authenticate the key is not the same as an
// unreachable one: it proves the key is unusable, so it must not be persisted.
func TestAuthLogin_ServerErrorDoesNotPersistKey(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "rate limited", status: http.StatusTooManyRequests, body: `{"error":{"code":"RATE_LIMITED","message":"slow down"}}`},
		{name: "server error", status: http.StatusInternalServerError, body: `{"error":{"code":"INTERNAL","message":"boom"}}`},
		{name: "forbidden", status: http.StatusForbidden, body: `{"error":{"code":"FORBIDDEN","message":"nope"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAuthFlags(t)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(srv.Close)
			cfgPath, before := seedConfig(t, srv.URL, "good-existing-key")

			_, _, err := runAuth(t, "auth", "login", "--config", cfgPath, "--api-key", "new-key")

			require.Error(t, err, "HTTP %d must not be reported as a successful login", tt.status)
			after, readErr := os.ReadFile(cfgPath)
			require.NoError(t, readErr)
			assert.Equal(t, before, string(after),
				"a key the server refused to accept must not replace a working one")
		})
	}
}

// os.Stdin bypasses cmd.SetIn, so an embedded caller or a table test could not
// supply the key at all.
func TestAuthLogin_ReadsKeyFromCommandInput(t *testing.T) {
	resetAuthFlags(t)
	srv := orgServer(t)
	cfgPath, _ := seedConfig(t, srv.URL, "")

	rootCmd.SetIn(strings.NewReader("key-from-cmd-in\n"))
	t.Cleanup(func() { rootCmd.SetIn(nil) })

	stdout, _, err := runAuth(t, "auth", "login", "--config", cfgPath)
	require.NoError(t, err)
	assert.Contains(t, stdout, cfgPath)

	saved, readErr := os.ReadFile(cfgPath)
	require.NoError(t, readErr)
	assert.Contains(t, string(saved), "key-from-cmd-in",
		"the key supplied through cmd.SetIn must be the one persisted")
}

// Defect 7: extra positional args are a usage error, not a silent success.
func TestAuth_RejectsExtraArgs(t *testing.T) {
	tests := []struct {
		sub     string
		seedKey string
		tail    []string
	}{
		{sub: "status", tail: []string{"extra", "junk", "args"}},
		{sub: "login", tail: []string{"--api-key", "k", "extra"}},
		{sub: "logout", seedKey: "some-key", tail: []string{"extra"}},
	}
	for _, tt := range tests {
		t.Run(tt.sub, func(t *testing.T) {
			resetAuthFlags(t)
			cfgPath, _ := seedConfig(t, "http://example.invalid", tt.seedKey)

			args := append([]string{"auth", tt.sub, "--config", cfgPath}, tt.tail...)
			_, _, err := runAuth(t, args...)
			require.Error(t, err, "auth %s must reject unexpected positional args", tt.sub)
		})
	}
}

// Defect 8: auth status must honour the global -o flag.
func TestAuthStatus_JSONOutput(t *testing.T) {
	resetAuthFlags(t)
	srv := orgServer(t)
	cfgPath, _ := seedConfig(t, srv.URL, "abcd12345678wxyz")

	stdout, _, err := runAuth(t, "auth", "status", "--config", cfgPath, "-o", "json")
	require.NoError(t, err)
	require.True(t, json.Valid([]byte(stdout)), "output is not valid JSON: %s", stdout)

	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, srv.URL, got["api_url"])
	assert.Equal(t, "abcd...wxyz", got["api_key_masked"])
	assert.Equal(t, "connected", got["status"])
	assert.Equal(t, "Acme", got["org_name"])
	assert.Equal(t, "org-42", got["org_id"])
	assert.Equal(t, "req-abc", got["request_id"])
}

func TestAuthStatus_YAMLOutput(t *testing.T) {
	resetAuthFlags(t)
	srv := orgServer(t)
	cfgPath, _ := seedConfig(t, srv.URL, "abcd12345678wxyz")

	stdout, _, err := runAuth(t, "auth", "status", "--config", cfgPath, "-o", "yaml")
	require.NoError(t, err)
	assert.Contains(t, stdout, "api_key_masked: abcd...wxyz")
	assert.Contains(t, stdout, "status: connected")
}

// Defect 10: `auth status` read the config file directly, so it reported the
// stored key while every other command used the env/flag override.
func TestAuthStatus_ReportsEffectiveCredential(t *testing.T) {
	tests := []struct {
		name       string
		env        string
		extraArgs  []string
		wantMasked string
	}{
		{name: "env var overrides config file", env: "envkey0000000envz", wantMasked: "envk...envz"},
		{name: "flag overrides config file", extraArgs: []string{"--api-key", "flagkey000000flgz"}, wantMasked: "flag...flgz"},
		{name: "config file when nothing overrides", wantMasked: "cfgk...cfgz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAuthFlags(t)
			srv := orgServer(t)
			cfgPath, _ := seedConfig(t, srv.URL, "cfgkey0000000cfgz")
			if tt.env != "" {
				t.Setenv("KUBEADAPT_API_KEY", tt.env)
			}

			args := append([]string{"auth", "status", "--config", cfgPath, "-o", "json"}, tt.extraArgs...)
			stdout, _, err := runAuth(t, args...)
			require.NoError(t, err)

			var got map[string]any
			require.NoError(t, json.Unmarshal([]byte(stdout), &got))
			assert.Equal(t, tt.wantMasked, got["api_key_masked"],
				"auth status must report the credential the CLI would actually send")
		})
	}
}

// Defect 11: `auth status` exited 0 even when the API rejected the key, so
// `kubeadapt auth status && deploy` sailed through unauthenticated.
func TestAuthStatus_ExitCodeReflectsUsability(t *testing.T) {
	tests := []struct {
		name       string
		url        func(*testing.T) string
		key        string
		wantErr    bool
		wantStdout string
	}{
		{
			name:       "connected",
			url:        func(t *testing.T) string { return orgServer(t).URL },
			key:        "abcd12345678wxyz",
			wantStdout: "Connected",
		},
		{
			name:       "unauthorized",
			url:        func(t *testing.T) string { return unauthorizedServer(t).URL },
			key:        "abcd12345678wxyz",
			wantErr:    true,
			wantStdout: "Unauthorized",
		},
		{
			name:       "transport error",
			url:        deadServerURL,
			key:        "abcd12345678wxyz",
			wantErr:    true,
			wantStdout: "Error:",
		},
		{
			name:       "no api key configured",
			url:        func(t *testing.T) string { return orgServer(t).URL },
			wantErr:    true,
			wantStdout: "(not set)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAuthFlags(t)
			cfgPath, _ := seedConfig(t, tt.url(t), tt.key)

			stdout, _, err := runAuth(t, "auth", "status", "--config", cfgPath)

			assert.Contains(t, stdout, tt.wantStdout,
				"the human-readable report must still be printed")
			if tt.wantErr {
				require.Error(t, err, "auth status must exit non-zero when the CLI cannot authenticate")
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestAuthStatus_UnauthorizedStillEmitsCompleteJSON(t *testing.T) {
	resetAuthFlags(t)
	srv := unauthorizedServer(t)
	cfgPath, _ := seedConfig(t, srv.URL, "abcd12345678wxyz")

	stdout, _, err := runAuth(t, "auth", "status", "--config", cfgPath, "-o", "json")

	require.Error(t, err, "a rejected key must not exit 0 in any output format")
	require.True(t, json.Valid([]byte(stdout)), "the JSON document must stay complete: %s", stdout)
	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, "unauthorized", got["status"])
	assert.Equal(t, "abcd...wxyz", got["api_key_masked"])
}

// A client that cannot even be constructed is the one failure that used to
// return before rendering, so the caller got an exit code and an empty report -
// exactly the case the diagnostic exists for.
func TestAuthStatus_ClientBuildFailureStillEmitsReport(t *testing.T) {
	resetAuthFlags(t)
	srv := orgServer(t)
	cfgPath, _ := seedConfig(t, srv.URL, "abcd12345678wxyz")

	// No RunContext: effectiveConfig falls back to the file and finds a key,
	// then newAPIClientFromCmd has nothing to build a client from.
	cfgFile = cfgPath
	outputFmt = formatJSON
	var stdout bytes.Buffer
	authStatusCmd.SetOut(&stdout)
	authStatusCmd.SetErr(&bytes.Buffer{})
	authStatusCmd.SetContext(t.Context())

	err := authStatusCmd.RunE(authStatusCmd, nil)

	require.Error(t, err, "an unusable client must still exit non-zero")
	require.True(t, json.Valid(stdout.Bytes()), "the report must still be emitted: %q", stdout.String())
	var got map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &got))
	assert.Equal(t, "error", got["status"])
	assert.Equal(t, "abcd...wxyz", got["api_key_masked"])
	assert.NotEmpty(t, got["error"], "the report must say why the client could not be built")
}

// Defect 12: any config.Load failure was reported as "No stored credentials
// found." with exit 0, while the key was still sitting on disk.
func TestAuthLogout_SurfacesUnreadableConfig(t *testing.T) {
	tests := []struct {
		name     string
		write    func(*testing.T, string)
		skipRoot bool
	}{
		{
			name: "corrupt yaml",
			write: func(t *testing.T, path string) {
				require.NoError(t, os.WriteFile(path, []byte("api_key: [unterminated\n"), 0600))
			},
		},
		{
			name:     "unreadable file",
			skipRoot: true,
			write: func(t *testing.T, path string) {
				require.NoError(t, os.WriteFile(path, []byte("api_key: k\n"), 0600))
				require.NoError(t, os.Chmod(path, 0000))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipRoot && os.Geteuid() == 0 {
				t.Skip("root bypasses file permissions")
			}
			resetAuthFlags(t)
			cfgPath := filepath.Join(t.TempDir(), "config.yaml")
			tt.write(t, cfgPath)

			stdout, _, err := runAuth(t, "auth", "logout", "--config", cfgPath)

			require.Error(t, err, "an unreadable config is not proof that the key is gone")
			assert.NotContains(t, stdout, "No stored credentials",
				"claiming there are no credentials while the key may still be on disk is the defect")
		})
	}
}

// Defect 6: logout must write through the cobra writer so tests and pipes can
// capture it.
func TestAuthLogout_UsesCmdWriter(t *testing.T) {
	tests := []struct {
		name    string
		seeded  bool
		wantMsg string
	}{
		{name: "with stored credentials", seeded: true, wantMsg: "Logged out"},
		{name: "config missing", seeded: false, wantMsg: "No stored credentials"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAuthFlags(t)
			cfgPath := filepath.Join(t.TempDir(), "nope.yaml")
			if tt.seeded {
				cfgPath, _ = seedConfig(t, "http://example.invalid", "some-key")
			}

			var stdout string
			osStdout := captureOSStdout(t, func() {
				var err error
				stdout, _, err = runAuth(t, "auth", "logout", "--config", cfgPath)
				require.NoError(t, err)
			})

			assert.Contains(t, stdout, tt.wantMsg, "logout must write to cmd.OutOrStdout()")
			assert.NotContains(t, osStdout, tt.wantMsg, "logout must not bypass the cobra writer")
		})
	}
}

// Defect 4: the verification call must observe cancellation. Against a healthy
// server a context.Background() call always succeeds and reports "Authenticated";
// only a request wired to cmd.Context() can fail on an already-cancelled context.
func TestAuthLogin_HonoursContextCancellation(t *testing.T) {
	resetAuthFlags(t)
	srv := orgServer(t)
	cfgPath, _ := seedConfig(t, srv.URL, "")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rootCmd.SetArgs([]string{"auth", "login", "--config", cfgPath, "--api-key", "whatever"})
	var out, errOut bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)

	require.NoError(t, rootCmd.ExecuteContext(ctx))
	assert.NotContains(t, out.String(), "Authenticated", "a cancelled context must abort verification")
	assert.Contains(t, errOut.String(), "context canceled")
}
