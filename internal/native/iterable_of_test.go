package native

import (
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
