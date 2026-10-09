package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestLibraryMethodAliasReadMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_method_values_dead_zone.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), func(value ir.Expression) ir.Expression {
		if read, yes := value.(ir.Read); yes && program.Locals[read.Local].Name == "search" && read.Checked {
			read.Checked = false
			changed = true
			return read
		}
		return value
	})
	if !changed {
		t.Fatal("mutant changed no checked method alias read")
	}
	result, binary := natively(t, program)
	source := onNode(t, path)
	if source.exitCode != 1 || result.exitCode != 0 || disagreement(source, result) != "exit codes differ" {
		t.Fatalf("want Node's temporal dead zone to catch an otherwise finishing mutant: node %d stderr %s; native %d stderr %s", source.exitCode, source.stderr, result.exitCode, result.stderr)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Log("caught by Node's uncaught ReferenceError, with exit 1")
}
