package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/4ddd17f_opt_3.a", "internal/oracle/testdata/047cb0d_narrowed_in_try.a", "internal/oracle/testdata/9984394_lib_dispatch.a", "internal/oracle/testdata/9984394_defined_in_try.a", "internal/oracle/testdata/catchability-limits/d96d304_try_stack.a", "internal/oracle/testdata/catchability-limits/d96d304_try_repeat.a", "internal/oracle/testdata/catchability-limits/d96d304_try_pad.a", "internal/oracle/testdata/catchability-limits/d96d304_try_finally_concat.a", "internal/oracle/testdata/57f2d04_with_frozen.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
	for _, path := range []string{"internal/oracle/testdata/catchability-refused/9984394_lib_codepoint.a", "internal/oracle/testdata/catchability-refused/9984394_lib_dispatch_codepoint.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, false, false})
	}
}

// Refusal is itself an oracle boundary: Node must catch the failure, and Lower
// must reject a program whose runtime primitive cannot reach that catch yet.
func TestCodePointCatchabilityBoundary(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"9984394_lib_codepoint.a", "9984394_lib_dispatch_codepoint.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/catchability-refused", name))
			if err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte("made 1\nfromCodePoint caught\n")}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatalf("Node probe: %s", difference)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "String.fromCodePoint") {
				t.Fatalf("want String.fromCodePoint catchability refusal, got %v", err)
			}
		})
	}
}

func TestIntegrationCatchabilityMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct {
		name, fixture string
		mutate        func(*ir.Program) bool
	}{
		{"undefined Error message default removed", "4ddd17f_opt_3.a", func(program *ir.Program) bool {
			return changeErrorFunction(program, "Error_initialize", func(value any) any {
				if coalesce, ok := value.(ir.Coalesce); ok {
					return coalesce.Value
				}
				return value
			})
		}},
		{"interface toFixed guard removed", "9984394_lib_dispatch.a", errorGuardMutant("toFixed() digits argument")},
		{"narrowed TypeError becomes panic", "9984394_defined_in_try.a", func(program *ir.Program) bool {
			return changeErrorFunction(program, "error_defined", func(value any) any {
				if thrown, ok := value.(ir.Throw); ok {
					call := thrown.Value.(ir.Call)
					return ir.Panic{Message: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: errorString(program, "TypeError: ")}, call.Arguments[0]}}}
				}
				return value
			})
		}},
		{"top-level narrowed TypeError becomes panic", "047cb0d_narrowed_in_try.a", func(program *ir.Program) bool {
			return changeErrorFunction(program, "error_defined", func(value any) any {
				if thrown, ok := value.(ir.Throw); ok {
					call := thrown.Value.(ir.Call)
					return ir.Panic{Message: ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: errorString(program, "TypeError: ")}, call.Arguments[0]}}}
				}
				return value
			})
		}},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", mutant.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if !mutant.mutate(program) {
				t.Fatal("mutant changed nothing")
			}
			got, _ := natively(t, program)
			if mutant.fixture == "4ddd17f_opt_3.a" {
				if !strings.Contains(string(got.stderr), "runtime error:") {
					t.Fatalf("want sanitizer to detect the missing message default, got %s", got.stderr)
				}
				t.Logf("sanitizer caught missing default: %s", got.stderr)
			}
			if difference := disagreement(want, got); difference == "" {
				t.Fatal("mutant survived Node comparison")
			} else {
				t.Logf("caught: %s (Node exit %d, mutant exit %d)", difference, want.exitCode, got.exitCode)
			}
		})
	}
}

func TestRuntimeRangeMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct {
		name, fixture string
		mutate        func(*ir.Program) bool
	}{
		{"stack guard becomes panic", "catchability-limits/d96d304_try_stack.a", func(program *ir.Program) bool {
			changed := false
			for index := range program.Functions {
				if program.Functions[index].StackGuarded {
					program.Functions[index].StackGuarded = false
					changed = true
				}
			}
			return changed
		}},
		{"repeat length guard removed", "catchability-limits/d96d304_try_repeat.a", runtimeOperationGuardMutant("repeat")},
		{"padding length guard removed", "catchability-limits/d96d304_try_pad.a", runtimeOperationGuardMutant("padStart")},
		{"concatenation length guard removed", "catchability-limits/d96d304_try_finally_concat.a", runtimeOperationGuardMutant("concat")},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", mutant.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if !mutant.mutate(program) {
				t.Fatal("mutant changed nothing")
			}
			got, _ := natively(t, program)
			if difference := disagreement(want, got); difference == "" {
				t.Fatal("mutant survived Node comparison")
			} else {
				t.Logf("caught: %s (Node exit %d, mutant exit %d)", difference, want.exitCode, got.exitCode)
			}
		})
	}
}

func runtimeOperationGuardMutant(method string) func(*ir.Program) bool {
	return func(program *ir.Program) bool {
		changed := false
		for index := range program.Functions {
			function := &program.Functions[index]
			if function.Name != "error_checked_library" || len(function.Body) < 2 {
				continue
			}
			operation := function.Body[1].(ir.Return).Value
			matching := false
			switch operation := operation.(type) {
			case ir.Concat:
				matching = method == "concat"
			case ir.StringCall:
				matching = operation.Method == method
			}
			if matching {
				function.Body[0] = ir.If{Condition: ir.BooleanConstant{}}
				changed = true
			}
		}
		return changed
	}
}

// Counts are deliberately not recorded for this supplemental probe: native
// stack depth differs from Node and from sanitized versus release frames.
// LeakSanitizer, output and exit status hold the cleanup behavior instead.
func TestRuntimeStackCleanup(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/catchability-limits/d96d304_stack_cleanup.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if difference := disagreement(run{stdout: []byte("caught with cleanup\nafter\n")}, want); difference != "" {
		t.Fatal(difference)
	}
	got, binary := natively(t, program)
	if difference := disagreement(want, got); difference != "" {
		t.Fatal(difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
}

func TestRuntimeStackCatchFinally(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/catchability-limits/d96d304_stack_catch_finally.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	expected := run{stdout: []byte("finally ran\n"), stderr: []byte("adamic: panic: RangeError: Maximum call stack size exceeded\n"), exitCode: 70}
	if difference := disagreement(expected, want); difference != "" {
		t.Fatal(difference)
	}
	got, _ := natively(t, program)
	if difference := disagreement(want, got); difference != "" {
		t.Fatal(difference)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
}
