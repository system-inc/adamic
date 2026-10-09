package configuration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersionRangeAdmitsExactlyWhatNpmWould pins each form against the versions on either side of its
// edges. Every row names a version that a nearby wrong reading would decide the other way.
func TestVersionRangeAdmitsExactlyWhatNpmWould(t *testing.T) {
	t.Parallel()
	cases := []struct {
		versionRange string
		admitted     []string
		refused      []string
	}{
		{"^1.0.0", []string{"1.0.0", "1.9.3", "1.0.1"}, []string{"0.9.9", "2.0.0", "2.0.0-beta.1", "1.1.0-beta.1"}},
		// Below 1.0.0 a caret holds the leftmost nonzero part, so it stops at the next minor or patch.
		{"^0.3.1", []string{"0.3.1", "0.3.9"}, []string{"0.4.0", "0.3.0", "1.0.0"}},
		{"^0.0.3", []string{"0.0.3"}, []string{"0.0.4", "0.1.0"}},
		{"~1.4.2", []string{"1.4.2", "1.4.9"}, []string{"1.4.1", "1.5.0"}},
		{">=1.2.0 <2.0.0", []string{"1.2.0", "1.99.0"}, []string{"1.1.9", "2.0.0"}},
		{">1.2.0 <=1.3.0", []string{"1.2.1", "1.3.0"}, []string{"1.2.0", "1.3.1"}},
		{"1.3.0 || 1.4.0", []string{"1.3.0", "1.4.0"}, []string{"1.3.1", "1.5.0"}},
		{"=1.3.0", []string{"1.3.0", "1.3.0+build.7"}, []string{"1.3.1"}},
		// A prerelease is admitted only where a comparator names its own major, minor and patch.
		{">=1.1.0-beta.0 <2.0.0", []string{"1.1.0-beta.1", "1.1.0", "1.5.0"}, []string{"1.2.0-beta.1", "1.1.0-alpha.9"}},
		{"1.0.0-rc.1", []string{"1.0.0-rc.1"}, []string{"1.0.0-rc.2", "1.0.0"}},
	}
	for _, testCase := range cases {
		parsed, err := ParseVersionRange(testCase.versionRange)
		if err != nil {
			t.Fatalf("%q: %v", testCase.versionRange, err)
		}
		for _, version := range testCase.admitted {
			if admitted, err := parsed.Admits(version); err != nil || !admitted {
				t.Errorf("%q should admit %s (got %v, %v)", testCase.versionRange, version, admitted, err)
			}
		}
		for _, version := range testCase.refused {
			if admitted, err := parsed.Admits(version); err != nil || admitted {
				t.Errorf("%q should refuse %s (got %v, %v)", testCase.versionRange, version, admitted, err)
			}
		}
	}
}

// TestVersionPrecedenceIsSemverOrder holds the comparison to semver.org's own example list, in both
// directions, so a comparison that only ever answered "less" or "equal" fails.
func TestVersionPrecedenceIsSemverOrder(t *testing.T) {
	t.Parallel()
	ordered := []string{"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0"}
	for index := 0; index+1 < len(ordered); index++ {
		lower, _ := parseSemanticVersion(ordered[index])
		higher, _ := parseSemanticVersion(ordered[index+1])
		if compareVersions(lower, higher) != -1 || compareVersions(higher, lower) != 1 {
			t.Errorf("%s should sort below %s", ordered[index], ordered[index+1])
		}
		if compareVersions(lower, lower) != 0 {
			t.Errorf("%s should equal itself", ordered[index])
		}
	}
}

func TestVersionRangeRefusesWhatItCannotReadExactly(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"", "  ", "1.x", "*", "1.0", "1.0.0 - 2.0.0", "01.0.0", "^1.0", "||", "^1.0.0 ||", "1.0.0-", "latest"} {
		if _, err := ParseVersionRange(text); err == nil {
			t.Errorf("%q should be refused", text)
		}
	}
}

func writeSettingsFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	directory := t.TempDir()
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(directory, "CohereSettings.json")
}

func TestOnlyTheProjectsOwnFilePinsCohere(t *testing.T) {
	t.Parallel()
	// The baseline: the project's own pin is read.
	loaded, err := Load(writeSettingsFiles(t, map[string]string{"CohereSettings.json": `{"cohere": "^1.0.0"}`}))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.CohereVersion == nil || loaded.CohereVersion.Text != "^1.0.0" {
		t.Fatalf("the project's pin was not read: %+v", loaded.CohereVersion)
	}

	unpinned, err := Load(writeSettingsFiles(t, map[string]string{"CohereSettings.json": `{"rules": {}}`}))
	if err != nil {
		t.Fatal(err)
	}
	if unpinned.CohereVersion != nil {
		t.Fatalf("a file with no pin read as pinned: %+v", unpinned.CohereVersion)
	}

	_, err = Load(writeSettingsFiles(t, map[string]string{
		"Base.json":           `{"cohere": "^1.0.0"}`,
		"CohereSettings.json": `{"extends": "./Base.json"}`,
	}))
	if err == nil || !strings.Contains(err.Error(), "only the project's own file may pin") {
		t.Fatalf("a pin in an extended file should be refused, got %v", err)
	}

	_, err = Load(writeSettingsFiles(t, map[string]string{"CohereSettings.json": `{"cohere": "1.x"}`}))
	if err == nil || !strings.Contains(err.Error(), "CohereSettings.json") {
		t.Fatalf("a malformed pin should be refused naming the file, got %v", err)
	}
}

// Not parallel: it swaps the package's releaseVersion function to stand in for published releases, which
// every test loading a config that pins cohere would see.
func TestAPublishedReleaseOutsideThePinRefusesNamingBothVersions(t *testing.T) {
	original := releaseVersion
	t.Cleanup(func() { releaseVersion = original })
	path := writeSettingsFiles(t, map[string]string{"CohereSettings.json": `{"cohere": "^1.0.0"}`})

	// The baseline: a release inside the range loads.
	releaseVersion = func() string { return "1.4.0" }
	if _, err := Load(path); err != nil {
		t.Fatalf("1.4.0 is inside ^1.0.0 and should load: %v", err)
	}

	releaseVersion = func() string { return "2.0.0" }
	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "^1.0.0") || !strings.Contains(err.Error(), "cohere 2.0.0") {
		t.Fatalf("2.0.0 is outside ^1.0.0 and should refuse naming both, got %v", err)
	}

	// A local build names no release, so the pin cannot judge it.
	releaseVersion = func() string { return "dev" }
	if _, err := Load(path); err != nil {
		t.Fatalf("a dev build should not be held to the pin: %v", err)
	}
}
