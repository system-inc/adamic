package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

// alignCase describes a minimal graph exercising one alignment mechanism.
//
// The shape mirrors the JavaScript probes these expectations were taken from, so that a reader
// checking a row against React's own output is comparing the same graph: one block whose terminal
// carries a fallthrough, one arm that uses the scoped value, and a fallthrough block whose first
// instruction sits at a known position.
type alignCase struct {
	// scope is the single scope's range before alignment.
	scope mutation_aliasing.MutableRange
	// terminalId is the evaluation order of the construct's terminal, which is the position the
	// fallthrough pop minimises a start back to.
	terminalId static_single_assignment.EvaluationOrder
	// useAt is where the scoped value is read. Placing it inside the arm rather than the entry block
	// is what keeps the scope active when the sweep reaches the fallthrough.
	useAt static_single_assignment.EvaluationOrder
	// fallthroughs is the evaluation order of the fallthrough block's first instruction, which is
	// the position the fallthrough push widens an end out to.
	fallthroughs static_single_assignment.EvaluationOrder
}

// buildAlignCase builds the case with an `If` terminal, which is the ordinary widening shape.
func buildAlignCase(t *testing.T, testCase alignCase) (*Function, *ReactiveScopes) {
	t.Helper()
	return buildAlignCaseWithTerminal(t, testCase, func(fallthroughBlock static_single_assignment.BlockId) Terminal {
		return &If{
			Consequent:  2,
			Alternate:   fallthroughBlock,
			Fallthrough: fallthroughBlock,
			Order:       testCase.terminalId,
		}
	})
}

// buildAlignCaseWithTerminal builds the case with a caller-supplied terminal, so that one graph can
// be run with the terminal kind as the only variable. That is what isolates the branch exclusion.
func buildAlignCaseWithTerminal(t *testing.T, testCase alignCase,
	makeTerminal func(fallthroughBlock static_single_assignment.BlockId) Terminal) (*Function, *ReactiveScopes) {
	t.Helper()

	function := &Function{}
	function.Identifiers = append(function.Identifiers, nil)
	scopes := &ReactiveScopes{
		byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{},
		ranges:       map[ScopeId]mutation_aliasing.MutableRange{1: testCase.scope},
		members:      map[ScopeId][]static_single_assignment.IdentifierId{},
		order:        []ScopeId{1},
	}

	nextIdentifier := static_single_assignment.IdentifierId(1)
	place := func(inScope bool) Place {
		id := nextIdentifier
		nextIdentifier++
		function.Identifiers = append(function.Identifiers, &Identifier{Id: id})
		if inScope {
			scopes.byIdentifier[id] = 1
			scopes.members[1] = append(scopes.members[1], id)
		}
		return Place{Identifier: id}
	}

	emit := func(block *BasicBlock, order static_single_assignment.EvaluationOrder, use Place) {
		instruction := &Instruction{
			Id:     InstructionId(len(function.Instructions)),
			Order:  order,
			LValue: place(false),
			Value:  &LoadLocal{Place: use},
		}
		function.Instructions = append(function.Instructions, instruction)
		block.Instructions = append(block.Instructions, instruction.Id)
	}

	scoped := place(true)

	// bb0 is the construct: its terminal carries the fallthrough. bb2 is the arm that reads the
	// scoped value. bb1 is the fallthrough.
	entry := &BasicBlock{Id: 0, Kind: BlockKindBlock}
	arm := &BasicBlock{Id: 2, Kind: BlockKindBlock}
	fallthroughBlock := &BasicBlock{Id: 1, Kind: BlockKindBlock}

	// A use in the entry block only when it precedes the terminal, so that a case whose `useAt`
	// sits after the terminal reads inside the arm instead. That distinction is what lets one
	// builder serve both the push (use before) and the pop (use after) mechanisms.
	if testCase.useAt < testCase.terminalId {
		emit(entry, testCase.useAt, scoped)
	} else {
		emit(arm, testCase.useAt, scoped)
	}
	entry.Terminal = makeTerminal(1)

	arm.Terminal = &Goto{Block: 1, Variant: GotoVariantBreak, Order: testCase.terminalId + 1}

	emit(fallthroughBlock, testCase.fallthroughs, place(false))
	fallthroughBlock.Terminal = &Return{Order: testCase.fallthroughs + 1}

	function.Blocks = append(function.Blocks, entry, arm, fallthroughBlock)
	return function, scopes
}

// buildNestedGotoCase builds two nested constructs and a goto that leaves BOTH at once, which is the
// only shape reaching the goto back-reference mechanism.
//
// Layout, with evaluation orders in brackets:
//
//	bb0  outer `if` [5]                 fallthrough bb10 (the outer continuation)
//	bb1  inner `if` [15]                fallthrough bb2  (the inner continuation)
//	bb3  reads the scoped value [21]
//	     goto bb10 [22]                 jumps past the INNER fallthrough to the OUTER one
//	bb2  inner continuation [30]
//	bb10 outer continuation [40]
//
// When the sweep reaches bb3's goto, both fallthrough ranges are open and bb10 is at index 0 rather
// than at the top, so the back-reference fires. A goto to bb2 instead would target the innermost and
// take the other path.
func buildNestedGotoCase(t *testing.T) (*Function, *ReactiveScopes) {
	t.Helper()

	function := &Function{}
	function.Identifiers = append(function.Identifiers, nil)
	scopes := &ReactiveScopes{
		byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{},
		// The scope opens at the read and closes after the inner continuation, so it is still active
		// at the goto and has not been retired by the activity filter.
		ranges:  map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 21, End: 31}},
		members: map[ScopeId][]static_single_assignment.IdentifierId{},
		order:   []ScopeId{1},
	}

	nextIdentifier := static_single_assignment.IdentifierId(1)
	place := func(inScope bool) Place {
		id := nextIdentifier
		nextIdentifier++
		function.Identifiers = append(function.Identifiers, &Identifier{Id: id})
		if inScope {
			scopes.byIdentifier[id] = 1
			scopes.members[1] = append(scopes.members[1], id)
		}
		return Place{Identifier: id}
	}
	emit := func(block *BasicBlock, order static_single_assignment.EvaluationOrder, use Place) {
		instruction := &Instruction{
			Id:     InstructionId(len(function.Instructions)),
			Order:  order,
			LValue: place(false),
			Value:  &LoadLocal{Place: use},
		}
		function.Instructions = append(function.Instructions, instruction)
		block.Instructions = append(block.Instructions, instruction.Id)
	}

	scoped := place(true)

	outer := &BasicBlock{Id: 0, Kind: BlockKindBlock}
	outer.Terminal = &If{Consequent: 1, Alternate: 10, Fallthrough: 10, Order: 5}

	inner := &BasicBlock{Id: 1, Kind: BlockKindBlock}
	inner.Terminal = &If{Consequent: 3, Alternate: 2, Fallthrough: 2, Order: 15}

	body := &BasicBlock{Id: 3, Kind: BlockKindBlock}
	emit(body, 21, scoped)
	body.Terminal = &Goto{Block: 10, Variant: GotoVariantBreak, Order: 22}

	innerContinuation := &BasicBlock{Id: 2, Kind: BlockKindBlock}
	emit(innerContinuation, 30, place(false))
	innerContinuation.Terminal = &Goto{Block: 10, Variant: GotoVariantBreak, Order: 31}

	outerContinuation := &BasicBlock{Id: 10, Kind: BlockKindBlock}
	emit(outerContinuation, 40, place(false))
	outerContinuation.Terminal = &Return{Order: 41}

	function.Blocks = append(function.Blocks, outer, inner, body, innerContinuation, outerContinuation)
	return function, scopes
}

// buildInnermostGotoCase builds one construct and a goto targeting its OWN fallthrough, which is the
// innermost open one.
//
// This is the shape that separates upstream's `start !== at(-1)` test from its absence. Upstream
// SKIPS this case, because the fallthrough pop on the very next block applies the same start
// widening. The pop does not touch the END, so a scope whose end falls short of the fallthrough is
// where skipping and not-skipping produce different answers.
//
//	bb0  `if` [5]                 fallthrough bb1
//	bb2  reads the scoped value [7]
//	     goto bb1 [8]             targets the INNERMOST open fallthrough
//	bb1  continuation [20]
//
// The scope is [7,9): active at the goto (9 > 8), and ending well before the fallthrough's first
// instruction at 20. Upstream leaves its end at 9; removing the test would drag it to 20.
func buildInnermostGotoCase(t *testing.T) (*Function, *ReactiveScopes) {
	t.Helper()

	function := &Function{}
	function.Identifiers = append(function.Identifiers, nil)
	scopes := &ReactiveScopes{
		byIdentifier: map[static_single_assignment.IdentifierId]ScopeId{},
		ranges:       map[ScopeId]mutation_aliasing.MutableRange{1: {Start: 7, End: 9}},
		members:      map[ScopeId][]static_single_assignment.IdentifierId{},
		order:        []ScopeId{1},
	}

	nextIdentifier := static_single_assignment.IdentifierId(1)
	place := func(inScope bool) Place {
		id := nextIdentifier
		nextIdentifier++
		function.Identifiers = append(function.Identifiers, &Identifier{Id: id})
		if inScope {
			scopes.byIdentifier[id] = 1
			scopes.members[1] = append(scopes.members[1], id)
		}
		return Place{Identifier: id}
	}
	emit := func(block *BasicBlock, order static_single_assignment.EvaluationOrder, use Place) {
		instruction := &Instruction{
			Id:     InstructionId(len(function.Instructions)),
			Order:  order,
			LValue: place(false),
			Value:  &LoadLocal{Place: use},
		}
		function.Instructions = append(function.Instructions, instruction)
		block.Instructions = append(block.Instructions, instruction.Id)
	}

	scoped := place(true)

	entry := &BasicBlock{Id: 0, Kind: BlockKindBlock}
	entry.Terminal = &If{Consequent: 2, Alternate: 1, Fallthrough: 1, Order: 5}

	arm := &BasicBlock{Id: 2, Kind: BlockKindBlock}
	emit(arm, 7, scoped)
	arm.Terminal = &Goto{Block: 1, Variant: GotoVariantBreak, Order: 8}

	continuation := &BasicBlock{Id: 1, Kind: BlockKindBlock}
	emit(continuation, 20, place(false))
	continuation.Terminal = &Return{Order: 21}

	function.Blocks = append(function.Blocks, entry, arm, continuation)
	return function, scopes
}
