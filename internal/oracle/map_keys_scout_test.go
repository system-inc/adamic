package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"scout_map_objects.a", "scout_map_numbers.a", "scout_map_references.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + path, true, false})
	}
}

// A lost lookup compiles and runs cleanly; source Node alone detects the changed result.
func TestScoutMapLookupMutants(t *testing.T) {
	for _, fixture := range []string{"scout_map_objects.a", "scout_map_numbers.a", "scout_map_references.a"} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			code := strings.ReplaceAll(original, "adamic_map_get(", "scout_missing_lookup(")
			if code == original {
				t.Fatal("mutant changed no lookup")
			}
			code = insertCollectionMutant(code, `static adamic_value *scout_missing_lookup(const adamic_map *map, adamic_value key) {
    (void)map;
    (void)key;
    return NULL;
}`)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must run cleanly: exit %d, stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want Node to catch lost lookup, got %q", difference)
			}
			t.Log("lost lookup: clean exit 0, no sanitizer or leak finding, caught by Node stdout")
		})
	}
}
