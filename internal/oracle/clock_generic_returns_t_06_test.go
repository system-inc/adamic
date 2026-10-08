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
	}{"internal/oracle/testdata/clock_generic_returns_t_06.a", true, false})
}

func TestClockGenericReturnsT06Mutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/clock_generic_returns_t_06.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	// Mutate the proven IR so the wrong discriminant is a semantic error,
	// rather than a TypeScript excess-property diagnostic in the source.
	changed := 0
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), func(value ir.Expression) ir.Expression {
		object, ok := value.(ir.ObjectLiteral)
		if !ok {
			return value
		}
		for i, field := range object.Fields {
			literal, ok := field.Value.(ir.StringConstant)
			if field.Name == "kind" && ok && program.Strings[literal.Index] == "number" {
				object.Fields[i].Value = ir.StringConstant{Index: len(program.Strings)}
				program.Strings = append(program.Strings, "wrong-number")
				changed++
			}
		}
		return object
	})
	if changed != 1 {
		t.Fatalf("want one numeric kind literal, got %d", changed)
	}
	want := onNode(t, path)
	native, binary := natively(t, program)
	for name, got := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Fatalf("%s mutant must execute cleanly: %+v", name, got)
		}
		if difference := disagreement(want, got); difference != "stdout differs" {
			t.Fatalf("%s: want stdout disagreement, got %q", name, difference)
		}
		t.Logf("clock_generic_returns_t_06_wrong_member_kind %s: Node %q, mutant %q", name, want.stdout, got.stdout)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
