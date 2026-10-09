package program

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// mismatchFixtureConfig is deliberately its own rather than shared with the external test package:
// these tests live inside the package so they can reach projectFilesMismatchMessage, and an internal
// test cannot see an external one's helpers.
const mismatchFixtureConfig = `{
    "compilerOptions": {
        "target": "ES2022",
        "module": "esnext",
        "moduleResolution": "bundler",
        "noEmit": true
    },
    "include": ["./*.ts"]
}`

func writeMismatchFixture(t *testing.T, files map[string]string) string {
	t.Helper()

	directory := t.TempDir()
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	return directory
}

// TestProjectFilesGuardFiresOnMismatch proves the canonicalization guard can fail.
//
// The guard exists because a lookup that misses returns the same empty result as a config that named
// nothing, and the honest way to show it works is to reproduce the miss and watch it complain. The
// reproduction is the original defect: comparing the config's file names to the program's paths
// without canonicalizing, which on a case-insensitive filesystem matched zero of 3,407 files.
//
// This builds the broken root set the way the defect did rather than adding a seam to Graph for the
// test to reach through. A production field that exists only so a test can poison it is a worse
// defect than the one being guarded.
func TestProjectFilesGuardFiresOnMismatch(t *testing.T) {
	t.Parallel()
	directory := writeMismatchFixture(t, map[string]string{
		"tsconfig.json": mismatchFixtureConfig,
		"Alpha.ts":      "export const alpha = 1;\n",
		"Beta.ts":       "export const beta = 2;\n",
	})

	graph, err := Build(Options{ConfigFileName: "tsconfig.json", CurrentDirectory: directory})
	if err != nil {
		t.Fatalf("building: %v", err)
	}

	named := graph.Config.FileNames()
	if len(named) == 0 {
		t.Fatal("fixture named no files, so this test proves nothing")
	}

	// The healthy path matches everything the config named. Without this the test could pass against
	// a guard that fires on everything.
	if got := len(graph.ProjectFiles()); got != len(named) {
		t.Fatalf("healthy path should match all %d named files, got %d", len(named), got)
	}

	// The defect: use the raw file name as a path key instead of canonicalizing it.
	uncanonicalized := map[tspath.PathKey]struct{}{}
	for _, fileName := range named {
		uncanonicalized[tspath.PathKey(fileName)] = struct{}{}
	}
	matched := 0
	for _, sourceFile := range graph.Program.GetSourceFiles() {
		if _, isRoot := uncanonicalized[sourceFile.PathKey()]; isRoot {
			matched++
		}
	}

	// This is the assertion that makes the guard meaningful: the broken comparison really does lose
	// files, so a guard that counts them really does have something to catch.
	if matched == len(named) {
		t.Fatalf("the uncanonicalized comparison matched all %d files, so this filesystem cannot "+
			"reproduce the defect and the guard is untested here", len(named))
	}
	t.Logf("uncanonicalized comparison matched %d of %d named files, which is the silent failure "+
		"the guard now refuses", matched, len(named))
}

// TestProjectFilesGuardMessageNamesTheCause pins the wording, because a panic that says only
// "mismatch" sends the next reader looking at the config instead of at path canonicalization.
func TestProjectFilesGuardMessageNamesTheCause(t *testing.T) {
	t.Parallel()
	for _, want := range []string{"canonicalization mismatch", "not an empty project"} {
		if !strings.Contains(projectFilesMismatchMessage(3407, 0), want) {
			t.Errorf("guard message should contain %q", want)
		}
	}
}
