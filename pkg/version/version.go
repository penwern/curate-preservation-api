// Package version provides information about the current build of the Curate Preservation API.
// It retrieves the version, commit hash, and build time from the Go build information.
// It is designed to be used in the command line interface to display version information.
// It is also used to provide version metadata for PREMIS Agent.
package version

import (
	"runtime/debug"
)

// Version returns the module version recorded by the Go linker.
// For a tagged build this is the tag (e.g. v1.0.2).
// For an un-tagged build it is the pseudo-version
// (e.g. v1.0.2-0.20250605-6d1e8239a3m).
func Version() string {
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "devel" // fallback for 'go run .' during local dev
}

// Commit returns the 12-char Git hash or "unknown".
func Commit() string { return buildSetting("vcs.revision") }

// BuildTime returns the commit time in RFC3339 or "unknown".
func BuildTime() string { return buildSetting("vcs.time") }

func buildSetting(key string) string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == key {
				return s.Value
			}
		}
	}
	return "unknown"
}
