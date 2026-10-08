package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	for _, name := range []string{"never", "context", "specialization", "branded"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/notyet_map_" + name + ".a", true, false})
	}
}
func TestMapKeyMutants(t *testing.T) {
	for _, name := range []string{"never", "context", "specialization", "branded"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_map_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			mutate := func(value ir.Expression) ir.Expression {
				if name == "never" {
					if m, ok := value.(ir.MapSize); ok {
						changed = true
						_ = m
						return ir.NumberConstant{Value: 1}
					}
				}
				if name != "never" {
					if m, ok := value.(ir.MapSet); ok && m.ValueType == ir.Number {
						changed = true
						m.Value = ir.NumberConstant{Value: -7}
						return m
					}
				}
				return value
			}
			mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
			for i := range program.Functions {
				mutateStringExpressions(reflect.ValueOf(&program.Functions[i].Body).Elem(), mutate)
			}
			if !changed {
				t.Fatal("missing mutant target")
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

func TestMapExplicitAnyRefused(t *testing.T) {
	for _, name := range []string{"explicit_any", "explicit_any_generic"} {
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_map_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			want := "0\n"
			if name == "explicit_any_generic" {
				want = "0 0\n"
			}
			if truth.exitCode != 0 || string(truth.stdout) != want {
				t.Fatalf("Node baseline: %+v", truth)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || refused.What != "an explicit any Map key" {
				t.Fatalf("want explicit any adaptation refusal, got %v", err)
			}
		})
	}
}
