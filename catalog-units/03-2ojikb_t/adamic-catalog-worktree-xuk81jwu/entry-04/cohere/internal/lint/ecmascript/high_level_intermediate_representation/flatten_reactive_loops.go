// Flattening reactive loops: a scope inside a loop is pruned, deliberately, before anything reads it.
//
// This is React's `flattenReactiveLoopsHIR`, run over this graph. oxc transcribes it at
// `oxc_react_compiler/src/react_compiler_inference/flatten_reactive_loops_hir.rs`; where the two
// disagree React wins, and each disagreement is recorded at the line that resolves it.
//
// # What it does, in upstream's own words
//
// "Prunes any reactive scopes that are within a loop (for, while, etc). We don't yet support
// memoization within loops because this would require an extra layer of reconciliation (plus a way
// to identify values across runs, similar to how we use `key` in JSX for lists). Eventually we may
// integrate more deeply into the runtime so that we can do a single level of reconciliation, but
// for now we've found it's sufficient to memoize *around* the loop."
//
// So a pruned scope here is a POLICY decision rather than a failure. That distinction is the whole
// reason this pass has to exist separately from the prunes that run later, and it is what the
// memoization validator reads: `visitFinishMemoize` skips a marker whose scope was pruned, because
// a memoization the compiler deliberately declined is not a memoization the developer lost.
//
// # Why its absence produced 287 false positives
//
// Without this pass a scope inside a loop still ends up pruned, just through a different door and
// three passes later: `PruneUnusedScopes` finds no declaration that originated in it and prunes it
// as unused. The end state looks identical and is not, because the marker was never flagged. The
// validator then reads a scope that is absent from the live set with an unflagged marker, which is
// exactly its definition of a lost memoization, and reports one.
//
// Measured on the ahra tree before this landed: 287 `preserve-manual-memoization` findings against
// eslint-plugin-react-hooks 7.1.1's zero, every one of them a memo that builds a collection in a
// loop. `babel-plugin-react-compiler` 1.0.0 over the same source emits a live memo block caching on
// the loop's own dependency, so the reference both prunes the inner scope and keeps the outer one.
//
// # Ordering
//
// Upstream runs this immediately after `buildReactiveScopeTerminalsHIR` and before
// `propagateScopeDependenciesHIR`, which is before every prune. That position is load-bearing in
// one direction: the pass reads `Scope` terminals, so it cannot run before the pass that creates
// them, and the dependency collector must see the pruned flag rather than discover the scope alive.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// FlattenReactiveLoops marks every reactive scope that begins inside a loop as pruned.
//
// Returns the scope ids it flattened, so a caller can assert the pass did something rather than
// trusting that it ran. The set is also the input `BuildReactiveFunction` reads to mark the blocks
// it emits, because this IR has no `PrunedScope` terminal to convert a `Scope` into: upstream
// rewrites the terminal in place, and the equivalent here is to carry the decision forward to the
// point where the reactive block is built. Recording ids rather than rewriting the graph keeps the
// terminal set at 21 variants and keeps every other pass reading the same shape it does today.
func FlattenReactiveLoops(function *Function) map[ScopeId]bool {
	if function == nil {
		return nil
	}

	flattened := map[ScopeId]bool{}

	// The blocks a loop's body spans, tracked by the block where control resumes after the loop.
	//
	// Upstream keys the same set on the loop's fallthrough and drops an entry when the walk reaches
	// that block, which is `retainWhere` in the TypeScript and `retain` in the Rust. The effect is a
	// scoped "we are inside a loop right now" flag that costs no dominator computation: a loop's
	// fallthrough is by construction the first block after the whole construct, so arriving there is
	// the same event as leaving the loop.
	activeLoops := map[static_single_assignment.BlockId]bool{}
	activeLoopCount := 0

	for _, block := range function.Blocks {
		if block == nil {
			continue
		}

		// Leaving a loop before reading this block's terminal, matching upstream's order. A loop
		// whose fallthrough is this block has ended, so a scope beginning here is outside it.
		if activeLoops[block.Id] {
			delete(activeLoops, block.Id)
			activeLoopCount--
		}

		switch terminal := block.Terminal.(type) {
		case *While:
			activeLoops[terminal.Fallthrough] = true
			activeLoopCount++
		case *DoWhile:
			activeLoops[terminal.Fallthrough] = true
			activeLoopCount++
		case *For:
			activeLoops[terminal.Fallthrough] = true
			activeLoopCount++
		case *ForOf:
			activeLoops[terminal.Fallthrough] = true
			activeLoopCount++
		case *ForIn:
			activeLoops[terminal.Fallthrough] = true
			activeLoopCount++
		case *Scope:
			// The guard is "any loop is active", not "this scope is inside a specific loop".
			// Upstream's is the same and the looseness is deliberate: the walk is in block order, so
			// every block between a loop's terminal and its fallthrough belongs to that loop's body.
			if activeLoopCount > 0 {
				flattened[terminal.Scope] = true
			}
		}
	}

	if len(flattened) == 0 {
		return nil
	}
	return flattened
}
