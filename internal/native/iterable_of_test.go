package native

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A generator seen as an object through IterableOf is never a fresh object a region can hold, so no
// region plan claims a frame's storage as an object's (Memory's condition on #qk8rztp S2c-1). Mutant:
// freshValue's IterableOf arm answering true, and this goes red.
func TestIterableOfIsNeverFresh(t *testing.T) {
	t.Parallel()
	plan := &regionPlan{program: &ir.Program{}}
	if plan.freshValue(ir.IterableOf{Value: ir.Call{Returns: ir.Generator}}) {
		t.Fatal("a generator seen as an object is fresh to the region plan")
	}
}

// Narrow's checked Object arm admits an object and nothing else, for every caller: an iterable
// view's union arm reads a generator through a plain cast and its own dispatch (lower's
// dispatchView, #qk8rztp S2c-1 commit 3), never by widening this test (Lower's (e)). The proven
// programs that reach a checked Object unbox never hand it a generator, so no oracle fixture can see
// it widened; this reads the C it emits. Mutant: the Object arm's shape test made to let a generator
// through, and this goes red.
func TestCheckedObjectUnboxAdmitsOnlyAnObject(t *testing.T) {
	t.Parallel()
	code := C(&ir.Program{Source: "unbox.a", Locals: []ir.Local{
		{Name: "held", Type: ir.Union, Function: -1},
		{Name: "unboxed", Type: ir.Object, Function: -1},
	}, Main: []ir.Statement{
		ir.Declare{Local: 0},
		ir.Declare{Local: 1, Value: ir.Narrow{Value: ir.Read{Local: 0, Of: ir.Union}, To: ir.Object, Checked: true, Message: "checked unbox"}},
	}})
	tests := 0
	for _, line := range strings.Split(code, "\n") {
		if !strings.Contains(line, "adamic_record_is(") || !strings.Contains(line, "adamic_panic(") {
			continue
		}
		tests++
		if !strings.Contains(line, "->kind != adamic_kind_object") || strings.Contains(line, "generator") {
			t.Errorf("the checked Object unbox's test admits more than an object: %s", strings.TrimSpace(line))
		}
	}
	if tests != 1 {
		t.Fatalf("%d checked Object unbox tests in the C, want 1:\n%s", tests, code)
	}
}
