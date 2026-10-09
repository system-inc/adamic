package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/library_object_replaced.a", true, true})
	for _, path := range []string{"library_object_intrinsic", "library_object_coercion", "library_object_enumeration"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + path + ".a", true, false})
	}
}

// Every mutant still compiles, exits successfully, and finishes leak-clean. Node's original
// source is the independent authority; neither backend agreement nor runtime checks kills it.
func TestObjectStep22NodeOnlyMutants(t *testing.T) {
	t.Parallel()
	for _, family := range []string{"tag", "metadata", "own", "order", "coercion"} {
		t.Run(family, func(t *testing.T) {
			fixture := "intrinsic"
			if family == "order" {
				fixture = "enumeration"
			}
			if family == "coercion" {
				fixture = "coercion"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_object_"+fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			switch family {
			case "tag":
				for index, text := range program.Strings {
					if text == "[object Null]" {
						program.Strings[index] = "[object Wrong]"
						changed = true
					}
				}
			case "metadata":
				for index, function := range program.Functions {
					if function.Name == "object_static_own" {
						program.Functions[index].Body = []ir.Statement{ir.Return{Value: ir.BooleanConstant{Value: false}}}
						changed = true
						break
					}
				}
			case "own":
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
					if own, yes := value.(ir.HasOwn); yes && !changed {
						changed = true
						return ir.Unary{Operator: ir.Not, Operand: own}
					}
					return value
				})
			case "order":
				mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
					if object, yes := value.(ir.ObjectLiteral); yes && !changed && len(object.Fields) > 2 {
						object.Fields = append([]ir.Field(nil), object.Fields...)
						slices.Reverse(object.Fields)
						changed = true
						return object
					}
					return value
				})
			case "coercion":
				for index, function := range program.Functions {
					if function.Name != "object_primitive" || function.Returns != ir.Number || len(function.Body) != 1 {
						continue
					}
					returned, yes := function.Body[0].(ir.Return)
					if !yes {
						continue
					}
					convert, yes := returned.Value.(ir.NumberCall)
					if !yes {
						continue
					}
					call, yes := convert.Arguments[0].(ir.CallClosure)
					if !yes {
						continue
					}
					property, yes := call.Closure.(ir.Property)
					if !yes || property.Name != "valueOf" {
						continue
					}
					property.Name = "toString"
					call.Closure = property
					call.Returns = ir.String
					convert.Arguments = []ir.Expression{call}
					returned.Value = convert
					program.Functions[index].Body = []ir.Statement{returned}
					changed = true
					break
				}
			}
			if !changed {
				t.Fatal("mutant did not change the lowered program")
			}
			expected := onNode(t, path)
			actual, binary := natively(t, program)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: %+v", actual)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
			for backend, observation := range map[string]run{"native": actual, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(expected, observation); difference != "stdout differs" {
					t.Fatalf("%s: want stdout alone, got %q", backend, difference)
				}
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "stdout differs" {
					t.Fatalf("WASI: want stdout alone, got %q", difference)
				}
			}
			t.Log("Node alone caught mutant on native and JavaScript; WASI checked when enabled")
		})
	}
}
