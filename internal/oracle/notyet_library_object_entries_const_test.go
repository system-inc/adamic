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
	}{"internal/oracle/testdata/notyet_library_object_entries_const.a", true, false})
}

func TestNotYetLibraryObjectEntriesConstMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/notyet_library_object_entries_const.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), func(value ir.Expression) ir.Expression {
		literal, ok := value.(ir.ObjectLiteral)
		if !ok || changed || len(literal.Fields) < 2 {
			return value
		}
		literal.Fields[0], literal.Fields[1] = literal.Fields[1], literal.Fields[0]
		changed = true
		return literal
	})
	if !changed {
		t.Fatal("mutant changed no object")
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
	t.Log("field-order mutant caught by Node stdout comparison")
}
