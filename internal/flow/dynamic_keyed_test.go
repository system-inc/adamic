package flow

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A descriptor selected at runtime can throw through the keyed read into the surrounding catch.
func TestDynamicKeyedGetterThrowEdge(t *testing.T) {
	t.Parallel()
	for _, throwing := range []bool{false, true} {
		read := ir.DynamicKeyedProperty{
			Object:  ir.Null{},
			Key:     ir.StringConstant{Index: 0},
			Reflect: true,
		}
		program := &ir.Program{
			Strings: []string{"answer"},
			Classes: []ir.Class{{Accessors: []ir.Accessor{{Name: "answer", Getter: 1, Setter: -1}}}},
			Functions: []ir.Function{
				{Body: []ir.Statement{ir.Try{
					HasCatch: true, CatchLocal: -1,
					Body:  []ir.Statement{ir.Evaluate{Value: read}},
					Catch: []ir.Statement{ir.Evaluate{Value: ir.NumberConstant{Value: 42}}},
				}}},
				{Returns: ir.Number, MayThrow: throwing},
			},
		}
		graph := Build(program, 0)
		var readBlock, catchBlock *BasicBlock
		for _, block := range graph.Blocks {
			for position, id := range block.Instructions {
				switch graph.Instructions[id].Expression.(type) {
				case ir.DynamicKeyedProperty:
					readBlock = block
					if throwing && position != len(block.Instructions)-1 {
						t.Fatal("the throwing read does not end its block")
					}
				case ir.NumberConstant:
					catchBlock = block
				}
			}
		}
		if readBlock == nil {
			t.Fatal("missing read instruction")
		}
		edge, hasThrow := readBlock.Terminal.(*MayThrow)
		if hasThrow != throwing {
			t.Fatalf("getter MayThrow=%t, read terminal %T", throwing, readBlock.Terminal)
		}
		if hasThrow && (catchBlock == nil || edge.Handler != catchBlock.Id) {
			t.Fatalf("keyed read has no edge to the catch: %+v", edge)
		}
	}
}
