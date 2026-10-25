package lower

import (
	"errors"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// viewLogicalWrite is a program whose one statement is viewed.count &&= 7 through a view that checks
// count, with the computed key and the optional field the refusal still holds back left to its caller.
func viewLogicalWrite(key ir.Expression, absent bool) *lowering {
	object := ir.Read{Local: 0, Of: ir.Object}
	assignment := ir.LogicalAssignment{
		Operator: "&&=",
		Of:       ir.Number,
		Read:     ir.Property{Object: object, Name: "count", Of: ir.Number, Absent: absent},
		Write:    ir.SetProperty{Object: object, Name: "count", Value: ir.NumberConstant{Value: 7}, ViewReceiverTypeID: 7},
	}
	if key != nil {
		assignment.Key, assignment.KeyNames = key, []string{"count"}
		assignment.Read = ir.Property{Object: object, Of: ir.Number, Absent: true}
		assignment.Write = ir.SetProperty{Object: object, Value: ir.NumberConstant{Value: 7}, ViewReceiverTypeID: 7}
	}
	return &lowering{result: &ir.Program{
		Source:        "main.a",
		Locals:        []ir.Local{{Name: "viewed", Type: ir.Object}, {Name: "key", Type: ir.String}},
		CheckedFields: ir.SortedMap[string, bool]{ir.CheckedFieldKey(7, "count"): true},
		Main:          []ir.Statement{ir.Evaluate{Value: assignment}},
	}}
}

// A logical assignment through a checked view by a computed key is refused even when its write records
// no site: the refusal asks CheckedField of each name the key can be first, and the site only places it.
// Mutant: skipping writes whose Site is 0 lowers this one.
func TestViewLogicalWriteWithoutASiteIsRefused(t *testing.T) {
	t.Parallel()
	var refused *Refused
	if err := viewLogicalWrite(ir.Read{Local: 1, Of: ir.String}, false).checkViewLogicalWrites(); !errors.As(err, &refused) {
		t.Fatalf("want a refusal, got %v", err)
	}
	if refused.Where != "main.a" || refused.What != "a logical assignment through a checked view" {
		t.Fatalf("refused at %q with %q", refused.Where, refused.What)
	}
}

// By name, a logical assignment through a checked view lowers, since both emitters read and write it as
// a plain read and = through the view do, and an optional field is still refused, since the checked read
// has no form for one. Mutants: dropping the early return lowers neither, and dropping its Absent test
// lowers both.
func TestViewLogicalWriteByNameLowersUnlessOptional(t *testing.T) {
	t.Parallel()
	if err := viewLogicalWrite(nil, false).checkViewLogicalWrites(); err != nil {
		t.Fatalf("a required field by name: want it lowered, got %v", err)
	}
	var refused *Refused
	if err := viewLogicalWrite(nil, true).checkViewLogicalWrites(); !errors.As(err, &refused) {
		t.Fatalf("an optional field: want a refusal, got %v", err)
	}
}
