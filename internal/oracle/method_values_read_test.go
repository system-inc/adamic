package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestLibraryMethodAliasReadMutant(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_method_values_dead_zone.a"))
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
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
		test.Fatal("mutant changed no checked method alias read")
	}
	result, binary := natively(test, program)
	source := onNode(test, path)
	if source.exitCode != 70 || result.exitCode != 0 || disagreement(source, result) != "exit codes differ" {
		test.Fatalf(
			"want Node's temporal dead zone to catch an otherwise finishing mutant: node %d stderr %s; native %d stderr %s",
			source.exitCode,
			source.stderr,
			result.exitCode,
			result.stderr,
		)
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Log("caught by Node's uncaught ReferenceError, with the exit-70 convention unchanged")
}
