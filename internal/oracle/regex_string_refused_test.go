package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, name := range []string{"metadata", "box_search", "conversion", "concat"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/regex_refused_" + name + ".a", true, false})
	}
}

// These mutations preserve well-typed IR, clean compilation, normal completion and leak freedom.
// Only the source run on Node can tell that a length, input, slash or printed number is wrong.
func TestRegexStringRefusalMutants(t *testing.T) {
	t.Parallel()
	for _, family := range []string{"metadata", "box_search", "conversion", "concat", "concat_boolean", "concat_undefined"} {
		t.Run(family, func(t *testing.T) {
			t.Parallel()
			fixture := family
			if family == "concat_boolean" || family == "concat_undefined" {
				fixture = "concat"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/regex_refused_"+fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				if changed {
					return value
				}
				switch node := value.(type) {
				case ir.NumberConstant:
					if family == "metadata" && node.Value == 2 {
						node.Value = 3
						changed = true
						return node
					}
				case ir.RegExpCall:
					if family == "box_search" && node.Method == "search" {
						node.Value = ir.StringCall{Method: "slice", Value: node.Value, Arguments: []ir.Expression{ir.NumberConstant{Value: 1}}}
						changed = true
						return node
					}
				case ir.Concat:
					if family == "conversion" && len(node.Parts) == 4 {
						if slash, ok := node.Parts[0].(ir.StringConstant); ok && program.Strings[slash.Index] == "/" {
							node.Parts = node.Parts[1:]
							changed = true
							return node
						}
					}
				case ir.BooleanToString:
					if family == "concat_boolean" {
						node.Value = ir.Unary{Operator: ir.Not, Operand: node.Value}
						changed = true
						return node
					}
				case ir.StringConstant:
					if family == "concat_undefined" && program.Strings[node.Index] == "undefined" {
						node.Index = len(program.Strings)
						program.Strings = append(program.Strings, "missing")
						changed = true
						return node
					}
				case ir.NumberToString:
					if family == "concat" {
						node.Value = ir.Binary{Operator: ir.Add, Left: node.Value, Right: ir.NumberConstant{Value: 1}}
						changed = true
						return node
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if !changed {
				t.Fatal("mutant target absent")
			}
			result, binary := natively(t, program)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: exit=%d stderr=%q", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want Node-only stdout difference, got %q", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("Node alone caught stdout difference; compiler, sanitizers and leaks passed")
		})
	}
}
