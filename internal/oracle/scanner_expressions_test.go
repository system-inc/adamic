package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"string-number-concatenation.a", "number_spellings.a", "primitive_concatenation.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/scanner_expressions/" + name, true, false})
	}
}

func TestScannerConcatenationMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scanner_expressions/number_spellings.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	mutate := func(value ir.Expression) ir.Expression {
		if number, ok := value.(ir.NumberToString); ok {
			changed = true
			number.Value = ir.Binary{Operator: ir.Add, Left: number.Value, Right: ir.NumberConstant{Value: 1}}
			return number
		}
		return value
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
	if !changed {
		t.Fatal("mutant changed nothing")
	}
	truth := onNode(t, path)
	got, binary := natively(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if leaked := leaks(t, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("backend mutant caught by %q", difference)
	}
}

func TestScannerPrimitiveConcatenationMutants(t *testing.T) {
	for _, name := range []string{"boolean", "undefined", "null"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scanner_expressions/primitive_concatenation.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			if name == "boolean" {
				mutate := func(value ir.Expression) ir.Expression {
					if boolean, ok := value.(ir.BooleanToString); ok {
						changed = true
						boolean.Value = ir.Unary{Operator: ir.Not, Operand: boolean.Value}
						return boolean
					}
					return value
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			} else {
				for index, text := range program.Strings {
					if text == name {
						program.Strings[index] = "wrong"
						changed = true
					}
				}
			}
			if !changed {
				t.Fatal("mutant changed nothing")
			}
			truth := onNode(t, path)
			got, binary := natively(t, program)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant failed outside comparison: %+v", got)
			}
			if leaked := leaks(t, program, binary); leaked != "" {
				t.Fatal(leaked)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				t.Fatalf("mutant caught by %q", difference)
			}
			if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "stdout differs" {
				t.Fatalf("backend mutant caught by %q", difference)
			}
			t.Logf("%s spelling mutant caught only by stdout in both backends", name)
		})
	}
}
