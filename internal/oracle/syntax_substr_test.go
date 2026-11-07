package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/syntax_substr.a", true, false})
}

// Each mutant remains valid IR and finishes cleanly. Source Node alone supplies the answers.
func TestSyntaxSubstrMutants(test *testing.T) {
	test.Parallel()
	for _, name := range []string{
		"relative-start",
		"truncate",
		"NaN",
		"undefined-length",
		"length-not-end",
		"length-clamp",
		"evaluation-order",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/syntax_substr.a"))
			if err != nil {
				test.Fatal(err)
			}
			program, err := lowered(test, path)
			if err != nil {
				test.Fatal(err)
			}
			changed := false
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.Name != "library_string_substr" {
					continue
				}
				mutateStringExpressions(reflect.ValueOf(
					&function.Body,
				).Elem(), func(value ir.Expression) ir.Expression {
					switch node := value.(type) {
					case ir.Binary:
						if name == "relative-start" && node.Operator == ir.Less {
							node.Operator = ir.Greater
							changed = true
							return node
						}
					case ir.MathCall:
						if name == "truncate" && node.Function == "trunc" {
							node.Function = "floor"
							changed = true
							return node
						}
						if name == "length-clamp" && node.Function == "min" && len(node.Arguments) == 2 {
							if _, remainder := node.Arguments[1].(ir.Binary); remainder {
								maximum := node.Arguments[0].(ir.MathCall)
								changed = true
								return maximum.Arguments[0]
							}
						}
					case ir.NumberCall:
						if name == "NaN" && node.Function == "isNaN" {
							changed = true
							return ir.BooleanConstant{Value: false}
						}
					case ir.Coalesce:
						if name == "undefined-length" {
							if _, length := node.Fallback.(ir.StringLength); length {
								node.Fallback = ir.NumberConstant{Value: 0}
								changed = true
								return node
							}
						}
					case ir.StringCall:
						if name == "length-not-end" && node.Method == "slice" {
							node.Arguments[1] = node.Arguments[1].(ir.Binary).Right
							changed = true
							return node
						}
					}
					return value
				})
			}
			if name == "evaluation-order" {
				mutate := func(value ir.Expression) ir.Expression {
					if node, call := value.(ir.Call); call &&
						program.Functions[node.Function].Name == "library_string_substr" &&
						len(node.Arguments) == 3 {
						if _, effect := node.Arguments[1].(ir.Call); effect {
							node.Arguments[1], node.Arguments[2] = node.Arguments[2], node.Arguments[1]
							changed = true
							return node
						}
					}
					return value
				}
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
				mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			}
			if !changed {
				test.Fatal("mutant changed no expression")
			}
			truth := onNode(test, path)
			generated := onJavaScriptBackend(test, program)
			if difference := disagreement(truth, generated); difference != "stdout differs" {
				test.Fatalf(
					"JavaScript mutant must fail only output comparison, got %q: %s",
					difference,
					generated.stderr,
				)
			}
			result, binary := natively(test, program)
			if result.exitCode != 0 {
				test.Fatalf("mutant must finish: %s", result.stderr)
			}
			if difference := disagreement(truth, result); difference != "stdout differs" {
				test.Fatalf("native mutant must fail only output comparison, got %q", difference)
			}
			if report := leaks(test, program, binary); report != "" {
				test.Fatal(report)
			}
			test.Log("caught in both backends by stdout comparison with source Node")
		})
	}
}
