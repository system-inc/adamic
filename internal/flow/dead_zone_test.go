package flow

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A write of a global in its temporal dead zone throws once its value has run (#bxmdk73), so the
// instruction holding the Assign ends its block with a MayThrow edge to the try's handler, and SSA
// still verifies: the handler sees the state before the instruction's define, as every throw edge
// does (#2yz9ra9). The value here reads kept, which the try's first statement wrote, so it runs before
// the store throws. A global that's hoisted has no dead zone, and the same store gets no edge.
// Mutant: CanThrow's ir.Assign branch dropped, and the dead-zone store gets no edge.
func TestDeadZoneStoreThrowsAfterItsValue(test *testing.T) {
	test.Parallel()
	build := func(hoisted bool) (*Function, *Instruction) {
		program := &ir.Program{
			Locals: []ir.Local{
				{Name: "later", Type: ir.Number, Global: true, Uninitialized: true, Hoisted: hoisted, Function: -1},
				{Name: "kept", Type: ir.Number, Function: 0},
			},
			Functions: []ir.Function{{Name: "fill", Returns: ir.Number, Body: []ir.Statement{
				ir.Declare{Local: 1, Value: ir.NumberConstant{Value: 1}},
				ir.Try{
					HasCatch:   true,
					CatchLocal: -1,
					Body: []ir.Statement{
						ir.Assign{Local: 1, Value: ir.NumberConstant{Value: 2}},
						ir.Assign{Local: 0, Value: ir.Read{Local: 1, Of: ir.Number}},
					},
					Catch: []ir.Statement{ir.Return{Value: ir.Read{Local: 1, Of: ir.Number}}},
				},
				ir.Return{Value: ir.Read{Local: 1, Of: ir.Number}},
			}}},
		}
		graph := Build(program, 0)
		Construct(graph)
		for _, instruction := range graph.Instructions {
			if assign, ok := (*instruction.At).(ir.Assign); ok && assign.Local == 0 {
				return graph, instruction
			}
		}
		test.Fatal("no instruction for the store to later")
		return nil, nil
	}
	endsInThrow := func(graph *Function, instruction *Instruction) bool {
		for _, block := range graph.Blocks {
			for position, id := range block.Instructions {
				if graph.Instructions[id] != instruction {
					continue
				}
				_, throws := block.Terminal.(*MayThrow)
				return throws && position == len(block.Instructions)-1
			}
		}
		test.Fatal("the store's instruction is in no block")
		return false
	}
	graph, store := build(false)
	if !endsInThrow(graph, store) {
		test.Fatal("a store in the dead zone has no throw edge after it")
	}
	if violations := VerifySSA(graph); len(violations) > 0 {
		test.Fatalf("SSA: %v", violations)
	}
	if hoisted, store := build(true); endsInThrow(hoisted, store) {
		test.Fatal("a store to a hoisted global has a throw edge")
	}
}
