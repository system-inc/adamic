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
func TestLibraryStringMutants(test *testing.T) {
	test.Parallel()
	cases := []struct{ name, fixture string }{
		{"conversion", "conversion"},
		{"prototype", "prototype"},
		{"charAt", "indices"},
		{"at", "indices"},
		{"codePointAt", "indices"},
		{"substring", "indices"},
		{"concat", "indices"},
		{"raw", "raw"},
		{"rawTemplate", "raw"},
		{"fromCharCode", "existing"},
		{"normalize", "existing"},
		{"replace", "existing"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/library_string_"+testCase.fixture+".a"),
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
				if changed && testCase.name != "raw" {
					return value
				}
				switch node := value.(type) {
				case ir.StringConstant:
					if testCase.name == "rawTemplate" && strings.Contains(program.Strings[node.Index], "\\n") {
						text := strings.ReplaceAll(program.Strings[node.Index], "\\n", "\n")
						node.Index = len(program.Strings)
						program.Strings = append(program.Strings, text)
						changed = true
						return node
					}
				case ir.NumberToString:
					if testCase.name == "conversion" {
						node.Value = ir.Unary{Operator: ir.Negate, Operand: node.Value}
						changed = true
						return node
					}
				case ir.Trim:
					if testCase.name == "prototype" {
						changed = true
						return ir.StringCall{Method: "trimStart", Value: node.Value}
					}
				case ir.StringIndex:
					if testCase.name == "charAt" {
						node.Index = ir.Binary{Operator: ir.Add, Left: node.Index, Right: ir.NumberConstant{Value: 1}}
						changed = true
						return node
					}
				case ir.StringCall:
					if testCase.name == "substring" && node.Method == "slice" {
						node.Arguments[0] = node.Arguments[1]
						changed = true
						return node
					}
					if (testCase.name == "at" && node.Method == "at") ||
						(testCase.name == "codePointAt" && node.Method == "codePointAt") {
						node.Arguments[0] = ir.Binary{
							Operator: ir.Add,
							Left:     node.Arguments[0],
							Right:    ir.NumberConstant{Value: 1},
						}
						changed = true
						return node
					}
					if testCase.name == "normalize" && node.Method == "normalize" {
						node.Arguments[0] = ir.StringConstant{Index: len(program.Strings)}
						program.Strings = append(program.Strings, "NFD")
						changed = true
						return node
					}
					if testCase.name == "replace" && node.Method == "replace" {
						node.Method = "replaceAll"
						changed = true
						return node
					}
				case ir.Concat:
					if testCase.name == "concat" && len(node.Parts) == 4 {
						if first, constant := node.Parts[0].(ir.StringConstant); constant &&
							program.Strings[first.Index] == "abc" {
							node.Parts = node.Parts[:3]
							changed = true
							return node
						}
					}
				case ir.StringFromCodes:
					if testCase.name == "fromCharCode" && !node.CodePoints {
						node.CodePoints = true
						changed = true
						return node
					}
				case ir.Binary:
					if testCase.name == "raw" && node.Operator == ir.Less {
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
				test.Fatal("mutant changed no expression")
			}
			result, binary := natively(test, program)
			if result.exitCode != 0 {
				test.Fatalf(
					"mutant must finish, not fail a compiler or sanitizer check: exit %d, stderr %s",
					result.exitCode,
					result.stderr,
				)
			}
			if difference := disagreement(onNode(test, path), result); difference != "stdout differs" {
				test.Fatalf("want only the Node comparison to kill mutant, got %q", difference)
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatalf("mutant must be leak-clean: %s", report)
			}
			test.Log("caught only by stdout comparison with source on Node")
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
