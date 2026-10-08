package native

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestForInStructuralSlotProof(t *testing.T) {
	read := ir.Read{Local: 0, Of: ir.Object}
	for _, property := range []bool{false, true} {
		program := &ir.Program{Locals: []ir.Local{{Type: ir.Object}, {Type: ir.Object}}, Main: []ir.Statement{
			ir.Declare{Local: 1, Value: read},
			ir.Evaluate{Value: ir.ObjectKeys{Object: ir.Read{Local: 1, Of: ir.Object}, Enumeration: true}},
		}}
		if property {
			program.Main = append(program.Main, ir.Evaluate{Value: ir.Property{Object: read, Name: "length", Of: ir.Number}})
		}
		emitter := &emitter{program: program}
		if got := emitter.forInOnly(0, map[int]bool{}); got == property {
			t.Fatalf("property use %t: enumeration-only proof %t", property, got)
		}
	}
}

func TestForInStructuralSlotVirtualTargets(t *testing.T) {
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Object}, {Type: ir.Object}, {Type: ir.Object}},
		Main:   []ir.Statement{ir.Evaluate{Value: ir.Call{Function: 0, Virtual: 1, Arguments: []ir.Expression{ir.Read{Local: 0, Of: ir.Object}}}}},
		Functions: []ir.Function{
			{Parameters: []int{1}, Body: []ir.Statement{ir.Evaluate{Value: ir.ObjectKeys{Object: ir.Read{Local: 1, Of: ir.Object}, Enumeration: true}}}},
			{Parameters: []int{2}, Body: []ir.Statement{ir.Evaluate{Value: ir.Property{Object: ir.Read{Local: 2, Of: ir.Object}, Name: "length", Of: ir.Number}}}},
		},
		MethodTargets: map[int][]int{0: {0, 1}},
	}
	emitter := &emitter{program: program}
	if emitter.forInOnly(0, map[int]bool{}) {
		t.Fatal("an override reads object fields, so the slot must keep object storage")
	}
}
