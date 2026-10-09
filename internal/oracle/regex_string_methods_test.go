package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"split", "match", "match_all", "search", "replace", "replace_all"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/regex_string_" + name + ".a", true, false,
		})
	}
}

// Each mutant is valid IR and C, exits normally and is sanitizer/leak clean. Only Node knows
// that dropping the first input code unit changes the specified result and lastIndex behavior.
func TestRegexStringMethodMutants(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ method, fixture string }{
		{"split", "split"}, {"match", "match"}, {"matchAll", "match_all"},
		{"search", "search"}, {"replace", "replace"}, {"replaceAll", "replace_all"},
	} {
		t.Run(test.method, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regex_string_"+test.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				if call, ok := value.(ir.RegExpCall); ok && call.Method == test.method && !changed {
					call.Value = ir.StringCall{Method: "slice", Value: call.Value, Arguments: []ir.Expression{ir.NumberConstant{Value: 1}}}
					changed = true
					return call
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if !changed {
				t.Fatal("mutant changed no expression")
			}
			result, binary := natively(t, program)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: exit %d stderr %q", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want Node-only stdout difference, got %q", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("Node alone caught wrong output; compiler, sanitizers and leak checks passed")
		})
	}
}

// Wrong exception identity must also be caught by Node, without relying on a panic or sanitizer.
func TestRegexStringErrorMutants(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ fixture, runtime, result, parameters, arguments string }{
		{"match_all", "adamic_regex_match_all", "adamic_object", "adamic_string *input, adamic_object *pattern", "input, pattern"},
		{"replace_all", "adamic_regex_replace", "adamic_string", "adamic_string *input, adamic_object *pattern, adamic_string *replacement, bool global", "input, pattern, replacement, global"},
	} {
		t.Run(test.fixture, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regex_string_"+test.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			source := native.C(program)
			if !strings.Contains(source, test.runtime+"(") {
				t.Fatal("mutant call absent")
			}
			source = strings.ReplaceAll(source, test.runtime+"(", "regex_string_mutant(")
			wrapper := "#include \"adamic.h\"\nstatic " + test.result + " *regex_string_mutant(" + test.parameters + ") {\n" +
				test.result + " *result = " + test.runtime + "(" + test.arguments + ");\n" +
				"if (adamic_thrown != NULL) { static adamic_string name = ADAMIC_STRING(\"RangeError\"); " +
				"adamic_release(adamic_thrown->slots[0].reference); adamic_thrown->slots[0].reference = adamic_retain(&name); }\nreturn result;\n}\n"
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(wrapper+source, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := execute(t, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: exit %d stderr %q", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want Node-only stdout difference, got %q", difference)
			}
			t.Log("Node alone caught TypeError changed to RangeError; sanitizer clean")
		})
	}
}
