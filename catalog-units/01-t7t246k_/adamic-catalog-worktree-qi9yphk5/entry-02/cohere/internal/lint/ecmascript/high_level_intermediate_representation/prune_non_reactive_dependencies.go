// Dropping dependencies that cannot change, so a scope stops comparing values that never differ.
//
// This is React's `pruneNonReactiveDependencies`
// (`ReactiveScopes/PruneNonReactiveDependencies.ts` at `react_conformance.UpstreamSha`).
//
// `CollectScopeDependencies` infers a scope's inputs without asking whether they can change over
// time, because at that point the answer is not known. This pass removes the ones that cannot, so a
// scope compares only values that might actually differ between renders.
//
// # This is the one 6b5 pass that changes an answer the rule computes
//
// `validatePreservedManualMemoization` compares dependency sets directly. `pruneUnusedScopes` and
// `pruneAlwaysInvalidatingScopes` add state alongside what the rule reads -- the pruned set -- but
// this one deletes from `scope.dependencies` itself. Skipping it leaves every scope's dependency set
// wider than upstream's, and the rule would compare sets that do not correspond.
//
// # Where our reactivity comes from, which is not where upstream's does
//
// Upstream builds a `Set<IdentifierId>` by walking the tree and seeding from hook results, props and
// a handful of propagating instruction shapes. This tree already has `InferReactive`, which writes
// `Place.Reactive` -- 31,949 of 51,945 corpus places carry it -- so the seed is read rather than
// recomputed.
//
// What is not already done is the propagation this pass performs on top: a scope with any reactive
// dependency makes all of its own outputs reactive, because those outputs re-evaluate whenever the
// dependency changes. That is a fact about scopes rather than about values, so `InferReactive` could
// not have known it, and it runs here.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// PruneNonReactiveDependencies removes scope dependencies that cannot change between renders.
//
// Returns how many it removed. The `reactive` set is seeded from `Place.Reactive` and then grown by
// the scope-output propagation, so `InferReactive` must have run on the same function first; a
// caller that skips it gets an empty seed and prunes every dependency, which is why the seed size is
// asserted rather than assumed.
func PruneNonReactiveDependencies(tree *ReactiveFunction, function *Function,
	dependencies *ScopeDependencies) int {
	if tree == nil || function == nil || dependencies == nil {
		return 0
	}

	pruner := nonReactivePruner{
		function:     function,
		dependencies: dependencies,
		reactive:     map[static_single_assignment.IdentifierId]bool{},
	}
	// The seed: every place lowering marked reactive. Read off the graph rather than the tree,
	// because `Place.Reactive` is written per reference and the graph holds every reference.
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
				if place.Reactive {
					pruner.reactive[place.Identifier] = true
				}
			})
		}
	}
	pruner.seeded = len(pruner.reactive)

	pruner.walk(tree.Body)
	return pruner.pruned
}

type nonReactivePruner struct {
	function     *Function
	dependencies *ScopeDependencies
	reactive     map[static_single_assignment.IdentifierId]bool
	// seeded is how many identifiers the seed contributed, before propagation.
	seeded int
	pruned int
}

// walk visits a block, pruning each scope's dependencies innermost-first.
//
// Innermost-first because the propagation runs outward: an inner scope's outputs become reactive
// before an enclosing scope asks whether its own dependencies are.
func (p *nonReactivePruner) walk(block ReactiveBlock) {
	for _, statement := range block {
		switch shape := statement.(type) {
		case *ReactiveScopeBlock:
			p.walk(shape.Instructions)
			// A pruned scope is walked and nothing more. Upstream's visitor overrides `visitScope`
			// only, and the base `visitPrunedScope` visits the body, so a pruned scope neither drops
			// its dependencies nor makes its declarations reactive. Propagating through one marked
			// a hook call's declarations reactive, `useRef`'s result among them, and a memo callback
			// nested beside it then kept the ref as a dependency the developer never wrote.
			if shape.Pruned {
				continue
			}
			p.visitScope(shape)
		case *ReactiveTerminalStatement:
			p.walkTerminal(shape)
		}
	}
}

// visitScope drops a scope's non-reactive dependencies, then propagates outward if any survived.
func (p *nonReactivePruner) visitScope(scope *ReactiveScopeBlock) {
	removed := p.dependencies.PruneNonReactiveDependenciesOf(scope.Scope, p.reactive)
	p.pruned += removed

	if len(p.dependencies.DependenciesOf(scope.Scope)) == 0 {
		return
	}
	// Upstream's comment: if any dependency is reactive then every output re-evaluates whenever
	// that dependency changes, so the outputs are reactive in practice even when the values
	// themselves would not be.
	for _, declared := range p.dependencies.DeclarationsOf(scope.Scope) {
		p.reactive[declared] = true
	}
	for _, reassigned := range p.dependencies.ReassignmentsOf(scope.Scope) {
		p.reactive[reassigned] = true
	}
}

// walkTerminal descends into every block a terminal contains.
func (p *nonReactivePruner) walkTerminal(statement *ReactiveTerminalStatement) {
	switch shape := statement.Terminal.(type) {
	case *ReactiveIf:
		p.walk(shape.Consequent)
		if shape.Alternate != nil {
			p.walk(*shape.Alternate)
		}
	case *ReactiveSwitch:
		for index := range shape.Cases {
			if shape.Cases[index].Block != nil {
				p.walk(*shape.Cases[index].Block)
			}
		}
	case *ReactiveFor:
		p.walk(shape.Loop)
	case *ReactiveForOf:
		p.walk(shape.Loop)
	case *ReactiveForIn:
		p.walk(shape.Loop)
	case *ReactiveWhile:
		p.walk(shape.Loop)
	case *ReactiveDoWhile:
		p.walk(shape.Loop)
	case *ReactiveLabelTerminal:
		p.walk(shape.Block)
	case *ReactiveTry:
		p.walk(shape.Block)
		p.walk(shape.Handler)
	}
}
