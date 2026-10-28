package lower

import (
	"errors"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// An optional checked read reaches emission only in a shape both emitters can give undefined for
// (#1yr7rdh): an absent field through the union adapter when its union holds undefined, and a ?.
// read through the adapter or in place, but neither for a callable. Every other optional read stops
// at the read, never at an emitter's compiler bug. Mutants: dropping either new refusal, or the
// in-place Absent one, which lowers the read for an emitter that panics on it.
func TestOptionalCheckedReadsReachOnlyWhatEmissionCanSay(t *testing.T) {
	t.Parallel()
	union := ir.ViewContract{Kind: ir.ViewUnion, Name: "Identifier | undefined", Of: ir.Object, Undefined: true}
	for _, test := range []struct {
		name     string
		contract ir.ViewContract
		read     ir.Property
		emitted  bool
	}{
		{"absent through a union holding undefined", union, ir.Property{Of: ir.Union, Absent: true}, true},
		{"?. through a union", ir.ViewContract{Kind: ir.ViewUnion, Name: "Identifier | StringLiteral", Of: ir.Object}, ir.Property{Of: ir.Union, Optional: true}, true},
		{"?. in place", ir.ViewContract{Kind: ir.ViewScalar, Name: "number", Of: ir.Number}, ir.Property{Of: ir.Number, Optional: true}, true},
		{"absent through a union without undefined", ir.ViewContract{Kind: ir.ViewUnion, Name: "Identifier | StringLiteral", Of: ir.Object}, ir.Property{Of: ir.Union, Absent: true}, false},
		{"absent in place", ir.ViewContract{Kind: ir.ViewObject, Name: "Identifier", Of: ir.Object}, ir.Property{Of: ir.Object, Absent: true}, false},
		{"?. of a callable", ir.ViewContract{Kind: ir.ViewCallable, Name: "() => void", Of: ir.Closure, ProducerCertified: true}, ir.Property{Of: ir.Closure, Optional: true}, false},
		{"absent callable", ir.ViewContract{Kind: ir.ViewUnion, Name: "(() => void) | undefined", Of: ir.Closure, Undefined: true}, ir.Property{Of: ir.Closure, Absent: true}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			read := test.read
			read.Object = ir.Null{}
			read.Name = "name"
			read.View = "declaration.name"
			read.ViewWhere = "fixture.ts:1:1"
			read.ViewType = test.contract.Name
			read.ViewTypeID = 7
			read.ViewReceiverTypeID = 3
			program := &ir.Program{
				ViewContracts:     []ir.ViewContract{test.contract},
				ViewContractTypes: map[int]ir.ViewContractID{7: 1},
				CheckedFields:     map[string]bool{ir.CheckedFieldKey(3, "name"): true},
				Main:              []ir.Statement{ir.Evaluate{Value: read}},
			}
			err := unemittableCheckedViewRead(program)
			var notYet *NotYet
			if test.emitted && err != nil {
				t.Fatalf("want the read handed to emission, got %v", err)
			}
			if !test.emitted && (!errors.As(err, &notYet) || notYet.Where != "fixture.ts:1:1") {
				t.Fatalf("want a NotYet at the read, got %v", err)
			}
		})
	}
}
