// Verification of single static assignment form: the property, checked rather than asserted.
//
// `Construct` is only useful if what it produces is actually SSA, and "the tests pass" is weak
// evidence when the tests were written by the same person as the construction. So the defining
// property is checked directly, over whatever function is handed in:
//
//  1. Every value is defined at most once.
//  2. EVERY USE IS DOMINATED BY ITS DEFINITION.
//
// The second is the real one. It is what makes SSA worth having: a pass that sees a use can reach
// the single definition and know it has already executed. A construction that places a phi at the
// wrong block, or renames a use to a definition sitting on a sibling branch, breaks exactly this
// and breaks nothing that a structural well-formedness check would notice.
//
// # Phi operands are checked against the PREDECESSOR, not the phi's own block
//
// A phi is not an ordinary instruction. Its operand for predecessor P is read on the edge from P,
// so the definition must dominate P's exit, not the block holding the phi. Checking it the naive
// way reports a false failure on every correct loop phi, since the back edge's value is defined
// below the header. This distinction is the single easiest thing to get wrong in a checker of this
// kind, and getting it wrong produces a checker that rejects correct code, which is at least loud.
// The reverse mistake - checking phis as though they were instructions in the predecessor - is the
// dangerous one, and it is why this is spelled out.
//
// # Dominance is computed here, over this graph
//
// `internal/utilities/control_flow_graph.AnalyzeDominators` is generic over that package's `Graph[E]`, not
// over `Function`, so it cannot be pointed at this. Rather than materialise a parallel graph to
// borrow it, the same Cooper-Harvey-Kennedy fixed point is run over `Function.Blocks`. It is twenty
// lines because reverse postorder is already established, and being a second implementation is
// acceptable for a checker: a checker that shares an implementation with the thing it checks can
// agree with it and both be wrong.
//
// The verifier and its dominance are static_single_assignment's, which this IR shares with Adamic's
// flow graph. The context-store exemption upstream's invariant implies is there too, asked of this IR
// through ssaGraph.ContextStoreDefines.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// VerifySSA checks that a function is in single static assignment form and returns every violation.
//
// An empty result means the property holds. It does NOT mean the function was worth checking: a
// function whose variable references all lowered to `LoadGlobal` has almost nothing to cohere and
// passes trivially. `SSAStats` is how a caller tells a real pass from a vacuous one.
//
// Nested functions are not descended into; call it per function.
func VerifySSA(function *Function) []static_single_assignment.SSAViolation {
	if function == nil {
		return nil
	}
	return static_single_assignment.VerifySSA(ssaGraph{}, function)
}

// CollectSSAStats measures one function.
func CollectSSAStats(function *Function) static_single_assignment.SSAStats {
	if function == nil {
		return static_single_assignment.SSAStats{}
	}
	return static_single_assignment.CollectSSAStats(ssaGraph{}, function)
}

// computeDominance runs Cooper-Harvey-Kennedy over the function's blocks, which are already in
// reverse postorder with unreachable blocks removed.
func computeDominance(function *Function) *static_single_assignment.Dominance {
	return static_single_assignment.ComputeDominance(ssaGraph{}, function)
}
