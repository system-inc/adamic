package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/scout_nullish_keys.a", "stage3/map-keys/nullish_set.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestScoutNullishKeyMutants(t *testing.T) {
	t.Parallel()
	for _, rule := range []struct{ name, body string }{
		{"null_undefined", `if(map->union_keys && key.reference==&adamic_null) key.reference=NULL;`},
		{"typeof_undefined", ""},
		{"undefined_string", `if(map->union_keys && key.reference==NULL) key.reference=adamic_retain(&adamic_string_undefined);`},
	} {
		t.Run(rule.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_nullish_keys.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			code := strings.ReplaceAll(original, "adamic_map_set(", "scout_nullish_store(")
			if code == original {
				t.Fatal("no insertion changed")
			}
			if rule.name == "typeof_undefined" {
				code = strings.ReplaceAll(original, "adamic_union_typeof(", "scout_nullish_typeof(")
				if code == original {
					t.Fatal("no typeof changed")
				}
				code = insertCollectionMutant(code, `static adamic_string *scout_nullish_typeof(const adamic_heap *value, bool null){(void)null;return adamic_union_typeof(value,true);}`)
			} else {
				code = insertCollectionMutant(code, `static void scout_nullish_store(adamic_map *map, adamic_value key, adamic_value value){`+rule.body+`adamic_map_set(map,key,value);}`)
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("unclean mutant: %d %s", actual.exitCode, actual.stderr)
			}
			if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
				t.Fatalf("want stdout disagreement: %q", difference)
			}
		})
	}
}

func TestScoutNullableKeyConversionMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_nullish_keys.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "nullable_union_key" {
			continue
		}
		returned := function.Body[0].(ir.Return)
		selection := returned.Value.(ir.Conditional)
		selection.WhenTrue = selection.WhenNot
		returned.Value = selection
		function.Body[0] = returned
		changed++
	}
	if changed == 0 {
		t.Fatal("no nullable conversion mutated")
	}
	want := onNode(t, path)
	actual, binary := nativelyUncached(t, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("unclean mutant: %d %s", actual.exitCode, actual.stderr)
	}
	if difference := disagreement(want, actual); difference != "stdout differs" {
		t.Fatalf("native: %q", difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatalf("JavaScript: %q", difference)
	}
}

func TestScoutNullishConditionalMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_nullish_keys.a"))
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
		if function.Name != "chooseNullish" {
			continue
		}
		returned := function.Body[0].(ir.Return)
		selection := returned.Value.(ir.Conditional)
		selection.WhenTrue = selection.WhenNot
		returned.Value = selection
		function.Body[0] = returned
		changed = true
	}
	if !changed {
		t.Fatal("no conditional changed")
	}
	want := onNode(t, path)
	actual, binary := nativelyUncached(t, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("unclean mutant: %d %s", actual.exitCode, actual.stderr)
	}
	if difference := disagreement(want, actual); difference != "stdout differs" {
		t.Fatalf("native: %q", difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript: %q", difference)
	}
}
