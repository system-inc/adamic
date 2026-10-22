package lower

import (
	"errors"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A logical assignment through a checked view is refused even when its write records no site: the
// refusal asks CheckedField first, and the site only places it. Mutant: skipping writes whose Site is
// 0 lowers this one.
func TestViewLogicalWriteWithoutASiteIsRefused(t *testing.T) {
	object := ir.Read{Local: 0, Of: ir.Object}
	l := &lowering{result: &ir.Program{
		Source:        "main.a",
		Locals:        []ir.Local{{Name: "viewed", Type: ir.Object}},
		CheckedFields: ir.SortedMap[string, bool]{ir.CheckedFieldKey(7, "count"): true},
		Main: []ir.Statement{ir.Evaluate{Value: ir.LogicalAssignment{
			Operator: "&&=",
			Of:       ir.Number,
			Read:     ir.Property{Object: object, Name: "count", Of: ir.Number},
			Write:    ir.SetProperty{Object: object, Name: "count", Value: ir.NumberConstant{Value: 7}, ViewReceiverTypeID: 7},
		}}},
	}}
	var refused *Refused
	if err := l.checkViewLogicalWrites(); !errors.As(err, &refused) {
		t.Fatalf("want a refusal, got %v", err)
	}
	if refused.Where != "main.a" || refused.What != "a logical assignment through a checked view" {
		t.Fatalf("refused at %q with %q", refused.Where, refused.What)
	}
}
