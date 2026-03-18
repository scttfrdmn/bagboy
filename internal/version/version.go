// Package version holds build-time version information injected via ldflags.
package version

// Version variables are set at build time via ldflags.
var (
	// Version is the semantic version of the binary.
	Version = "0.8.0-dev"
	// Commit is the short git commit hash at build time.
	Commit = "none"
	// Date is the UTC build timestamp in RFC3339 format.
	Date = "unknown"
	// BuiltBy identifies who or what built the binary (e.g., "goreleaser", "source").
	BuiltBy = "source"
)

// UserAgent returns the HTTP User-Agent string used by bagboy clients.
func UserAgent() string { return "bagboy/" + Version }
