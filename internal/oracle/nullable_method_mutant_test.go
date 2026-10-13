package oracle

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Each mutant preserves the ABI, compiles, and finishes without a sanitizer or
// leak failure. Only comparison with the original source on Node can catch it.
func nullableNodeMutant(t *testing.T, fixture string, mutate func(*ir.Program) int) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	nullablePathNodeMutant(t, path, mutate)
}

func nullablePathNodeMutant(t *testing.T, path string, mutate func(*ir.Program) int) {
	t.Helper()
	truth := onNode(t, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("source Node: %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	correct, binary := natively(t, program)
	if difference := disagreement(truth, correct); difference != "" {
		t.Fatalf("unmutated native: %s", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatalf("unmutated JavaScript: %s", difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(truth, onWASI(t, native.C(program))); difference != "" {
			t.Fatalf("unmutated WASI: %s", difference)
		}
	}
	if mutate(program) == 0 {
		t.Fatal("mutant changed nothing")
	}
	actual, binary := natively(t, program)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", actual)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatalf("mutant must not leak: %s", report)
	}
	if difference := disagreement(truth, actual); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q, want only stdout differs", difference)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		actual := onWASI(t, native.C(program))
		if difference := disagreement(truth, actual); difference != "stdout differs" {
			t.Fatalf("WASI mutant caught by %q, want only stdout differs", difference)
		}
	}
	t.Log("mutant caught only by Node stdout; clean exit, stderr and leak checks")
}

func nullableMutateExpressions(program *ir.Program, mutate func(ir.Expression) ir.Expression) {
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
}

func TestNullableSpellingNodeMutant(t *testing.T) {
	t.Parallel()
	nullableNodeMutant(t, "nullable_coverage_calls.a", func(program *ir.Program) int {
		changed := 0
		for index, text := range program.Strings {
			if text == "null" {
				program.Strings[index] = "undefined"
				changed++
			}
		}
		return changed
	})
}

func TestNullableNumberNodeMutant(t *testing.T) {
	t.Parallel()
	nullableNodeMutant(t, "nullable_coverage_calls.a", func(program *ir.Program) int {
		changed := 0
		for index := range program.Functions {
			function := &program.Functions[index]
			if function.Name != "nullable_number" {
				continue
			}
			returned := function.Body[0].(ir.Return)
			conditional := returned.Value.(ir.Conditional)
			conditional.WhenTrue = ir.NumberConstant{Value: math.NaN()}
			returned.Value = conditional
			function.Body[0] = returned
			changed++
		}
		return changed
	})
}

func TestNullableTypeOfNodeMutant(t *testing.T) {
	t.Parallel()
	nullableNodeMutant(t, "nullable_references_typeof_match.a", func(program *ir.Program) int {
		changed := 0
		nullableMutateExpressions(program, func(expression ir.Expression) ir.Expression {
			if observed, ok := expression.(ir.TypeOf); ok && observed.Null {
				observed.Null = false
				changed++
				return observed
			}
			return expression
		})
		return changed
	})
}

func TestNullableJSONNodeMutant(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "nullable-json.a")
	// Coalescing makes this JSON observation valid with the current string-only
	// console declaration; direct nullable console witnesses wait for host-5.
	const source = `function show(n: string | null, u: string | undefined): void {
    console.log(JSON.stringify({n: n, u: u}) ?? 'missing');
    console.log(JSON.stringify([n]) ?? 'missing');
    console.log(JSON.stringify([u]) ?? 'missing');
    console.log(JSON.stringify(n) ?? 'missing');
}
show(null, undefined);
show('owned' + 1, 'owned' + 2);
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	nullablePathNodeMutant(t, path, func(program *ir.Program) int {
		changed := 0
		var schema func(*ir.JSONSchema)
		schema = func(value *ir.JSONSchema) {
			if value == nil {
				return
			}
			if value.Null {
				value.Null = false
				changed++
			}
			schema(value.Element)
			for _, field := range value.Fields {
				schema(field.Schema)
			}
		}
		nullableMutateExpressions(program, func(expression ir.Expression) ir.Expression {
			if stringify, ok := expression.(ir.JSONStringify); ok {
				schema(stringify.Schema)
			}
			return expression
		})
		return changed
	})
}

func TestNullableGenericNodeMutant(t *testing.T) {
	t.Parallel()
	nullableNodeMutant(t, "nullable_coverage_nested_generics.a", func(program *ir.Program) int {
		first, second := -1, -1
		for index, function := range program.Functions {
			if strings.HasPrefix(function.Name, "arrayKinds_") {
				if first < 0 {
					first = index
				} else {
					second = index
				}
			}
		}
		if first < 0 || second < 0 {
			return 0
		}
		changed := 0
		nullableMutateExpressions(program, func(expression ir.Expression) ir.Expression {
			if call, ok := expression.(ir.Call); ok && call.Function == second {
				call.Function = first
				changed++
				return call
			}
			return expression
		})
		return changed
	})
}

func TestNullableObservationOnceNodeMutant(t *testing.T) {
	t.Parallel()
	nullableNodeMutant(t, "nullable_coverage_calls.a", func(program *ir.Program) int {
		changed := 0
		nullableMutateExpressions(program, func(expression ir.Expression) ir.Expression {
			call, ok := expression.(ir.Call)
			if ok && program.Functions[call.Function].Name == "next" && len(call.Arguments) == 1 {
				if argument, ok := call.Arguments[0].(ir.BooleanConstant); ok && !argument.Value {
					changed++
					return ir.Comma{Left: call, Right: call}
				}
			}
			return expression
		})
		return changed
	})
}

func TestMethodPresenceNodeMutant(t *testing.T) {
	t.Parallel()
	nullableNodeMutant(t, "method_presence.a", func(program *ir.Program) int {
		changed := 0
		nullableMutateExpressions(program, func(expression ir.Expression) ir.Expression {
			if presence, ok := expression.(ir.MethodPresence); ok {
				presence.Name = "absent"
				changed++
				return presence
			}
			return expression
		})
		return changed
	})
}
