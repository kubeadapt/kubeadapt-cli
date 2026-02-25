// Package version holds build-time version metadata injected via ldflags.
// GoReleaser injects these values automatically during the release build.
// For local dev builds (task build), values default to "dev", "none", "unknown".
package version

var (
	// Version is the semantic version tag (e.g. "v1.2.3").
	Version = "dev"

	// Commit is the short git commit SHA at build time.
	Commit = "none"

	// Date is the UTC build timestamp in ISO 8601 format.
	Date = "unknown"
)
