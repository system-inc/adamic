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

func TestNotYetLibraryObjectEntriesConstMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/notyet_library_object_entries_const.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
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
		test.Fatal("mutant changed no object")
	}
	result, binary := natively(test, program)
	if result.exitCode != 0 {
		test.Fatalf("mutant failed outside comparison: %s", result.stderr)
	}
	if difference := disagreement(onNode(test, path), result); difference != "stdout differs" {
		test.Fatalf("want stdout mismatch, got %q", difference)
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Log("field-order mutant caught by Node stdout comparison")
}
