// Single static assignment form: rename every definition so each value is written exactly once,
// and insert phi nodes where control from several predecessors rejoins.
//
// This is Braun et al., "Simple and Efficient Construction of Static Single Assignment Form"
// (CC 2013), in the sealed-block form upstream uses. React's own specification of the pass is at
// `compiler/packages/babel-plugin-react-compiler/docs/passes/02-enterSSA.md`, and the shape here
// follows it closely enough that the two can be read side by side.
//
// The algorithm is static_single_assignment.Construct, which this IR shares with Adamic's flow graph,
// with its reasoning: why Braun rather than Cytron, the reverse postorder it rests on, and loops.
// What stays here is what is this IR's own.
//
// # Why this does not consume control_flow_graph's dominators
//
// Braun needs no dominance frontier at all, which is why this does not consume
// `internal/lint/ecmascript/control_flow_graph`'s dominator tree. `AnalyzeDominators` is generic over `control_flow_graph.Graph[E]` and `control_flow_graph.Block[E]`, which are
// that package's own types; this graph is `Function` and `BasicBlock`. There is no conversion, and
// writing one would mean materialising a second graph purely to compute a frontier this algorithm
// does not use. Dominance is still what CORRECTNESS is stated against - see `VerifySSA` - but it is
// computed there, over this graph, for checking rather than for construction.
//
// Stated plainly for the next pass built on this IR, because it is easy to plan around the wrong
// answer: NOTHING in `internal/lint/ecmascript/control_flow_graph` applies to this graph without a conversion that
// does not exist. Its dominators, its dataflow solver, and its path analysis are all generic over
// `Graph[E]`. A pass here that wants any of them either writes the conversion or, as this does,
// computes what it needs over `Function.Blocks`, which is already in reverse postorder.
//
// # The doubled finally, and why it does not reach here
//
// `controlflow` lays a `finally` body out TWICE, and three consumers have been bitten by it. For
// this pass it would be a correctness bug rather than a nuisance: two layouts of one body give a
// binding two definitions that are not a merge, and construction over that mints a phi where the
// source has no join.
//
// It does not reach here, and that was verified rather than assumed. Lowering emits the finally
// body ONCE and jumps to it from every path that must run it; see `lowerTryStatement`. Checked
// empirically on `try/finally` and `try/catch/finally`, both of which produce a single finally
// block whose predecessors are the normal and abrupt paths. That is a real join, and a phi placed
// there is correct.
//
// # What is deliberately not here
//
// Redundant-phi elimination is a separate pass, `EliminateRedundantPhis`, which construction runs.
// Reclassifying `const`/`let` after renaming - upstream's
// `rewrite_instruction_kinds_based_on_reassignment` - is NOT implemented, because nothing in cohere
// reads `InstructionKind` yet and a reclassification no pass consumes is untested by construction.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// Construct converts a function to single static assignment form, in place, and recursively for
// every nested function.
//
// After it returns: every identifier that a source binding takes is written exactly once, every use
// names the definition that actually reaches it, and `BasicBlock.Phis` holds a phi wherever a
// binding's value depends on which predecessor control arrived from.
//
// # What this cannot do for you, and how to tell
//
// A variable reference lowering left as `LoadGlobal` carries a name rather than a value, so there
// is nothing here to rename and no phi to place for it. That is correct for a true global and wrong
// for a local, and lowering produces the latter for EVERY reference when it was given no type
// checker. `Construct` cannot distinguish the two cases. `Lower`'s comment carries the measurement;
// the short version is that SSA over an unresolved function is well-formed, empty of phis, and
// meaningless.
//
// # Captures, and what renaming does and does not do at a function boundary
//
// A variable a nested function closes over IS tracked: lowering populates `FunctionExpression.
// Captures` and `Function.Context`, and the nested body reads it as `LoadContext` rather than
// `LoadGlobal`. `Construct` recurses into nested functions, so both sides are in single-assignment
// form.
//
// What renaming does NOT do is unify the two sides, and a pass must not assume it did. The capture
// and the captured value are separate identifiers in separate tables, because a `Place` is only
// meaningful against the table of the function holding it. The edge between them is positional:
// `FunctionExpression.Captures[i]` is the enclosing function's value and `nested.Context[i]` is the
// nested function's, naming one source binding from the two sides. A pass following a value into a
// closure walks that pairing.
//
// The capture is bound at the point the closure is CREATED, which is what makes it a value rather
// than a name. For `let n = 1; if (c) { n = 2; } const g = () => n;` the capture names the phi that
// merges the two versions, so `g` closes over what actually reaches it. A later write through
// `StoreContext` is a definition inside the nested function and is versioned there; it is NOT
// reflected back into the enclosing function's versions, so a pass asking what the OUTER `n` holds
// after calling `g` must still treat that as unknown.
func Construct(function *Function) {
	if function == nil {
		return
	}

	// Re-establish the invariants rather than trusting them. Reverse postorder is the property the
	// whole algorithm rests on and it is cheap to recompute; a caller who restructured the graph and
	// forgot would otherwise get a wrong answer with no symptom.
	Finalize(function)

	// Construction, with redundant-phi elimination, which is part of it rather than an optional
	// follow-up. A phi from a previous run is dropped first, so a second run after inlining an
	// immediately invoked function expression does not duplicate them.
	static_single_assignment.Construct(ssaGraph{}, function)

	for _, nested := range function.Functions {
		Construct(nested)
	}
}

// PhiOperandsInOrder returns a phi's predecessor block ids, in ascending block order.
//
// `Phi.Operands` is kept sorted by predecessor block id, so this is its predecessor ids in that
// order, read straight off the slice. `TestLowerIsDeterministic` in lower_corpus_test.go is what
// catches output that varies between runs.
func PhiOperandsInOrder(phi *Phi) []static_single_assignment.BlockId {
	return static_single_assignment.PhiOperandsInOrder(phi)
}
