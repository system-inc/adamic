package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/notyet_generic_return.a",
		"internal/oracle/testdata/notyet_generic_optional_return.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

// Change only a concrete specialization's semantics, keeping its ABI valid. Node
// must catch these mutants even when native builds and sanitizers finish cleanly.
func TestNotYetGenericReturnsMutants(t *testing.T) {
	for _, rule := range []string{"return", "optional", "value"} {
		t.Run(rule, func(t *testing.T) {
			fixture := "notyet_generic_return.a"
			if rule == "optional" {
				fixture = "notyet_generic_optional_return.a"
			}
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for index := range program.Functions {
				function := &program.Functions[index]
				switch rule {
				case "return":
					if strings.HasPrefix(function.Name, "runWithoutCaching_") && function.Returns == ir.Number {
						function.Body = []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: -7}}}
						changed = true
					}
				case "optional":
					if strings.HasPrefix(function.Name, "withContext_") && function.Returns == ir.MaybeNumber {
						mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
							if maybe, ok := value.(ir.MaybeOf); ok && maybe.Value == nil {
								maybe.Value = ir.NumberConstant{Value: 23}
								changed = true
								return maybe
							}
							return value
						})
					}
				case "value":
					if strings.HasPrefix(function.Name, "withValue_") {
						mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
							if call, ok := value.(ir.CallClosure); ok && len(call.Arguments) == 1 && call.Arguments[0].Type() == ir.Number {
								call.Arguments[0] = ir.NumberConstant{Value: -7}
								changed = true
								return call
							}
							return value
						})
					}
				}
			}
			if !changed {
				t.Fatal("mutant target absent")
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
			t.Logf("%s mutant caught by Node stdout in both backends", rule)
		})
	}
}
