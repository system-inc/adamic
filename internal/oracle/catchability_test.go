package oracle

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/coverage_error_nullable_cause.a", "internal/oracle/testdata/coverage_error_cause_null.a", "internal/oracle/testdata/coverage_error_prototype_null.a", "internal/oracle/testdata/4ddd17f_opt_3.a", "internal/oracle/testdata/047cb0d_narrowed_in_try.a", "internal/oracle/testdata/9984394_lib_dispatch.a", "internal/oracle/testdata/9984394_defined_in_try.a", "internal/oracle/testdata/catchability-limits/d96d304_try_stack.a", "internal/oracle/testdata/catchability-limits/d96d304_try_repeat.a", "internal/oracle/testdata/catchability-limits/d96d304_try_pad.a", "internal/oracle/testdata/catchability-limits/d96d304_try_finally_concat.a", "internal/oracle/testdata/57f2d04_with_frozen.a"} {
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
			changed := false
			rewrite := func(value any) any {
				if literal, ok := value.(ir.ObjectLiteral); ok && literal.Class != 0 && program.Classes[literal.Class-1].Name == "Error" {
					for index, field := range literal.Fields {
						if coalesce, ok := field.Value.(ir.Coalesce); ok && field.Name == "message" {
							literal.Fields[index].Value = coalesce.Value
							changed = true
						}
					}
					return literal
				}
				return value
			}
			program.Main = errorRewrite(reflect.ValueOf(program.Main), rewrite).Interface().([]ir.Statement)
			for index := range program.Functions {
				program.Functions[index].Body = errorRewrite(reflect.ValueOf(program.Functions[index].Body), rewrite).Interface().([]ir.Statement)
			}
			return changed
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

// These mutants preserve successful compilation and exit 0. Only Node's
// observable null identity or incompatible-receiver output distinguishes them.
func TestReaderNullMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct {
		name, fixture string
		change        func(any) any
	}{
		{"drop nullable reference tag", "coverage_error_nullable_cause.a", func(value any) any {
			if boxed, ok := value.(ir.Box); ok && boxed.Nullable {
				boxed.Nullable = false
				return boxed
			}
			return value
		}},
		{"drop boxed null tag", "coverage_error_cause_null.a", func(value any) any {
			if null, ok := value.(ir.Null); ok && null.Of == ir.Union {
				return ir.Undefined{Of: ir.Union}
			}
			return value
		}},
		{"fold unknown null comparison", "coverage_error_cause_null.a", func(value any) any {
			if test, ok := value.(ir.IsNull); ok && test.Value.Type() == ir.Union {
				test.AlwaysFalse = true
				return test
			}
			return value
		}},
		{"null prototype receiver returns Error", "coverage_error_prototype_null.a", func(value any) any {
			if _, ok := value.(ir.Throw); ok {
				return ir.Return{Value: ir.StringConstant{Index: 0}}
			}
			return value
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
			changed := false
			rewrite := func(value any) any {
				next := mutant.change(value)
				if mutant.fixture == "coverage_error_prototype_null.a" {
					if returned, ok := next.(ir.Return); ok {
						returned.Value = ir.StringConstant{Index: errorString(program, "Error")}
						next = returned
					}
				}
				changed = changed || !reflect.DeepEqual(value, next)
				return next
			}
			program.Main = errorRewrite(reflect.ValueOf(program.Main), rewrite).Interface().([]ir.Statement)
			for index := range program.Functions {
				if mutant.fixture == "coverage_error_prototype_null.a" && program.Functions[index].Name != "Error_prototype_toString" {
					continue
				}
				program.Functions[index].Body = errorRewrite(reflect.ValueOf(program.Functions[index].Body), rewrite).Interface().([]ir.Statement)
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			got, _ := natively(t, program)
			for backend, observation := range map[string]run{"native": got, "JavaScript": onJavaScriptBackend(t, program)} {
				if backend == "JavaScript" && mutant.fixture == "coverage_error_nullable_cause.a" {
					continue
				}
				if observation.exitCode != 0 || string(observation.stdout) == string(want.stdout) {
					t.Fatalf("%s mutant must be caught by stdout at exit 0: %+v", backend, observation)
				}
				t.Logf("%s caught by stdout: Node %q, mutant %q", backend, want.stdout, observation.stdout)
			}
		})
	}
}

func TestMergedGeneratedErrorMutants(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ kind, fixture string }{
		{"TypeError", "class_features_static_private.a"},
		{"ReferenceError", "class_inheritance_conditional.a"},
	} {
		t.Run(probe.kind, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", probe.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			changed := false
			for index := range program.Classes {
				if program.Classes[index].Name == probe.kind {
					program.Classes[index].Base = 0
					changed = true
				}
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			got, _ := natively(t, program)
			for backend, observation := range map[string]run{"native": got, "JavaScript": onJavaScriptBackend(t, program)} {
				if observation.exitCode != 0 || string(observation.stdout) == string(want.stdout) {
					t.Fatalf("%s lost ancestry must be caught by stdout at exit 0: %+v", backend, observation)
				}
				t.Logf("%s lost %s ancestry caught by Node stdout", backend, probe.kind)
			}
		})
	}
}
