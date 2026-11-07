package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A private method belongs to its declaring class even when called on a derived receiver.
// Both methods have the same ABI, so confusing their owners must be caught by Node output.
func TestPrivateGenericMethodOwnerMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/class_private_members.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	base, derived := -1, -1
	for index, function := range program.Functions {
		if strings.HasPrefix(function.Name, "Base_#tag@") {
			base = index
		}
		if strings.HasPrefix(function.Name, "Derived_#tag@") {
			derived = index
		}
	}
	if base < 0 || derived < 0 {
		test.Fatal("missing private method owners")
	}
	changed := false
	mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), func(value ir.Expression) ir.Expression {
		if call, yes := value.(ir.Call); yes && call.Function == base {
			call.Function = derived
			changed = true
			return call
		}
		return value
	})
	if !changed {
		test.Fatal("missing private base method call")
	}
	got, binary := natively(test, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		test.Fatalf("mutant must compile and finish cleanly: %+v", got)
	}
	want := onNode(test, path)
	if difference := disagreement(want, got); difference != "stdout differs" {
		test.Fatalf("private method owner mutant survived: %q", difference)
	}
	if report := leaks(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Logf("caught by Node output: source %q; mutant %q", want.stdout, got.stdout)
}
