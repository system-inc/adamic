package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

var scout19Fixtures = []string{"map_entries", "multimap", "identity", "numeric_ids", "mutation", "set_intersection"}

func init() {
	for _, name := range scout19Fixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/scout19_" + name + ".a", lowers: true})
	}
}

// Mutants preserve valid C, ownership, and successful execution. Source Node alone
// decides whether the tsc excerpt's collection behavior changed.
func TestScout19TSCCollectionMutants(t *testing.T) {
	t.Parallel()
	for _, name := range scout19Fixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout19_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			before := code
			helper := ""
			switch name {
			case "map_entries":
				code = strings.ReplaceAll(code, "adamic_map_set(", "scout19_set(")
				helper = `static void scout19_set(adamic_map *map, adamic_value key, adamic_value value) {
    value.number += 1;
    adamic_map_set(map, key, value);
}`
			case "multimap":
				code = strings.ReplaceAll(code, "adamic_array_push(", "scout19_push(")
				helper = `static void scout19_push(adamic_array *array, adamic_value value) {
    value.number += 1;
    adamic_array_push(array, value);
}`
			case "identity":
				for _, line := range strings.Split(code, "\n") {
					if strings.Contains(line, "bool ") && strings.Contains(line, "adamic_map_get(") && strings.Contains(line, " != NULL)") {
						code = strings.Replace(code, line, strings.Replace(line, " != NULL)", " == NULL)", 1), 1)
						break
					}
				}
			case "numeric_ids":
				code = strings.ReplaceAll(code, "adamic_map_set(", "scout19_first(")
				helper = `static void scout19_first(adamic_map *map, adamic_value key, adamic_value value) {
    if (adamic_map_get(map, key) != NULL) {
        adamic_release(value.reference);
        return;
    }
    adamic_map_set(map, key, value);
}`
			case "mutation", "set_intersection":
				code = strings.ReplaceAll(code, "adamic_map_delete(", "scout19_delete(")
				helper = `static bool scout19_delete(adamic_map *map, adamic_value key) {
    (void)map;
    (void)key;
    return false;
}`
			}
			if code == before {
				t.Fatal("mutant changed no code")
			}
			if helper != "" {
				code = insertCollectionMutant(code, helper)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant must exit 0 and be sanitizer-clean: %d %s", actual.exitCode, actual.stderr)
			}
			if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
				t.Fatalf("Node must catch the mutant's output: %q", difference)
			}
			t.Log("clean exit 0; ASan/UBSan/LSan clean; Node stdout catches mutant")
		})
	}
}
