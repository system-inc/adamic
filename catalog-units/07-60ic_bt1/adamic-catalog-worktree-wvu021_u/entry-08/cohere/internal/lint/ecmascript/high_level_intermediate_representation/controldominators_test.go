package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"sort"
	"testing"
)

// ControlDominators finds the branch a block is control-dependent on.
//
// The question `set-state-in-effect` asks: a setter inside `if (reference.current !== value)` is
// exempt upstream because a ref decides whether the call happens at all. Answering it needs the
// post-dominator frontier rather than the post-dominator chain `UnconditionalBlocks` walks, and the
// two are easy to confuse because they read the same tree.
func TestControlDominatorsFindsTheDecidingBranch(t *testing.T) {
	t.Parallel()

	function := lowerSource(t, `function Component(x, flag) {
  if (flag) {
    sink(1);
  }
  return x;
}`)
	if function == nil {
		t.Fatal("the source did not lower")
	}

	// Every branching terminal counts, so the guarded block is controlled and the others are not.
	controlled := ControlDominators(function, func(Place) bool { return true })

	var controlledBlocks, freeBlocks []static_single_assignment.BlockId
	for _, block := range function.Blocks {
		if controlled(block.Id) {
			controlledBlocks = append(controlledBlocks, block.Id)
		} else {
			freeBlocks = append(freeBlocks, block.Id)
		}
	}
	sort.Slice(controlledBlocks, func(a, b int) bool { return controlledBlocks[a] < controlledBlocks[b] })

	if len(controlledBlocks) == 0 {
		t.Fatal("no block is control dependent on the `if`, so the frontier found nothing")
	}
	if len(freeBlocks) == 0 {
		t.Fatal("every block reads as controlled, which a predicate returning true for the entry " +
			"would also produce; the entry and the join must not be control dependent")
	}
	t.Logf("controlled=%v free=%v", controlledBlocks, freeBlocks)
}

// A predicate that matches no test exempts nothing.
//
// The half that makes the test above mean something. `ControlDominators` takes a predicate over the
// branch's test, and a version that ignored it would report every guarded block as controlled.
func TestControlDominatorsConsultsTheTest(t *testing.T) {
	t.Parallel()

	function := lowerSource(t, `function Component(x, flag) {
  if (flag) {
    sink(1);
  }
  return x;
}`)
	if function == nil {
		t.Fatal("the source did not lower")
	}

	never := ControlDominators(function, func(Place) bool { return false })
	for _, block := range function.Blocks {
		if never(block.Id) {
			t.Errorf("bb%d reads as controlled while no test satisfies the predicate", block.Id)
		}
	}
}

// A function with no branches has nothing to be control dependent on.
//
// Straight-line code is the case a frontier implementation gets wrong by returning the entry block
// for everything, which would exempt every setter in the codebase.
func TestControlDominatorsIsEmptyWithoutBranches(t *testing.T) {
	t.Parallel()

	function := lowerSource(t, `function Component(x) {
  sink(1);
  return x;
}`)
	if function == nil {
		t.Fatal("the source did not lower")
	}

	controlled := ControlDominators(function, func(Place) bool { return true })
	for _, block := range function.Blocks {
		if controlled(block.Id) {
			t.Errorf("bb%d is control dependent in a function with no branching terminal", block.Id)
		}
	}
}

// A nil function yields a predicate rather than a panic.
func TestControlDominatorsIsNilSafe(t *testing.T) {
	t.Parallel()

	if ControlDominators(nil, func(Place) bool { return true })(static_single_assignment.BlockId(1)) {
		t.Error("a nil function reported a controlled block")
	}
	function := lowerSource(t, "function Component(x) { return x; }")
	if ControlDominators(function, nil)(static_single_assignment.BlockId(1)) {
		t.Error("a nil predicate reported a controlled block")
	}
}
