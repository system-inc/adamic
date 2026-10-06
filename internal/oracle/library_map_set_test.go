package oracle

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Each mutant still compiles with -Werror, runs without a sanitizer finding, and exits successfully.
// Only comparison with the program's source on Node kills it. Counts are deliberately not involved.
func TestLibraryMapSetMutants(t *testing.T) {
	families := []string{"union", "intersection", "difference", "symmetricDifference", "isSubsetOf", "isSupersetOf", "isDisjointFrom", "keys", "iterators", "constructors", "forEach", "size", "groupBy"}
	for _, family := range families {
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			fixture := "library_map_set.a"
			switch family {
			case "keys", "size":
				fixture = "library_map_set_keys.a"
			case "iterators":
				fixture = "library_map_set_iterators.a"
			case "constructors", "forEach":
				fixture = "library_map_set_construct.a"
			case "groupBy":
				fixture = "library_map_set_group_by.a"
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
				if family == "groupBy" && function.Name == "map_groupBy" {
					declaration := function.Body[1].(ir.Declare)
					declaration.Value = ir.NumberConstant{Value: 1}
					function.Body[1] = declaration
					changed = true
				}
				if family == "constructors" && function.Name == "collection_iterator_array" {
					parameter := ir.Read{Local: function.Parameters[0], Of: ir.Object}
					step := ir.Evaluate{Value: ir.CallClosure{Closure: ir.Property{Object: parameter, Name: "next", Of: ir.Closure}, Returns: ir.Object}}
					function.Body = append([]ir.Statement{step}, function.Body...)
					changed = true
				}
			}
			code := native.C(program)
			switch family {
			case "keys":
				original := code
				for _, references := range []string{"true", "false"} {
					code = strings.ReplaceAll(code, "adamic_map_new_maybe_numbers("+references+")", "adamic_map_new(false, "+references+")")
				}
				changed = code != original
			case "size":
				original := code
				code = strings.ReplaceAll(code, "->count", "->used")
				changed = code != original
			case "iterators":
				changed = strings.Contains(code, "adamic_collection_iterator(")
				code = strings.ReplaceAll(code, "adamic_collection_iterator(", "library_mutant_iterator(")
				code = insertCollectionMutant(code, `static adamic_object *library_mutant_iterator(adamic_map *map, int part, int key, int value, bool set) {
	adamic_object *object = adamic_collection_iterator(map, part, key, value, set);
	adamic_closure *next = object->slots[0].reference;
	adamic_object *state = next->cells[0]->value.reference;
	adamic_map_iterator *iterator = state->slots[0].reference;
	iterator->next = 1;
	return object;
}`)
			case "forEach":
				changed = strings.Contains(code, "adamic_map_iterate(")
				code = strings.ReplaceAll(code, "adamic_map_iterate(", "library_mutant_iterate(")
				code = strings.ReplaceAll(code, "adamic_map_iterator_next(", "library_mutant_next(")
				code = insertCollectionMutant(code, `static size_t library_mutant_limit;
static adamic_map_iterator *library_mutant_iterate(adamic_map *map) {
	library_mutant_limit = map->used;
	return adamic_map_iterate(map);
}
static bool library_mutant_next(adamic_map_iterator *iterator, adamic_value *key, adamic_value *value) {
	return iterator->next < library_mutant_limit && adamic_map_iterator_next(iterator, key, value);
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

func insertCollectionMutant(code, helper string) string {
	include := "#include \"adamic.h\""
	if !strings.Contains(code, include) {
		panic(fmt.Sprintf("mutant: runtime include missing from %q", code[:80]))
	}
	return strings.Replace(code, include, include+"\n"+helper, 1)
}
