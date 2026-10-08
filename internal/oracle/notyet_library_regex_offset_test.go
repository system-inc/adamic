package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/notyet_library_regex_offset.a", true, false})
}

func TestNotYetLibraryRegexOffsetMutants(t *testing.T) {
	for _, rule := range []string{"offset", "input"} {
		t.Run(rule, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_library_regex_offset.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for i := range program.Functions {
				if program.Functions[i].Name != "regex_replace_callback" {
					continue
				}
				mutateStringExpressions(reflect.ValueOf(&program.Functions[i].Body).Elem(), func(value ir.Expression) ir.Expression {
					call, ok := value.(ir.CallClosure)
					if !ok || len(call.Arguments) < 3 || changed {
						return value
					}
					if rule == "offset" {
						call.Arguments[1] = ir.NumberConstant{Value: 0}
					} else {
						call.Arguments[2] = ir.StringConstant{Index: 0}
					}
					changed = true
					return call
				})
			}
			if !changed {
				t.Fatal("mutant changed no callback")
			}
			result, binary := natively(t, program)
			if result.exitCode != 0 {
				t.Fatalf("mutant failed outside output comparison: %s", result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want stdout mismatch, got %q", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("caught by Node stdout comparison")
		})
	}
}
