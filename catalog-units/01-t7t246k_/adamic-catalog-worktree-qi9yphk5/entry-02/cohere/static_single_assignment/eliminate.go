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
package static_single_assignment

// EliminateRedundantPhis removes phis that are not merges, rewriting every reference to a removed
// phi's result to the value it collapsed to.
//
// Runs to a fixed point over one function; an IR recurses into nested functions itself.
func EliminateRedundantPhis[G Graph[F, B, P], F any, B comparable, P any](graph G, function F) {
	for {
		// rewrites maps a removed phi's result to what it collapsed to.
		rewrites := map[IdentifierId]IdentifierId{}

		for _, block := range graph.Blocks(function) {
			phis := graph.Phis(block)
			kept := phis[:0]
			for _, phi := range phis {
				if collapsed, redundant := redundantPhiValue(graph, phi); redundant {
					rewrites[graph.IdentifierOf(phi.Place)] = collapsed
					continue
				}
				kept = append(kept, phi)
			}
			graph.SetPhis(block, kept)
		}

		if len(rewrites) == 0 {
			return
		}

		// A removed phi may collapse to another removed phi's result, so chase each chain to its
		// end before rewriting. The chain is acyclic because a phi only collapses to a value that
		// is not itself.
		resolve := func(id IdentifierId) IdentifierId {
			seen := 0
			for {
				next, ok := rewrites[id]
				if !ok {
					return id
				}
				id = next
				seen++
				if seen > len(rewrites) {
					// Defensive: a cycle cannot arise from the redundancy test above, but a caller
					// that hand-built phis could create one, and looping forever is the worst
					// possible response.
					return id
				}
			}
		}

		applyRewrites(graph, function, resolve)
	}
}

// redundantPhiValue reports the single value a phi collapses to, if it collapses at all.
//
// The rule, which is upstream's: ignoring operands that are the phi's own result, if every
// remaining operand names one value, the phi is that value. A phi with no operands other than
// itself cannot arise from a reachable block and is treated as not redundant so it stays visible.
func redundantPhiValue[G Graph[F, B, P], F any, B comparable, P any](graph G, phi *Phi[P]) (IdentifierId, bool) {
	result := graph.IdentifierOf(phi.Place)
	var candidate IdentifierId
	found := false
	for _, entry := range phi.Operands {
		operand := graph.IdentifierOf(entry.Place)
		if operand == result {
			continue
		}
		if !found {
			candidate = operand
			found = true
			continue
		}
		if operand != candidate {
			return 0, false
		}
	}
	if !found {
		return 0, false
	}
	return candidate, true
}

// applyRewrites rewrites every place in the function through resolve.
func applyRewrites[G Graph[F, B, P], F any, B comparable, P any](graph G, function F, resolve func(IdentifierId) IdentifierId) {
	rewrite := func(place *P, role Role) {
		*place = graph.WithIdentifier(*place, resolve(graph.IdentifierOf(*place)))
	}
	for _, block := range graph.Blocks(function) {
		for _, phi := range graph.Phis(block) {
			rewrite(&phi.Place, Define)
			for index := range phi.Operands {
				rewrite(&phi.Operands[index].Place, Use)
			}
		}
		count := graph.InstructionCount(function, block)
		for index := 0; index < count; index++ {
			graph.EachInstructionPlace(function, block, index, rewrite)
		}
		graph.EachTerminalPlace(block, rewrite)
	}
	params := graph.Params(function)
	for index := range params {
		rewrite(&params[index], Define)
	}
	if returns := graph.Returns(function); returns != nil {
		rewrite(returns, Use)
	}
}
