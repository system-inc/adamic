// Redundant phi elimination.
//
// A phi is redundant when it is not actually a merge: every operand is the same value, or every
// operand is either the same value or the phi's own result. The second case is what a loop produces
// when the body does not reassign the variable - the back edge feeds the phi its own output - and it
// is the reason this cannot be a simple "are all operands equal" test.
//
// # Why this is here rather than left to the next agent
//
// Braun's construction places a phi whenever a lookup crosses a join, before it knows whether the
// operands will agree. On a loop it must: the header's phi is minted before the back edge's value
// exists. So the construction cannot avoid producing redundant phis, and a redundant phi is not
// cosmetic - it is a value that appears to have several reaching definitions when it has one, which
// is precisely the fact every pass above this will key on. Shipping construction without it would
// mean every consumer re-deriving "is this phi real".
//
// Removal cascades: deleting one phi can make another redundant, since a phi's operand may be the
// removed phi's result. So this iterates to a fixed point rather than making one pass.
//
// Upstream's equivalent is `eliminate_redundant_phi.rs`, 177 lines.
//
// The pass is static_single_assignment.EliminateRedundantPhis, which this IR shares with Adamic's flow
// graph; the reasoning above is carried there too.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// EliminateRedundantPhis removes phis that are not merges, rewriting every reference to a removed
// phi's result, the Returns place included, to the value it collapsed to.
//
// Runs to a fixed point over one function. Construct runs it for each function it converts.
func EliminateRedundantPhis(function *Function) {
	if function == nil {
		return
	}
	static_single_assignment.EliminateRedundantPhis(ssaGraph{}, function)
}
