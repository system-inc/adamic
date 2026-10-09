//go:build unix

package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The modules a run covers are go.work's uses, minus any inside a submodule, found by its gitlink in the
// index rather than by its name; a nested module's pattern is its module path, which reaches it through the
// workspace from the root.
func TestReadWorkspaceLeavesOutAModuleInsideASubmodule(t *testing.T) {
	t.Parallel()
	fixture := newLandFixture(t)
	library := t.TempDir()
	fixture.git(library, "init", "-q", "-b", "main")
	fixture.commit(library, "first.txt")
	files := map[string]string{
		"go.work":      "go 1.21\n\nuse (\n\t.\n\t./inner\n\t./library/mod\n)\n",
		"go.mod":       "module example.com/root\n\ngo 1.21\n",
		"inner/go.mod": "module example.com/root/inner\n\ngo 1.21\n",
	}
	for name, contents := range files {
		path := filepath.Join(fixture.repository, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fixture.git(fixture.repository, "add", ".")
	fixture.git(fixture.repository, "-c", "protocol.file.allow=always", "submodule", "add", "-q", library, "library")
	fixture.git(fixture.repository, "commit", "-q", "-m", "a workspace")

	covered, err := readWorkspace(fixture.repository)
	if err != nil {
		t.Fatal(err)
	}
	if got := covered.patterns(); !slices.Equal(got, []string{"./...", "example.com/root/inner/..."}) {
		t.Errorf("patterns are %q, want the root's and inner's", got)
	}
	if !slices.Equal(covered.inSubmodule, []string{"library/mod"}) {
		t.Errorf("left to a submodule: %q, want library/mod", covered.inSubmodule)
	}
}

// A go.mod go.work doesn't use is refused, by path and with the use line to add, rather than left ungated,
// which is how the shims went untested. One in testdata, in a dot directory or inside a submodule is not.
func TestReadWorkspaceRefusesAModuleGoWorkDoesNotUse(t *testing.T) {
	t.Parallel()
	fixture := newLandFixture(t)
	library := t.TempDir()
	fixture.git(library, "init", "-q", "-b", "main")
	for name, contents := range map[string]string{"go.mod": "module example.com/library\n\ngo 1.21\n"} {
		if err := os.WriteFile(filepath.Join(library, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fixture.git(library, "add", ".")
	fixture.git(library, "commit", "-q", "-m", "a module of the submodule's own")
	files := map[string]string{
		"go.work":                  "go 1.21\n\nuse .\n",
		"go.mod":                   "module example.com/root\n\ngo 1.21\n",
		"unlisted/go.mod":          "module example.com/root/unlisted\n\ngo 1.21\n",
		"testdata/fixture/go.mod":  "module example.com/fixture\n\ngo 1.21\n",
		".cache/downloaded/go.mod": "module example.com/downloaded\n\ngo 1.21\n",
	}
	for name, contents := range files {
		path := filepath.Join(fixture.repository, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fixture.git(fixture.repository, "-c", "protocol.file.allow=always", "submodule", "add", "-q", library, "library")

	_, err := readWorkspace(fixture.repository)
	if err == nil || !strings.Contains(err.Error(), "unlisted/go.mod: add `./unlisted` to go.work's use") {
		t.Fatalf("an unlisted module was not refused by path: %v", err)
	}
	for _, quiet := range []string{"testdata", ".cache", "library"} {
		if strings.Contains(err.Error(), quiet) {
			t.Errorf("the refusal names %s, which is not a module of the repository's: %v", quiet, err)
		}
	}

	if err := os.WriteFile(filepath.Join(fixture.repository, "go.work"), []byte("go 1.21\n\nuse (\n\t.\n\t./unlisted\n)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readWorkspace(fixture.repository); err != nil {
		t.Errorf("with the module used, reading the workspace still failed: %v", err)
	}
}
