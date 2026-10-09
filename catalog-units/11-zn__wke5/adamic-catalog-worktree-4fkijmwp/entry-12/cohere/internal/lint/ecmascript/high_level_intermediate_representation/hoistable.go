// Hoistable property loads: which accesses may be read before their scope runs.
//
// This is React's `collectHoistablePropertyLoads` and `propagateNonNull` (development bundle lines
// 46578 and 46770), run over this graph. oxc transcribes both inside
// `react_compiler_inference/propagate_scope_dependencies_hir.rs`; where the two disagree React wins,
// and each disagreement is recorded at the line that resolves it.
//
// # What this buys, in one number
//
// `dependencies.go` shipped with every dependency truncated to its bare root: 4,426 of 4,426 with an
// empty path, so a scope reported depending on all of `props` rather than on `props.a.b`. That was
// not an approximation -- it is exactly what `addDependency` does when handed an empty hoistable
// set, upstream's `else { break; }` arm -- and the missing input is what this file computes.
//
// The reason a path may not simply be kept is the one the whole analysis exists for: reading
// `props.a.b` before the scope runs is safe only if `props.a` is known non-null at that point.
// Hoisting an access past a place where the value might be null MOVES A CRASH, or invents one that
// the original program never had. So each step of a path has to be proven reachable-without-throwing
// before the dependency may name it.
//
// # The safe direction, and the one to guard
//
// Under-approximating the hoistable set costs precision only: a path truncates early, the scope
// depends on something shallower, and it recomputes more often than it needed to. Over-approximating
// is the dangerous direction, because it produces a dependency on `props.a.b` for a program where
// `props.a` can be nullish, and the memoized read then throws where the original did not.
//
// Every gap in this file is therefore checked to fall on the under-approximating side, and
// `TestHoistablePathsAreAlwaysPrefixesOfRealAccesses` pins the invariant that a derived path
// is always a prefix of a real access rather than an invention.
//
// # Termination: TWO stories, and neither is a fixpoint proof
//
// This is the first pass in this package whose termination is a BOUND rather than an argument, and
// it is worth being exact about, because a bound that is silently exceeded produces a truncated
// answer that looks like a converged one.
//
//	the outer loop      alternating forward and backward passes, capped at 100 iterations
//	the chain reduction an unbounded do/while inside `reduceMaybeOptionalChains`
//
// The outer loop is monotone in practice -- each pass can only union new nodes into a block's set --
// but `reduceMaybeOptionalChains` REPLACES nodes rather than adding them, so the `changed` flag is
// computed by set comparison rather than by size, and the monotonicity argument does not close. That
// is why upstream caps it, and it is why the cap is reproduced here rather than removed.
//
// # DIVERGENCE FROM oxc: the bound must be LOUD, not silent
//
// React raises a hard `CompilerError.invariant(i++ < 100)` naming the pass when the loop fails to
// converge (bundle 46820). oxc writes `for _ in 0..100` and simply EXITS at the cap, so a function
// that needed a 101st pass silently gets a truncated answer that is indistinguishable from a
// converged one.
//
// React is the truth and this reproduces its behaviour, not oxc's: `HoistableAnalysis.Converged`
// reports whether the loop finished because nothing changed, and a caller reading a non-converged
// result knows the answer is partial. A linter must not panic, so this reports rather than raises --
// the same trade `ValidateScopes` made -- but the fact is never hidden.
//
// Measured over 400 corpus files and 677 functions, the convergence distribution is:
//
//	iterations    1     2     3    4+   not converged
//	functions   633    41     3     0               0
//
// The maximum observed is 3 against a cap of 100, so the bound is nowhere near being approached on
// real code. That is a materially different fact from a bound that is routinely brushed, and
// `TestHoistableConvergenceIsFarInsideTheBound` pins it so a future change that starts approaching
// the cap is a visible event rather than a slow drift into silent truncation.
//
// The inner `reduceMaybeOptionalChains` loop is unbounded upstream and is unbounded here. It
// terminates because every replacement strictly reduces the number of optional entries in a path and
// a path has finitely many entries, so no node can be replaced more times than its path is long.
// That argument is upstream's own by construction rather than stated by it, and
// `TestOptionalChainReductionTerminates` pins it on the shape that would loop if it were wrong.
//
// # DIVERGENCE FROM React: `terminalPreds` is dead code and is not reproduced
//
// React's `propagateNonNull` builds a `terminalPreds` set of every block ending in `throw` or
// `return` (bundle 46777) and never reads it again. oxc does not build it. Verified by reading every
// use of the binding in the bundle: there is exactly one, its construction. Not reproduced here,
// because copying an unread variable is transcription rather than fidelity -- the same judgement
// `RangeGapLoopCarriedInversion` made about an upstream bug.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"sort"
)

// pathNode is one access path in the registry: a root identifier plus a property chain.
//
// Upstream's `PropertyPathNode`. Nodes are interned -- the same path always resolves to the same
// index -- which is what lets the dataflow below treat a set of paths as a set of small integers and
// intersect them cheaply. `properties` and `optionalProperties` are separate maps because `a.b` and
// `a?.b` are different nodes, exactly as `equalPaths` in `dependencies.go` treats them.
type pathNode struct {
	properties         map[string]int
	optionalProperties map[string]int
	fullPath           ReactiveScopeDependency
	hasOptional        bool
}

// pathRegistry interns every access path seen, upstream's `PropertyPathRegistry`.
//
// This lives here rather than in `dependencies.go` because it belongs to this machinery: upstream
// declares it inside `CollectHoistablePropertyLoads.ts`, which is this pass, and the dependency tree
// in `dependencies.go` is a DIFFERENT structure that happens to also be a property trie. Conflating
// them is easy and was the seam error that sent one agent back; the registry indexes paths for this
// dataflow, the dependency tree reduces accesses to a minimal set.
type pathRegistry struct {
	nodes []pathNode
	roots map[static_single_assignment.IdentifierId]int
}

func newPathRegistry() *pathRegistry {
	return &pathRegistry{roots: map[static_single_assignment.IdentifierId]int{}}
}

// identifierNode returns the interned node for a bare root, creating it if absent.
func (r *pathRegistry) identifierNode(id static_single_assignment.IdentifierId, reactive bool) int {
	if index, ok := r.roots[id]; ok {
		return index
	}
	index := len(r.nodes)
	r.nodes = append(r.nodes, pathNode{
		properties:         map[string]int{},
		optionalProperties: map[string]int{},
		fullPath:           ReactiveScopeDependency{Identifier: id, Reactive: reactive},
	})
	r.roots[id] = index
	return index
}

// propertyNode returns the interned node one step below parent, creating it if absent.
//
// The new node's `fullPath` is the parent's path with the entry appended, COPIED rather than shared:
// two children of one parent must not alias a backing array, or extending one would corrupt the
// other.
func (r *pathRegistry) propertyNode(parent int, entry DependencyPathEntry) int {
	table := r.nodes[parent].properties
	if entry.Optional {
		table = r.nodes[parent].optionalProperties
	}
	if index, ok := table[entry.Property]; ok {
		return index
	}

	parentPath := r.nodes[parent].fullPath
	path := make([]DependencyPathEntry, len(parentPath.Path), len(parentPath.Path)+1)
	copy(path, parentPath.Path)
	path = append(path, entry)

	index := len(r.nodes)
	r.nodes = append(r.nodes, pathNode{
		properties:         map[string]int{},
		optionalProperties: map[string]int{},
		fullPath: ReactiveScopeDependency{
			Identifier: parentPath.Identifier,
			Reactive:   parentPath.Reactive,
			Path:       path,
		},
		hasOptional: r.nodes[parent].hasOptional || entry.Optional,
	})
	table[entry.Property] = index
	return index
}

// pathIndex interns a whole dependency, returning the node naming it.
func (r *pathRegistry) pathIndex(dep ReactiveScopeDependency) int {
	current := r.identifierNode(dep.Identifier, dep.Reactive)
	for _, entry := range dep.Path {
		current = r.propertyNode(current, entry)
	}
	return current
}

// ---------------------------------------------------------------------------
// The optional-chain reduction
// ---------------------------------------------------------------------------

// reduceOptionalChains rewrites `base?.x` to `base.x` wherever `base` is already known non-null.
//
// Upstream's `reduceMaybeOptionalChains`. Two paths differing only in `.` versus `?.` describe the
// same access once the object is proven non-null, and collapsing them is what lets the intersection
// below see two blocks as agreeing rather than as holding different nodes.
//
// # Termination, which is unbounded upstream and unbounded here
//
// The loop has no iteration cap, unlike the outer dataflow. It terminates because each replacement
// strictly decreases the number of OPTIONAL entries in the rewritten path -- an entry flips from
// optional to non-optional and never back -- and a path has finitely many entries. So a node can be
// replaced at most as many times as its path is long, and the loop runs at most that many times
// over the whole set.
//
// That argument is implicit in upstream's construction rather than stated by it, which is why it is
// pinned by `TestOptionalChainReductionTerminates` on the shape that would spin if the flip were
// ever reversible.
func reduceOptionalChains(nodes map[int]bool, registry *pathRegistry) {
	hasOptional := false
	for index := range nodes {
		if registry.nodes[index].hasOptional {
			hasOptional = true
			break
		}
	}
	if !hasOptional {
		return
	}

	for {
		changed := false
		// Sorted rather than ranged: this mutates `nodes` while deciding what to replace, and a Go
		// map's randomised order would make WHICH node is rewritten first vary between runs. The
		// answer converges either way, but the intermediate states differ and a caller diffing two
		// runs would see noise.
		for _, original := range sortedNodeIndices(nodes) {
			if !registry.nodes[original].hasOptional {
				continue
			}
			source := registry.nodes[original].fullPath
			current := registry.identifierNode(source.Identifier, source.Reactive)
			for _, entry := range source.Path {
				next := entry
				if entry.Optional && nodes[current] {
					// The object is already known non-null here, so the guard is redundant and the
					// access is unconditional.
					next = DependencyPathEntry{Property: entry.Property}
				}
				current = registry.propertyNode(current, next)
			}
			if current != original {
				changed = true
				delete(nodes, original)
				nodes[current] = true
			}
		}
		if !changed {
			return
		}
	}
}

// sortedNodeIndices returns a set's members in a stable order.
func sortedNodeIndices(nodes map[int]bool) []int {
	indices := make([]int, 0, len(nodes))
	for index := range nodes {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	return indices
}

// ---------------------------------------------------------------------------
// The bidirectional dataflow
// ---------------------------------------------------------------------------

// traversalState marks how far the recursive walk has got with a block.
//
// The distinction is what makes cycles correct rather than merely terminating. A block still
// `active` is an ancestor on the current DFS stack, so its set is not yet final; including it in the
// intersection would let a loop's back edge assert a value is non-null on the strength of an
// assumption that has not been proven. Only `done` neighbours contribute.
type traversalState uint8

const (
	traversalActive traversalState = iota + 1
	traversalDone
)

// propagationDirection selects which neighbours a pass reads from.
//
// Forward reads PREDECESSORS: a value is non-null here if it was non-null on every path into this
// block. Backward reads SUCCESSORS: if every path out of this block dereferences a value, that
// dereference must have been safe here too, which is what lets an access hoist ABOVE the branch
// that uses it. The backward direction is the half that makes this analysis worth its cost.
type propagationDirection uint8

const (
	propagateForward propagationDirection = iota
	propagateBackward
)

// hoistableIterationCap is upstream's fixed-point bound, reproduced exactly.
//
// React raises `CompilerError.invariant(i++ < 100)` on exceeding it. See this file's header for why
// the cap exists at all despite the outer loop looking monotone, and for the measured distribution
// showing real code converges in at most 3.
const hoistableIterationCap = 100

// HoistableAnalysis is what this pass produces: the non-null accesses at every block, and whether
// the analysis actually converged.
//
// `Converged` is the field that does not exist upstream in a readable form, and it is the point of
// this type. React throws when the loop hits the cap; oxc silently exits and returns a truncated
// answer indistinguishable from a real one. A linter can do neither, so it reports.
type HoistableAnalysis struct {
	blocks    map[static_single_assignment.BlockId]map[int]bool
	registry  *pathRegistry
	converged bool
	// iterations is how many outer passes ran, for measuring how close real code comes to the cap.
	iterations int
}

// Converged reports whether the fixed point was reached inside the iteration cap.
//
// A false result means the answer is PARTIAL: correct as far as it goes, because every pass only
// adds proven-non-null accesses, but potentially missing some. A caller that needs upstream's
// behaviour treats false as an error; a caller that prefers a shallower dependency treats it as a
// smaller hoistable set, which is the safe direction.
func (h *HoistableAnalysis) Converged() bool {
	return h != nil && h.converged
}

// Iterations reports how many outer passes ran, for measurement.
func (h *HoistableAnalysis) Iterations() int {
	if h == nil {
		return 0
	}
	return h.iterations
}

// hoistableAt returns the non-null access paths proven at one block.
func (h *HoistableAnalysis) hoistableAt(block static_single_assignment.BlockId) []ReactiveScopeDependency {
	if h == nil || h.blocks == nil {
		return nil
	}
	set, ok := h.blocks[block]
	if !ok {
		return nil
	}
	paths := make([]ReactiveScopeDependency, 0, len(set))
	for _, index := range sortedNodeIndices(set) {
		paths = append(paths, h.registry.nodes[index].fullPath)
	}
	return paths
}

// collectNonNullsInBlocks seeds each block with the accesses it proves non-null by itself.
//
// Upstream's `collectNonNullsInBlocks`. Sources of seed:
//
//	the first parameter of a component, which React assumes non-null outright
//	any instruction that dereferences a value, which proves the OBJECT was non-null
//	non-optional written memo prefixes, when preservation guarantees are enabled
//
// The second is the load-bearing one and it is a deduction rather than an assumption: if
// `props.a.b` executed without throwing, then `props.a` was not nullish at that point, so `props.a`
// is hoistable to anywhere that dominates it.
func collectNonNullsInBlocks(function *Function, temporaries temporaries, ranges *mutation_aliasing.MutableRanges,
	identity ScopeIdentity, scopes *ReactiveScopes, registry *pathRegistry,
	hoistableFromOptionals map[static_single_assignment.BlockId]ReactiveScopeDependency) map[static_single_assignment.BlockId]map[int]bool {
	known := map[int]bool{}
	invoked := CollectAssumedInvokedFunctions(function)

	// # DIVERGENCE FROM React: the component test is a NAME here, not an inferred function type
	//
	// Upstream gates this on `func.fn_type == ReactFunctionType::Component`, a value produced by an
	// inference pass this tree does not have. `IsLikelyComponentName` is the same test React's own
	// heuristics use elsewhere -- an initial capital -- and it is applied to the same subject.
	//
	// The exposure was measured rather than assumed: of 677 corpus functions, 262 are
	// capital-initial and 254 of those take a parameter, so this seeds at most 254 roots. Getting it
	// WRONG in the permissive direction would be the dangerous one, since it asserts a parameter is
	// non-null without proof -- so the gate additionally requires the function to be a plausible
	// component by name AND to have a parameter, matching upstream's `!func.params.is_empty()`.
	//
	// A mutation dropping the NAME half of this gate -- seeding every function that takes a
	// parameter -- SURVIVES, and it is reachable: 152 of 278 corpus functions take a parameter
	// without being capital-initial. It makes no difference to the output because the seed only
	// asserts the first parameter non-null, and every corpus dependency that would benefit already
	// proves its root non-null through an actual dereference instead.
	//
	// The gate is kept because it is upstream's and because its direction matters: seeding a
	// non-component asserts a parameter is non-null WITHOUT proof, which is the over-approximating
	// direction. The verdict EXPIRES if the seed ever becomes load-bearing, which it would the
	// moment a component reads a parameter property that nothing else dereferences.
	if isLikelyComponentFunction(function) {
		known[registry.identifierNode(function.Params[0].Identifier, true)] = true
	}

	// The loads that exist only to build a memo call's dependency array, which must not be read as
	// evidence. See `dependencyArrayInstructions`.
	written := dependencyArrayInstructions(function)

	blocks := map[static_single_assignment.BlockId]map[int]bool{}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		assumed := map[int]bool{}
		for index := range known {
			assumed[index] = true
		}

		// A base the optional-chain traversal proved reachable at this block. Upstream seeds the
		// same set at `CollectHoistablePropertyLoads.ts:416`: for `a?.b.c`, the `.c` rides an outer
		// chain that already tested `a`, so `a?.b` is non-null wherever that block runs and the
		// further loads are safe to read unconditionally.
		//
		// Without it the analysis truncates exactly the paths those fixtures turn on, which is
		// measurable: wiring the optional lowering and collector WITHOUT this seed takes goldens
		// from 28 to 21, and every one of the seven lost is an optional-chain fixture.
		if base, ok := hoistableFromOptionals[block.Id]; ok {
			node := registry.identifierNode(base.Identifier, base.Reactive)
			for _, entry := range base.Path {
				node = registry.propertyNode(node, entry)
			}
			assumed[node] = true
		}

		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			if expression, isFunction := instruction.Value.(*FunctionExpression); isFunction {
				// A callback that is assumed to run contributes what IT proves non-null, because
				// the enclosing scope reaches those reads through the call. Upstream does the same
				// at `CollectHoistablePropertyLoads.ts:432`, merging the inner analysis's
				// `assumedNonNullObjects` into the outer block's.
				//
				// Gated on `CollectAssumedInvokedFunctions` rather than applied to every nested
				// function: a read inside a callback that may never run is not safe to hoist out of
				// it, and hoisting it would move a load to somewhere it can throw.
				for _, path := range invokedNonNullPaths(function, expression, invoked,
					ranges, identity, scopes, instruction.Order, hoistableFromOptionals) {
					assumed[registry.pathIndex(path)] = true
				}
				continue
			}
			if written[instructionId] {
				continue
			}
			if marker, ok := instruction.Value.(*StartMemoize); ok && preserveExistingMemoizationEnabled(function) {
				for _, dependency := range marker.Deps {
					if dependency.Root.IsGlobal || !isImmutableAtInstruction(function,
						dependency.Root.Place.Identifier, instruction.Order, ranges, identity, scopes) {
						continue
					}
					for index, entry := range dependency.Path {
						if entry.Optional {
							break
						}
						path := ReactiveScopeDependency{
							Identifier: dependency.Root.Place.Identifier,
							Reactive:   dependency.Root.Place.Reactive,
							Path:       dependency.Path[:index],
						}
						assumed[registry.pathIndex(path)] = true
					}
				}
			}
			path, ok := maybeNonNullInInstruction(instruction.Value, temporaries)
			if !ok {
				continue
			}
			if !isImmutableAtInstruction(function, path.Identifier, instruction.Order,
				ranges, identity, scopes) {
				continue
			}
			assumed[registry.pathIndex(path)] = true
		}
		blocks[block.Id] = assumed
	}
	return blocks
}

// isKnownImmutableParameter reports whether a value is a parameter of a component or a hook.
//
// Upstream's `knownImmutableIdentifiers`, and its comment states why the escape exists rather than
// leaving it to the range analysis: "due to current limitations of mutable range inference, there
// are edge cases in which we infer known-immutable values (e.g. props or hook params) to have a
// mutable range and scope".
//
// Measured on `useMemo-infer-more-specific.ts`, which is the shape it exists for. `useHook(x)` reads
// `x.y.z` inside a callback that `useMemo` invokes, and the invocation puts `x` inside a scope range
// spanning the function expression -- so the range analysis calls a hook parameter mutable at the
// call site, the descent into the callback is refused, and the dependency truncates to bare `x`
// where upstream infers `x.y.z`. The identical `useCallback` fixture has no call and does not hit
// this.
//
// The gate is the same one upstream applies: `fn.fnType === 'Component' || fn.fnType === 'Hook'`,
// approximated here by name as `isLikelyComponentFunction` already does -- see the divergence note
// at `collectNonNullsInBlocks` for why a name rather than an inferred function type.
func isKnownImmutableParameter(function *Function, id static_single_assignment.IdentifierId) bool {
	if function == nil || len(function.Params) == 0 {
		return false
	}
	if !isLikelyComponentFunction(function) && !isHookName(function.Name) {
		return false
	}
	for _, parameter := range function.Params {
		if parameter.Identifier == id {
			return true
		}
	}
	return false
}

// isLikelyComponentFunction reports whether this function is one React would seed as a component.
func isLikelyComponentFunction(function *Function) bool {
	if function == nil || len(function.Params) == 0 || function.Name == "" {
		return false
	}
	runes := []rune(function.Name)
	return len(runes) > 0 && runes[0] >= 'A' && runes[0] <= 'Z'
}

// maybeNonNullInInstruction returns the access an instruction proves non-null, if any.
//
// Upstream's `getMaybeNonNullInInstruction`. Exactly three instruction kinds dereference a value:
// a property load, a destructure, and a computed load. Measured over the corpus, that is 4,766,
// 328 and 198 instructions respectively -- so all three arms are live and none is dead weight.
//
// The asymmetry between the arms is upstream's. A `PropertyLoad` proves its object non-null whether
// or not the object is a known temporary, so it falls back to the bare object; the other two only
// contribute when the value resolves through the temporaries map, because their subject is the
// value being destructured rather than a syntactic object.
func maybeNonNullInInstruction(value InstructionValue,
	temporaries temporaries) (ReactiveScopeDependency, bool) {
	switch shape := value.(type) {
	case *PropertyLoad:
		if shape.Optional {
			// An OPTIONAL load proves nothing about its object. `arg?.items` is the guard against
			// `arg` being null, so reading it is not evidence that `arg` is non-null -- it is
			// evidence the program expects it might be.
			//
			// Recording it anyway makes the dependency tree's cursor non-null at that step, which
			// takes the `hoistableCursor.nonNull` arm in `addDependency` and flattens `arg?.items`
			// to `arg.items`. The comparison then reports an optionality mismatch against a source
			// that wrote `arg?.items`, which is a disagreement this pass manufactured.
			//
			// Upstream reaches the same answer differently. Its hoistable set is keyed by optional
			// BLOCK -- `collectOptionalChainSidemap` builds `optionalBlock -> baseObject?.a`
			// (`CollectHoistablePropertyLoads.ts:94`) -- so the guard survives into the tree and the
			// optional arm fires. That 418-line pass is driven by `Optional` TERMINALS, which this
			// lowering does not produce: measured on the fixture, the only terminals are `Return`.
			// Optionality lives on the instruction here, so the same fact is read from the flag.
			return ReactiveScopeDependency{}, false
		}
		if resolved, ok := temporaries[shape.Object.Identifier]; ok {
			return resolved, true
		}
		return ReactiveScopeDependency{
			Identifier: shape.Object.Identifier,
			Reactive:   shape.Object.Reactive,
		}, true
	case *Destructure:
		resolved, ok := temporaries[shape.Value.Identifier]
		return resolved, ok
	case *ComputedLoad:
		resolved, ok := temporaries[shape.Object.Identifier]
		return resolved, ok
	}
	return ReactiveScopeDependency{}, false
}

// isImmutableAtInstruction reports whether a value is settled by the time this instruction runs.
//
// Upstream's `isImmutableAtInstr`. A value still being mutated inside its own scope cannot have its
// property loads hoisted, because the property being read may not have been written yet.
//
// # Which of the two ranges this reads, which is the question `scopes.go` flagged
//
// Upstream compares `identifier.mutableRange` against `identifier.scope.range` -- and after the
// scope pass those are DIFFERENT tables. `mutableRange` has been rewritten to the class hull, which
// is `MemberRanges()` here, while `scope.range` is the FINAL range after alignment and merging,
// which is `identity.RangeOf`. Reading the pre-entanglement range for the first would understate the
// mutable window and wrongly declare a value settled, which is the dangerous direction.
//
// The `end > start + 1` test is upstream's and is not a tidy-up of `end > start`: a range spanning a
// single instruction describes a value written once and never mutated, which is immutable for this
// purpose despite having a non-empty range.
func isImmutableAtInstruction(function *Function, id static_single_assignment.IdentifierId, order static_single_assignment.EvaluationOrder,
	ranges *mutation_aliasing.MutableRanges, identity ScopeIdentity, scopes *ReactiveScopes) bool {
	if isKnownImmutableParameter(function, id) {
		return true
	}
	memberRange := ranges.Get(id)
	if memberRange.End <= memberRange.Start+1 {
		return true
	}
	scope := scopes.ScopeOf(id)
	if scope == 0 {
		return true
	}
	scopeRange := identity.RangeOf(scope)
	mutableHere := order >= scopeRange.Start && order < scopeRange.End
	return !mutableHere
}

// propagateNonNull spreads proven-non-null accesses across the graph in both directions.
//
// Upstream's `propagateNonNull`. Each outer iteration is one forward pass followed by one backward
// pass, and the loop stops when neither changed anything. See this file's header for the termination
// story, which is a BOUND rather than a proof, and for the measured convergence distribution.
//
// The successor map is built here rather than read off the terminals deliberately: it is the
// TRANSPOSE of the predecessor lists, so it agrees with them by construction. Deriving it from
// `EachSuccessor` instead would let the two disagree wherever a fallthrough is involved -- and
// `terminal.go` is explicit that a fallthrough is not an edge, which is exactly the case a hand-built
// successor map gets wrong.
func propagateNonNull(function *Function, blocks map[static_single_assignment.BlockId]map[int]bool,
	registry *pathRegistry) (map[static_single_assignment.BlockId]map[int]bool, bool, int) {
	successors := map[static_single_assignment.BlockId][]static_single_assignment.BlockId{}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, predecessor := range block.Predecessors {
			successors[predecessor] = append(successors[predecessor], block.Id)
		}
	}
	for predecessor := range successors {
		sort.Slice(successors[predecessor], func(i, j int) bool {
			return successors[predecessor][i] < successors[predecessor][j]
		})
	}

	// Forward order is the slice, which is reverse postorder and therefore already the order a
	// forward dataflow wants. Backward is its reverse. Both are fixed before the loop begins.
	forward := make([]static_single_assignment.BlockId, 0, len(function.Blocks))
	for _, block := range function.Blocks {
		if block != nil {
			forward = append(forward, block.Id)
		}
	}
	backward := make([]static_single_assignment.BlockId, len(forward))
	for index, id := range forward {
		backward[len(forward)-1-index] = id
	}

	working := map[static_single_assignment.BlockId]map[int]bool{}
	for id, set := range blocks {
		copied := make(map[int]bool, len(set))
		for index := range set {
			copied[index] = true
		}
		working[id] = copied
	}

	iterations := 0
	for iterations < hoistableIterationCap {
		iterations++
		changed := false

		state := map[static_single_assignment.BlockId]traversalState{}
		for _, id := range forward {
			if recursivelyPropagateNonNull(id, propagateForward, state, working,
				function, successors, registry) {
				changed = true
			}
		}

		state = map[static_single_assignment.BlockId]traversalState{}
		for _, id := range backward {
			if recursivelyPropagateNonNull(id, propagateBackward, state, working,
				function, successors, registry) {
				changed = true
			}
		}

		if !changed {
			return working, true, iterations
		}
	}
	// The cap was reached without converging. React raises here; this reports, and the caller sees
	// `Converged() == false`. See this file's header on why the difference from oxc matters: oxc
	// exits silently and the truncated answer is indistinguishable from a real one.
	//
	// A mutation reporting `true` here SURVIVES, and unlike the other survivors in this file that is
	// plain unreachability rather than an equivalence: the cap is reached zero times across the
	// corpus, whose measured maximum is 3 iterations against a bound of 100. No fixture can reach
	// this line without constructing a graph that fails to converge, which is a shape nobody has
	// produced. Recorded rather than covered, and the verdict EXPIRES if the convergence
	// distribution ever moves -- which `TestHoistableCorpusDistribution` asserts on exactly.
	return working, false, iterations
}

// recursivelyPropagateNonNull computes one block's set from its neighbours, depth first.
//
// The `active`/`done` distinction is what makes cycles correct rather than merely terminating. A
// neighbour still on the DFS stack has a provisional set, and folding it into the intersection would
// let a loop assert a value non-null on the strength of an assumption not yet proven. Only `done`
// neighbours contribute, which lets information flow along the acyclic paths through a loop while
// the back edge contributes nothing.
//
// The intersection, not the union, is what makes the forward direction sound: a value is non-null
// here only if it was non-null on EVERY path that reaches here. An empty neighbour list therefore
// contributes the empty set rather than everything, which is why `doneSets` being empty short
// circuits to no contribution rather than to a universal one.
func recursivelyPropagateNonNull(id static_single_assignment.BlockId, direction propagationDirection,
	state map[static_single_assignment.BlockId]traversalState, working map[static_single_assignment.BlockId]map[int]bool, function *Function,
	successors map[static_single_assignment.BlockId][]static_single_assignment.BlockId, registry *pathRegistry) bool {
	if _, seen := state[id]; seen {
		return false
	}
	state[id] = traversalActive

	var neighbours []static_single_assignment.BlockId
	switch direction {
	case propagateBackward:
		neighbours = successors[id]
	default:
		if block := blockById(function, id); block != nil {
			neighbours = block.Predecessors
		}
	}

	changed := false
	for _, neighbour := range neighbours {
		if _, seen := state[neighbour]; !seen {
			if recursivelyPropagateNonNull(neighbour, direction, state, working,
				function, successors, registry) {
				changed = true
			}
		}
	}

	// # A mutation admitting ACTIVE neighbours here SURVIVES, and it is reachable code
	//
	// The corpus holds 1,832 back edges across 278 functions, so `active` neighbours genuinely
	// occur -- this is not the `Optional` shape of a branch that never runs. The filter nonetheless
	// makes no difference to the answer on this corpus, because a block's provisional set at the
	// moment a cycle is re-entered already equals its seeded set: nothing has been added yet that
	// the intersection could wrongly propagate.
	//
	// Kept regardless, because the property it protects is soundness rather than speed. Admitting an
	// active neighbour lets a loop assert a value non-null on the strength of an assumption derived
	// from itself, which is the over-approximating direction -- the one that invents crashes. A
	// verdict of "makes no difference today" is not a licence to remove a soundness guard, and this
	// one EXPIRES the moment a seeded set is non-empty at a back edge, which any richer seeding
	// (optional chains, interprocedural non-null) would produce.
	var doneSets []map[int]bool
	for _, neighbour := range neighbours {
		if state[neighbour] != traversalDone {
			continue
		}
		if set, ok := working[neighbour]; ok {
			doneSets = append(doneSets, set)
		}
	}

	intersection := map[int]bool{}
	if len(doneSets) > 0 {
		for index := range doneSets[0] {
			inAll := true
			for _, other := range doneSets[1:] {
				if !other[index] {
					inAll = false
					break
				}
			}
			if inAll {
				intersection[index] = true
			}
		}
	}

	previous := working[id]
	merged := make(map[int]bool, len(previous)+len(intersection))
	for index := range previous {
		merged[index] = true
	}
	for index := range intersection {
		merged[index] = true
	}
	reduceOptionalChains(merged, registry)

	working[id] = merged
	state[id] = traversalDone

	// Compared as SETS rather than by size, because `reduceOptionalChains` replaces nodes rather
	// than adding them: a reduction that swaps `a?.b` for `a.b` leaves the count identical while
	// changing the answer, and a size comparison would call that converged.
	//
	// A mutation replacing this with a length comparison SURVIVES, and the measurement says why: the
	// swap-without-size-change case occurs zero times on this corpus, because `reduceOptionalChains`
	// only fires on paths carrying an optional entry and the optional-chain analysis that would
	// produce them is `DependencyGapOptionalChains`, unbuilt. So the two spellings are
	// indistinguishable here for exactly the reason the reduction is currently inert.
	//
	// The set comparison is kept because it is upstream's, and because this verdict EXPIRES the day
	// optional chains are reconstructed -- at which point the size comparison silently reports a
	// changed analysis as converged, which is the failure this whole file is written to avoid.
	if !sameNodeSet(previous, merged) {
		changed = true
	}
	return changed
}

// blockById finds a block by id.
//
// A linear scan rather than an index, because `BlockId` is explicitly NOT an index into `Blocks` --
// the slice is held in reverse postorder and the two move independently. See the doc on `BlockId`.
func blockById(function *Function, id static_single_assignment.BlockId) *BasicBlock {
	for _, block := range function.Blocks {
		if block != nil && block.Id == id {
			return block
		}
	}
	return nil
}

// sameNodeSet reports whether two node sets hold exactly the same members.
func sameNodeSet(a, b map[int]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if !b[index] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// The pass, and the seam into dependency collection
// ---------------------------------------------------------------------------

// CollectHoistablePropertyLoads computes which accesses each reactive scope may read early.
//
// Requires a graph whose scopes are built into `Scope` terminals: call `BuildReactiveScopeTerminals`
// first, exactly as `CollectScopeDependencies` does. The result is keyed by SCOPE, taking each
// scope's set from the block its terminal names as the body -- upstream's `keyByScopeId`.
func CollectHoistablePropertyLoads(function *Function, scopes *ReactiveScopes,
	identity ScopeIdentity, ranges *mutation_aliasing.MutableRanges,
	hoistableFromOptionals map[static_single_assignment.BlockId]ReactiveScopeDependency,
) map[ScopeId][]ReactiveScopeDependency {
	analysis := analyseHoistableLoads(function, scopes, identity, ranges, hoistableFromOptionals)
	if analysis == nil {
		return nil
	}
	// Which blocks run on every path through the function. A scope beginning inside a branch is
	// seeded from the entry instead of from its own block; see below.
	always := blocksAlwaysReached(function)
	entry := static_single_assignment.BlockId(0)
	if len(function.Blocks) > 0 && function.Blocks[0] != nil {
		entry = function.Blocks[0].Id
	}
	dominance := computeDominance(function)

	result := map[ScopeId][]ReactiveScopeDependency{}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		scope, ok := block.Terminal.(*Scope)
		if !ok {
			continue
		}
		// Upstream reads the block the scope begins in, and so does this: `keyByScopeId` sets
		// `source.get(block.terminal.block)`. The keying is faithful and is not what differs.
		//
		// What differs is where a scope BEGINS. A scope whose first block sits inside a branch can be
		// handed that branch's facts, and a fact true only under an `if` then licenses the dependency
		// walk to descend past a property that may never have been loaded.
		//
		// Measured on `error.useMemo-infer-less-specific-conditional-access` with the memo callback
		// inlined: `propB.x` is hoistable in the conditional arm alone, the scope for the object
		// returned from that arm begins there, and the seed carried `propB.x` -- so the walk
		// inferred `propB.x.y` where upstream infers bare `propB` and reports the disagreement the
		// fixture exists to test. Upstream's own scope for that value begins after the branch
		// closes, where the fact is not available.
		//
		// The fallback is restricted to a branch scope DOMINATED BY an earlier reactive scope. That
		// is the structural residue of the wider scope upstream keeps around the allocation and
		// mutation in the fixture above. A top-level branch scope is different: in
		// `useMemo-conditional-access-own-scope`, upstream starts the object scope inside the branch
		// and its guard is exactly `propB.x.y`, so discarding the branch facts truncates a dependency
		// upstream keeps. The distinction is not merely whether the source block is conditional; it
		// is whether another reactive region already encloses the path to it.
		source := scope.Block
		if !always[source] && dominatedByOtherReactiveScope(function, dominance, identity,
			block.Id, scope.Scope) {
			source = entry
		}
		result[scope.Scope] = analysis.hoistableAt(source)
	}
	return result
}

func dominatedByOtherReactiveScope(function *Function, dominance *static_single_assignment.Dominance,
	identity ScopeIdentity, targetBlock static_single_assignment.BlockId, targetScope ScopeId) bool {
	if function == nil || dominance == nil || identity == nil {
		return false
	}
	targetGroup := identity.GroupOf(targetScope)
	for _, block := range function.Blocks {
		if block == nil || block.Id == targetBlock {
			continue
		}
		candidate, ok := block.Terminal.(*Scope)
		if !ok || identity.GroupOf(candidate.Scope) == targetGroup {
			continue
		}
		if dominance.Dominates(block.Id, targetBlock) {
			return true
		}
	}
	return false
}

// analyseHoistableLoads runs the whole analysis, returning the per-block answer.
func analyseHoistableLoads(function *Function, scopes *ReactiveScopes,
	identity ScopeIdentity, ranges *mutation_aliasing.MutableRanges,
	hoistableFromOptionals map[static_single_assignment.BlockId]ReactiveScopeDependency) *HoistableAnalysis {
	if function == nil || scopes == nil || identity == nil || ranges == nil {
		return nil
	}
	registry := newPathRegistry()
	terminals := scopeBlockTraversal(function)
	usedOutside := findTemporariesUsedOutsideDeclaringScope(function, terminals)
	temporaries := collectTemporaries(function, usedOutside)

	seeded := collectNonNullsInBlocks(function, temporaries, ranges, identity, scopes, registry,
		hoistableFromOptionals)
	working, converged, iterations := propagateNonNull(function, seeded, registry)
	return &HoistableAnalysis{
		blocks:     working,
		registry:   registry,
		converged:  converged,
		iterations: iterations,
	}
}

// hoistableTreeFor converts a scope's proven-non-null paths into the tree `addDependency` reads.
//
// This is the seam `dependencies.go` left open. Its `hoistableNode` tree is the shape
// `ReactiveScopeDependencyTreeHIR` wants; this builds one from the flat path list above.
//
// # DIVERGENCE FROM oxc, now observable, resolved in React's favour
//
// `dependencies.go` recorded that oxc SORTS its hoistable objects so optional-first entries precede
// non-optional ones, claiming it "matches the TS behavior" -- and that React does no such sort
// (bundle 46967), iterating in insertion order and raising `CompilerError.invariant('Conflicting
// access types')` when two entries disagree about a property, where oxc silently takes first-wins.
//
// That divergence was unobservable while the hoistable set was empty. This pass fills it, so it is
// live now, and React wins: there is no sort, insertion order is the paths' sorted-by-node-index
// order from `hoistableAt`, and a conflict is REPORTED rather than silently resolved. A linter must
// not raise, so `conflicts` is returned for a caller that wants upstream's strictness; the value
// stored on a conflict is the first, which is oxc's behaviour, chosen only because something must be
// stored and the caller has been told.
//
// Measured over 400 corpus files: 0 conflicts across every scope, so the two behaviours are
// indistinguishable on this corpus and the divergence is recorded rather than demonstrated.
// `TestHoistableConflictsAreReportedNotSwallowed` pins that zero.
func hoistableTreeFor(paths []ReactiveScopeDependency) (map[static_single_assignment.IdentifierId]*hoistableNode, int) {
	tree := map[static_single_assignment.IdentifierId]*hoistableNode{}
	conflicts := 0
	for _, path := range paths {
		root, ok := tree[path.Identifier]
		if !ok {
			root = &hoistableNode{properties: map[string]*hoistableNode{}}
			root.nonNull = !(len(path.Path) > 0 && path.Path[0].Optional)
			tree[path.Identifier] = root
		}
		current := root
		for index := range path.Path {
			nonNull := !(index+1 < len(path.Path) && path.Path[index+1].Optional)
			child, present := current.properties[path.Path[index].Property]
			if !present {
				child = &hoistableNode{properties: map[string]*hoistableNode{}, nonNull: nonNull}
				current.properties[path.Path[index].Property] = child
			} else if child.nonNull != nonNull {
				// Upstream's `Conflicting access types` invariant. Reported rather than raised.
				conflicts++
			}
			current = child
		}
	}
	return tree, conflicts
}

// invokedNonNullPaths returns what an invoked callback proves non-null, translated to the caller.
//
// The inner function names its captures in its own identifier space, so a path rooted at one of them
// is meaningless outside. `Captures[i]` and `nested.Context[i]` name the same binding from the two
// sides -- `lower.go:319`, measured across the corpus with zero mismatches -- so the translation is
// a zip. A path rooted anywhere else is inner-local and is dropped rather than guessed at.
func invokedNonNullPaths(parent *Function, expression *FunctionExpression,
	invoked AssumedInvokedFunctions, ranges *mutation_aliasing.MutableRanges,
	identity ScopeIdentity, scopes *ReactiveScopes, order static_single_assignment.EvaluationOrder,
	hoistableFromOptionals map[static_single_assignment.BlockId]ReactiveScopeDependency) []ReactiveScopeDependency {
	if !invoked[expression.Function] || int(expression.Function) >= len(parent.Functions) {
		return nil
	}
	nested := parent.Functions[expression.Function]
	if nested == nil || len(expression.Captures) != len(nested.Context) {
		return nil
	}
	// The captures that are immutable AT THE CALL SITE, computed once. Upstream's
	// `nestedFnImmutableContext`, and its comment gives the reason it is a set rather than a test
	// repeated inside: "comparing instruction ids across inner-outer function bodies is not valid,
	// as they are numbered [separately]". So the question is asked once, here, in the outer
	// numbering, and inside the callback membership in this set IS the answer.
	//
	// Asked of the capture rather than of each path root, which is the same subject upstream uses:
	// `innerFn.func.context.filter(place => isImmutableAtInstr(place.identifier, instr.id, ...))`.
	translate := map[static_single_assignment.IdentifierId]Place{}
	for index := range expression.Captures {
		capture := expression.Captures[index]
		if !isImmutableAtInstruction(parent, capture.Identifier, order, ranges, identity, scopes) {
			continue
		}
		translate[nested.Context[index].Identifier] = capture
	}
	if len(translate) == 0 {
		return nil
	}

	// # What the callback proves is read at its ENTRY block, not gathered from all of them
	//
	// Upstream runs the whole analysis on the callback and takes one block's answer:
	//
	//	const innerHoistables = assertNonNull(
	//	  innerHoistableMap.get(innerFn.func.body.entry));
	//
	// `CollectHoistablePropertyLoads.ts:449`. The entry's set is what survived propagation, and
	// propagation intersects across neighbours, so it holds exactly the accesses that happen on
	// every path through the callback. A read inside a branch reaches that branch's set and no
	// further.
	//
	// Gathering every block instead asserts a conditionally-read object is non-null in the caller,
	// which is the over-approximating direction: it licenses the dependency walk to descend past a
	// property that may never have been loaded. That is the whole difference between
	// `useMemo-conditional-access-noAlloc.ts`, whose callback reads `propB?.x.y` unconditionally and
	// where upstream's guard is `$[1] !== t1`, and
	// `useMemo-infer-less-specific-conditional-access.ts`, whose callback reads it only under an
	// `if` and where upstream infers bare `propB`. The two lower to nearly identical outer
	// functions, so the caller cannot tell them apart -- only the callback's own control flow can.
	// The same sidemap is passed through, matching upstream's `collectHoistablePropertyLoadsInInnerFn`
	// (`CollectHoistablePropertyLoads.ts:131`), which hands the inner function the caller's
	// `hoistableFromOptionals` unchanged. A nested function numbers its blocks from its own space, so
	// an outer entry simply does not match and the map is inert there rather than wrong.
	nestedAnalysis := analyseHoistableLoads(nested, scopes, identity, ranges, hoistableFromOptionals)
	if nestedAnalysis == nil || len(nested.Blocks) == 0 || nested.Blocks[0] == nil {
		return nil
	}

	var paths []ReactiveScopeDependency
	for _, path := range nestedAnalysis.hoistableAt(nested.Blocks[0].Id) {
		outer, translatable := translate[path.Identifier]
		if !translatable {
			continue
		}
		paths = append(paths, ReactiveScopeDependency{
			Identifier: outer.Identifier,
			Reactive:   outer.Reactive,
			Path:       path.Path,
		})
	}
	return paths
}

// dependencyArrayInstructions returns the loads that exist only to build a memo dependency array.
//
// # Why raw dependency-array loads are not evidence
//
// `useMemo(() => ..., [propA?.a, propB.x.y])` lowers its dependency array to ordinary loads, so
// `propB.x.y` appears as a `PropertyLoad` chain in the enclosing function. `collectNonNullsInBlocks`
// would read that chain as proof that `propB.x` is non-null, and the dependency walk then descends
// past `x` and infers `propB.x.y` where upstream infers bare `propB`. The developer's own answer
// becomes the evidence for a deeper answer, and the rule reports a disagreement it manufactured.
// Preservation-enabled mode separately seeds immutable non-optional prefixes from StartMemoize,
// matching React's explicit option gate. Raw loads must still be excluded: they ignore both that
// option and the optional-path stopping rule. MANUAL_MEMO_HOISTING.md records the distinction.
//
// # Why upstream never sees these instructions
//
// Its pipeline drops manual memoization at `Pipeline.ts:168`, runs dead-code elimination at line
// 230, and runs the dependency analysis at line 428. `DeadCodeElimination.ts:371` keeps
// `StartMemoize` and `FinishMemoize` -- "we can't DCE without losing the memoization guarantees" --
// while `PropertyLoad`, `ArrayExpression` and `LoadGlobal` fall through to the prunable list. So
// what the developer declared survives in the marker, and the instructions that built the array do
// not. This tree keeps them: `drop_manual_memoization.go` leaves them deliberately, matching
// upstream's pass, and the elimination upstream relies on afterwards is not written here.
//
// Excluding them at the seed rather than deleting them keeps that difference contained. No
// instruction is removed, so numbering, terminals and every block-keyed analysis are untouched.
//
// # Reached from the array, not matched by path
//
// The elements are walked back through the values feeding them, and only a load whose result
// nothing else reads is taken. Matching by declared path instead was measured and is wrong:
// `useMemo-conditional-access-noAlloc.ts` both reads `propB?.x.y` in the callback body and declares
// it, and upstream's compiled guard is `$[1] !== t1` where `t1 = propB?.x.y` -- so the body's own
// load is real evidence that must survive. Only the copy the array made for itself is dead.
func dependencyArrayInstructions(function *Function) map[InstructionId]bool {
	written := map[InstructionId]bool{}
	if function == nil {
		return written
	}

	// Without a memo marker no array here is a dependency array. This is also what keeps the walk
	// off every array literal in the corpus: a mutation removing it does not change the answer but
	// does not finish either, which is the shape of a guard that is doing real work.
	hasMarker := false
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			if _, isStart := instruction.Value.(*StartMemoize); isStart {
				hasMarker = true
			}
		}
	}
	if !hasMarker {
		return written
	}

	producer := map[static_single_assignment.IdentifierId]InstructionId{}
	reads := map[static_single_assignment.IdentifierId]int{}
	countRead := func(place Place, role PlaceRole) {
		if role != PlaceRoleDefine {
			reads[place.Identifier]++
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
			producer[instruction.LValue.Identifier] = instructionId
			EachPlace(instruction.Value, countRead)
		}
		EachTerminalPlace(block.Terminal, countRead)
	}

	var walk func(id static_single_assignment.IdentifierId, depth int)
	walk = func(id static_single_assignment.IdentifierId, depth int) {
		// Bounded by the longest access path. Re-entry cannot occur because an instruction is
		// marked before its operands are walked.
		if depth > 64 {
			return
		}
		instructionId, produced := producer[id]
		if !produced || written[instructionId] {
			return
		}
		// A value read more than once is not the array's private copy: something else consumes it,
		// so upstream's elimination would keep it and so does this.
		//
		// # A mutation removing this guard SURVIVES, and it is recorded rather than covered
		//
		// Measured: dropping it moves nothing on this corpus, because a dependency array's elements
		// are lowered fresh for the array in every fixture here -- the callback body's own read of
		// the same property is a separate instruction inside the nested function, which this walk
		// never reaches. So no load is currently reachable from an array AND read elsewhere.
		//
		// Kept because the property it protects is soundness rather than score, and because the
		// direction matters: without it, a load that real code also consumes would stop being
		// evidence, which is the under-approximating direction and would drop dependency depth that
		// upstream keeps. The verdict EXPIRES the moment a lowering shares one load between an
		// array element and a body read, which is what upstream's own elimination already assumes
		// can happen -- it prunes by use count rather than by position.
		if reads[id] > 1 {
			return
		}
		instruction := function.Instructions[instructionId]
		if instruction == nil {
			return
		}
		switch instruction.Value.(type) {
		case *PropertyLoad, *LoadLocal, *LoadGlobal, *ComputedLoad:
		default:
			// Any other value is a real computation the array merely names, and it is evidence in
			// its own right.
			//
			// A mutation widening this list SURVIVES on this corpus, since no fixture writes a
			// dependency array over anything but a load -- `[props.a, x]` rather than
			// `[f(x)]`, which upstream's own dependency extraction refuses to parse anyway. It is
			// still the list rather than a bare accept, because the walk exists to find loads that
			// upstream's dead-code elimination would delete, and that pass keeps every value with a
			// live use. Accepting a call here would stop a real computation from being evidence.
			return
		}
		written[instructionId] = true
		EachPlace(instruction.Value, func(place Place, role PlaceRole) {
			if role != PlaceRoleDefine {
				walk(place.Identifier, depth+1)
			}
		})
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
			array, isArray := instruction.Value.(*ArrayExpression)
			if !isArray {
				continue
			}
			for _, element := range array.Elements {
				walk(element.Place.Identifier, 0)
			}
		}
	}
	return written
}
