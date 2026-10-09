package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Ranked shapes have separate reductions so failures identify the actual shape.
// The optional find reduction is a compiler refusal, held in lower's tests.
func init() {
	for index := 1; index <= 20; index++ {
		if index == 16 {
			continue
		}
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{fmt.Sprintf("internal/oracle/testdata/library_tsc_census_%02d.a", index), true, false})
	}
	for _, name := range []string{"substr", "push_many"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_tsc_census_" + name + ".a", true, false})
	}
}

// These mutants all finish cleanly. The independent source on Node is the only
// check that rejects them; JavaScript and WASI exercise the same mutated IR too.
func TestLibraryTscCensusMutants(t *testing.T) {
	for _, name := range []string{"substr_start", "substr_count", "push_order", "push_length", "push_snapshot"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := "substr"
			if name == "push_order" || name == "push_length" || name == "push_snapshot" {
				fixture = "push_many"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_tsc_census_"+fixture+".a"))
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
				if fixture == "substr" && function.Name == "library_string_substr" {
					mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
						call, ok := value.(ir.StringCall)
						if !ok || call.Method != "slice" {
							return value
						}
						if name == "substr_start" {
							call.Arguments[0] = ir.Binary{Operator: ir.Add, Left: call.Arguments[0], Right: ir.NumberConstant{Value: 1}}
						} else {
							call.Arguments[1] = call.Arguments[1].(ir.Binary).Right
						}
						changed = true
						return call
					})
				}
				if fixture == "push_many" && function.Name == "array_push_many" {
					if name == "push_order" && len(function.Body) > 2 {
						function.Body[0], function.Body[1] = function.Body[1], function.Body[0]
						changed = true
					}
					if name == "push_length" {
						function.Body[len(function.Body)-1] = ir.Return{Value: ir.NumberConstant{Value: -1}}
						changed = true
					}
				}
			}
			if name == "push_snapshot" {
				mutate := func(value ir.Expression) ir.Expression {
					call, ok := value.(ir.Call)
					if !ok || program.Functions[call.Function].Name != "array_push_many" || len(call.Arguments) < 2 {
						return value
					}
					later, ok := call.Arguments[len(call.Arguments)-1].(ir.Call)
					if !ok || program.Functions[later.Function].Name != "changeLater" {
						return value
					}
					for index, argument := range call.Arguments {
						if snapshot, ok := argument.(ir.ArraySlice); ok {
							call.Arguments[index] = snapshot.Array
							changed = true
						}
					}
					return call
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			}
			if !changed {
				t.Fatal("mutant changed no family operation")
			}
			truth := onNode(t, path)
			actual, binary := natively(t, program)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: %+v", actual)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			if difference := disagreement(truth, actual); difference != "stdout differs" {
				t.Fatalf("native catcher: %q", difference)
			}
			js := onJavaScriptBackend(t, program)
			if js.exitCode != 0 || len(js.stderr) != 0 || disagreement(truth, js) != "stdout differs" {
				t.Fatalf("JavaScript must fail only Node stdout: %+v", js)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasi := onWASI(t, native.C(program))
				if wasi.exitCode != 0 || len(wasi.stderr) != 0 || disagreement(truth, wasi) != "stdout differs" {
					t.Fatalf("WASI must fail only Node stdout: %+v", wasi)
				}
			}
			t.Log("clean exit, no sanitizer or leak failure; caught only by Node stdout")
		})
	}
}
