package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"testing"
)

// The Node control refines a boxed union into an object. Its constituent slots
// disagree, even when an earlier expression stop masks the storage proof.
func TestClassSetPropertyIntersectionSlotAgreement(t *testing.T) {
	program, err := load.Load([]string{"../oracle/testdata/class_set_property_intersection_boundary.a"})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{checker: checked}
	refined := checked.GetTypeAtLocation(file.Statements.Nodes[1].Name())
	if l.intersectionKeptSlot(refined, "name", ir.Object) {
		t.Fatal("boxed union and plain object constituents must not share a slot proof")
	}
}
