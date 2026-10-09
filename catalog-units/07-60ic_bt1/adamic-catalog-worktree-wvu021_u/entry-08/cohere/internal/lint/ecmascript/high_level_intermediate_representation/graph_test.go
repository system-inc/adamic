package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

// TestReversePostorderKeepsLoopBodyBeforeContinuation pins the ordering that mutable ranges read.
//
// React visits a structured terminal's fallthrough before its real successors while constructing
// postorder, then reverses the result. That puts a loop mutation before a read after the loop. A
// conventional DFS that visits the loop edge first produces another mathematically valid reverse
// postorder, but it puts the continuation first and widens every loop mutation through the rest of
// the function.
func TestReversePostorderKeepsLoopBodyBeforeContinuation(t *testing.T) {
	t.Parallel()

	function := lowerSource(t, `
		function f(items) {
			for (const item of items) {
				sink(item);
			}
			return done(items);
		}
	`)

	type location struct {
		block int
		order static_single_assignment.EvaluationOrder
	}
	locations := map[string]location{}
	for blockIndex, block := range function.Blocks {
		for _, instructionID := range block.Instructions {
			instruction := function.Instructions[instructionID]
			load, ok := instruction.Value.(*LoadGlobal)
			if !ok || (load.Name != "sink" && load.Name != "done") {
				continue
			}
			locations[load.Name] = location{block: blockIndex, order: instruction.Order}
		}
	}

	sink, hasSink := locations["sink"]
	done, hasDone := locations["done"]
	if !hasSink || !hasDone {
		t.Fatalf("fixture did not retain both calls: locations=%v\n%s", locations, Print(function))
	}
	if sink.block >= done.block || sink.order >= done.order {
		t.Fatalf("loop body must precede its continuation: sink=%+v done=%+v\n%s",
			sink, done, Print(function))
	}
}
