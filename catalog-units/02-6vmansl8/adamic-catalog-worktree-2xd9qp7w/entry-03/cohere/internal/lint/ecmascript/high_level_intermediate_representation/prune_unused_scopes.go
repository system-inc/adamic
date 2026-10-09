// Turning scopes with no outputs into ordinary blocks.
//
// This is React's `pruneUnusedScopes` (`ReactiveScopes/PruneUnusedScopes.ts` at
// `react_conformance.UpstreamSha`) plus `updateScopeDeclarations`, which is
// `MergeReactiveScopesThatInvalidateTogether`'s helper and lands here because this pass is what
// makes it load-bearing.
//
// # Why the rule needs this
//
// `validatePreservedManualMemoization` keeps two live sets, `scopes` and `prunedScopes`, and a
// memoized value whose scope was pruned is treated differently from one whose scope survived. This
// pass is what puts ids into the second set, so without it every scope looks live to the rule.
//
// # Pruned is a flag here, not a second statement type
//
// Upstream emits a `pruned-scope` statement distinct from `scope`. This tree carries `Pruned` on
// `ReactiveScopeBlock` instead, which was already declared and always false, with a comment naming
// the four upstream passes that set it. This is the first of those four to exist. The shapes are
// equivalent -- both say "this was a scope and its memoization is discarded" -- and the flag avoids
// a variant every visitor and transform would have to switch on.
package high_level_intermediate_representation

// PruneUnusedScopes converts scopes with no outputs of their own into ordinary blocks.
//
// Returns how many it pruned, so a caller can tell a no-op from an unrun pass.
//
// A scope is pruned when it holds no return statement, reassigns nothing, and either declares
// nothing or declares only values that originated in scopes nested inside it. The last clause is
// upstream's `hasOwnDeclaration` and it is the reason `OriginOf` exists: a declaration recorded
// against a scope may have been produced within it and bubbled up, and a scope whose entire
// declaration set arrived that way has produced nothing itself.
//
// The return-statement condition is upstream's and it is not obvious. A scope containing a `return`
// is kept even with no declarations, because the value being memoized is early-returned from inside
// it -- `propagateEarlyReturns` rewrites that shape later and needs the scope still standing.
func PruneUnusedScopes(tree *ReactiveFunction, dependencies *ScopeDependencies) int {
	if tree == nil || dependencies == nil {
		return 0
	}
	pruner := scopePruner{dependencies: dependencies}
	pruner.walk(tree.Body)
	return pruner.pruned
}

type scopePruner struct {
	dependencies *ScopeDependencies
	pruned       int
}

// walk visits a block, pruning any scope in it, and reports whether it held a return statement.
//
// The return flag propagates outward: a `return` anywhere inside a scope keeps that scope, and a
// scope's own walk must therefore see returns from every block nested within it.
func (p *scopePruner) walk(block ReactiveBlock) bool {
	sawReturn := false
	for _, statement := range block {
		switch shape := statement.(type) {
		case *ReactiveScopeBlock:
			// The nested walk's return flag decides this scope, and also propagates outward: a
			// return inside a nested scope is still a return inside this block.
			if p.visitScope(shape) {
				sawReturn = true
			}

		case *ReactiveTerminalStatement:
			if p.walkTerminal(shape) {
				sawReturn = true
			}
			if _, isReturn := shape.Terminal.(*ReactiveReturn); isReturn {
				sawReturn = true
			}
		}
	}
	return sawReturn
}

// visitScope prunes one scope if it qualifies, and reports whether it contained a return.
func (p *scopePruner) visitScope(scope *ReactiveScopeBlock) bool {
	sawReturn := p.walk(scope.Instructions)
	if scope.Pruned {
		return sawReturn
	}
	if sawReturn {
		return sawReturn
	}
	if len(p.dependencies.ReassignmentsOf(scope.Scope)) != 0 {
		return sawReturn
	}
	if p.hasOwnDeclaration(scope) {
		return sawReturn
	}
	scope.Pruned = true
	p.pruned++
	return sawReturn
}

// hasOwnDeclaration reports whether a scope declares any value that originated in the scope itself.
//
// Upstream's predicate of the same name, and the whole reason `ScopeDependencies` records an origin.
// A scope holding only declarations that bubbled up from nested scopes has produced nothing, so
// memoizing it buys nothing that memoizing the inner scopes does not already buy.
//
// A declaration with no recorded origin counts as the scope's own, which is the conservative
// direction: an unknown origin keeps the scope rather than pruning it on missing information.
func (p *scopePruner) hasOwnDeclaration(scope *ReactiveScopeBlock) bool {
	for _, declared := range p.dependencies.DeclarationsOf(scope.Scope) {
		origin, found := p.dependencies.OriginOf(declared)
		if !found || origin == scope.Scope {
			return true
		}
	}
	return false
}

// walkTerminal descends into every block a terminal contains, reporting whether any held a return.
func (p *scopePruner) walkTerminal(statement *ReactiveTerminalStatement) bool {
	sawReturn := false
	visit := func(block ReactiveBlock) {
		if p.walk(block) {
			sawReturn = true
		}
	}
	switch shape := statement.Terminal.(type) {
	case *ReactiveIf:
		visit(shape.Consequent)
		if shape.Alternate != nil {
			visit(*shape.Alternate)
		}
	case *ReactiveSwitch:
		for index := range shape.Cases {
			if shape.Cases[index].Block != nil {
				visit(*shape.Cases[index].Block)
			}
		}
	case *ReactiveFor:
		visit(shape.Loop)
	case *ReactiveForOf:
		visit(shape.Loop)
	case *ReactiveForIn:
		visit(shape.Loop)
	case *ReactiveWhile:
		visit(shape.Loop)
	case *ReactiveDoWhile:
		visit(shape.Loop)
	case *ReactiveLabelTerminal:
		visit(shape.Block)
	case *ReactiveTry:
		visit(shape.Block)
		visit(shape.Handler)
	}
	return sawReturn
}
