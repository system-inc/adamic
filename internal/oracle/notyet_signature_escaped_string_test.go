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
	}{"internal/oracle/testdata/notyet_signature_escaped_string.a", true, false})
}

func TestEscapedStringReturnMutant(t *testing.T) {
	escapedStringReturnMutant(t, "notyet_signature_escaped_string.a", "escapedName", false)
}

func escapedStringReturnMutant(t *testing.T, fixture, name string, absent bool) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	wrong := len(program.Strings)
	program.Strings = append(program.Strings, "wrong")
	changed := false
	for i := range program.Functions {
		function := &program.Functions[i]
		if function.Name != name {
			continue
		}
		if function.Returns != ir.String {
			t.Fatalf("want string base signature, got %v", function.Returns)
		}
		mutateStringExpressions(reflect.ValueOf(&function.Body).Elem(), func(value ir.Expression) ir.Expression {
			if absent {
				if _, ok := value.(ir.Undefined); ok {
					changed = true
					return ir.StringConstant{Index: wrong}
				}
			} else {
				if property, ok := value.(ir.Property); ok && property.Of == ir.String && (property.Name == "Call" || property.Name == "New") {
					changed = true
					return ir.StringConstant{Index: wrong}
				}
			}
			return value
		})
	}
	if !changed {
		t.Fatal("missing branded return mutant target")
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
}
