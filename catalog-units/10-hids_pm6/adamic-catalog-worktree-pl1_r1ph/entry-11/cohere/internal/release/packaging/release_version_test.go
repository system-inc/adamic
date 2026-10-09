package release

import (
	"strings"
	"testing"
)

// TestReleaseVersionFloor holds both halves of the version rule: semver, and nothing below 1.0.0.
// The accepted cases matter as much as the refused ones, because a gate that refused everything
// would pass every refusal below.
func TestReleaseVersionFloor(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"1.0.0", "1.0.0-rc.1", "1.2.3", "12.0.0", "2.10.31-beta.2"} {
		if err := requireReleaseVersion(version); err != nil {
			t.Errorf("%s was refused: %v", version, err)
		}
	}

	refused := map[string]string{
		"0.3.1":       "below 1.0.0",
		"0.0.0":       "below 1.0.0",
		"0.9.9-rc.1":  "below 1.0.0",
		"1.0":         "not a semantic version",
		"v1.0.0":      "not a semantic version",
		"01.0.0":      "not a semantic version",
		"1.0.0+build": "not a semantic version",
		"1.0.0-":      "not a semantic version",
		" 1.0.0":      "not a semantic version",
	}
	for version, reason := range refused {
		err := requireReleaseVersion(version)
		if err == nil {
			t.Errorf("%q was accepted", version)
			continue
		}
		if !strings.Contains(err.Error(), reason) {
			t.Errorf("%q was refused for the wrong reason: %v", version, err)
		}
	}
}

// TestBuildRefusesAVersionBelowOne proves Build calls the check, which the table above cannot: it
// drives requireReleaseVersion directly and would pass with the call deleted from Build.
func TestBuildRefusesAVersionBelowOne(t *testing.T) {
	t.Parallel()

	_, err := Build(Options{Version: "0.3.1", ModuleDirectory: t.TempDir(), OutputDirectory: t.TempDir()})
	if err == nil {
		t.Fatal("Build accepted 0.3.1")
	}
	if !strings.Contains(err.Error(), "start at 1.0.0") {
		t.Fatalf("Build failed, but not on the version: %v", err)
	}
}
