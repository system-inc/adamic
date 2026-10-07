package flow

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// The first body statement must not turn a loop header into function entry.
// Entry is also where the parameter's initial value is defined for SSA.
func TestBuildLeadingLoopKeepsEntryWithoutIncomingEdges(t *testing.T) {
	t.Parallel()
	for _, checkAfter := range []bool{false, true} {
		name := "while"
		if checkAfter {
			name = "do while"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			read := ir.Read{Local: 0, Of: ir.Number}
			program := &ir.Program{
				Locals: []ir.Local{{Name: "count", Type: ir.Number, Function: 0}},
				Functions: []ir.Function{{Name: "leadingLoop", Parameters: []int{0}, Returns: ir.Number, Body: []ir.Statement{
					ir.Loop{CheckAfter: checkAfter, Condition: ir.Binary{Operator: ir.Less, Left: read, Right: ir.NumberConstant{Value: 3}}, Body: []ir.Statement{
						ir.Assign{Local: 0, Value: ir.Binary{Operator: ir.Add, Left: read, Right: ir.NumberConstant{Value: 1}}},
					}},
					ir.Return{Value: read},
				}}},
			}
			graph := Build(program, 0)
			entry, ok := graph.Block(graph.Entry)
			if !ok {
				t.Fatal("missing function entry")
			}
			for _, block := range graph.Blocks {
				EachSuccessor(block.Terminal, func(target BlockId) {
					if target == graph.Entry {
						t.Fatalf("block %d has an incoming edge to entry %d", block.Id, graph.Entry)
					}
				})
			}
			if len(entry.Predecessors) != 0 {
				t.Fatalf("entry predecessors: %v", entry.Predecessors)
			}
			if len(entry.Instructions) != 0 {
				t.Fatal("leading loop reused entry for body instructions")
			}
			jump, ok := entry.Terminal.(*Goto)
			if !ok {
				t.Fatalf("entry terminal %T, want a preheader jump", entry.Terminal)
			}
			first, ok := graph.Block(jump.Block)
			if !ok || first.Id == entry.Id {
				t.Fatal("entry must jump to a separate loop block")
			}
			if len(first.Predecessors) != 2 {
				t.Fatalf("loop entry predecessors: %v", first.Predecessors)
			}
			reaching := reachingDefinitions(graph)
			entryId := graph.Entry
			Construct(graph)
			if graph.Entry != entryId {
				t.Fatal("SSA replaced the fresh entry")
			}
			if violations := VerifySSA(graph); len(violations) != 0 {
				t.Fatalf("invalid loop SSA: %v", violations)
			}
			checkReaching(t, name, graph, reaching)
			if CollectSSAStats(graph).Phis == 0 {
				t.Fatal("loop parameter update produced no phi")
			}
		})
	}
}
