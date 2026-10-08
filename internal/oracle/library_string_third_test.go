package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// These mutants still emit valid C, exit zero, and satisfy the sanitizer and
// ownership gates. Only the independent source run on Node detects the error.
func TestLibraryStringThirdMutants(t *testing.T) {
	for _, test := range []struct{ name, fixture string }{
		{"boxLength", "boxed"}, {"boxIndex", "boxed"}, {"boxOwn", "boxed"},
		{"match", "regex_dispatch"}, {"matchAll", "regex_dispatch"},
		{"replace", "regex_dispatch"}, {"replaceAll", "regex_dispatch"},
		{"search", "regex_dispatch"}, {"RegExpCreate", "regex_dispatch"}, {"split", "regex_dispatch"},
		{"fromCodePoint", "ranges"}, {"repeat", "ranges"},
	} {
		t.Run(test.name, func(t *testing.T) {
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
				if changed && test.name != "fromCodePoint" && test.name != "boxOwn" && test.name != "repeat" {
					return value
				}
				switch node := value.(type) {
				case ir.ObjectLiteral:
					if test.name == "boxLength" {
						for index := range node.Fields {
							if node.Fields[index].Name == "length" {
								node.Fields[index].Value = ir.Binary{Operator: ir.Add, Left: node.Fields[index].Value, Right: ir.NumberConstant{Value: 1}}
								changed = true
								return node
							}
						}
					}
				case ir.StringIndex:
					if test.name == "boxIndex" {
						node.Index = ir.Binary{Operator: ir.Add, Left: node.Index, Right: ir.NumberConstant{Value: 1}}
						changed = true
						return node
					}
				case ir.RegExpCall:
					if test.name == "RegExpCreate" && len(node.Arguments) > 0 {
						if made, ok := node.Arguments[0].(ir.RegExpNew); ok && !made.Invalid && program.Regexps[made.Index].Pattern == "(a)" && program.Regexps[made.Index].Flags == "" {
							node.Value = ir.StringCall{Method: "slice", Value: node.Value, Arguments: []ir.Expression{ir.NumberConstant{Value: 1}}}
							changed = true
							return node
						}
					}
					if node.Method == test.name || node.Method == test.name+"Callback" {
						node.Value = ir.StringCall{Method: "slice", Value: node.Value, Arguments: []ir.Expression{ir.NumberConstant{Value: 1}}}
						changed = true
						return node
					}
				case ir.Binary:
					if test.name == "boxOwn" && node.Operator == ir.Less {
						node.Operator = ir.LessOrEqual
						changed = true
						return node
					}
					if test.name == "fromCodePoint" && node.Operator == ir.Greater {
						if constant, ok := node.Right.(ir.NumberConstant); ok && constant.Value == 0x10ffff {
							node.Operator = ir.GreaterOrEqual
							changed = true
							return node
						}
					}
					if test.name == "repeat" && node.Operator == ir.Less {
						if constant, ok := node.Right.(ir.NumberConstant); ok && constant.Value == 0 {
							node.Operator = ir.LessOrEqual
							changed = true
							return node
						}
					}
				}
				return value
			}
			for index := range program.Functions {
				name := program.Functions[index].Name
				if test.name == "boxLength" && name != "library_string_box" {
					continue
				}
				if test.name == "boxOwn" && name != "library_string_box_hasOwn" {
					continue
				}
				if test.name == "boxIndex" && name != "library_string_box_index" {
					continue
				}
				if test.name == "fromCodePoint" && !strings.HasPrefix(name, "library_string_codepoints_checked") {
					continue
				}
				if test.name == "repeat" && !strings.HasPrefix(name, "library_string_repeat_checked") {
					continue
				}
				mutateStringExpressions(reflect.ValueOf(&program.Functions[index].Body).Elem(), mutate)
			}
			if !changed {
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
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
