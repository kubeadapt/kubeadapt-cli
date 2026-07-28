// Values are injected via ldflags by GoReleaser; local dev builds fall back to
// "dev", "none", "unknown".
package version

var (
	Version = "dev"

	Commit = "none"

	Date = "unknown"
)
