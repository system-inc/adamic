// The identifier graph `pruneNonEscapingScopes` walks, and the propagation over it.
//
// This is React's `State`, `IdentifierNode`, `ScopeNode` and `computeMemoizedIdentifiers`
// (`ReactiveScopes/PruneNonEscapingScopes.ts` at `react_conformance.UpstreamSha`).
//
// # Keyed on DeclarationId, deliberately, and not to be improved
//
// Upstream's comment is explicit and worth keeping: this pass models data flow rather than control
// flow, so it wants every value that may flow into one program variable regardless of the path
// taken. `DeclarationId` groups exactly those; `IdentifierId` would split them by
// single-assignment version and the pass would then reason about paths it does not model.
//
// The comment also says the change becomes correct once the pass is control-flow aware, which it is
// not. Do not "fix" this.
//
// # Cycles, and why a node is marked non-memoized while its dependencies are visited
//
// The walk is depth-first from the escaping roots. A cycle would otherwise recurse forever, so a
// node sets `seen` and answers `false` before descending -- a temporary lie that is corrected on the
// way back up. Upstream does the same and says so; the alternative, a fixpoint, would be slower for
// no gain because the answer only ever moves one way.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"sort"
)

// MemoizationGraph is what one walk of the tree produces: a node per declaration and per scope.
type MemoizationGraph struct {
	// identifiers is the node table, keyed by declaration.
	identifiers map[static_single_assignment.DeclarationId]*memoizationNode
	// scopes maps a scope to the declarations it depends on.
	scopes map[ScopeId][]static_single_assignment.DeclarationId
	// escaping are the roots: values returned, or passed to a hook.
	escaping map[static_single_assignment.DeclarationId]bool
}

type memoizationNode struct {
	level        MemoizationLevel
	memoized     bool
	dependencies map[static_single_assignment.DeclarationId]bool
	scopes       map[ScopeId]bool
	seen         bool
}

// NewMemoizationGraph returns an empty graph ready to have nodes declared into it.
func NewMemoizationGraph() *MemoizationGraph {
	return &MemoizationGraph{
		identifiers: map[static_single_assignment.DeclarationId]*memoizationNode{},
		scopes:      map[ScopeId][]static_single_assignment.DeclarationId{},
		escaping:    map[static_single_assignment.DeclarationId]bool{},
	}
}

// Declare adds a node for a declaration if it has none, at the weakest level.
//
// Upstream's `declare`, used for the function's own id and its parameters. Later classification
// raises the level through `Record`; a declaration that is never classified stays at `Never`, which
// is the right answer for a parameter nothing writes to.
func (g *MemoizationGraph) Declare(declaration static_single_assignment.DeclarationId) {
	if g == nil || g.identifiers == nil {
		return
	}
	if _, exists := g.identifiers[declaration]; exists {
		return
	}
	g.identifiers[declaration] = &memoizationNode{
		level:        MemoizationNever,
		dependencies: map[static_single_assignment.DeclarationId]bool{},
		scopes:       map[ScopeId]bool{},
	}
}

// Record raises a declaration's level and adds its dependencies.
//
// The level is joined rather than assigned, because one declaration can be an lvalue more than once
// -- reassigned, or destructured alongside another binding -- and the final level is the strongest
// any assignment demanded. Assigning instead would make the answer depend on visit order.
func (g *MemoizationGraph) Record(declaration static_single_assignment.DeclarationId, level MemoizationLevel,
	dependencies []static_single_assignment.DeclarationId) {
	if g == nil {
		return
	}
	g.Declare(declaration)
	node := g.identifiers[declaration]
	node.level = JoinMemoizationLevels(node.level, level)
	for _, dependency := range dependencies {
		if dependency == declaration {
			// A self-dependency carries no information and would make the cycle guard the only
			// thing preventing infinite recursion on an ordinary value.
			continue
		}
		node.dependencies[dependency] = true
	}
}

// AssociateScope records that a declaration belongs to a scope, and that scope's own dependencies.
//
// Both halves matter. The declaration-to-scope edge is how marking a value memoized forces the
// scope's other dependencies; the scope-to-dependencies edge is what that forcing walks.
func (g *MemoizationGraph) AssociateScope(declaration static_single_assignment.DeclarationId, scope ScopeId,
	scopeDependencies []static_single_assignment.DeclarationId) {
	if g == nil {
		return
	}
	g.Declare(declaration)
	g.identifiers[declaration].scopes[scope] = true
	if _, recorded := g.scopes[scope]; !recorded {
		g.scopes[scope] = append([]static_single_assignment.DeclarationId(nil), scopeDependencies...)
	}
}

// MarkEscaping records a declaration as a root: returned, or passed to a hook.
//
// Upstream collects two kinds. A returned value escapes because the caller holds it. A value passed
// to a hook escapes because React may retain it -- the closure handed to `useEffect` is the standard
// case, and it is why `IsHookCallee` was lifted to the shelf.
func (g *MemoizationGraph) MarkEscaping(declaration static_single_assignment.DeclarationId) {
	if g == nil {
		return
	}
	g.Declare(declaration)
	g.escaping[declaration] = true
}

// ComputeMemoized walks outward from the escaping roots and returns everything that must be held.
//
// Upstream's `computeMemoizedIdentifiers`. A node is memoized when its level is `Memoized`, or
// `Conditional` with a memoized dependency, or `Unmemoized` under forcing. Marking a node forces
// every dependency of every scope it belongs to, which is the second case the pass exists for: a
// value that does not itself escape still needs holding when it feeds a scope whose output does.
func (g *MemoizationGraph) ComputeMemoized() map[static_single_assignment.DeclarationId]bool {
	memoized := map[static_single_assignment.DeclarationId]bool{}
	if g == nil {
		return memoized
	}
	// Reset traversal state so the graph can be walked more than once.
	for _, node := range g.identifiers {
		node.seen = false
		node.memoized = false
	}
	seenScopes := map[ScopeId]bool{}

	var visit func(declaration static_single_assignment.DeclarationId, force bool) bool
	var forceScope func(scope ScopeId)

	visit = func(declaration static_single_assignment.DeclarationId, force bool) bool {
		node, exists := g.identifiers[declaration]
		if !exists {
			// Upstream raises an invariant. A linter answers false: a declaration nothing recorded
			// cannot be shown to need memoizing, and refusing to hold it is the direction that
			// costs granularity rather than correctness.
			return false
		}
		if node.seen {
			return node.memoized
		}
		node.seen = true
		// The temporary lie that terminates a cycle; corrected below.
		node.memoized = false

		hasMemoizedDependency := false
		// Sorted, because Go map iteration is randomised and this walk is order-dependent: the
		// `seen` guard locks in whichever answer a node was first reached with, and a node reached
		// with `force` set answers differently from the same node reached without it. Measured
		// before this was added -- three identical runs over the corpus returned 506, 507 and 508
		// pruned scopes and memoized counts spanning 1,922 to 2,054.
		//
		// A pass whose output varies run to run makes a cache non-reproducible without ever
		// producing an answer anyone can point at as wrong, which is the same reason
		// `ScopeDependencies.Ids` returns a slice rather than ranging a map.
		for _, dependency := range sortedDeclarations(node.dependencies) {
			if visit(dependency, false) {
				hasMemoizedDependency = true
			}
		}

		switch {
		case node.level == MemoizationMemoized,
			node.level == MemoizationConditional && (hasMemoizedDependency || force),
			node.level == MemoizationUnmemoized && force:
			node.memoized = true
			memoized[declaration] = true
			for _, scope := range sortedScopes(node.scopes) {
				forceScope(scope)
			}
		}
		return node.memoized
	}

	forceScope = func(scope ScopeId) {
		if seenScopes[scope] {
			return
		}
		seenScopes[scope] = true
		for _, dependency := range g.scopes[scope] {
			visit(dependency, true)
		}
	}

	for _, declaration := range sortedDeclarations(g.escaping) {
		visit(declaration, false)
	}
	return memoized
}

// sortedDeclarations returns a map's keys in a stable order. See ComputeMemoized for why.
func sortedDeclarations(set map[static_single_assignment.DeclarationId]bool) []static_single_assignment.DeclarationId {
	keys := make([]static_single_assignment.DeclarationId, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// sortedScopes returns a map's keys in a stable order. See ComputeMemoized for why.
func sortedScopes(set map[ScopeId]bool) []ScopeId {
	keys := make([]ScopeId, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// Len reports how many declarations the graph holds, for measurement.
func (g *MemoizationGraph) Len() int {
	if g == nil {
		return 0
	}
	return len(g.identifiers)
}

// EscapingCount reports how many roots the graph holds, for measurement.
func (g *MemoizationGraph) EscapingCount() int {
	if g == nil {
		return 0
	}
	return len(g.escaping)
}
