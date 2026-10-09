package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// These mutants keep valid IR and valid C and finish without sanitizer findings. Only the source
// run on Node knows their answers are wrong. Each changes a function family's real behavior.
func TestLibraryStringMutants(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, fixture string }{
		{"conversion", "conversion"}, {"prototype", "prototype"}, {"charAt", "indices"},
		{"at", "indices"}, {"codePointAt", "indices"}, {"substring", "indices"}, {"concat", "indices"}, {"raw", "raw"}, {"rawTemplate", "raw"},
		{"fromCharCode", "existing"}, {"normalize", "existing"}, {"replace", "existing"},
	}
	for _, test := range cases {
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
				if changed && test.name != "raw" {
					return value
				}
				switch node := value.(type) {
				case ir.StringConstant:
					if test.name == "rawTemplate" && strings.Contains(program.Strings[node.Index], "\\n") {
						text := strings.ReplaceAll(program.Strings[node.Index], "\\n", "\n")
						node.Index = len(program.Strings)
						program.Strings = append(program.Strings, text)
						changed = true
						return node
					}
				case ir.NumberToString:
					if test.name == "conversion" {
						node.Value = ir.Unary{Operator: ir.Negate, Operand: node.Value}
						changed = true
						return node
					}
				case ir.Trim:
					if test.name == "prototype" {
						changed = true
						return ir.StringCall{Method: "trimStart", Value: node.Value}
					}
				case ir.StringIndex:
					if test.name == "charAt" {
						node.Index = ir.Binary{Operator: ir.Add, Left: node.Index, Right: ir.NumberConstant{Value: 1}}
						changed = true
						return node
					}
				case ir.StringCall:
					if test.name == "substring" && node.Method == "slice" {
						node.Arguments[0] = node.Arguments[1]
						changed = true
						return node
					}
					if (test.name == "at" && node.Method == "at") || (test.name == "codePointAt" && node.Method == "codePointAt") {
						node.Arguments[0] = ir.Binary{Operator: ir.Add, Left: node.Arguments[0], Right: ir.NumberConstant{Value: 1}}
						changed = true
						return node
					}
					if test.name == "normalize" && node.Method == "normalize" {
						node.Arguments[0] = ir.StringConstant{Index: len(program.Strings)}
						program.Strings = append(program.Strings, "NFD")
						changed = true
						return node
					}
					if test.name == "replace" && node.Method == "replace" {
						node.Method = "replaceAll"
						changed = true
						return node
					}
				case ir.Concat:
					if test.name == "concat" && len(node.Parts) == 4 {
						if first, constant := node.Parts[0].(ir.StringConstant); constant && program.Strings[first.Index] == "abc" {
							node.Parts = node.Parts[:3]
							changed = true
							return node
						}
					}
				case ir.StringFromCodes:
					if test.name == "fromCharCode" && !node.CodePoints {
						node.CodePoints = true
						changed = true
						return node
					}
				case ir.Binary:
					if test.name == "raw" && node.Operator == ir.Less {
						// The substitution limit, rather than the loop bound: its left operand is nextIndex + 1.
						if _, next := node.Left.(ir.Binary); next {
							node.Operator = ir.LessOrEqual
							changed = true
							return node
						}
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if !changed {
				t.Fatal("mutant changed no expression")
			}
			result, binary := natively(t, program)
			if result.exitCode != 0 {
				t.Fatalf("mutant must finish, not fail a compiler or sanitizer check: exit %d, stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want only the Node comparison to kill mutant, got %q", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatalf("mutant must be leak-clean: %s", report)
			}
			t.Log("caught only by stdout comparison with source on Node")
		})
	}
}

func mutateStringExpressions(value reflect.Value, mutate func(ir.Expression) ir.Expression) {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return
		}
		if expression, yes := value.Interface().(ir.Expression); yes {
			replacement := mutate(expression)
			if !reflect.DeepEqual(expression, replacement) {
				value.Set(reflect.ValueOf(replacement))
				return
			}
		}
		copy := reflect.New(value.Elem().Type()).Elem()
		copy.Set(value.Elem())
		mutateStringExpressions(copy, mutate)
		value.Set(copy)
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			mutateStringExpressions(value.Field(index), mutate)
		}
	case reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			mutateStringExpressions(value.Index(index), mutate)
		}
	}
}
