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
	}{"internal/oracle/testdata/notyet_library_regex_callback.a", true, false})
}

func TestNotYetLibraryRegexCallbackMutants(t *testing.T) {
	for _, rule := range []string{"match", "literal", "unicode", "global", "offset", "reset"} {
		t.Run(rule, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_library_regex_callback.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				if changed && rule != "unicode" {
					return value
				}
				switch n := value.(type) {
				case ir.CallClosure:
					if rule == "match" && len(n.Arguments) == 1 {
						n.Arguments[0] = ir.StringConstant{Index: len(program.Strings)}
						program.Strings = append(program.Strings, "WRONG")
						changed = true
						return n
					}
					if rule == "literal" && len(n.Arguments) == 0 {
						changed = true
						index := len(program.Strings)
						program.Strings = append(program.Strings, "a")
						text := ir.StringConstant{Index: index}
						return ir.StringCall{Value: text, Method: "replace", Arguments: []ir.Expression{text, n}}
					}
				case ir.Conditional:
					if rule == "unicode" {
						if v, ok := n.WhenTrue.(ir.NumberConstant); ok && v.Value == 2 {
							n.Condition = ir.BooleanConstant{Value: true}
							changed = true
							return n
						}
					}
				case ir.Property:
					if rule == "global" && n.Name == "global" {
						changed = true
						return ir.BooleanConstant{Value: false}
					}
				case ir.RegExpProperty:
					if rule == "offset" && n.Name == "index" {
						changed = true
						return ir.NumberConstant{Value: 0}
					}
				}
				return value
			}
			if rule == "reset" {
				for i := range program.Functions {
					if program.Functions[i].Name != "regex_replace_callback" {
						continue
					}
					for j, statement := range program.Functions[i].Body {
						branch, ok := statement.(ir.If)
						if !ok || len(branch.Then) != 1 {
							continue
						}
						set, ok := branch.Then[0].(ir.SetProperty)
						if !ok || set.Name != "lastIndex" {
							continue
						}
						set.Value = ir.NumberConstant{Value: 1}
						branch.Then[0] = set
						program.Functions[i].Body[j] = branch
						changed = true
					}
				}
			}
			mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
			if !changed {
				t.Fatal("mutant changed no expression")
			}
			result, binary := natively(t, program)
			if result.exitCode != 0 {
				t.Fatalf("mutant failed outside comparison: %s", result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want stdout mismatch, got %q", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("caught by source Node stdout comparison")
		})
	}
}
