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

func TestScannerConcatenationMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/scanner_expressions/number_spellings.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
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
		test.Fatal("mutant changed nothing")
	}
	truth := onNode(test, path)
	got, binary := natively(test, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant failed outside comparison: %+v", got)
	}
	if leaked := leaks(test, program, binary); leaked != "" {
		test.Fatal(leaked)
	}
	if difference := disagreement(truth, got); difference != "stdout differs" {
		test.Fatalf("mutant caught by %q", difference)
	}
	if difference := disagreement(truth, onJavaScriptBackend(test, program)); difference != "stdout differs" {
		test.Fatalf("backend mutant caught by %q", difference)
	}
}

func TestScannerPrimitiveConcatenationMutants(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"boolean", "undefined", "null"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/scanner_expressions/primitive_concatenation.a"),
			)
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
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
				test.Fatal("mutant changed nothing")
			}
			truth := onNode(test, path)
			got, binary := natively(test, program)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				test.Fatalf("mutant failed outside comparison: %+v", got)
			}
			if leaked := leaks(test, program, binary); leaked != "" {
				test.Fatal(leaked)
			}
			if difference := disagreement(truth, got); difference != "stdout differs" {
				test.Fatalf("mutant caught by %q", difference)
			}
			if difference := disagreement(truth, onJavaScriptBackend(test, program)); difference != "stdout differs" {
				test.Fatalf("backend mutant caught by %q", difference)
			}
			test.Logf("%s spelling mutant caught only by stdout in both backends", name)
		})
	}
}
