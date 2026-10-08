package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Registration stays with the String slice; the shared harness is unchanged.
func init() {
	for _, entry := range []struct {
		name   string
		lowers bool
	}{
		{"range_catch", true}, {"typeof_conversion", true}, {"primitive", true}, {"affixes", true}, {"wellformed", true}, {"null_receiver", true}, {"raw_primitive", true},
	} {
		path := "internal/oracle/testdata/library_string_" + entry.name + ".a"
		if !entry.lowers {
			// The flow gate globs runnable top-level fixtures; refusal probes live separately.
			path = "internal/oracle/testdata/library_string_refusals/library_string_" + entry.name + ".a"
		}
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			path, entry.lowers, false,
		})
	}
}

// Every mutant remains valid C, exits zero and is sanitizer- and leak-clean.
// Only the independently executed source on Node rejects its answer.
func TestLibraryStringSecondMutants(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, fixture string }{
		{"identity", "typeof_conversion"}, {"primitive", "primitive"},
		{"arrayConversion", "primitive"}, {"mapConversion", "primitive"},
		{"startsWith", "affixes"}, {"endsWith", "affixes"},
		{"isWellFormed", "wellformed"}, {"toWellFormed", "wellformed"},
		{"nullReceiver", "null_receiver"}, {"rawLimit", "raw_primitive"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_string_"+test.fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				if changed && test.name != "identity" && test.name != "rawLimit" {
					return value
				}
				switch node := value.(type) {
				case ir.CallClosure:
					if test.name == "primitive" {
						if property, yes := node.Closure.(ir.Property); yes && property.Name == "toString" {
							property.Name = "valueOf"
							node.Closure = property
							changed = true
							return node
						}
					}
				case ir.ArrayJoin:
					if test.name == "arrayConversion" {
						node.Separator = ir.StringConstant{Index: len(program.Strings)}
						program.Strings = append(program.Strings, ";")
						changed = true
						return node
					}
				case ir.BuiltinError:
					if test.name == "nullReceiver" && node.Kind == 1 {
						node.Kind = 3
						changed = true
						return node
					}
				case ir.StringConstant:
					if test.name == "mapConversion" && program.Strings[node.Index] == "[object Map]" {
						program.Strings[node.Index] = "[object Set]"
						changed = true
					}
				case ir.StringCall:
					if test.name == "isWellFormed" && node.Method == "isWellFormed" {
						changed = true
						return ir.Unary{Operator: ir.Not, Operand: node}
					}
					if test.name == "toWellFormed" && node.Method == "toWellFormed" {
						changed = true
						return node.Value
					}
					if node.Method == "slice" && (test.name == "startsWith" || test.name == "endsWith") {
						index := 0
						if test.name == "endsWith" {
							index = 1
						}
						node.Arguments[index] = ir.Binary{Operator: ir.Add, Left: node.Arguments[index], Right: ir.NumberConstant{Value: 1}}
						changed = true
						return node
					}
				case ir.Read:
					if test.name == "identity" {
						changed = true
						return ir.StringCall{Method: "slice", Value: node, Arguments: []ir.Expression{ir.NumberConstant{Value: 1}}}
					}
				case ir.Binary:
					if test.name == "rawLimit" && node.Operator == ir.Less {
						if _, next := node.Left.(ir.Binary); next {
							node.Operator = ir.LessOrEqual
							changed = true
							return node
						}
					}
				}
				return value
			}
			if test.name == "identity" || test.name == "startsWith" || test.name == "endsWith" {
				for index := range program.Functions {
					if program.Functions[index].Name == "library_string_"+test.name {
						mutateStringExpressions(reflect.ValueOf(&program.Functions[index].Body).Elem(), mutate)
					}
				}
			} else {
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			}
			if !changed {
				t.Fatal("mutant changed no expression")
			}
			result, binary := natively(t, program)
			if result.exitCode != 0 {
				t.Fatalf("mutant must finish cleanly: exit %d stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want only Node stdout comparison to reject mutant, got %q", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatalf("mutant leaked: %s", report)
			}
			t.Log("caught only by stdout comparison with source on Node")
		})
	}
}
