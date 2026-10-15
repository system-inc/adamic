package lint

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

var compilerCorpusFixtureFolders = []string{"testdata", "validation", "evidence"}

// Agreement covers tracked production sources. Fixture and proof directories
// contain deliberately invalid inputs owned by their separate parser tests. The
// walk never enters one (corpusfiles.RepositoryExcept): it listed every rule's
// evidence/ before dropping the files, so a doc-only change there moved this
// corpus's read sets and the products that read it (#hp8gq3e).
func compilerStage1Sources(t *testing.T, checkout string) []string {
	t.Helper()
	sources := corpusfiles.RepositoryExcept(t, checkout, []string{"stage1"}, []string{"*.ts", "*.a"}, compilerCorpusFixtureFolders)
	t.Logf("compiler corpus stage1: %d sources, none under testdata/, validation/ or evidence/", len(sources))
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

// The registry must be in the stage1 sources exactly once: missing loses its coverage, and twice is every one of its
// file/rule pairs a duplicate (landable-17).
func TestCompilerCorpusListsTheRegistryOnce(t *testing.T) {
	t.Parallel()
	registry := "/tree/stage1/cohere/lint/.generated/registry.ts"
	other := "/tree/stage1/cohere/lint/rules.ts"
	for _, check := range []struct {
		sources []string
		ok      bool
	}{
		{[]string{other}, false},
		{[]string{other, registry}, true},
		{[]string{registry, other, registry}, false},
	} {
		if err := compilerRegistryListedOnce(check.sources, registry); (err == nil) != check.ok {
			t.Errorf("%q: %v, want ok %t", check.sources, err, check.ok)
		}
	}
}
