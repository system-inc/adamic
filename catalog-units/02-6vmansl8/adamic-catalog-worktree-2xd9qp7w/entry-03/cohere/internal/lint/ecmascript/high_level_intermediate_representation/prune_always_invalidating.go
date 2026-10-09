// Pruning scopes that will always invalidate, so they stop paying for comparisons they cannot win.
//
// This is React's `pruneAlwaysInvalidatingScopes` (`ReactiveScopes/PruneAlwaysInvalidatingScopes.ts`
// at `react_conformance.UpstreamSha`).
//
// # The idea
//
// A scope compares its dependencies each render and recomputes when one changed. If a dependency is
// a value that gets a fresh identity every render anyway -- an array literal, an object literal,
// JSX, a `new` -- then the comparison can never say "unchanged", so the scope recomputes
// unconditionally and the comparison is pure cost. Better to drop the memoization than to pay for a
// check whose answer is fixed.
//
// The unmemoized-ness spreads: a scope pruned this way produces values that are themselves
// unmemoized, so a scope depending on those is in the same position, and so on outward.
//
// # The edge case upstream calls out, transcribed rather than reasoned about
//
// A function call may return a primitive, so upstream optimistically assumes it does and an
// unmemoized call does not prune downstream memoization. Only guaranteed allocations do. That is a
// deliberate imprecision in the safe direction and it is why `CallExpression` is absent from the
// seeding switch below despite being an obvious member of the same family.
//
// # Why this pass is what made `PruneDeclarationsLastUsedBefore` owed
//
// It reads a scope's declarations as a propagation channel: every declaration naming an
// always-invalidating value joins the unmemoized set. So a declaration the merge should have dropped
// -- one last read inside an absorbed range -- would prune scopes upstream keeps. That is the deny
// direction, and it is the reason `updateScopeDeclarations` stopped being optional.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// PruneAlwaysInvalidatingScopes marks scopes whose dependencies can never compare equal.
//
// Returns how many it pruned, so a caller can tell a no-op from an unrun pass.
func PruneAlwaysInvalidatingScopes(tree *ReactiveFunction, function *Function,
	dependencies *ScopeDependencies) int {
	if tree == nil || function == nil || dependencies == nil {
		return 0
	}
	pruner := alwaysInvalidatingPruner{
		function:           function,
		dependencies:       dependencies,
		alwaysInvalidating: map[static_single_assignment.IdentifierId]bool{},
		unmemoized:         map[static_single_assignment.IdentifierId]bool{},
	}
	pruner.walk(tree.Body, false)
	return pruner.pruned
}

type alwaysInvalidatingPruner struct {
	function     *Function
	dependencies *ScopeDependencies
	// alwaysInvalidating are values that allocate fresh identity on every evaluation.
	alwaysInvalidating map[static_single_assignment.IdentifierId]bool
	// unmemoized are the subset of those that no scope is memoizing, so they really do change every
	// render rather than merely being capable of it.
	unmemoized map[static_single_assignment.IdentifierId]bool
	pruned     int
}

// walk visits a block, seeding value sets from instructions and pruning scopes that qualify.
//
// `withinScope` is upstream's traversal state: an allocation inside a scope is memoized by that
// scope, so it is always-invalidating in principle but not unmemoized in fact. The distinction is
// the whole pass -- without it every array literal anywhere would prune every scope reading it.
func (p *alwaysInvalidatingPruner) walk(block ReactiveBlock, withinScope bool) {
	for _, statement := range block {
		switch shape := statement.(type) {
		case *ReactiveInstructionStatement:
			p.visitInstruction(shape.Instruction, withinScope)

		case *ReactiveScopeBlock:
			p.walk(shape.Instructions, true)
			p.visitScope(shape)

		case *ReactiveTerminalStatement:
			p.walkTerminal(shape, withinScope)
		}
	}
}

// visitInstruction seeds and propagates the two value sets.
func (p *alwaysInvalidatingPruner) visitInstruction(instruction *ReactiveInstruction,
	withinScope bool) {
	if instruction == nil {
		return
	}
	plain, isPlain := instruction.Value.(*ReactiveInstructionValue)
	if !isPlain || plain.Value == nil {
		return
	}

	switch value := plain.Value.(type) {
	case *ArrayExpression, *ObjectExpression, *JsxExpression, *JsxFragment, *NewExpression:
		// The seeding set, and `CallExpression` is deliberately not here; see the package comment.
		if instruction.LValue != nil {
			p.alwaysInvalidating[instruction.LValue.Identifier] = true
			if !withinScope {
				p.unmemoized[instruction.LValue.Identifier] = true
			}
		}

	case *StoreLocal:
		// Both sets flow through an assignment, separately: a value can be always-invalidating
		// without being unmemoized, and copying it must preserve that difference.
		if p.alwaysInvalidating[value.Value.Identifier] {
			p.alwaysInvalidating[value.LValue.Identifier] = true
		}
		if p.unmemoized[value.Value.Identifier] {
			p.unmemoized[value.LValue.Identifier] = true
		}

	case *LoadLocal:
		if instruction.LValue == nil {
			return
		}
		if p.alwaysInvalidating[value.Place.Identifier] {
			p.alwaysInvalidating[instruction.LValue.Identifier] = true
		}
		if p.unmemoized[value.Place.Identifier] {
			p.unmemoized[instruction.LValue.Identifier] = true
		}
	}
}

// visitScope prunes a scope depending on any unmemoized value, and propagates outward.
func (p *alwaysInvalidatingPruner) visitScope(scope *ReactiveScopeBlock) {
	if scope.Pruned {
		return
	}
	depends := false
	for _, dependency := range p.dependencies.DependenciesOf(scope.Scope) {
		if p.unmemoized[dependency.Identifier] {
			depends = true
			break
		}
	}
	if !depends {
		return
	}

	// The propagation channel. Every always-invalidating value this scope declares or reassigns
	// becomes unmemoized, because the scope that was memoizing it is about to stop.
	for _, declared := range p.dependencies.DeclarationsOf(scope.Scope) {
		if p.alwaysInvalidating[declared] {
			p.unmemoized[declared] = true
		}
	}
	for _, reassigned := range p.dependencies.ReassignmentsOf(scope.Scope) {
		if p.alwaysInvalidating[reassigned] {
			p.unmemoized[reassigned] = true
		}
	}

	scope.Pruned = true
	p.pruned++
}

// walkTerminal descends into every block a terminal contains.
func (p *alwaysInvalidatingPruner) walkTerminal(statement *ReactiveTerminalStatement,
	withinScope bool) {
	switch shape := statement.Terminal.(type) {
	case *ReactiveIf:
		p.walk(shape.Consequent, withinScope)
		if shape.Alternate != nil {
			p.walk(*shape.Alternate, withinScope)
		}
	case *ReactiveSwitch:
		for index := range shape.Cases {
			if shape.Cases[index].Block != nil {
				p.walk(*shape.Cases[index].Block, withinScope)
			}
		}
	case *ReactiveFor:
		p.walk(shape.Loop, withinScope)
	case *ReactiveForOf:
		p.walk(shape.Loop, withinScope)
	case *ReactiveForIn:
		p.walk(shape.Loop, withinScope)
	case *ReactiveWhile:
		p.walk(shape.Loop, withinScope)
	case *ReactiveDoWhile:
		p.walk(shape.Loop, withinScope)
	case *ReactiveLabelTerminal:
		p.walk(shape.Block, withinScope)
	case *ReactiveTry:
		p.walk(shape.Block, withinScope)
		p.walk(shape.Handler, withinScope)
	}
}
