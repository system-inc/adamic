package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func scout22MutantMatchesOnlyNode(t *testing.T, path string, program *ir.Program) {
	t.Helper()
	expected := onNode(t, path)
	result, binary := natively(t, program)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant must exit cleanly: %+v", result)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(expected, result); difference != "stdout differs" {
		t.Fatalf("want Node-only stdout mismatch, got %q", difference)
	}
	if difference := disagreement(expected, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant: %q", difference)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "stdout differs" {
			t.Fatalf("WASI mutant: %q", difference)
		}
	}
	t.Log("mutant exited zero without stderr or leaks; only Node caught stdout differs on each backend")
}

func TestBorrowedOwnNodeMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_lifetime/scout_22_tsc_own_keys.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for index := range program.Functions {
		mutateStringExpressions(reflect.ValueOf(&program.Functions[index].Body).Elem(), func(value ir.Expression) ir.Expression {
			if _, ok := value.(ir.HasOwn); ok {
				changed = true
				return ir.Unary{Operator: ir.Not, Operand: value}
			}
			return value
		})
	}
	if !changed {
		t.Fatal("no borrowed own-property query to mutate")
	}
	scout22MutantMatchesOnlyNode(t, path, program)
}
