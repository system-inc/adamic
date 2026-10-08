package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Each mutant still compiles with -Werror, runs without a sanitizer finding, and exits successfully.
// Only comparison with the program's source on Node kills it. Counts are deliberately not involved.
func TestLibraryMapSet2Mutants(t *testing.T) {
	families := []string{"union", "intersection", "difference", "symmetricDifference", "isSubsetOf", "isSupersetOf", "isDisjointFrom", "forEach"}
	for _, family := range families {
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			fixture := "library_map_set_2_map_operands.a"
			if family == "forEach" {
				fixture = "library_map_set_2_foreach_never.a"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.Name == "set_"+family {
					if family == "isDisjointFrom" {
						function.Body[len(function.Body)-1] = ir.Return{Value: ir.BooleanConstant{Value: false}}
					} else {
						function.Parameters[0], function.Parameters[1] = function.Parameters[1], function.Parameters[0]
					}
					changed = true
				}
			}

			code := native.C(program)
			if family == "forEach" {
				changed = strings.Contains(code, "adamic_map_iterator_next(")
				code = strings.ReplaceAll(code, "adamic_map_iterator_next(", "library_mutant_skip_next(")
				code = insertCollectionMutant(code, `static bool library_mutant_skip_next(adamic_map_iterator *iterator, adamic_value *key, adamic_value *value) {
    (void)iterator; (void)key; (void)value;
    return false;
}`)
			}

			if !changed {
				t.Fatal("the mutant changed no code")
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must run cleanly so only Node catches it: exit %d, stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("mutant was not caught by stdout comparison: %q", difference)
			}
			t.Logf("%s: clean exit 0, no sanitizer finding, Node stdout comparison caught it", family)
		})
	}
}

func init() {
	for _, name := range []string{"map_operands", "foreach_never"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/library_map_set_2_" + name + ".a", lowers: true})
	}
}
