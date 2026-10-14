package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/array_generic/library_array_heterogeneous.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/array_generic/library_tsc_census_push_many.a", true, false})
}

func arrayGenericNodeMutant(t *testing.T, fixture string, mutate func(*ir.Program) bool) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if !mutate(program) {
		t.Fatal("mutant changed nothing")
	}
	expected := onNode(t, path)
	actual, binary := natively(t, program)
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	results := map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		results["WASI"] = onWASI(t, native.C(program))
	}
	for backend, result := range results {
		if result.exitCode != 0 || len(result.stderr) != 0 || disagreement(expected, result) != "stdout differs" {
			t.Fatalf("%s: mutant must finish cleanly and differ only from Node stdout: %+v", backend, result)
		}
		t.Logf("%s: caught only by Node stdout; clean exit and no leak report", backend)
	}
}

func TestArrayGenericForwardNodeMutant(t *testing.T) {
	t.Parallel()
	arrayGenericNodeMutant(t, "library_array_search.a", func(program *ir.Program) bool {
		changed := false
		for i := range program.Functions {
			function := &program.Functions[i]
			if function.Name != "array_generic_indexOf" {
				continue
			}
			for j, statement := range function.Body {
				branch, ok := statement.(ir.If)
				if !ok {
					continue
				}
				eligible, ok := branch.Condition.(ir.Binary)
				if !ok || eligible.Operator != ir.And {
					continue
				}
				bound := eligible.Right.(ir.Binary)
				bound.Operator = ir.LessOrEqual
				eligible.Right = bound
				branch.Condition = eligible
				function.Body[j] = branch
				changed = true
			}
		}
		return changed
	})
}

func TestArrayGenericBackwardNodeMutant(t *testing.T) {
	t.Parallel()
	arrayGenericNodeMutant(t, "library_array_search.a", func(program *ir.Program) bool {
		changed := false
		for i := range program.Functions {
			function := &program.Functions[i]
			if function.Name != "array_generic_lastIndexOf" {
				continue
			}
			for j, statement := range function.Body {
				declaration, ok := statement.(ir.Declare)
				if !ok || program.Locals[declaration.Local].Name != "from" {
					continue
				}
				value := declaration.Value.(ir.Conditional)
				negative := value.WhenTrue.(ir.Binary)
				negative.Operator = ir.Subtract
				value.WhenTrue = negative
				declaration.Value = value
				function.Body[j] = declaration
				changed = true
			}
		}
		return changed
	})
}

func TestArrayPushManyNodeMutant(t *testing.T) {
	t.Parallel()
	arrayGenericNodeMutant(t, "array_generic/library_tsc_census_push_many.a", func(program *ir.Program) bool {
		changed := false
		for i := range program.Functions {
			function := &program.Functions[i]
			if function.Name != "array_push_many" {
				continue
			}
			for j, statement := range function.Body {
				evaluated, ok := statement.(ir.Evaluate)
				if !ok {
					continue
				}
				if _, ok := evaluated.Value.(ir.ArrayPush); !ok {
					continue
				}
				evaluated.Value = ir.NumberConstant{}
				function.Body[j] = evaluated
				changed = true
				break
			}
		}
		return changed
	})
}

func TestArrayIntrinsicBrandNodeMutant(t *testing.T) {
	t.Parallel()
	arrayGenericNodeMutant(t, "library_array_metadata.a", func(program *ir.Program) bool {
		changed := false
		mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
			if fact, ok := value.(ir.BooleanConstant); ok && !changed {
				fact.Value = !fact.Value
				changed = true
				return fact
			}
			return value
		})
		return changed
	})
}

func TestArraySlotsIdentityNodeMutant(t *testing.T) {
	t.Parallel()
	arrayGenericNodeMutant(t, "library_array_search.a", func(program *ir.Program) bool {
		changed := false
		mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
			if search, ok := value.(ir.ArraySearch); ok && search.Element == ir.Object {
				search.Last = !search.Last
				changed = true
				return search
			}
			return value
		})
		return changed
	})
}

func TestArrayCopyWithinUndefinedNodeMutant(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/library_array_copy_within.a"))
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(source), "console.log([1, 2, 3, 4].copyWithin(0, 2, undefined)")
	if start < 0 {
		t.Fatal("missing undefined-end witness")
	}
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, source[start:], 0644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	original := native.C(program)
	mutant := strings.ReplaceAll(original, "HUGE_VAL", "0.0")
	if mutant == original {
		t.Fatal("mutant changed nothing")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if got.exitCode != 0 || len(got.stderr) != 0 || disagreement(onNode(t, path), got) != "stdout differs" {
		t.Fatalf("must fail only Node stdout: %+v", got)
	}
	t.Log("undefined end changed to zero; caught only by Node stdout")
}
