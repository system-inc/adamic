package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"math"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/notyet_library_string_bounds.a", true, false})
}

// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
func TestNotYetLibraryStringBoundsMutants(t *testing.T) {
	for _, rule := range []string{"slice", "substring", "position"} {
		// Not parallel: oracle helpers write the shared os.UserCacheDir()/adamic/gate and os.UserCacheDir()/adamic/runtime caches.
		t.Run(rule, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_library_string_bounds.a"))
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
				switch n := value.(type) {
				case ir.Coalesce:
					if rule == "slice" || rule == "substring" {
						if v, ok := n.Fallback.(ir.NumberConstant); ok && math.IsInf(v.Value, 1) {
							n.Fallback = ir.NumberConstant{Value: 0}
							changed = true
							return n
						}
					}
				case ir.StringCall:
					if rule == "position" && n.Method == "slice" && len(n.Arguments) == 2 {
						n.Arguments[1] = ir.NumberConstant{Value: math.Inf(1)}
						changed = true
						return n
					}
				}
				return value
			}
			if rule == "slice" {
				mutateStringExpressions(reflect.ValueOf(&program.Functions[0].Body).Elem(), mutate)
			} else if rule == "substring" {
				for i := range program.Functions {
					if program.Functions[i].Name == "library_string_substring" {
						mutateStringExpressions(reflect.ValueOf(&program.Functions[i].Body).Elem(), mutate)
					}
				}
				// Coalescing is performed at the call site, before the substring helper receives its end.
				if !changed {
					mutateStringExpressions(reflect.ValueOf(&program.Functions[0].Body).Elem(), func(value ir.Expression) ir.Expression {
						n, ok := value.(ir.Call)
						if !ok || program.Functions[n.Function].Name != "library_string_substring" || len(n.Arguments) != 3 {
							return value
						}
						if end, ok := n.Arguments[2].(ir.Coalesce); ok {
							end.Fallback = ir.NumberConstant{Value: 0}
							n.Arguments[2] = end
							changed = true
							return n
						}
						return value
					})
				}
			} else {
				for i := range program.Functions {
					if program.Functions[i].Name == "string_last_index_of_position" {
						mutateStringExpressions(reflect.ValueOf(&program.Functions[i].Body).Elem(), mutate)
					}
				}
			}
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
