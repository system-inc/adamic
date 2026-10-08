package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/notyet_generic_explicit_returns.a", true, false})
}
func TestExplicitGenericReturnMutants(t *testing.T) {
	for _, rule := range []string{"nonNull", "default", "field", "watcher"} {
		t.Run(rule, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_generic_explicit_returns.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			wrong := len(program.Strings)
			program.Strings = append(program.Strings, "wrong")
			for i := range program.Functions {
				function := &program.Functions[i]
				if rule == "nonNull" && strings.HasPrefix(function.Name, "nonNull_") && function.Returns == ir.Number {
					function.Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: -7}}}
					changed = true
				}
				if rule == "default" && strings.HasPrefix(function.Name, "nonNullDefault_") && function.Returns == ir.Number {
					function.Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: -7}}}
					changed = true
				}
				if rule == "field" && strings.HasPrefix(function.Name, "field_") && function.Returns == ir.MaybeNumber {
					function.Body = []ir.Statement{ir.Return{Value: ir.MaybeOf{Of: ir.MaybeNumber, Value: ir.NumberConstant{Value: 19}}}}
					changed = true
				}
				if rule == "watcher" && function.Closure && function.Returns == ir.String {
					mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
						if _, ok := value.(ir.Concat); ok {
							changed = true
							return ir.StringConstant{Index: wrong}
						}
						return value
					})
				}
			}
			if !changed {
				t.Fatal("missing generic mutant target")
			}
			want := onNode(t, path)
			got, binary := natively(t, program)
			if got.exitCode != 0 || len(got.stderr) != 0 {
				t.Fatalf("mutant must finish cleanly: %+v", got)
			}
			if difference := disagreement(want, got); difference != "stdout differs" {
				t.Fatalf("native mutant survived: %q", difference)
			}
			if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "stdout differs" {
				t.Fatalf("JavaScript mutant survived: %q", difference)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// These calls have the same nullable ABI but different proven type arguments.
// The fixture must retain one specialization per concrete checker identity.
func TestExplicitNullableArgumentsHaveDistinctInstances(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_generic_explicit_returns.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	instances := 0
	for _, function := range program.Functions {
		if strings.HasPrefix(function.Name, "nullableIdentity_") {
			instances++
		}
	}
	if instances != 2 {
		t.Fatalf("fixture needs two distinct explicit instantiations, got %d", instances)
	}
}
