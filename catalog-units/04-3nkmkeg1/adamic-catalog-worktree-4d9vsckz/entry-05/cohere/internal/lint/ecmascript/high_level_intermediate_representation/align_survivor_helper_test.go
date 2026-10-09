package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"sort"
)

// mergeWithReversedEndOrder is the merge sweep with `sortByEndDescending` INVERTED.
//
// A local reimplementation rather than a flag on `mergeSweepState`, so that the committed pass is
// not edited to host a counterfactual. It tracks `merge_scopes.go` and the only line that differs is
// the one marked below; `TestAlignReversedMergeTracksTheRealOne` asserts the two agree on the
// unaligned table, which is what makes a difference on the ALIGNED table attributable to the
// comparator rather than to this copy having drifted.
func mergeWithReversedEndOrder(function *Function, scopes *ReactiveScopes) *MergedScopes {
	result := &MergedScopes{}
	if function == nil || scopes == nil || scopes.Len() == 0 {
		return result
	}
	state := &mergeSweepState{rangeOf: scopes.RangeOf}
	state.starts, state.ends = collectScopeInfo(function, scopes)
	memberRanges := scopes.MemberRanges()

	visit := func(id static_single_assignment.EvaluationOrder) {
		state.visits++
		if len(state.ends) > 0 && state.ends[len(state.ends)-1].position <= id {
			top := state.ends[len(state.ends)-1]
			state.ends = state.ends[:len(state.ends)-1]
			closing := append([]ScopeId(nil), top.scopes...)
			sortByStartDescending(closing, state.rangeOf)
			for _, scope := range closing {
				index := indexOfScope(state.activeScopes, scope)
				if index == -1 {
					continue
				}
				if index != len(state.activeScopes)-1 {
					state.joined.union(append([]ScopeId{scope}, state.activeScopes[index+1:]...))
				}
				state.activeScopes = append(state.activeScopes[:index], state.activeScopes[index+1:]...)
			}
		}
		if len(state.starts) > 0 && state.starts[len(state.starts)-1].position <= id {
			top := state.starts[len(state.starts)-1]
			state.starts = state.starts[:len(state.starts)-1]
			opening := append([]ScopeId(nil), top.scopes...)
			// THE ONE DIFFERING LINE: ascending rather than descending by end.
			sort.SliceStable(opening, func(i, j int) bool {
				return state.rangeOf(opening[i]).End < state.rangeOf(opening[j]).End
			})
			state.activeScopes = append(state.activeScopes, opening...)
			for i := 1; i < len(opening); i++ {
				previous, current := opening[i-1], opening[i]
				if state.rangeOf(previous).End == state.rangeOf(current).End {
					state.joined.union([]ScopeId{previous, current})
				}
			}
		}
	}

	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			visit(instruction.Order)
			EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
				state.visitPlace(instruction.Order, place, scopes, memberRanges)
			})
		}
		order := TerminalOrder(block.Terminal)
		visit(order)
		EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) {
			state.visitPlace(order, place, scopes, memberRanges)
		})
	}
	return state.apply(scopes)
}
