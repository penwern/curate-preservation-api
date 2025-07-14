package version

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	version := Version()

	// Version should not be empty
	if version == "" {
		t.Error("Version() should not return empty string")
	}

	// For testing environment, version is likely to be "devel"
	// In a real build it would be a proper version string
	if version != "devel" && !strings.HasPrefix(version, "v") {
		t.Logf("Version: %s (this may be expected in test environment)", version)
	}
}

func TestCommit(t *testing.T) {
	commit := Commit()

	// Commit should not be empty
	if commit == "" {
		t.Error("Commit() should not return empty string")
	}

	// Should return either a commit hash or "unknown"
	if commit != "unknown" {
		// If it's not "unknown", it should look like a git hash
		// Git hashes are hexadecimal and typically 7-40 characters
		if len(commit) < 7 || len(commit) > 40 {
			t.Errorf("Commit hash '%s' has unexpected length", commit)
		}

		// Check if it contains only valid hex characters
		for _, char := range commit {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				t.Errorf("Commit hash '%s' contains non-hex character: %c", commit, char)
			}
		}
	}
}

func TestBuildTime(t *testing.T) {
	buildTime := BuildTime()

	// BuildTime should not be empty
	if buildTime == "" {
		t.Error("BuildTime() should not return empty string")
	}

	// Should return either an RFC3339 timestamp or "unknown"
	if buildTime != "unknown" {
		// Basic check for RFC3339 format (should contain 'T' and end with 'Z' or timezone)
		if !strings.Contains(buildTime, "T") {
			t.Errorf("BuildTime '%s' does not appear to be in RFC3339 format", buildTime)
		}
	}
}

func TestBuildSetting_ValidKey(t *testing.T) {
	// Test buildSetting with a key that might exist
	revision := buildSetting("vcs.revision")
	if revision == "" {
		t.Error("buildSetting should not return empty string")
	}

	// Should return either a value or "unknown"
	if revision != "unknown" {
		t.Logf("VCS revision: %s", revision)
	}
}

func TestBuildSetting_InvalidKey(t *testing.T) {
	// Test buildSetting with a key that doesn't exist
	result := buildSetting("nonexistent.key")
	
	if result != "unknown" {
		t.Errorf("Expected 'unknown' for nonexistent key, got '%s'", result)
	}
}

func TestBuildSetting_EmptyKey(t *testing.T) {
	// Test buildSetting with empty key
	result := buildSetting("")
	
	if result != "unknown" {
		t.Errorf("Expected 'unknown' for empty key, got '%s'", result)
	}
}

func TestVersionConsistency(t *testing.T) {
	// Call Version() multiple times and ensure it returns the same value
	version1 := Version()
	version2 := Version()
	
	if version1 != version2 {
		t.Errorf("Version() returned different values: '%s' vs '%s'", version1, version2)
	}
}

func TestCommitConsistency(t *testing.T) {
	// Call Commit() multiple times and ensure it returns the same value
	commit1 := Commit()
	commit2 := Commit()
	
	if commit1 != commit2 {
		t.Errorf("Commit() returned different values: '%s' vs '%s'", commit1, commit2)
	}
}

func TestBuildTimeConsistency(t *testing.T) {
	// Call BuildTime() multiple times and ensure it returns the same value
	buildTime1 := BuildTime()
	buildTime2 := BuildTime()
	
	if buildTime1 != buildTime2 {
		t.Errorf("BuildTime() returned different values: '%s' vs '%s'", buildTime1, buildTime2)
	}
}

func TestVersionInfo_Integration(t *testing.T) {
	// Test that all version functions work together without panicking
	version := Version()
	commit := Commit()
	buildTime := BuildTime()

	t.Logf("Version: %s", version)
	t.Logf("Commit: %s", commit)
	t.Logf("BuildTime: %s", buildTime)

	// All should return non-empty strings
	if version == "" || commit == "" || buildTime == "" {
		t.Error("All version functions should return non-empty strings")
	}
}

func TestBuildInfoAvailability(t *testing.T) {
	// Test if build info is available
	_, ok := debug.ReadBuildInfo()
	if !ok {
		t.Log("Build info not available (this is expected in some test environments)")
	}

	// The functions should still work even if build info is not available
	version := Version()
	commit := Commit()
	buildTime := BuildTime()

	if version == "" || commit == "" || buildTime == "" {
		t.Error("Version functions should handle missing build info gracefully")
	}
}

func TestVersionNotDevelInBuild(t *testing.T) {
	// This test documents the expected behavior
	// In development, version should be "devel"
	// In a proper build, it should be a version tag
	version := Version()
	
	if version == "devel" {
		t.Log("Running in development mode (version is 'devel')")
	} else if strings.HasPrefix(version, "v") {
		t.Logf("Running with tagged version: %s", version)
	} else if version == "(devel)" {
		t.Log("Running with Go's default devel version")
	} else {
		t.Logf("Running with version: %s", version)
	}
}

func TestBuildSettingEdgeCases(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{"vcs.revision", "vcs.revision"},
		{"vcs.time", "vcs.time"},
		{"vcs.modified", "vcs.modified"},
		{"GOOS", "GOOS"},
		{"GOARCH", "GOARCH"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := buildSetting(tc.key)
			// Should not panic and should return a string
			if result == "" {
				t.Errorf("buildSetting('%s') returned empty string", tc.key)
			}
			t.Logf("%s: %s", tc.key, result)
		})
	}
}