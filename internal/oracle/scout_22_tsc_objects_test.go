package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/scout_22_tsc_own_keys.a", "internal/oracle/testdata/scout_22_tsc_keyword_entries.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// Treating Object.prototype.constructor as own is safe and leak-clean. Only
// source Node distinguishes an inherited property from the receiver's own keys.
func TestScout22OwnPropertyMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_22_tsc_own_keys.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	constructor := len(program.Strings)
	program.Strings = append(program.Strings, "constructor")
	changed := 0
	mutate := func(value ir.Expression) ir.Expression {
		if own, ok := value.(ir.HasOwn); ok {
			changed++
			return ir.Binary{Operator: ir.Or, Left: own, Right: ir.Binary{Operator: ir.Equal, Left: own.Key, Right: ir.StringConstant{Index: constructor}}}
		}
		return value
	}
	for index := range program.Functions {
		mutateStringExpressions(reflect.ValueOf(&program.Functions[index].Body).Elem(), mutate)
	}
	if changed == 0 {
		t.Fatal("no own-property queries mutated")
	}
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
	t.Logf("Node alone caught %d inherited-as-own queries; native exited zero without leaks", changed)
}

func TestScout22KeywordOrderMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_22_tsc_keyword_entries.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
		if call, ok := value.(ir.ObjectCall); ok && call.Method == "entries" {
			changed++
			return ir.ArrayReverse{Array: call}
		}
		return value
	})
	if changed != 1 {
		t.Fatalf("want one entries mutation, got %d", changed)
	}
	scout22MutantMatchesOnlyNode(t, path, program)
}

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
