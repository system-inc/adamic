// Scope dependencies: what each reactive scope reads from outside itself.
//
// This is React's `propagateScopeDependenciesHIR` (development bundle line 47327) run over this
// graph. oxc transcribes it at
// `oxc_react_compiler/src/react_compiler_inference/propagate_scope_dependencies_hir.rs`; where the
// two disagree React wins, and each disagreement is recorded at the line that resolves it.
//
// # What this pass computes, and what a "dependency" actually is
//
// A reactive scope memoizes a group of values together. To know when to recompute them, the scope
// needs its INPUTS: the values it reads that were produced before it began. Those inputs are
// dependencies, and a dependency is not an identifier -- it is an ACCESS PATH rooted at one.
//
//	ReactiveScopeDependency { Identifier, Reactive, Path []DependencyPathEntry }
//	DependencyPathEntry     { Property string, Optional bool }
//
// So `props.a.b` is one dependency: root `props`, path `[a, b]`. This matters because a scope
// reading only `props.a.b` should not invalidate when `props.c` changes, and because the dedup
// predicate compares the whole path. Upstream's `areEqualPaths` (bundle line 47352) compares each
// entry's property AND its `Optional` flag, so `a?.b` and `a.b` are DIFFERENT dependencies. That
// asymmetry is reproduced here rather than simplified; see `DependencyPathEntry.Optional`.
//
// # The three fields this pass writes, and why they are all one pass
//
// `ReactiveScope` has eight fields. `scopes.go` filled `id`, `range` and `loc` and stopped, naming
// `dependencies`, `declarations` and `reassignments` as belonging here. That boundary was checked
// mechanically before this file was written rather than taken on trust: every writer of the three
// fields across all of oxc was enumerated, and the initial producer of each is this pass alone
// (`propagate_scope_dependencies_hir.rs` lines 111, 1812 and 1857). Every other write is a later
// `prune_*` or `merge_*` pass mutating what this one created, and none of those exists here. The one
// exception found was `propagate_early_returns.rs:199`, a second `declarations` producer, which is
// the pass `scopes.go` had already assigned elsewhere.
//
// # Its input is the GRAPH, not the scope table, which is why three stages had to land first
//
// The single most surprising thing about this pass is that it never reads a scope side table to
// find its subjects. It recovers scopes by walking blocks for a `Scope` TERMINAL, through
// `ScopeBlockTraversal` (bundle 20758). React's `keyByScopeId` (46602) is six lines of
// `if (block.terminal.kind === 'scope')`.
//
// Measured before `BuildReactiveScopeTerminals` existed, this graph held 3,306 scopes in the side
// table and 0 reachable from the CFG. That zero was the denominator, not a gap: `collectDependencies`
// keys its output map on scope ids obtained only from that traversal, and `visitDependency` and
// `visitReassignment` are both gated on the scope stack being non-empty. A port written against the
// side table would have terminated, passed every fixture, and produced nothing -- the characteristic
// failure of this pass is an empty output that looks plausible. Three stages (alignment, merging,
// terminal construction) closed it, and the count is now 2,874 reachable.
//
// The lesson generalised into the instrument this package now uses: a writer-check confirms the
// OUTPUT field belongs to this file and is silent on whether the file's INPUTS exist. Both levels
// have to be checked, and the second level is "is the type ever CONSTRUCTED", not merely declared.
// `Optional` used to be the standing example; optional-chain lowering now constructs it and its
// sidemap is an input to this pass.
//
// # DIVERGENCE FROM React: oxc sorts a collection React does not
//
// `ReactiveScopeDependencyTreeHIR::new` in oxc sorts its hoistable objects so optional-first entries
// precede non-optional ones, with a comment claiming this "matches the TS behavior". React does no
// such sort (bundle 46967): it iterates `hoistableObjects` in insertion order and raises a hard
// `CompilerError.invariant('Conflicting access types')` when two entries disagree, where oxc
// silently takes first-wins via `or_insert_with`.
//
// React is the truth, so no sort is performed here. The divergence is recorded rather than split
// because the conflict behavior is already reproduced directly: inputs retain insertion order and
// `ScopeDependencies.Conflicts` exposes disagreements rather than silently choosing the first.
//
// # Termination is a single ordered walk, with no fixpoint anywhere
//
// There is no worklist and no iteration to convergence. `CollectScopeDependencies` makes ONE pass
// over the blocks in order, maintaining a stack of active scopes. Every block is visited once, every
// instruction once, and every operand once. The dependency tree is built by insertion and read once
// by `deriveMinimalDependencies`, whose recursion is over a tree whose depth is the longest property
// path and whose nodes are finite and fixed before the walk begins.
//
// This is deliberately narrower than the passes that prepare its inputs. Non-null propagation is a
// separate bounded dataflow pass and optional-chain recovery is a separate structural traversal;
// both finish before this collector receives their sidemaps. `TestDependencyCollectionIsASinglePass`
// pins the walk here, not those producers.
//
// # The distribution that proves these dependencies are real
//
// A count cannot separate a working pass from one that collects nothing, so the headline is a
// CROSS-TABULATION of dependencies per scope against scope member count, over 400 corpus files and
// 677 outermost functions. The control cell is scopes carrying ZERO dependencies: a pass that
// collected nothing would put every scope there, and a pass that over-collected would empty it.
//
// See `TestDependencyDistributionIsReal`, which pins the shape rather than the corpus numbers.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"strconv"
	"strings"
)

// DependencyPathEntry is one step in an access path: the `.b` in `props.a.b`.
//
// `Optional` records that this step came from `?.` rather than `.`. It is part of the entry's
// IDENTITY rather than a decoration: upstream's `areEqualPaths` compares it, so `a?.b` and `a.b` are
// two different dependencies and both can be present on one scope. Reproduced rather than collapsed
// because collapsing them would silently merge a null-guarded read with an unguarded one.
type DependencyPathEntry struct {
	Property string
	Optional bool
}

// ReactiveScopeDependency is one input to a reactive scope: an access path rooted at a value.
//
// The root is an `IdentifierId` and the path may be empty, which is the ordinary case for a bare
// value used directly. `Reactive` is carried from the `Place` the dependency was read through, and
// is upstream's `reactive` field; it is not consulted by this pass and exists for the pruning pass
// that would consume it (`prune_non_reactive_dependencies.rs:230` retains on exactly this).
type ReactiveScopeDependency struct {
	Identifier static_single_assignment.IdentifierId
	Reactive   bool
	Path       []DependencyPathEntry
}

// equalPaths reports whether two access paths are identical, upstream's `areEqualPaths`.
//
// Compares `Optional` as well as `Property`, which is what makes `a?.b` and `a.b` distinct. A plain
// property-name comparison passes every fixture built from unguarded accesses and silently merges
// the guarded ones, so the flag is compared here rather than at the call site.
func equalPaths(a, b []DependencyPathEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index].Property != b[index].Property || a[index].Optional != b[index].Optional {
			return false
		}
	}
	return true
}

// DependencyGap names a dependency rule this pass cannot apply, for a caller that needs to know.
//
// Modelled on `RangeGap`, `ScopeGap` and `ScopeTerminalsGap` deliberately and for the same reason: a
// gap answerable through the API is one a consumer can decline on, while a gap living only in a
// comment is one the next reader inherits by accident.
type DependencyGap uint8

const (
	// DependencyGapOptionalChains is the residual optional-chain surface not yet lowered into the
	// control-flow shape this pass consumes.
	//
	// Dot-property chains are implemented: lowering emits Optional/Branch/value blocks and
	// `CollectOptionalChainSidemap` reconstructs their paths, processed instructions, test terminals
	// and hoistable objects. Computed optional links and optional calls still use the older
	// instruction flags/call lowering, so they do not enter that protocol. A consumer requiring full
	// optional-chain coverage must still decline on this gap until those two forms join the same CFG.
	DependencyGapOptionalChains DependencyGap = iota

	// DependencyGapTypeExclusions is upstream's two type-based dependency filters, not applied.
	//
	// `checkValidDependency` rejects two kinds of value outright: a ref's `.current` (via
	// `isRefValueType`) and an object method (`Type::ObjectMethod`). Both are EXCLUSIONS, so their
	// absence can only ADD dependencies, never drop one.
	//
	// This IR carries no type (measured when it still had an always-nil type field: 0 of 110,265
	// identifiers over the corpus carried one). So neither predicate is expressible. The
	// exposure was measured rather than estimated: 7 `ObjectMethod` instructions and 3 `.current`
	// property loads across 4,766 loads in 677 functions. The object-method half is recoverable
	// structurally -- an `ObjectMethod` INSTRUCTION is identifiable without types -- and is applied
	// here; only the ref half is genuinely absent, and it is three accesses.
	DependencyGapTypeExclusions
)

// DependencyGaps are the rules CollectScopeDependencies does not apply. See DependencyGap.
//
// Returned as a value rather than documented alone so a test can assert on it, which makes closing a
// gap a visible event rather than a silent improvement.
func DependencyGaps() []DependencyGap {
	return []DependencyGap{
		DependencyGapOptionalChains,
		DependencyGapTypeExclusions,
	}
}

// ScopeDependencies is the table this pass produces: the three scope fields, keyed by scope.
//
// A side table for the reason `ReactiveScopes` is one: the scope lives in a table rather than on the
// identifier, so its outputs live beside it. The zero value is empty and ready to read.
type ScopeDependencies struct {
	// temporaries maps a temporary to the access path it holds, kept so the memoization validator
	// can normalize an inferred dependency into the shape a source dependency is written in.
	//
	// Upstream's `ValidatePreservedManualMemoization` holds its own copy of this map, built by the
	// same walk that collects dependencies. Recording it here rather than rebuilding it is what
	// keeps the two answers identical: a second `collectTemporaries` over a tree three passes later
	// would see a rewritten graph and resolve some paths differently.
	temporaries temporaries

	dependencies map[ScopeId][]ReactiveScopeDependency
	declarations map[ScopeId][]static_single_assignment.IdentifierId
	// declarationOrigin names the scope a declared value actually originated in.
	//
	// `declarations` records a value into every enclosing scope on its stack, which is upstream's
	// behaviour and is what lets an outer scope see a value produced within it. The key therefore
	// says which scope holds the declaration, not which one produced it, and those differ whenever
	// a value bubbles up.
	//
	// Upstream keeps the origin in the map's value: `scope.declarations.set(id, {identifier, scope:
	// originalDeclaration.scope.value})`. This table dropped it on the reasoning that "the scope is
	// the map key, so the pair would store it twice" -- which is true for the key and false for the
	// origin. `pruneUnusedScopes` is the pass that needs the difference: its `hasOwnDeclaration`
	// prunes a scope whose declarations all bubbled up from inner ones, and without the origin that
	// question cannot be asked at all.
	declarationOrigin map[static_single_assignment.IdentifierId]ScopeId
	reassignments     map[ScopeId][]static_single_assignment.IdentifierId
	order             []ScopeId
	// conflicts counts hoistable entries that disagreed about an access type. Upstream raises an
	// invariant on these; see `hoistableTreeFor`. Measured at zero on the corpus.
	conflicts int
}

// DependenciesOf returns a scope's inputs, or nil for a scope with none.
//
// The slice is the table's own and must not be modified by a caller.
func (d *ScopeDependencies) DependenciesOf(scope ScopeId) []ReactiveScopeDependency {
	if d == nil || d.dependencies == nil {
		return nil
	}
	return d.dependencies[scope]
}

// DeclarationsOf returns the values a scope produces that are used outside it.
//
// Upstream stores `(IdentifierId, ReactiveScopeDeclaration)` pairs where the second names the
// declaring scope. The scope is the map key here, so the pair would store it twice.
func (d *ScopeDependencies) DeclarationsOf(scope ScopeId) []static_single_assignment.IdentifierId {
	if d == nil || d.declarations == nil {
		return nil
	}
	return d.declarations[scope]
}

// OriginOf returns the scope a declared value was produced in, and whether it is known.
//
// Distinct from the key of `DeclarationsOf`, which names a scope that holds the declaration and may
// be an enclosing one. `pruneUnusedScopes` prunes a scope whose declarations all came from within
// it, and that is the only question this answers.
func (d *ScopeDependencies) OriginOf(identifier static_single_assignment.IdentifierId) (ScopeId, bool) {
	if d == nil || d.declarationOrigin == nil {
		return 0, false
	}
	origin, found := d.declarationOrigin[identifier]
	return origin, found
}

// PruneDeclarationsLastUsedBefore drops a scope's declarations that are not read after it ends.
//
// Upstream's `updateScopeDeclarations`, which runs after a merge widens a scope's range: a value
// declared by the survivor but last read inside the range it absorbed is no longer an output of the
// merged scope, because nothing outside reads it.
//
// # Why this is a mutator on a table that is otherwise read-only
//
// Everything else here is written once by the collector and read thereafter. This is the exception
// because the scope range it depends on is not final when the collector runs -- the merge widens it
// afterwards -- so the pruning cannot happen at collection time. Upstream has the same shape and the
// same reason.
//
// # When this matters, since it did not for the first consumer
//
// A wider declaration set is safe wherever declarations are read to permit something:
// `pruneUnusedScopes` keeps a scope that declares a value of its own, so extra declarations keep
// extra scopes and cost only granularity.
//
// It is not safe where they are read to deny. `pruneAlwaysInvalidatingScopes` uses declarations as a
// propagation channel -- every declaration naming an always-invalidating value adds that value to
// the unmemoized set, which prunes further scopes downstream -- so a declaration upstream would have
// dropped prunes a scope upstream keeps, and the rule reads the pruned set.
//
// A declaration with no recorded last usage is kept, which is the conservative direction: a value
// this pass knows nothing about is not a value proven dead.
func (d *ScopeDependencies) PruneDeclarationsLastUsedBefore(scope ScopeId, end static_single_assignment.EvaluationOrder,
	usage *LastUsage, function *Function) int {
	if d == nil || d.declarations == nil || usage == nil || function == nil {
		return 0
	}
	held := d.declarations[scope]
	if len(held) == 0 {
		return 0
	}

	kept := make([]static_single_assignment.IdentifierId, 0, len(held))
	for _, declared := range held {
		lastUsedAt, found := usage.LastUsedAt(declarationOf(function, declared))
		if found && lastUsedAt < end {
			continue
		}
		kept = append(kept, declared)
	}
	removed := len(held) - len(kept)
	if removed > 0 {
		d.declarations[scope] = kept
	}
	return removed
}

// PruneNonReactiveDependenciesOf drops a scope's dependencies whose root is not reactive.
//
// Upstream's deletion inside `pruneNonReactiveDependencies`. A mutator on an otherwise read-only
// table for the same reason `PruneDeclarationsLastUsedBefore` is one: the fact it depends on --
// whether a value can change between renders -- is not known when the collector runs.
//
// This is the only pass in phase 6 that changes a value the rule reads directly. The rule compares
// dependency sets, so a set left wider than upstream's is a set that will not correspond.
func (d *ScopeDependencies) PruneNonReactiveDependenciesOf(scope ScopeId,
	reactive map[static_single_assignment.IdentifierId]bool) int {
	if d == nil || d.dependencies == nil || reactive == nil {
		return 0
	}
	held := d.dependencies[scope]
	if len(held) == 0 {
		return 0
	}

	kept := make([]ReactiveScopeDependency, 0, len(held))
	for _, dependency := range held {
		if reactive[dependency.Identifier] {
			kept = append(kept, dependency)
		}
	}
	removed := len(held) - len(kept)
	if removed > 0 {
		d.dependencies[scope] = kept
	}
	return removed
}

// ReassignmentsOf returns the bindings a scope reassigns.
func (d *ScopeDependencies) ReassignmentsOf(scope ScopeId) []static_single_assignment.IdentifierId {
	if d == nil || d.reassignments == nil {
		return nil
	}
	return d.reassignments[scope]
}

// Ids returns every scope this pass recorded, in the order it closed them.
//
// Returned as a slice rather than by ranging a map so a caller iterating scopes gets the same
// sequence every run. Go map iteration is randomised, and a pass whose output order varies makes a
// cache non-reproducible without ever producing a wrong answer.
func (d *ScopeDependencies) Ids() []ScopeId {
	if d == nil {
		return nil
	}
	return d.order
}

// Len reports how many scopes carry an entry, for measurement.
func (d *ScopeDependencies) Len() int {
	if d == nil {
		return 0
	}
	return len(d.order)
}

// ---------------------------------------------------------------------------
// The dependency tree, and the minimal set it derives
// ---------------------------------------------------------------------------

// propertyAccessType is how a node in the dependency tree was reached.
//
// Two independent bits packed into upstream's four-variant enum: whether the access was guarded by
// `?.` (optional) or not (unconditional), and whether the node is itself a DEPENDENCY or merely a
// waypoint on the path to one. `mergeAccess` below is what combines two readings of the same node.
type propertyAccessType uint8

const (
	optionalAccess propertyAccessType = iota
	unconditionalAccess
	optionalDependency
	unconditionalDependency
)

func isOptionalAccess(access propertyAccessType) bool {
	return access == optionalAccess || access == optionalDependency
}

func isDependencyAccess(access propertyAccessType) bool {
	return access == optionalDependency || access == unconditionalDependency
}

// mergeAccess combines two readings of one tree node, upstream's `merge`.
//
// The two bits combine differently and that asymmetry is the point. "Dependency" is a union: reached
// as a dependency on either path, it is one. "Unconditional" is an INTERSECTION expressed by De
// Morgan -- the result is optional only when BOTH readings were optional, so a value read once
// guarded and once unguarded is unguarded, because the unguarded read already happened.
func mergeAccess(a, b propertyAccessType) propertyAccessType {
	unconditional := !(isOptionalAccess(a) && isOptionalAccess(b))
	dependency := isDependencyAccess(a) || isDependencyAccess(b)
	switch {
	case unconditional && dependency:
		return unconditionalDependency
	case unconditional:
		return unconditionalAccess
	case dependency:
		return optionalDependency
	default:
		return optionalAccess
	}
}

// dependencyNode is one value in the access-path tree a scope's dependencies form.
//
// `properties` is ordered rather than a bare map because `deriveMinimalDependencies` walks it and
// the walk's order becomes the output's order. Upstream uses a JS `Map`, which is insertion-ordered;
// a Go map is randomised, so insertion order is kept explicitly in `keys`.
type dependencyNode struct {
	properties map[string]*dependencyNode
	keys       []string
	access     propertyAccessType
	reactive   bool
}

func newDependencyNode(access propertyAccessType, reactive bool) *dependencyNode {
	return &dependencyNode{properties: map[string]*dependencyNode{}, access: access, reactive: reactive}
}

// makeOrMergeProperty is upstream's `makeOrMergeProperty`: descend one step, creating or merging.
func (n *dependencyNode) makeOrMergeProperty(property string, access propertyAccessType) *dependencyNode {
	if existing, ok := n.properties[property]; ok {
		existing.access = mergeAccess(existing.access, access)
		return existing
	}
	child := newDependencyNode(access, n.reactive)
	n.properties[property] = child
	n.keys = append(n.keys, property)
	return child
}

// hoistableNode is one value in the tree of accesses proven safe to hoist out of a scope.
//
// This was empty when the type was written, and the comment saying so outlived the fact by several
// changes. `CollectHoistablePropertyLoads` in `hoistable.go` populates it, called from
// `CollectScopeDependenciesWithHoistable`, so a path descends here wherever that analysis proved
// the object non-null. `CollectScopeDependencies` -- the spelling that passes nil -- is the one
// that still sees an empty set, and truncates every path to its root as a result.
type hoistableNode struct {
	properties map[string]*hoistableNode
	// nonNull is upstream's `HoistableAccessType::NonNull`, meaning this access is known not to
	// throw and may therefore be read before the scope runs.
	nonNull bool
}

// dependencyTree accumulates a scope's accesses and reduces them to a minimal set.
//
// Upstream's `ReactiveScopeDependencyTreeHIR`. Two trees, not one: `hoistable` is what may be read
// early, `roots` is what was actually read, and `addDependency` walks them in lockstep.
type dependencyTree struct {
	hoistable map[static_single_assignment.IdentifierId]*hoistableNode
	roots     map[static_single_assignment.IdentifierId]*dependencyNode
	rootOrder []static_single_assignment.IdentifierId
}

// newDependencyTree builds the tree, seeding it with the accesses proven hoistable.
//
// # DIVERGENCE FROM oxc, recorded at the line that resolves it
//
// oxc sorts `hoistable_objects` here so that optional-first entries precede non-optional ones,
// commenting that this "matches the TS behavior". It does not. React iterates the collection in
// insertion order with no sort (bundle 46967) and raises `CompilerError.invariant('Conflicting
// access types')` when two entries disagree about a property, where oxc silently keeps the first.
//
// React is the truth, so there is no sort. The divergence WAS unobservable because `hoistable` was
// always empty, and that is no longer the state: `CollectHoistablePropertyLoads` is built and
// supplies real entries. Measured on `useMemo-alias-property-load-dep.ts`, the tree seeds three
// roots, one carrying a property and marked non-null. So ordering and conflicts are now reachable
// and this note is a live caveat rather than a dormant one.
func newDependencyTree(hoistable map[static_single_assignment.IdentifierId]*hoistableNode) *dependencyTree {
	if hoistable == nil {
		hoistable = map[static_single_assignment.IdentifierId]*hoistableNode{}
	}
	return &dependencyTree{hoistable: hoistable, roots: map[static_single_assignment.IdentifierId]*dependencyNode{}}
}

// addDependency records one access path, truncating it where the path stops being safe to hoist.
//
// # How a path truncates, and why truncation is the safe direction
//
// The loop walks the access path and the hoistable tree together. An entry survives only if the
// hoistable cursor says the object is already known non-null, or the entry itself is optional; a
// non-optional entry under an unknown object hits `break` and the path stops there.
//
// Under an empty hoistable set the cursor is nil at every step, so `props.a.b` truncates to bare
// `props`. That is not an approximation of upstream -- it is exactly what React does when handed an
// empty hoistable set (bundle 47019, the `else { break; }` arm), and it is the SAFE direction: the
// scope is reported as depending on all of `props` rather than on `props.a.b`, so it invalidates
// more often than strictly necessary and never less. `CollectScopeDependencies` is the spelling
// that still sees that; `CollectScopeDependenciesWithHoistable` supplies a real set and descends.
//
// # The open question this comment used to pose is answered, and it was not about this tree
//
// The question was why a BARE access is recorded alongside the deep path, given the tree then emits
// the root and prunes the deep path under it as redundant -- upstream's own shape in
// `collectMinimalInSubtree`, and correct given the root is there.
//
// The root should not have been there. `t = propA` records `t -> propA` in the temporaries sidemap,
// and the instruction reading `t` resolves through that sidemap and records the FULL path. Visiting
// the intermediate instruction's operands as well recorded the prefix a second time, as a bare root.
// Upstream skips such an instruction entirely; see `isDeferredDependency`, now transcribed.
//
// Measured on the corpus, that duplicate submission was the whole gap: 71 bare roots were replaced
// by 147 specific paths, every added path an extension of a removed root and every removed root
// replaced. The path distribution moved from 237 deep / 2,205 flat to 386 / 2,136.
func (t *dependencyTree) addDependency(dep ReactiveScopeDependency) {
	root, ok := t.roots[dep.Identifier]
	if !ok {
		root = newDependencyNode(unconditionalAccess, dep.Reactive)
		t.roots[dep.Identifier] = root
		t.rootOrder = append(t.rootOrder, dep.Identifier)
	}

	cursor := root
	var hoistableCursor *hoistableNode
	if seed, present := t.hoistable[dep.Identifier]; present {
		hoistableCursor = seed
	}

	for _, entry := range dep.Path {
		var nextHoistable *hoistableNode
		var access propertyAccessType
		switch {
		case hoistableCursor != nil && hoistableCursor.nonNull:
			nextHoistable = hoistableCursor.properties[entry.Property]
			access = unconditionalAccess
		case entry.Optional:
			if hoistableCursor != nil {
				nextHoistable = hoistableCursor.properties[entry.Property]
			}
			access = optionalAccess
		default:
			// The path stops being hoistable here. See this function's comment: with no null
			// analysis this is reached on the first non-optional entry, which is upstream's own
			// behaviour under an empty hoistable set rather than a shortcut taken here.
			cursor.access = mergeAccess(cursor.access, optionalDependency)
			return
		}
		cursor = cursor.makeOrMergeProperty(entry.Property, access)
		hoistableCursor = nextHoistable
	}

	cursor.access = mergeAccess(cursor.access, optionalDependency)
}

// deriveMinimalDependencies returns the shallowest set of paths covering every access.
//
// A node marked as a dependency is emitted and its children are NOT walked, because depending on
// `props.a` already covers `props.a.b`. That pruning is the whole point of the tree: a scope reading
// `props.a`, `props.a.b` and `props.a.c` reports one dependency rather than three.
func (t *dependencyTree) deriveMinimalDependencies() []ReactiveScopeDependency {
	var results []ReactiveScopeDependency
	for _, rootId := range t.rootOrder {
		root := t.roots[rootId]
		collectMinimalInSubtree(root, rootId, nil, &results)
	}
	return results
}

// collectMinimalInSubtree is upstream's `collectMinimalDependenciesInSubtree`.
//
// Recursion depth is the longest access path, and the tree is finite and fully built before this
// runs, so this terminates without any convergence argument. See the package comment on termination.
func collectMinimalInSubtree(node *dependencyNode, root static_single_assignment.IdentifierId,
	path []DependencyPathEntry, results *[]ReactiveScopeDependency) {
	if isDependencyAccess(node.access) {
		copied := make([]DependencyPathEntry, len(path))
		copy(copied, path)
		*results = append(*results, ReactiveScopeDependency{
			Identifier: root,
			Reactive:   node.reactive,
			Path:       copied,
		})
		return
	}
	for _, key := range node.keys {
		child := node.properties[key]
		collectMinimalInSubtree(child, root, append(path, DependencyPathEntry{
			Property: key,
			Optional: isOptionalAccess(child.access),
		}), results)
	}
}

// ---------------------------------------------------------------------------
// The temporaries sidemap: resolving a value back to the access path that made it
// ---------------------------------------------------------------------------

// temporaries maps a value to the access path it holds, so `t = props.a; t.b` reads as `props.a.b`.
//
// Lowering breaks every expression into single instructions, so `props.a.b` becomes a `PropertyLoad`
// into a temporary and a second `PropertyLoad` off it. Without this map the second load's object is
// an anonymous value and the dependency would be recorded as that temporary rather than as a path
// through `props`. This is what reassembles them.
type temporaries map[static_single_assignment.IdentifierId]ReactiveScopeDependency

// getProperty resolves one property access into a dependency, upstream's `getProperty`.
//
// When the object is itself a known temporary the access EXTENDS that temporary's path; otherwise it
// starts a new path rooted at the object. The extension copies the path rather than appending in
// place, because two accesses off one temporary must not share a backing array.
func (t temporaries) getProperty(object Place, property string, optional bool) ReactiveScopeDependency {
	if resolved, ok := t[object.Identifier]; ok {
		path := make([]DependencyPathEntry, len(resolved.Path), len(resolved.Path)+1)
		copy(path, resolved.Path)
		path = append(path, DependencyPathEntry{Property: property, Optional: optional})
		return ReactiveScopeDependency{
			Identifier: resolved.Identifier,
			Reactive:   resolved.Reactive,
			Path:       path,
		}
	}
	return ReactiveScopeDependency{
		Identifier: object.Identifier,
		Reactive:   object.Reactive,
		Path:       []DependencyPathEntry{{Property: property, Optional: optional}},
	}
}

// resolve returns the dependency a place names, following the temporaries map.
func (t temporaries) resolve(place Place) ReactiveScopeDependency {
	if resolved, ok := t[place.Identifier]; ok {
		path := make([]DependencyPathEntry, len(resolved.Path))
		copy(path, resolved.Path)
		return ReactiveScopeDependency{
			Identifier: resolved.Identifier,
			Reactive:   resolved.Reactive,
			Path:       path,
		}
	}
	return ReactiveScopeDependency{Identifier: place.Identifier, Reactive: place.Reactive}
}

// collectTemporaries builds the sidemap, upstream's `collectTemporariesSidemap`.
//
// A value is recorded only when it is NOT used outside the scope that declares it: a temporary read
// from elsewhere is a real value with its own identity rather than a rename of the thing it loaded.
// `usedOutside` carries that answer, computed by `findTemporariesUsedOutsideDeclaringScope`.
//
// The `LoadLocal` arm requires the destination to be ANONYMOUS and the source to be NAMED, which is
// upstream's `lvalue.identifier.name === null && place.identifier.name !== null`. That asymmetry is
// what keeps this a map from temporaries to real values rather than a general alias table.
func collectTemporaries(function *Function, usedOutside map[static_single_assignment.DeclarationId]bool) temporaries {
	result := temporaries{}
	collectTemporariesInto(function, usedOutside, result)
	return result
}

// collectTemporariesInto walks one function, recursing into nested ones.
//
// # DIVERGENCE FROM React: nested functions get their own map here
//
// Upstream threads ONE `temporaries` map through the recursion into nested functions, keyed by
// `IdentifierId`. That is safe there because ids are unique across the whole compilation. They are
// per-function here -- `FindDisjointMutableValues` records why -- so a shared map would conflate
// values: measured on a two-closure function, 14 collisions across 30 distinct ids.
//
// Nested functions are therefore walked with a FRESH map, whose results are discarded. That is
// narrower than upstream, and the narrowing is in the safe direction: a dependency inside a nested
// function resolves to the nested value rather than to a path through the outer one, so a path is
// reported shallower rather than wrong. Recorded because a reader diffing against the bundle will
// see the shared map missing.
func collectTemporariesInto(function *Function, usedOutside map[static_single_assignment.DeclarationId]bool, into temporaries) {
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			lvalue := function.Identifiers[instruction.LValue.Identifier]
			if lvalue == nil {
				continue
			}
			if usedOutside[lvalue.Declaration] {
				continue
			}
			switch value := instruction.Value.(type) {
			case *PropertyLoad:
				into[instruction.LValue.Identifier] =
					into.getProperty(value.Object, value.Property, value.Optional)
			case *LoadLocal:
				source := function.Identifiers[value.Place.Identifier]
				if lvalue.Name == "" && source != nil && source.Name != "" {
					into[instruction.LValue.Identifier] = ReactiveScopeDependency{
						Identifier: value.Place.Identifier,
						Reactive:   value.Place.Reactive,
					}
				}
			case *LoadContext:
				source := function.Identifiers[value.Place.Identifier]
				if lvalue.Name == "" && source != nil && source.Name != "" {
					into[instruction.LValue.Identifier] = ReactiveScopeDependency{
						Identifier: value.Place.Identifier,
						Reactive:   value.Place.Reactive,
					}
				}
			}
		}
	}
	// See this function's comment: a fresh map per nested function, because IdentifierIds are
	// per-function here and a shared one would conflate values across the boundary.
	for _, nested := range function.Functions {
		if nested != nil {
			collectTemporariesInto(nested, usedOutside, temporaries{})
		}
	}
}

// findTemporariesUsedOutsideDeclaringScope returns the bindings read outside the scope that made
// them.
//
// Upstream's function of the same name (bundle 47355). A temporary read outside its declaring scope
// cannot be treated as a rename of what it loaded, because the scope boundary means the two are
// separately memoized. Keyed by `DeclarationId` rather than `IdentifierId` so that a binding
// reassigned inside a scope is one subject rather than several.
func findTemporariesUsedOutsideDeclaringScope(function *Function,
	terminals map[static_single_assignment.BlockId]scopeBlockInfo) map[static_single_assignment.DeclarationId]bool {
	declaringScope := map[static_single_assignment.DeclarationId]ScopeId{}
	usedOutside := map[static_single_assignment.DeclarationId]bool{}
	var active []ScopeId

	isActive := func(scope ScopeId) bool {
		for _, candidate := range active {
			if candidate == scope {
				return true
			}
		}
		return false
	}
	handlePlace := func(place Place) {
		identifier := function.Identifiers[place.Identifier]
		if identifier == nil {
			return
		}
		if scope, ok := declaringScope[identifier.Declaration]; ok && !isActive(scope) {
			usedOutside[identifier.Declaration] = true
		}
	}

	// Iterating the slice rather than sorting by id: `Function.Blocks` is held in reverse postorder
	// after lowering, so slice order IS execution order and `block.Id` is the addressing key. Using
	// the slice INDEX as a BlockId would be a silent correctness bug, because the two move
	// independently -- see the doc on `BlockId`.
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		if info, ok := terminals[block.Id]; ok {
			if info.begins {
				active = append(active, info.scope)
			} else if len(active) > 0 {
				active = active[:len(active)-1]
			}
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			EachInstructionPlace(instruction, func(place Place, role PlaceRole) {
				if role != PlaceRoleDefine {
					handlePlace(place)
				}
			})
			if len(active) == 0 {
				continue
			}
			switch instruction.Value.(type) {
			case *LoadLocal, *LoadContext, *PropertyLoad:
				if lvalue := function.Identifiers[instruction.LValue.Identifier]; lvalue != nil {
					declaringScope[lvalue.Declaration] = active[len(active)-1]
				}
			}
		}
		EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) {
			handlePlace(place)
		})
	}
	return usedOutside
}

// ---------------------------------------------------------------------------
// The scope traversal, and the walk that collects
// ---------------------------------------------------------------------------

// scopeBlockInfo marks a block as the start or the end of a reactive scope's body.
//
// Upstream's `ScopeBlockInfo`, a two-variant union of `begin` and `end`. The `pruned` flag it also
// carries is absent here: it is set only for a `PrunedScope` terminal, which this tree does not
// declare because nothing constructs one (see `ScopeTerminalsGapPrunedScope`). Its two uses upstream
// are both SKIPS -- exclude the scope from `usedOutsideDeclaringScope`, and do not record its deps
// -- so a uniformly-false flag records every scope rather than wrongly dropping any.
type scopeBlockInfo struct {
	scope  ScopeId
	begins bool
}

// scopeBlockTraversal maps each block to the scope it begins or ends, upstream's
// `ScopeBlockTraversal.blockInfos`.
//
// This is the function that made three earlier stages necessary. It finds scopes by looking for a
// `Scope` TERMINAL in the graph rather than by consulting a table, so before
// `BuildReactiveScopeTerminals` landed it returned an empty map on every function and every
// consumer of it silently produced nothing.
func scopeBlockTraversal(function *Function) map[static_single_assignment.BlockId]scopeBlockInfo {
	infos := map[static_single_assignment.BlockId]scopeBlockInfo{}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		if scope, ok := block.Terminal.(*Scope); ok {
			infos[scope.Block] = scopeBlockInfo{scope: scope.Scope, begins: true}
			infos[scope.Fallthrough] = scopeBlockInfo{scope: scope.Scope}
		}
	}
	return infos
}

// declaration records where a value was declared and which scopes were open at the time.
//
// Upstream's `Decl`. The scope stack is what `visitDependency` consults to decide that a value
// produced inside one scope and read inside another must be recorded as a DECLARATION of the first.
type declaration struct {
	order      static_single_assignment.EvaluationOrder
	scopeStack []ScopeId
}

// dependencyCollector is the walk's state, upstream's `DependencyCollectionContext`.
type dependencyCollector struct {
	function    *Function
	temporaries temporaries

	// processedOptionalInstructions are the StoreLocal instructions whose values were recovered by
	// the optional-chain traversal. Visiting their operands here would submit a chain prefix in
	// addition to the full path submitted at the phi or another site of use. This is the instruction
	// half of upstream's `processedInstrsInOptional` set.
	processedOptionalInstructions map[InstructionId]bool

	// processedOptionalTests are the blocks whose `Branch` terminal an optional chain was recovered
	// from. Upstream gates its terminal operand walk on `isDeferredDependency`, whose set is typed
	// `Instruction | Terminal` and holds the matched test (`CollectOptionalChainDependencies.ts:411`,
	// `PropagateScopeDependenciesHIR.ts:833`). Without the gate the test's operand is submitted as a
	// dependency in its own right, which is a shallower access than the chain it belongs to and wins
	// the shallowest-node reduction.
	processedOptionalTests map[static_single_assignment.BlockId]bool
	scopeStack             []ScopeId
	dependencies           [][]ReactiveScopeDependency
	declarations           map[static_single_assignment.DeclarationId]declaration
	reassignments          map[static_single_assignment.IdentifierId]declaration
	objectMethods          map[static_single_assignment.IdentifierId]bool
	result                 *ScopeDependencies
	// scopeRange answers a scope's final range, which `checkValidDependency` compares against.
	scopeRange func(ScopeId) mutation_aliasing.MutableRange
	// hoistable is the per-scope set of accesses proven safe to read before the scope runs, from
	// `CollectHoistablePropertyLoads`. Nil is a valid and meaningful value: it truncates every
	// dependency path to its root, which is exactly what this pass did before that analysis existed.
	hoistable map[ScopeId][]ReactiveScopeDependency
	// nestedHoistable is the prototype seam: paths that exist only inside a nested function, which
	// `CollectHoistablePropertyLoads` never sees because it does not descend either.
	nestedHoistable map[ScopeId][]ReactiveScopeDependency
	assumedInvoked  AssumedInvokedFunctions
}

func (c *dependencyCollector) currentScope() (ScopeId, bool) {
	if len(c.scopeStack) == 0 {
		return 0, false
	}
	return c.scopeStack[len(c.scopeStack)-1], true
}

func (c *dependencyCollector) enterScope(scope ScopeId) {
	c.scopeStack = append(c.scopeStack, scope)
	c.dependencies = append(c.dependencies, nil)
}

// exitScope closes a scope, recording its dependencies and propagating them outward.
//
// The propagation is what makes a nested scope's inputs also inputs of its parent: a value the inner
// scope needs must be available before the outer one runs. Only dependencies that are still VALID
// at the outer level propagate, which `checkValidDependency` decides.
func (c *dependencyCollector) exitScope(scope ScopeId) {
	if len(c.scopeStack) == 0 {
		return
	}
	scoped := c.dependencies[len(c.dependencies)-1]
	c.dependencies = c.dependencies[:len(c.dependencies)-1]
	c.scopeStack = c.scopeStack[:len(c.scopeStack)-1]

	for _, dep := range scoped {
		if len(c.dependencies) > 0 && c.checkValidDependency(dep) {
			c.dependencies[len(c.dependencies)-1] =
				append(c.dependencies[len(c.dependencies)-1], dep)
		}
	}

	if c.result.dependencies == nil {
		c.result.dependencies = map[ScopeId][]ReactiveScopeDependency{}
	}
	if _, seen := c.result.dependencies[scope]; !seen {
		c.result.order = append(c.result.order, scope)
	}
	c.result.dependencies[scope] = append(c.result.dependencies[scope], scoped...)
}

func (c *dependencyCollector) declare(id static_single_assignment.IdentifierId, decl declaration) {
	identifier := c.function.Identifiers[id]
	if identifier == nil {
		return
	}
	if _, exists := c.declarations[identifier.Declaration]; !exists {
		c.declarations[identifier.Declaration] = decl
	}
}

// checkValidDependency decides whether a value read inside a scope is an INPUT to it.
//
// A value is a dependency only when it was declared BEFORE the scope began. A value produced inside
// the scope is not an input to it -- it is part of what the scope computes. That comparison is the
// whole predicate, and it is why the scope's final range matters: comparing against a pre-alignment
// range would misclassify values declared in the widened region.
//
// The two type-based exclusions upstream also applies are `DependencyGapTypeExclusions`. The object
// method half is recovered structurally here rather than through a type, because an `ObjectMethod`
// instruction is identifiable without one; the ref half is not expressible and is 3 accesses.
func (c *dependencyCollector) checkValidDependency(dep ReactiveScopeDependency) bool {
	if c.objectMethods[dep.Identifier] {
		return false
	}
	scope, ok := c.currentScope()
	if !ok {
		return false
	}
	identifier := c.function.Identifiers[dep.Identifier]
	if identifier == nil {
		return false
	}
	decl, found := c.reassignments[dep.Identifier]
	if !found {
		decl, found = c.declarations[identifier.Declaration]
	}
	if !found {
		return false
	}
	return decl.order < c.scopeRange(scope).Start
}

// visitDependency records one access, and records a DECLARATION when it crosses a scope boundary.
//
// The second half is the `declarations` field: a value produced inside a scope and read outside it
// must be exported from that scope, because the memoized block has to return it.
func (c *dependencyCollector) visitDependency(dep ReactiveScopeDependency) {
	// A `.current` read is not a dependency; the object holding it is.
	//
	// Upstream's `visitDependency` does the same, with the comment "ref.current access is not a
	// valid dep", and truncates the path to empty. Its guard is
	// `isUseRefType(identifier) && path[0].property === 'current'`, and the type half is not
	// expressible here -- the IR carries no type, which is what `DependencyGapTypeExclusions`
	// records.
	//
	// Taking the name half alone is a DIVERGENCE and is measured rather than assumed safe: a value
	// named `current` on a non-ref object would be truncated where upstream keeps the path, which
	// costs precision in the safe direction -- the scope depends on the whole object and
	// invalidates more often, never less.
	if len(dep.Path) > 0 && dep.Path[0].Property == "current" {
		dep = ReactiveScopeDependency{
			Identifier: dep.Identifier,
			Reactive:   dep.Reactive,
		}
	}
	identifier := c.function.Identifiers[dep.Identifier]
	if identifier == nil {
		return
	}
	if original, ok := c.declarations[identifier.Declaration]; ok && len(original.scopeStack) > 0 {
		for _, declaringScope := range original.scopeStack {
			if c.scopeIsActive(declaringScope) {
				continue
			}
			if c.result.declarations == nil {
				c.result.declarations = map[ScopeId][]static_single_assignment.IdentifierId{}
			}
			if !containsIdentifier(c.result.declarations[declaringScope], dep.Identifier) {
				c.result.declarations[declaringScope] =
					append(c.result.declarations[declaringScope], dep.Identifier)
			}
			// The origin is the innermost scope on the stack recorded at declaration time, which is
			// the last entry: the stack is pushed outermost-first as scopes open.
			//
			// Written unconditionally rather than first-writer-wins. A guard was there and a
			// mutation removing it changed nothing, which is correct: `original.scopeStack` is
			// fixed per identifier, so every enclosing scope in this loop writes the same value.
			// The guard read as protection against a later scope overwriting the origin with
			// itself, and that cannot happen because the value written does not depend on
			// `declaringScope` at all.
			if c.result.declarationOrigin == nil {
				c.result.declarationOrigin = map[static_single_assignment.IdentifierId]ScopeId{}
			}
			c.result.declarationOrigin[dep.Identifier] =
				original.scopeStack[len(original.scopeStack)-1]
		}
	}
	if len(c.dependencies) == 0 {
		return
	}
	if c.checkValidDependency(dep) {
		c.dependencies[len(c.dependencies)-1] =
			append(c.dependencies[len(c.dependencies)-1], dep)
	}
}

// visitReassignment records that the current scope writes to a binding declared outside it.
func (c *dependencyCollector) visitReassignment(place Place) {
	scope, ok := c.currentScope()
	if !ok {
		return
	}
	if !c.checkValidDependency(ReactiveScopeDependency{
		Identifier: place.Identifier, Reactive: place.Reactive}) {
		return
	}
	if c.result.reassignments == nil {
		c.result.reassignments = map[ScopeId][]static_single_assignment.IdentifierId{}
	}
	target := c.function.Identifiers[place.Identifier]
	if target == nil {
		return
	}
	for _, existing := range c.result.reassignments[scope] {
		if other := c.function.Identifiers[existing]; other != nil &&
			other.Declaration == target.Declaration {
			return
		}
	}
	c.result.reassignments[scope] = append(c.result.reassignments[scope], place.Identifier)
}

func (c *dependencyCollector) scopeIsActive(scope ScopeId) bool {
	for _, candidate := range c.scopeStack {
		if candidate == scope {
			return true
		}
	}
	return false
}

func containsIdentifier(list []static_single_assignment.IdentifierId, id static_single_assignment.IdentifierId) bool {
	for _, existing := range list {
		if existing == id {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// The pass
// ---------------------------------------------------------------------------

// CollectScopeDependencies computes every reactive scope's inputs, declarations and reassignments.
//
// Requires a graph whose reactive scopes have been built into `Scope` TERMINALS: call
// `BuildReactiveScopeTerminals` first. A function with no such terminals gets an empty result, which
// is the correct answer for it and is also the characteristic failure of a caller that forgot --
// `TestDependenciesRequireScopeTerminals` pins the distinction.
//
// `identity` resolves a pre-merge scope id to the scope that actually survives and answers its final
// range, exactly as `BuildReactiveScopeTerminals` requires. Passing the pre-merge table instead
// produces plausible-looking output keyed on scopes that no longer exist.
func CollectScopeDependencies(function *Function, identity ScopeIdentity) *ScopeDependencies {
	result := &ScopeDependencies{}
	if function == nil || identity == nil {
		return result
	}

	terminals := scopeBlockTraversal(function)
	if len(terminals) == 0 {
		return result
	}

	usedOutside := findTemporariesUsedOutsideDeclaringScope(function, terminals)
	collected := collectTemporaries(function, usedOutside)

	// The optional chains. `PropagateScopeDependenciesHIR.ts:61` merges the traversal's temporaries
	// into the collector's map, so a value written inside a chain resolves to the full path with its
	// optionality recorded per step rather than to a bare root with a plain one.
	optional := CollectOptionalChainSidemap(function)
	if optional != nil {
		for id, dependency := range optional.TemporariesReadInOptional {
			collected[id] = dependency
		}
	}
	result.temporaries = collected
	collector := &dependencyCollector{
		function:                      function,
		temporaries:                   collected,
		declarations:                  map[static_single_assignment.DeclarationId]declaration{},
		reassignments:                 map[static_single_assignment.IdentifierId]declaration{},
		objectMethods:                 objectMethodValues(function),
		result:                        result,
		scopeRange:                    identity.RangeOf,
		processedOptionalInstructions: optionalProcessedInstructions(optional),
		processedOptionalTests:        optionalProcessedTests(optional),
	}

	// Parameters are declared before any instruction runs, so they are available to every scope.
	// Upstream declares them with `id: UNSET`, which is zero here and compares below every scope
	// start -- which is the intent: a parameter is always declared before the scope.
	for _, param := range function.Params {
		collector.declare(param.Identifier, declaration{})
	}

	collector.walk(function, terminals)
	collector.reduce(identity)
	return result
}

// CollectScopeDependenciesWithHoistable is the pass with the null analysis supplying path depth.
//
// `CollectScopeDependencies` truncates every dependency to its bare root, because `addDependency`
// stops at the first non-optional entry it cannot prove safe to hoist -- upstream's own behaviour
// under an empty hoistable set. This spelling runs `CollectHoistablePropertyLoads` first and hands
// its answer in, which is what lets a dependency name `props.a.b` rather than `props`.
//
// Both spellings are kept, and the difference between them is the measurement that shows this
// analysis works at all: the path-length distribution moves from every path at length zero to a real
// spread. See `TestHoistableAnalysisDeepensDependencyPaths`.
//
// The `ranges` argument must be the same table the scopes were assigned from. `Conflicts` on the
// result reports hoistable entries that disagreed about an access type, which upstream raises on.
func CollectScopeDependenciesWithHoistable(function *Function, scopes *ReactiveScopes,
	identity ScopeIdentity, ranges *mutation_aliasing.MutableRanges) *ScopeDependencies {
	result := &ScopeDependencies{}
	if function == nil || scopes == nil || identity == nil || ranges == nil {
		return result
	}

	terminals := scopeBlockTraversal(function)
	if len(terminals) == 0 {
		return result
	}

	usedOutside := findTemporariesUsedOutsideDeclaringScope(function, terminals)
	collected := collectTemporaries(function, usedOutside)

	// The optional chains. `PropagateScopeDependenciesHIR.ts:61` merges the traversal's temporaries
	// into the collector's map, so a value written inside a chain resolves to the full path with its
	// optionality recorded per step rather than to a bare root with a plain one.
	optional := CollectOptionalChainSidemap(function)
	if optional != nil {
		for id, dependency := range optional.TemporariesReadInOptional {
			// Later wins, matching `new Map([...temporaries, ...temporariesReadInOptional])`.
			// The optional traversal is the authoritative reconstruction for values it consumed.
			collected[id] = dependency
		}
	}
	result.temporaries = collected
	collector := &dependencyCollector{
		function:                      function,
		temporaries:                   collected,
		declarations:                  map[static_single_assignment.DeclarationId]declaration{},
		reassignments:                 map[static_single_assignment.IdentifierId]declaration{},
		objectMethods:                 objectMethodValues(function),
		result:                        result,
		scopeRange:                    identity.RangeOf,
		hoistable:                     CollectHoistablePropertyLoads(function, scopes, identity, ranges, optionalHoistable(optional)),
		processedOptionalInstructions: optionalProcessedInstructions(optional),
		processedOptionalTests:        optionalProcessedTests(optional),
	}
	for _, param := range function.Params {
		collector.declare(param.Identifier, declaration{})
	}
	collector.walk(function, terminals)
	collector.reduce(identity)
	return result
}

// Conflicts reports hoistable entries that disagreed about whether an access was optional.
//
// Upstream raises `CompilerError.invariant('Conflicting access types')` on any of these. Measured at
// zero across 400 corpus files, so this exists to make a future non-zero visible rather than to
// describe current behaviour.
func (d *ScopeDependencies) Conflicts() int {
	if d == nil {
		return 0
	}
	return d.conflicts
}

// objectMethodValues returns the values produced by an ObjectMethod instruction.
//
// This is the recoverable half of `DependencyGapTypeExclusions`. Upstream rejects an object method
// as a dependency by consulting `Type::ObjectMethod`; types are nil in this IR, but the INSTRUCTION
// that produces one is identifiable, and its lvalue is the value upstream's type would have named.
// Seven such instructions exist across the corpus, so this is a small but exact recovery rather than
// an approximation of the type check.
func objectMethodValues(function *Function) map[static_single_assignment.IdentifierId]bool {
	methods := map[static_single_assignment.IdentifierId]bool{}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			if _, ok := instruction.Value.(*ObjectMethod); ok {
				methods[instruction.LValue.Identifier] = true
			}
		}
	}
	return methods
}

// walk is the single ordered pass over the graph, upstream's `handleFunctionDeps`.
//
// One traversal of the blocks in execution order, maintaining the scope stack. Every block, every
// instruction and every operand is visited exactly once, which is the whole termination argument:
// the work is linear in the graph and the loops are over slices fixed before the walk begins.
func (c *dependencyCollector) walk(function *Function, terminals map[static_single_assignment.BlockId]scopeBlockInfo) {
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		if info, ok := terminals[block.Id]; ok {
			if info.begins {
				c.enterScope(info.scope)
			} else {
				c.exitScope(info.scope)
			}
		}

		// Phi operands are visited before the instructions, because a phi is conceptually evaluated
		// on entry to the block, and in ascending predecessor order so the output order is the same
		// between runs.
		for _, phi := range block.Phis {
			// `PhiOperandsInOrder` returns the PREDECESSOR block ids in sorted order, the order
			// `Phi.Operands` keeps.
			//
			// # The order cannot be observed here today, and the reason is not that a fixture is
			// missing
			//
			// The real reason, measured: every phi in this tree sits OUTSIDE every reactive scope.
			// On a fixture with a multi-operand phi carrying two distinct operand values, phis
			// inside a scope = 0 and phis outside every scope = 1. `visitDependency` returns early
			// when the dependency stack is empty, so no phi operand currently reaches the output at
			// all and their order cannot be observed.
			//
			// The fixed order is kept anyway, for the reason `Sets` sorts in `scopes.go`: the
			// alternative is a latent nondeterminism that appears the day the scopes widen. This
			// verdict EXPIRES the moment a scope contains a phi -- which `AlignReactiveScopes`
			// widening a scope across a control-flow join would produce.
			for _, predecessor := range PhiOperandsInOrder(phi) {
				c.visitDependency(c.temporaries.resolve(phi.Operands.At(predecessor)))
			}
		}

		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			c.handleInstruction(instruction)
		}

		if !c.processedOptionalTests[block.Id] {
			EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) {
				if role != PlaceRoleDefine {
					c.visitDependency(c.temporaries.resolve(place))
				}
			})
		}
	}

	// Anything still open at the end is closed in reverse, so a scope whose fallthrough was never
	// reached still records its dependencies rather than discarding them.
	for len(c.scopeStack) > 0 {
		c.exitScope(c.scopeStack[len(c.scopeStack)-1])
	}
}

// handleInstruction visits one instruction, upstream's `handleInstruction`.
//
// The per-kind arms are upstream's and the order within each matters: a store visits its VALUE
// before declaring its lvalue, so `x = x + 1` reads the old `x` as a dependency rather than seeing
// its own definition.
//
// The DECLARATION is recorded before the deferral check, and before the visit, because a deferred
// instruction still defines its lvalue -- `checkValidDependency` compares against that order, so
// skipping the declare would make every later read of the temporary look undeclared.
func (c *dependencyCollector) handleInstruction(instruction *Instruction) {
	decl := declaration{order: instruction.Order, scopeStack: c.copyScopeStack()}
	c.declare(instruction.LValue.Identifier, decl)

	if c.isDeferredDependency(instruction) {
		return
	}

	switch value := instruction.Value.(type) {
	case *FunctionExpression:
		// The recursion REPLACES the bare capture visit: `EachPlace` would record each capture as
		// a rootless access, and that bare root prunes away every deeper path in the same subtree.
		c.visitNestedFunction(value.Function, value.Captures)
	case *StartMemoize:
		// The marker's operands are the dependencies the developer WROTE, and reading them as
		// dependencies of the enclosing scope is what makes writing `[x]` change what we infer.
		//
		// `EachPlace` visits `dep.Root.Place` for every entry (`visitor.go:175`), so
		// `useMemo(() => [x.y.z], [x])` visits bare `x` here. That bare access marks the dependency
		// tree's ROOT, and `collectMinimalInSubtree` then prunes every deeper path beneath a node
		// already marked -- correct pruning, given the root is there. The result is that the
		// developer's own dependency array determines the inference it is supposed to be checked
		// against, and the rule reports a disagreement it manufactured.
		//
		// # This is a divergence, and the reason it is taken rather than ported
		//
		// Upstream visits these places. `eachInstructionValueOperand` yields `dep.root.value` for
		// every `NamedLocal` (`visitors.ts:260`), their `handleInstruction` has no `StartMemoize`
		// arm so it reaches the same default operand loop, and the markers survive as far as
		// codegen. So the arm below is not a transcription of anything.
		//
		// What justifies it is upstream's own compiled output on the fixtures it moves. Measured,
		// two goldens go from miss to hit and nothing is traded:
		//
		//	useMemo-inner-decl.ts             upstream `$[0] !== data.a`   ours data.a
		//	useMemo-alias-property-load-dep.ts upstream `$[0] !== propA.x` ours propA.x
		//
		// The oracle moves 76 to 78 of 88 with `ours` unchanged at 158 and scope under-production
		// unchanged at 5 fixtures. A per-golden dump before and after differs by exactly those two
		// rows.
		//
		// Why upstream does not need it is not established. Their marker sits in the same place
		// ours does -- after the callee load, `DropManualMemoization.ts:499` -- and their
		// `checkValidDependency` is this file's line for line. Something downstream of collection
		// resolves the bare root for them, and until that is found this arm is the measured answer
		// rather than the understood one.
		_ = value
	case *PropertyLoad:
		c.visitDependency(c.temporaries.getProperty(value.Object, value.Property, value.Optional))
	case *StoreLocal:
		c.visitDependency(c.temporaries.resolve(value.Value))
		if value.Kind == InstructionKindReassign {
			c.visitReassignment(value.LValue)
			c.reassignments[value.LValue.Identifier] = decl
		}
		c.declare(value.LValue.Identifier, decl)
	case *StoreContext:
		c.visitDependency(c.temporaries.resolve(value.Value))
		if value.Kind == InstructionKindReassign {
			c.visitReassignment(value.LValue)
			c.reassignments[value.LValue.Identifier] = decl
		}
		c.declare(value.LValue.Identifier, decl)
	case *DeclareLocal:
		c.declare(value.LValue.Identifier, decl)
	case *DeclareContext:
		c.declare(value.LValue.Identifier, decl)
	case *Destructure:
		c.visitDependency(c.temporaries.resolve(value.Value))
		// The shared `eachPatternPlace` yields defaults and computed keys as USES alongside the
		// bindings, which is richer than upstream's `eachPatternOperand` and exactly right here: a
		// default value is read by the destructure and is a genuine dependency.
		eachPatternPlace(value.LValue, func(place Place, role PlaceRole) {
			if role != PlaceRoleDefine {
				c.visitDependency(c.temporaries.resolve(place))
				return
			}
			if value.Kind == InstructionKindReassign {
				c.visitReassignment(place)
				c.reassignments[place.Identifier] = decl
			}
			c.declare(place.Identifier, decl)
		})
	default:
		EachPlace(instruction.Value, func(place Place, role PlaceRole) {
			if role != PlaceRoleDefine {
				c.visitDependency(c.temporaries.resolve(place))
			}
		})
	}
}

// visitNestedFunction records what a nested function reads through its captures.
//
// PROTOTYPE. The dependency a closure creates lives inside the closure body: `useCallback(() =>
// ref.current)` produces the `.current` load only in the nested function's instruction table, and
// nothing else in this file walks in there.
//
// The translation at the boundary is `lowerNestedFunction`'s pairing: `Captures[i]` and
// `nested.Context[i]` name the same source binding from the two sides. Measured across the corpus,
// 324 pairs with zero length or name mismatches, so the map is total.
//
// Roots that do not translate are nested-local values and are dropped rather than guessed at.
func (c *dependencyCollector) visitNestedFunction(id FunctionId, captures []Place) {
	if int(id) >= len(c.function.Functions) {
		return
	}
	nested := c.function.Functions[id]
	if nested == nil || len(captures) != len(nested.Context) {
		return
	}

	// Inner identifier -> the place in THIS function naming the same binding.
	translate := map[static_single_assignment.IdentifierId]Place{}
	for i := range captures {
		translate[nested.Context[i].Identifier] = captures[i]
	}

	// The nested function's own temporaries, which `collectTemporariesInto` builds and discards.
	nestedTemporaries := temporaries{}
	collectTemporariesInto(nested, map[static_single_assignment.DeclarationId]bool{}, nestedTemporaries)

	// The nested function's own optional chains, in the nested function's own key space.
	//
	// `IdentifierId` is per-function here, so these entries cannot be lifted into the parent's map
	// -- see `traverseNested`. They belong beside the nested temporaries they were built from, where
	// `nestedAccessesByBlock` resolves reads through them and the capture translation below carries
	// the resulting access out.
	if nestedOptional := CollectOptionalChainSidemap(nested); nestedOptional != nil {
		for identifier, dependency := range nestedOptional.TemporariesReadInOptional {
			// Later wins for the same reason as the outer map: optional-chain reconstruction
			// carries information a plain PropertyLoad transcription cannot recover.
			nestedTemporaries[identifier] = dependency
		}
	}

	accesses, accessBlocks := nestedAccessesByBlock(nested, nestedTemporaries)
	// Which of the nested function's blocks run on every path through it. A deep path read only in a
	// branch still becomes a dependency below -- the value is read there -- but it does not seed the
	// hoistable set, because `nestedHoistable` is keyed by scope and would otherwise offer a
	// branch's fact to the whole scope. The CFG-derived set beside it is keyed by block already.
	//
	// # This is inert alone and worth five true positives beside the hoistable change
	//
	// Measured on its own, this moves nothing: the array loads in the enclosing function were still
	// seeding `propB.x` as non-null, so the walk descended past `x` whatever the seed here held.
	// With `dependencyArrayInstructions` excluding those loads, removing this drops the board from
	// 25 to 20. Latent rather than inert, and recorded because it read as a failed change for a
	// whole cycle before the other half landed.
	alwaysReached := blocksAlwaysReached(nested)
	if c.assumedInvoked == nil {
		c.assumedInvoked = CollectAssumedInvokedFunctions(c.function)
	}
	// A root is suppressed only when some OTHER access names it with a deeper path: that bare
	// root is the redundant evidence that prunes the path away in `collectMinimalInSubtree`.
	deepRoots := map[static_single_assignment.IdentifierId]bool{}
	for _, dep := range accesses {
		if len(dep.Path) > 0 {
			if outer, ok := translate[dep.Identifier]; ok {
				deepRoots[outer.Identifier] = true
			}
		}
	}
	for index, dep := range accesses {
		outer, ok := translate[dep.Identifier]
		if !ok {
			continue
		}
		if len(dep.Path) == 0 && deepRoots[outer.Identifier] {
			continue
		}
		translated := ReactiveScopeDependency{
			Identifier: outer.Identifier,
			Reactive:   outer.Reactive,
			Path:       dep.Path,
		}
		if c.assumedInvoked[id] && len(dep.Path) > 0 && alwaysReached[accessBlocks[index]] &&
			c.checkValidDependency(translated) {
			if scope, ok := c.currentScope(); ok {
				if c.nestedHoistable == nil {
					c.nestedHoistable = map[ScopeId][]ReactiveScopeDependency{}
				}
				c.nestedHoistable[scope] = append(c.nestedHoistable[scope], translated)
			}
		}
		c.visitDependency(translated)
	}
}

// nestedAccesses returns every access a nested function makes, rooted at its own identifiers.
func nestedAccesses(function *Function, temps temporaries) []ReactiveScopeDependency {
	accesses, _ := nestedAccessesByBlock(function, temps)
	return accesses
}

// nestedAccessesByBlock is `nestedAccesses` plus the block each access was read in.
//
// The block matters because a deep path read inside a branch is a dependency -- the value is read
// there -- but is not evidence the object is safe to load early. Upstream keys its non-null sets by
// block for exactly this reason, and a scope reads the set at its own block, so a fact established
// inside a branch is available in that branch and nowhere else.
func nestedAccessesByBlock(function *Function, temps temporaries) ([]ReactiveScopeDependency,
	[]static_single_assignment.BlockId) {
	var accesses []ReactiveScopeDependency
	var blocks []static_single_assignment.BlockId
	processedOptionalTests := map[static_single_assignment.BlockId]bool{}
	processedInstructions := map[InstructionId]bool{}
	if optional := CollectOptionalChainSidemap(function); optional != nil {
		processedOptionalTests = optional.ProcessedOptionalTests
		processedInstructions = nestedOptionalInstructionsToDefer(function, optional)
	}
	record := func(block static_single_assignment.BlockId, dependency ReactiveScopeDependency) {
		accesses = append(accesses, dependency)
		blocks = append(blocks, block)
	}
	deferred := func(instruction *Instruction) bool {
		_, ok := temps[instruction.LValue.Identifier]
		return ok
	}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil || processedInstructions[instruction.Id] || deferred(instruction) {
				continue
			}
			switch value := instruction.Value.(type) {
			case *FunctionExpression:
				// A closure inside a closure. The pairing composes, so translate through this
				// function's own context and let the caller translate again.
				if int(value.Function) >= len(function.Functions) {
					continue
				}
				inner := function.Functions[value.Function]
				if inner == nil || len(value.Captures) != len(inner.Context) {
					continue
				}
				translate := map[static_single_assignment.IdentifierId]Place{}
				for i := range value.Captures {
					translate[inner.Context[i].Identifier] = value.Captures[i]
				}
				innerTemporaries := temporaries{}
				collectTemporariesInto(inner, map[static_single_assignment.DeclarationId]bool{}, innerTemporaries)
				if innerOptional := CollectOptionalChainSidemap(inner); innerOptional != nil {
					for identifier, dependency := range innerOptional.TemporariesReadInOptional {
						innerTemporaries[identifier] = dependency
					}
				}
				for _, dep := range nestedAccesses(inner, innerTemporaries) {
					outer, ok := translate[dep.Identifier]
					if !ok {
						continue
					}
					record(block.Id, temps.resolveWithPath(outer, dep.Path))
				}
			case *PropertyLoad:
				record(block.Id,
					temps.getProperty(value.Object, value.Property, value.Optional))
			default:
				EachPlace(instruction.Value, func(place Place, role PlaceRole) {
					if role != PlaceRoleDefine {
						record(block.Id, temps.resolve(place))
					}
				})
			}
		}
		if !processedOptionalTests[block.Id] {
			EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) {
				if role != PlaceRoleDefine {
					record(block.Id, temps.resolve(place))
				}
			})
		}
	}
	return accesses, blocks
}

// nestedOptionalInstructionsToDefer returns the optional-chain stores whose operand would only
// contribute a redundant prefix while translating a closure's captures.
//
// A consecutive optional link such as the first `?.` in `x?.y?.z` is consumed by the optional
// traversal: recording its StoreLocal operand as an access would add `x?.y`, which then prunes the
// real `x?.y?.z` dependency. A non-optional continuation is deliberately different. Upstream's
// current inference keeps the prefix in `propB?.x.y` and in the known-bug `x.a.b?.c.d?.e` case;
// those prefixes are what make preserved-memoization validation reject the less-specific inferred
// dependency. Only an immediate OPTIONAL extension is therefore deferred here.
func nestedOptionalInstructionsToDefer(function *Function,
	optional *OptionalChainSidemap) map[InstructionId]bool {
	result := map[InstructionId]bool{}
	if function == nil || optional == nil {
		return result
	}
	for instructionID := range optional.ProcessedInstructions {
		instruction := function.Instructions[instructionID]
		if instruction == nil {
			continue
		}
		store, ok := instruction.Value.(*StoreLocal)
		if !ok {
			continue
		}
		prefix, ok := optional.TemporariesReadInOptional[store.LValue.Identifier]
		if !ok {
			continue
		}
		for _, candidate := range optional.TemporariesReadInOptional {
			if candidate.Identifier != prefix.Identifier || len(candidate.Path) != len(prefix.Path)+1 ||
				!candidate.Path[len(prefix.Path)].Optional ||
				!equalPaths(candidate.Path[:len(prefix.Path)], prefix.Path) {
				continue
			}
			result[instructionID] = true
			break
		}
	}
	return result
}

// blocksAlwaysReached returns the blocks of a function that run on every path through it.
//
// Post-dominance is the exact relation: a block runs on every path from entry to exit precisely when
// every such path passes through it.
func blocksAlwaysReached(function *Function) map[static_single_assignment.BlockId]bool {
	always := map[static_single_assignment.BlockId]bool{}
	if function == nil || len(function.Blocks) == 0 || function.Blocks[0] == nil {
		return always
	}
	entry := function.Blocks[0].Id
	always[entry] = true
	tree := computePostDominance(function)
	if tree == nil {
		return always
	}
	for current := entry; ; {
		next, found := tree.immediate[current]
		if !found || next == current {
			return always
		}
		always[next] = true
		current = next
	}
}

// resolveWithPath resolves a place and appends a suffix path to whatever it resolved to.
func (t temporaries) resolveWithPath(place Place, suffix []DependencyPathEntry) ReactiveScopeDependency {
	base := t.resolve(place)
	if len(suffix) == 0 {
		return base
	}
	path := make([]DependencyPathEntry, 0, len(base.Path)+len(suffix))
	path = append(path, base.Path...)
	path = append(path, suffix...)
	base.Path = path
	return base
}

// isDeferredDependency reports whether this instruction's accesses are recorded at the site of USE
// rather than here, upstream's `isDeferredDependency`.
//
// # Why an instruction that produces a dependency is skipped entirely
//
// An instruction whose lvalue is in the temporaries sidemap is not a value in its own right, it is
// one link in an access path that some later instruction will read whole. `t = propA` records
// `t -> propA`, and the instruction that reads `t` resolves through the sidemap and records the
// FULL path. Visiting the operands here as well would record the prefix a second time, as a bare
// root with an empty path.
//
// That double-recording is what truncated every path in this tree. `addDependency` submits both
// `propA` and `propA.x`, and `collectMinimalInSubtree` then emits the root and prunes the deep path
// under it as redundant -- correct pruning given the root is there, and the root should not have
// been there. Skipping the deferred instruction is upstream's answer to the open question this
// file's `addDependency` comment used to pose: the bare access is recorded because nothing was
// stopping it.
//
// Upstream also defers instructions consumed by an optional chain, via `processedInstrsInOptional`.
// The optional traversal records those StoreLocal instructions explicitly because their own
// instruction lvalues are not the stored values keyed in the temporaries sidemap.
func (c *dependencyCollector) isDeferredDependency(instruction *Instruction) bool {
	if c.processedOptionalInstructions[instruction.Id] {
		return true
	}
	_, deferred := c.temporaries[instruction.LValue.Identifier]
	return deferred
}

func (c *dependencyCollector) copyScopeStack() []ScopeId {
	if len(c.scopeStack) == 0 {
		return nil
	}
	copied := make([]ScopeId, len(c.scopeStack))
	copy(copied, c.scopeStack)
	return copied
}

// reduce turns each scope's raw accesses into the minimal covering set of dependencies.
//
// This is the second half of upstream's entry point: build a tree per scope, add every access, take
// the minimal set, and append each candidate that is not already present. The dedup predicate is
// upstream's -- same DECLARATION and equal paths -- so two values of one reassigned binding
// deduplicate against each other, which comparing identifiers would not do.
func (c *dependencyCollector) reduce(identity ScopeIdentity) {
	if c.result.dependencies == nil {
		return
	}
	for _, scope := range c.result.order {
		accesses := c.result.dependencies[scope]
		if len(accesses) == 0 {
			continue
		}
		// The hoistable set for this scope decides how DEEP each dependency path may go. An explicit
		// caller can still omit it, in which case the dependency tree conservatively truncates paths
		// to their roots; the production entry point supplies `CollectHoistablePropertyLoads`.
		seed := c.hoistable[scope]
		if extra := c.nestedHoistable[scope]; len(extra) > 0 {
			seed = appendWithoutOptionalDuplicates(seed, extra)
		}
		hoistable, conflicts := hoistableTreeFor(seed)
		c.result.conflicts += conflicts
		tree := newDependencyTree(hoistable)
		for _, access := range accesses {
			tree.addDependency(access)
		}
		var reduced []ReactiveScopeDependency
		for _, candidate := range tree.deriveMinimalDependencies() {
			if !c.alreadyPresent(reduced, candidate) {
				reduced = append(reduced, candidate)
			}
		}
		c.result.dependencies[scope] = reduced
	}
}

// alreadyPresent is upstream's dedup predicate: same declaration and equal paths.
func (c *dependencyCollector) alreadyPresent(existing []ReactiveScopeDependency,
	candidate ReactiveScopeDependency) bool {
	candidateIdentifier := c.function.Identifiers[candidate.Identifier]
	if candidateIdentifier == nil {
		return false
	}
	for _, present := range existing {
		presentIdentifier := c.function.Identifiers[present.Identifier]
		if presentIdentifier == nil {
			continue
		}
		if presentIdentifier.Declaration == candidateIdentifier.Declaration &&
			equalPaths(present.Path, candidate.Path) {
			return true
		}
	}
	return false
}

// NormalizeInferredDependency rewrites an inferred dependency into the shape a source one is in.
//
// Upstream's normalization step at the top of `validateInferredDep`
// (`ValidatePreservedManualMemoization.ts:236-264`). An inferred dependency is rooted at whatever
// identifier the collector recorded, which for `props.a.b` is the temporary holding `props.a`. A
// source dependency is rooted at what the developer wrote. Comparing the two without this step
// compares a temporary against a name and reports every dependency as a root difference.
//
// When the root is a known temporary, its path is PREPENDED to the inferred path, which is what
// reassembles `t = props.a; t.b` into `props.a.b`. Otherwise the dependency is already rooted at a
// real binding and only needs rewrapping.
//
// Returns false when the root is neither a known temporary nor a named binding. Upstream raises an
// invariant there; this declines instead, because an unnamed root that is not in the map cannot be
// compared against anything a developer could have written, and reporting it would be a finding
// about the lowering rather than about the source.
func (d *ScopeDependencies) NormalizeInferredDependency(function *Function,
	dependency ReactiveScopeDependency) (ManualMemoDependency, bool) {
	if d == nil || function == nil {
		return ManualMemoDependency{}, false
	}

	if resolved, found := d.temporaries[dependency.Identifier]; found {
		path := make([]DependencyPathEntry, 0, len(resolved.Path)+len(dependency.Path))
		path = append(path, resolved.Path...)
		path = append(path, dependency.Path...)
		return ManualMemoDependency{
			Root: ManualMemoRoot{Place: Place{Identifier: resolved.Identifier,
				Reactive: resolved.Reactive}},
			Path: path,
		}, true
	}

	if int(dependency.Identifier) >= len(function.Identifiers) {
		return ManualMemoDependency{}, false
	}
	identifier := function.Identifiers[dependency.Identifier]
	if identifier == nil || identifier.Name == "" {
		return ManualMemoDependency{}, false
	}
	path := make([]DependencyPathEntry, len(dependency.Path))
	copy(path, dependency.Path)
	return ManualMemoDependency{
		Root: ManualMemoRoot{Place: Place{Identifier: dependency.Identifier,
			Reactive: dependency.Reactive}},
		Path: path,
	}, true
}

// appendWithoutOptionalDuplicates adds paths to a hoistable seed, dropping any whose prefixes
// disagree with one already committed about optionality.
//
// `ReactiveScopeDependencyTreeHIR`'s constructor states the precondition this maintains: "we expect
// these to not contain duplicates (e.g. both `a?.b` and `a.b`) only because
// `CollectHoistablePropertyLoads` merges duplicates when traversing the CFG". Upstream gets the
// invariant for free because every path it hands the tree came through `reduceMaybeOptionalChains`
// first.
//
// Paths recovered from inside a nested function do not. They are collected by walking the closure
// body directly, so `y.a?.b` from a callback can meet `y.a.b` from the enclosing function in one
// seed and the tree then reports a conflicting access type on an access the two agree about.
// Measured: with the nested seed disabled the corpus reports zero conflicts, with it enabled and
// unguarded eleven. Every conflict is this seam rather than the collector.
//
// The comparison is per PREFIX rather than per whole path, which is what the tree keys on. Two
// paths can share a prefix and diverge after it -- `[a][b?]` against `[a][b][c?]` -- and the tree
// walks entry by entry, so it sees `b` twice with different nullability while a whole-path key sees
// two unrelated strings. Measured: keying on the whole path took conflicts 11 to 5, and every
// survivor was a prefix disagreement of exactly that shape.
//
// The existing path wins. `reduceOptionalChains` has already run over the CFG-derived set and
// flipped every guard it could prove redundant, so its answer for a shared prefix is at least as
// resolved as anything the closure walk produces.
func appendWithoutOptionalDuplicates(seed []ReactiveScopeDependency,
	extra []ReactiveScopeDependency) []ReactiveScopeDependency {
	if len(extra) == 0 {
		return seed
	}
	// Every prefix already committed, and whether its last step reads optionally. Grown as paths
	// are admitted rather than built once, so two paths within `extra` are checked against each
	// other as well as against the seed.
	committed := map[string]bool{}
	for _, path := range seed {
		recordPrefixOptionality(path, committed)
	}

	combined := append([]ReactiveScopeDependency{}, seed...)
	for _, path := range extra {
		if prefixesDisagree(path, committed) {
			continue
		}
		recordPrefixOptionality(path, committed)
		combined = append(combined, path)
	}
	return combined
}

// recordPrefixOptionality notes, for every prefix of a path, whether its last step was optional.
//
// First writer wins per prefix, matching `hoistableTreeFor`, which keeps the node it already built
// and only counts a disagreement.
func recordPrefixOptionality(path ReactiveScopeDependency, into map[string]bool) {
	var builder strings.Builder
	builder.WriteString(strconv.Itoa(int(path.Identifier)))
	for _, entry := range path.Path {
		builder.WriteByte('.')
		builder.WriteString(entry.Property)
		key := builder.String()
		if _, seen := into[key]; !seen {
			into[key] = entry.Optional
		}
	}
}

// prefixesDisagree reports whether any prefix of a path reads its step with the opposite
// optionality to what is already committed for that same prefix.
func prefixesDisagree(path ReactiveScopeDependency, committed map[string]bool) bool {
	var builder strings.Builder
	builder.WriteString(strconv.Itoa(int(path.Identifier)))
	for _, entry := range path.Path {
		builder.WriteByte('.')
		builder.WriteString(entry.Property)
		if optional, seen := committed[builder.String()]; seen && optional != entry.Optional {
			return true
		}
	}
	return false
}

// optionalHoistable is the sidemap's per-block hoistable set, or nil when there was no traversal.
func optionalHoistable(sidemap *OptionalChainSidemap) map[static_single_assignment.BlockId]ReactiveScopeDependency {
	if sidemap == nil {
		return nil
	}
	return sidemap.HoistableObjects
}

// optionalProcessedInstructions names the instructions whose operands were consumed while
// reconstructing an optional chain, or nil when no chains were found.
func optionalProcessedInstructions(sidemap *OptionalChainSidemap) map[InstructionId]bool {
	if sidemap == nil {
		return nil
	}
	return sidemap.ProcessedInstructions
}

// optionalProcessedTests names the blocks whose branch terminal an optional chain was recovered
// from, or nil when no chains were found.
func optionalProcessedTests(sidemap *OptionalChainSidemap) map[static_single_assignment.BlockId]bool {
	if sidemap == nil {
		return nil
	}
	return sidemap.ProcessedOptionalTests
}
