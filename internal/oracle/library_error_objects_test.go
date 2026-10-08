package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

var errorObjectFixtures = []string{"constructors", "cause", "identity", "format", "aggregate", "stack_header", "stack_frames"}

func init() {
	for _, name := range errorObjectFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_error_" + name + ".a", true, false})
	}
}

// This is an explicit recorded disagreement, not a shortened oracle. Every
// observation still goes through disagreement with full stdout/stderr/exit.
func knownErrorStackFrames(t *testing.T, path string, expected, actual run, backend string) bool {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(source), "// oracle-known-difference: stack frames\n") {
		return false
	}
	difference := disagreement(expected, actual)
	if difference != "stdout differs" || expected.exitCode != 0 || actual.exitCode != 0 || len(expected.stderr) != 0 || len(actual.stderr) != 0 {
		t.Fatalf("known stack frames difference must be full stdout disagreement with clean exits; %s: %s", backend, difference)
	}
	nodeLines, actualLines := strings.Split(string(expected.stdout), "\n"), strings.Split(string(actual.stdout), "\n")
	if nodeLines[0] != actualLines[0] {
		t.Fatalf("stack first line differs: Node %q %s %q", nodeLines[0], backend, actualLines[0])
	}
	if backend != "JavaScript" && (len(actualLines) != 3 || !strings.HasPrefix(actualLines[1], "Adamic frame: adamic_function_") || strings.Contains(actualLines[1], "\n    at ")) {
		t.Fatalf("expected one named Adamic frame with limit 1: %q", actual.stdout)
	}
	t.Logf("known named difference: stack frames (%s); full comparison: %s; Node stdout %q; %s stdout %q; exits %d/%d", backend, difference, expected.stdout, backend, actual.stdout, expected.exitCode, actual.exitCode)
	return true
}

func TestErrorObjectsMutants(t *testing.T) {
	for _, mutant := range []struct{ name, fixture string }{
		{"native constructor becomes Error", "constructors"},
		{"cause omitted", "cause"},
		{"instanceof always false", "identity"},
		{"isPrototypeOf always false", "identity"},
		{"constructor own probe always false", "identity"},
		{"native prototype skips Error", "identity"},
		{"toString returns only message", "format"},
		{"AggregateError loses its elements", "aggregate"},
		{"stack header ignores message", "stack_header"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_error_"+mutant.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			mutate := func(expression ir.Expression) ir.Expression {
				if makeError, ok := expression.(ir.MakeError); ok {
					if mutant.fixture == "stack_header" && makeError.Family != "" {
						makeError.Message = ir.StringConstant{Index: len(program.Strings) - 1}
						changed = true
						return makeError
					}
					if mutant.fixture == "constructors" && makeError.Family == "TypeError" {
						makeError.Family = "Error"
						changed = true
						return makeError
					}
					if mutant.fixture == "cause" && makeError.Cause != nil {
						makeError.Cause = nil
						changed = true
						return makeError
					}
					if mutant.fixture == "aggregate" && makeError.Errors != nil {
						makeError.Errors = ir.ArrayLiteral{Element: makeError.ErrorElement}
						changed = true
						return makeError
					}
				}
				if mutant.name == "constructor own probe always false" {
					if constant, ok := expression.(ir.BooleanConstant); ok && constant.Value {
						changed = true
						return ir.BooleanConstant{Value: false}
					}
				}
				if call, ok := expression.(ir.ObjectCall); ok {
					if mutant.name == "isPrototypeOf always false" && call.Method == "errorIsPrototypeOf" {
						changed = true
						return ir.BooleanConstant{Value: false}
					}
					if mutant.name == "instanceof always false" && call.Method == "errorInstanceOf" {
						changed = true
						return ir.BooleanConstant{Value: false}
					}
					if mutant.name == "native prototype skips Error" && call.Method == "errorGetPrototype" {
						if box, ok := call.Arguments[0].(ir.Box); ok {
							if proto, ok := box.Value.(ir.ObjectCall); ok && proto.Method == "errorPrototype" && proto.Arguments[0].(ir.NumberConstant).Value == 2 {
								changed = true
								return ir.Box{Value: ir.ObjectCall{Method: "errorPrototype", Returns: ir.Object, Arguments: []ir.Expression{ir.NumberConstant{Value: 0}}}}
							}
						}
					}
					if mutant.fixture == "format" && call.Method == "errorToString" {
						changed = true
						return ir.ObjectCall{Method: "errorMember", Returns: ir.String, Arguments: []ir.Expression{call.Arguments[0], ir.NumberConstant{Value: 1}}}
					}
				}
				return expression
			}
			program.Strings = append(program.Strings, "")
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			expected := onNode(t, path)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			observations := []struct {
				name   string
				result run
			}{{"native", executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)}, {"JavaScript", onJavaScriptBackend(t, program)}}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				observations = append(observations, struct {
					name   string
					result run
				}{"WASI", onWASI(t, native.C(program))})
			}
			for _, observation := range observations {
				actual := observation.result
				if actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(expected, actual) != "stdout differs" {
					t.Fatalf("%s mutant must be caught only by clean Node comparison: exit %d stderr %q stdout %q", observation.name, actual.exitCode, actual.stderr, actual.stdout)
				}
				t.Logf("%s: clean exit 0; caught only by Node stdout comparison", observation.name)
			}
		})
	}
}

func TestErrorPrototypeCleanupWithoutOutput(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "quiet.a")
	if err := os.WriteFile(path, []byte("Error.prototype.name = 'dynamic'.repeat(7);"), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	binary := filepath.Join(directory, "quiet")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	observations := []run{executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary), onJavaScriptBackend(t, program)}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		observations = append(observations, onWASI(t, native.C(program)))
	}
	for _, actual := range observations {
		if difference := disagreement(expected, actual); difference != "" {
			t.Fatalf("quiet prototype cleanup: %s stdout %q stderr %q exit %d", difference, actual.stdout, actual.stderr, actual.exitCode)
		}
	}
}

func TestErrorPrototypeParallelReads(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "shared.a")
	source := `import {parallelMap} from 'adamic';
Error.prototype.name = 'shared'.repeat(3);
const values:readonly number[]=[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100,101,102,103,104,105,106,107,108,109,110,111,112,113,114,115,116,117,118,119,120,121,122,123,124,125,126,127];
const names=parallelMap(values,(value:number):string=>new Error(value.toString()).name);
console.log(names.join(','));`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if actual := onJavaScriptBackend(t, program); disagreement(expected, actual) != "" {
		t.Fatal("JavaScript parallel prototype read differs")
	}
	checkParallelVariants(t, program, expected)
}
