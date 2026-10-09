package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	for _, fixture := range []struct {
		path            string
		lowers, checked bool
	}{
		{"internal/oracle/testdata/scout22_prototype_create.a", true, false},
		{"internal/oracle/testdata/scout22_cached_intrinsic.a", true, false},
		{"internal/oracle/testdata/scout22_prototype_refused/scout22_prototype_links.a", false, false},
		{"internal/oracle/testdata/scout22_prototype_refused/scout22_prototype_lifetimes.a", false, false},
	} {
		fixtures = append(fixtures, fixture)
	}
}
func TestScout22PrototypeLinks(t *testing.T) {
	original, _ := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout22_prototype_refused/scout22_prototype_links.a"))
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "main.ts")
	if err = os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, original)
	result, binary := natively(t, program)
	if difference := disagreement(expected, result); difference != "" {
		t.Fatal(difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(expected, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "" {
			t.Fatal(difference)
		}
	}
	changed := false
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
		if call, ok := value.(ir.ObjectCall); ok && call.Method == "setPrototypeOf" && !changed {
			call.Arguments[1] = ir.ObjectCall{Method: "intrinsicObjectPrototype", Returns: ir.Object}
			changed = true
			return call
		}
		return value
	})
	if !changed {
		t.Fatal("no prototype link to mutate")
	}
	scout22MutantMatchesOnlyNode(t, original, program)
}

func TestScout22PrototypeLifetimes(t *testing.T) {
	original, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout22_prototype_refused/scout22_prototype_lifetimes.a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "main.ts")
	if err = os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, original)
	result, binary := natively(t, program)
	if difference := disagreement(expected, result); difference != "" {
		t.Fatal(difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(expected, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatal(difference)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if difference := disagreement(expected, onWASI(t, native.C(program))); difference != "" {
			t.Fatal(difference)
		}
	}
}
