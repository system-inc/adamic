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
func TestPrivateGenericMethodOwnerMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_private_members.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal("missing private method owners")
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
		t.Fatal("missing private base method call")
	}
	got, binary := natively(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must compile and finish cleanly: %+v", got)
	}
	want := onNode(t, path)
	if difference := disagreement(want, got); difference != "stdout differs" {
		t.Fatalf("private method owner mutant survived: %q", difference)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Logf("caught by Node output: source %q; mutant %q", want.stdout, got.stdout)
}
