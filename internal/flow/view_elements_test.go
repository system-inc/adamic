package flow

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestViewElementPreservesAliasing(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Locals: []ir.Local{{Type: ir.Array}}}
	n := &inference{function: &Function{Program: program, Identifiers: []*Identifier{nil, {Declaration: 1}}}, current: map[DeclarationId]IdentifierId{1: 1}, escaped: map[IdentifierId]bool{}}
	got := n.value(ir.ViewElement{Array: ir.Read{Local: 0, Of: ir.Array}, Index: ir.NumberConstant{}, Contract: 1, Of: ir.Object})
	if len(got.part) != 1 || got.part[0].Identifier != 1 || got.unknown || got.fresh || len(n.list) != 0 || len(n.escaped) != 0 {
		t.Fatalf("a check changed the value's aliases or introduced effects: %#v, effects %#v", got, n.list)
	}
	if writes(ir.ViewElement{}) || callsBack(ir.ViewElement{}) {
		t.Fatal("an element check writes program storage or calls user code")
	}
}
