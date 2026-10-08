package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var stringSortFixtures = []string{"library_array_sort_strings.a", "library_array_sort_strings_holes.a", "library_array_sort_strings_stability.a"}

func init() {
	for _, name := range stringSortFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

func TestArrayStringSort(t *testing.T) {
	for _, name := range stringSortFixtures {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("Node: %+v", truth)
			}
			got, binary := natively(t, program)
			for _, actual := range []run{got, released(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, actual); difference != "" {
					t.Fatalf("%s: Node %q; got %q stderr %q", difference, truth.stdout, actual.stdout, actual.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if difference := disagreement(truth, onWASI(t, native.C(program))); difference != "" {
					t.Fatal(difference)
				}
			}
			t.Log("Node, native sanitized/release, JavaScript and enabled WASI agree")
		})
	}
}

// These mutations remain valid, clean C programs. Only Node's observation rejects
// their results; no allocation assertion or built-in expected-output table does.
func TestArrayStringSortMutants(t *testing.T) {
	for _, mutation := range []struct{ name, fixture, symbol, helper string }{
		{"code-point", stringSortFixtures[0], "adamic_string_compare", `static int mutant_compare(const adamic_string *left, const adamic_string *right) {
    size_t i = 0, j = 0;
    size_t nl = (size_t)adamic_string_length(left), nr = (size_t)adamic_string_length(right);
    while (i < nl && j < nr) {
        double a = adamic_string_code_point_at(left, (double)i).number;
        double b = adamic_string_code_point_at(right, (double)j).number;
        if (a != b) return a < b ? -1 : 1;
        i += a > 65535 ? 2 : 1; j += b > 65535 ? 2 : 1;
    }
    return i < nl ? 1 : j < nr ? -1 : 0;
}`},
		{"descending", stringSortFixtures[0], "adamic_string_compare", `static int mutant_compare(const adamic_string *left, const adamic_string *right) {
    return -adamic_string_compare(left, right);
}`},
		{"unstable", stringSortFixtures[2], "adamic_array_sort", `static void mutant_compare(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context) {
    adamic_array_sort(array, compare, context);
    if (adamic_thrown != NULL) return;
    for (size_t start = 0; start < array->length;) {
        size_t end = start + 1;
        while (end < array->length && compare(array->elements[start], array->elements[end], context) == 0) end++;
        for (size_t i = start, j = end - 1; i < j; i++, j--) {
            adamic_value value = array->elements[i]; array->elements[i] = array->elements[j]; array->elements[j] = value;
        }
        start = end;
    }
}`},
		{"undefined-as-holes", stringSortFixtures[1], "adamic_array_sort_strings", `static void mutant_compare(adamic_array *array, int (*compare)(adamic_value, adamic_value, void *), void *context) {
    adamic_array_sort_strings(array, compare, context);
    if (array->sparse == NULL) return;
    adamic_array *keys = adamic_map_keys(array->sparse);
    for (size_t i = 0; i < keys->length; i++) {
        adamic_value key = keys->elements[i];
        adamic_value *slot = adamic_map_get(array->sparse, key);
        if (slot->reference == NULL) {
            adamic_map_delete(array->sparse, key);
            array->capacity++;
        }
    }
    adamic_release(keys);
}`},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", mutation.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			code := native.C(program)
			if !strings.Contains(code, mutation.symbol+"(") {
				t.Fatal("mutation site missing")
			}
			code = strings.ReplaceAll(code, mutation.symbol+"(", "mutant_compare(")
			code = strings.Replace(code, "#include \"adamic.h\"", "#include \"adamic.h\"\n"+mutation.helper, 1)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := execute(t, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant did not run cleanly: %+v", actual)
			}
			if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
				t.Fatalf("Node must catch stdout, got %q", difference)
			}
			t.Log("compiled, exited 0, sanitizers clean; caught only by Node: stdout differs")
		})
	}
}

func TestArrayStringSortRefusals(t *testing.T) {
	for _, elements := range []string{"[10, 9, 1]", "[true, false]", "[{x: 1}, {x: 0}]", "[1, 'a']", "[undefined, 1]"} {
		t.Run(elements, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "refused.a")
			if err := os.WriteFile(path, []byte("const values = "+elements+"; values.sort();\n"), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			// This library-only area predates boxed mixed array elements. Its
			// earlier representation refusal is retained until that compiler
			// feature arrives; once it does, the default-sort refusal below applies.
			var notYet *lower.NotYet
			if elements == "[1, 'a']" && errors.As(err, &notYet) && notYet.What == "an array of string | number" {
				t.Log("mixed elements refused before sort: boxed union representation is absent")
				return
			}
			var refused *lower.Refused
			if !errors.As(err, &refused) || refused.What != "sort without a comparator" || refused.Fix != "pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9]" {
				t.Fatalf("want unchanged default-sort refusal, got %v", err)
			}
		})
	}
}
