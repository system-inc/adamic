package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

// These formerly runnable marker controls are now rejected by the later ruling.
var impossibleAssertionFixtures = []string{
	"scanner_initialized", "static_uninitialized", "uninitialized_catch", "uninitialized_field", "uninitialized_map_entry", "initialized", "literal_statement", "static_initialized", "literal_return", "uninitialized_append", "uninitialized_default", "uninitialized_iteration", "uninitialized_spread", "uninitialized_const", "uninitialized_interface", "uninitialized_optional",
}

func impossibleAssertionFixture(name string) bool {
	for _, refused := range impossibleAssertionFixtures {
		if name == refused {
			return true
		}
	}
	return false
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/non_null_possible_read.a", true, true})
}

func TestImpossibleNonNullFixturesAreRefused(t *testing.T) {
	t.Parallel()
	paths := []string{}
	for _, name := range impossibleAssertionFixtures {
		paths = append(paths, "non_null_"+name+".a")
	}
	for _, extension := range []string{"a", "ts"} {
		for _, operand := range []string{"undefined", "null"} {
			paths = append(paths, "non_null_refuse_"+operand+"."+extension)
		}
	}
	for _, form := range []string{"argument", "return", "expression", "comparison"} {
		paths = append(paths, "non_null_refuse_"+form+".a")
	}
	for _, name := range paths {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want impossible assertion Refused, got %v", err)
			}
			if refused.What != "a non-null assertion whose operand is exactly undefined" && refused.What != "a non-null assertion whose operand is exactly null" {
				t.Fatalf("wrong refusal: %v", err)
			}
			if refused.Fix != "declare the variable optional and assign undefined" {
				t.Fatalf("wrong fix: %v", err)
			}
			if !strings.HasSuffix(refused.Error(), ": Adamic 0.1 refuses "+refused.What+"; declare the variable optional and assign undefined") {
				t.Fatalf("wrong diagnostic: %v", err)
			}
		})
	}
}

func TestPossibleNonNullStopsAtAssertion(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_possible_read.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	checks, ready := 0, 0
	mutateReadiness(program.Main, func(node any) any {
		switch value := node.(type) {
		case ir.Coalesce:
			if value.Panic != nil {
				checks++
			}
		case ir.Read:
			if value.Readiness != "" {
				ready++
			}
		case ir.Property:
			if value.Readiness != "" {
				ready++
			}
		}
		return node
	})
	if checks != 1 || ready != 0 {
		t.Fatalf("want one eager non-null check, no readiness check: checks=%d ready=%d", checks, ready)
	}
	want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: non-null assertion failed: values.get('missing')! is null or undefined\n"), exitCode: 70}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatal(difference)
		}
	}
	path, err = filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/non_null_possible_typeerror.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err = lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ = nativelyUncached(t, program)
	if difference := disagreement(want, native); difference != "" {
		t.Fatal(difference)
	}
	node := onNode(t, path)
	if node.exitCode == 0 || string(node.stdout) != "before\n" || !strings.Contains(string(node.stderr), "TypeError") {
		t.Fatalf("want Node's missing-value TypeError after before: %#v", node)
	}
}
