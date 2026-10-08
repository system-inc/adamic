package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/notyet_signature_pattern_default.a", true, false})
}

func TestPatternDefaultMutants(t *testing.T) {
	for _, rule := range []string{"always-default", "wrong-field"} {
		t.Run(rule, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_signature_pattern_default.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for i := range program.Functions {
				function := &program.Functions[i]
				if function.Name != "render" {
					continue
				}
				mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
					if rule == "always-default" {
						if selected, ok := value.(ir.Coalesce); ok && selected.Of == ir.Object {
							selected.Value = ir.Undefined{Of: ir.Object}
							changed = true
							return selected
						}
					}
					if rule == "wrong-field" {
						if property, ok := value.(ir.Property); ok && property.Name == "prefix" {
							property.Name = "suffix"
							changed = true
							return property
						}
					}
					return value
				})
			}
			if !changed {
				t.Fatal("missing parameter mutant target")
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
