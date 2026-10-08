package lint

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

var compilerCorpusFixtureFolders = []string{"testdata", "validation", "evidence"}

// Agreement covers tracked production sources. Fixture and proof directories
// contain deliberately invalid inputs owned by their separate parser tests.
func compilerStage1Sources(t *testing.T, checkout string) []string {
	t.Helper()
	files := corpusfiles.Repository(t, checkout, []string{"stage1"}, []string{"*.ts", "*.a"})
	root, err := filepath.Abs(checkout)
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	excluded := 0
	for _, file := range files {
		relative, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		fixture := false
		for _, directory := range strings.Split(filepath.ToSlash(filepath.Dir(relative)), "/") {
			if slices.Contains(compilerCorpusFixtureFolders, directory) {
				fixture = true
			}
		}
		if fixture {
			excluded++
			continue
		}
		sources = append(sources, file)
	}
	t.Logf("compiler corpus stage1: excluded %d tracked files under testdata/, validation/, evidence/; retained %d sources", excluded, len(sources))
	return sources
}

// Exact directory names are the entire exclusion contract. Sources next to
// named folders, similar folder names, and files with those names stay covered.
// An added folder exclusion fails this test instead of silently shrinking coverage.
func TestCompilerCorpusSourcesExcludeOnlyNamedFolders(t *testing.T) {
	t.Parallel()
	folders := append([]string(nil), compilerCorpusFixtureFolders...)
	sort.Strings(folders)
	if !reflect.DeepEqual(folders, []string{"evidence", "testdata", "validation"}) {
		t.Fatalf("compiler corpus exclusions grew beyond the named folders: %q", folders)
	}
	checkout := t.TempDir()
	retained := []string{
		"stage1/source.ts", "stage1/lib/types.a",
		"stage1/testdata.ts", "stage1/validation.a", "stage1/evidence.ts",
		"stage1/testdatabase/source.ts", "stage1/validation-tools/source.a",
		"stage1/evidence-based/source.ts", "stage1/tests/source.ts",
		"stage1/fixtures/source.a", "stage1/lib/data/source.ts",
	}
	excluded := []string{
		"stage1/testdata/invalid.ts", "stage1/lib/testdata/nested/invalid.a",
		"stage1/validation/invalid.ts", "stage1/lib/validation/nested/invalid.a",
		"stage1/evidence/invalid.ts", "stage1/lib/evidence/nested/invalid.a",
		"stage1/testdata/evidence/invalid.ts",
	}
	for selected, group := range [][]string{retained, excluded} {
		for _, name := range group {
			file := filepath.Join(checkout, name)
			if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
				t.Fatal(err)
			}
			// No oracle should ever parse the deliberately invalid fixture content.
			content := "export const value = 1;\n"
			if selected == 1 {
				content = "new Promise(@dec async () => {})\n"
			}
			if err := os.WriteFile(file, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	execute(t, checkout, "git", "init", "-q", "-b", "main")
	execute(t, checkout, "git", "config", "user.name", "Corpus contract")
	execute(t, checkout, "git", "config", "user.email", "corpus@example.invalid")
	execute(t, checkout, "git", "add", "-f", "stage1")
	execute(t, checkout, "git", "commit", "-qm", "Corpus contract fixture")
	got := compilerStage1Sources(t, checkout)
	want := make([]string, len(retained))
	for i, file := range retained {
		want[i] = filepath.Join(checkout, file)
	}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("exclusion exceeded named directory contract or retained a fixture:\ngot %q\nwant %q", got, want)
	}
	t.Logf("all %d sources retained; only %d named-folder fixtures excluded", len(retained), len(excluded))
}
