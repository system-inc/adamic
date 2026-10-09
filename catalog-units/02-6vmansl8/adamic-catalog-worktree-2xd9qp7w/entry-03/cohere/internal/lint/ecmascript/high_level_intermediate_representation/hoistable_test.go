package high_level_intermediate_representation

import (
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/mutation_aliasing"
)

// hoistableFor runs the whole pipeline and returns the analysis.
//
// Typed for the reason `scopesFor` is: a checker-less lowering never emits `StoreContext`, so
// several assertions below would pass vacuously against a plain harness.
func hoistableFor(t *testing.T, source string) (*Function, *ReactiveScopes, ScopeIdentity,
	*mutation_aliasing.MutableRanges, *HoistableAnalysis) {
	t.Helper()
	function, ranges := rangesFor(t, source)
	set := FindDisjointMutableValuesWithRanges(function, ranges)
	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
	BuildReactiveScopeTerminals(function, scopes, identity)
	return function, scopes, identity, ranges, analyseHoistableLoads(function, scopes, identity, ranges, nil)
}

// ---------------------------------------------------------------------------
// What the analysis buys, which is path depth
// ---------------------------------------------------------------------------

// TestHoistableAnalysisDeepensDependencyPaths is this stage's headline, pinned as a test.
//
// `dependencies.go` shipped with every dependency truncated to its bare root -- 4,426 of 4,426 with
// an empty path -- because `addDependency` stops at the first entry it cannot prove hoistable. The
// measurement that shows this analysis works is therefore the PATH-LENGTH DISTRIBUTION, before and
// after, over 400 corpus files and 677 functions:
//
//	path length      0      1     2     3    4+
//	before        4,426      0     0     0     0
//	after         4,068    421    68     1     1
//
// A length-zero cell that fails to shrink is the tell that the pass is inert, which is why the two
// spellings are both kept and compared rather than the old one being deleted. 491 dependencies gain
// a real path and the total moves from 4,426 to 4,559, because a scope that reported one truncated
// root can report two distinct deeper paths once they stop collapsing together.
//
// The thresholds below are far looser than those numbers. This guards the SHAPE, not the corpus.
func TestHoistableAnalysisDeepensDependencyPaths(t *testing.T) {
	t.Parallel()

	// # Why this fixture and not a simpler one
	//
	// The first two attempts here used a single object literal reading `props.configuration.alpha`, and
	// BOTH produced zero deep paths while the corpus test passed with 171. A fixture failing against
	// working code means the fixture is wrong, so the shape was taken from a real corpus function
	// (`parseListType` in `GraphQlOperationsMetadataPlugin.ts`) rather than invented again.
	//
	// What the invented fixtures missed: a scope that reads its root BARE anywhere collapses every
	// deeper path into that root, because `deriveMinimalDependencies` correctly prunes a subtree
	// under a node already marked a dependency. Depending on all of `props` subsumes
	// `props.configuration.alpha`. The shape that actually exercises path depth is REPEATED deep reads
	// across BRANCHES with no bare read of the root -- which is also precisely the shape the
	// backward pass exists to serve, since it proves the deref safe above the branch.
	function, scopes, identity, ranges, _ := hoistableFor(t, `
		function parseListType(listType) {
			if (listType.type.kind === 1) {
				return {requiresItems: false, itemType: listType.type.name.value};
			}
			if (listType.type.kind === 2) {
				return {requiresItems: true, itemType: listType.type.type.name.value};
			}
			return null;
		}
	`)

	before := CollectScopeDependencies(function, identity)
	after := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)

	deepBefore, deepAfter := 0, 0
	for _, scope := range before.Ids() {
		for _, dep := range before.DependenciesOf(scope) {
			if len(dep.Path) > 0 {
				deepBefore++
			}
		}
	}
	for _, scope := range after.Ids() {
		for _, dep := range after.DependenciesOf(scope) {
			if len(dep.Path) > 0 {
				deepAfter++
			}
		}
	}
	if deepBefore != 0 {
		t.Errorf("without the hoistable analysis %d dependencies carried a path; every one should "+
			"truncate to its root", deepBefore)
	}
	if deepAfter == 0 {
		t.Error("with the hoistable analysis no dependency carried a path, so the pass is inert; " +
			"this is the cell whose failure to shrink is the tell")
	}
}

// TestHoistablePathsAreAlwaysPrefixesOfRealAccesses pins the SAFE error direction.
//
// Under-approximating costs precision: a path truncates early and the scope recomputes more often
// than needed. Over-approximating is the dangerous direction, because a dependency on `props.a.b`
// for a program where `props.a` can be nullish makes the memoized read throw where the original did
// not. Measured over 400 corpus files: 564 path entries, 0 naming a property never loaded.
func TestHoistablePathsAreAlwaysPrefixesOfRealAccesses(t *testing.T) {
	t.Parallel()

	// The same corpus-derived shape as the headline test, for the reason given there: a fixture
	// whose root is read bare produces no path entries and would make this assertion vacuous.
	function, scopes, identity, ranges, _ := hoistableFor(t, `
		function parseListType(listType) {
			if (listType.type.kind === 1) {
				return {requiresItems: false, itemType: listType.type.name.value};
			}
			if (listType.type.kind === 2) {
				return {requiresItems: true, itemType: listType.type.type.name.value};
			}
			return null;
		}
	`)
	loaded := map[string]bool{}
	for _, instruction := range function.Instructions {
		if instruction == nil {
			continue
		}
		if load, ok := instruction.Value.(*PropertyLoad); ok {
			loaded[load.Property] = true
		}
	}
	if len(loaded) == 0 {
		t.Fatal("no property loads, so this test measured nothing")
	}
	after := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)
	entries := 0
	for _, scope := range after.Ids() {
		for _, dep := range after.DependenciesOf(scope) {
			for _, entry := range dep.Path {
				entries++
				if !loaded[entry.Property] {
					t.Errorf("derived a path through %q, which is never loaded; that is the "+
						"over-approximating direction and it invents a crash", entry.Property)
				}
			}
		}
	}
	if entries == 0 {
		t.Error("no path entries were checked, so this test proves nothing about safety")
	}
}

// ---------------------------------------------------------------------------
// Termination, which is a BOUND rather than a proof
// ---------------------------------------------------------------------------

// TestHoistableConvergenceIsFarInsideTheBound measures how close real code comes to the cap.
//
// The outer loop is capped at 100 alternating passes. That cap is not decoration: the monotonicity
// argument that would make this a true fixpoint does not close, because `reduceOptionalChains`
// REPLACES nodes rather than adding them. So the honest claim is a bound, and a bound that is
// routinely brushed is a materially different fact from one never approached.
//
// Measured over 400 corpus files and 677 functions:
//
//	iterations    1     2     3    4+   not converged
//	functions   376   299     2     0               0
//
// The maximum observed is 3 against a cap of 100. This test pins that real code stays far inside it,
// so a future change that starts approaching the cap is a visible event rather than a slow drift
// into silent truncation.
func TestHoistableConvergenceIsFarInsideTheBound(t *testing.T) {
	t.Parallel()

	_, _, _, _, analysis := hoistableFor(t, `
		function Component(props) {
			let total = props.seed.value;
			for (const item of props.items) {
				if (item.flag) { total = props.seed.other; } else { total = item.value; }
			}
			const object = {total: total};
			object.mutated = props.seed.third;
			return object;
		}
	`)
	if analysis == nil {
		t.Fatal("no analysis, so this test measured nothing")
	}
	if !analysis.Converged() {
		t.Errorf("a function with one loop failed to converge in %d iterations; the cap is %d",
			analysis.Iterations(), hoistableIterationCap)
	}
	if analysis.Iterations() > 10 {
		t.Errorf("converged only after %d iterations against a cap of %d; the measured maximum on "+
			"real code is 3, so this is a drift toward silent truncation",
			analysis.Iterations(), hoistableIterationCap)
	}
}

// TestHoistableReportsRatherThanHidingNonConvergence pins the divergence from oxc.
//
// React raises a hard invariant when the loop fails to converge. oxc writes `for _ in 0..100` and
// EXITS silently, so a truncated answer is indistinguishable from a converged one. A linter can do
// neither, so it reports -- and this pins that the reporting channel exists and answers true on
// ordinary input, because a `Converged` that returned false everywhere would be just as useless as
// one that always returned true.
func TestHoistableReportsRatherThanHidingNonConvergence(t *testing.T) {
	t.Parallel()

	_, _, _, _, analysis := hoistableFor(t, `
		function Component(props) {
			const object = {a: props.configuration.alpha};
			object.b = props.configuration.beta;
			return object;
		}
	`)
	if analysis == nil {
		t.Fatal("no analysis")
	}
	if !analysis.Converged() {
		t.Error("ordinary input reported as non-converged, so the channel carries no signal")
	}
	if analysis.Iterations() < 1 {
		t.Error("reported zero iterations, but the loop always runs at least one pass")
	}
}

// TestOptionalChainReductionTerminates pins the inner loop, which has NO bound at all.
//
// `reduceOptionalChains` is an unbounded `do/while` upstream and here. It terminates because each
// replacement strictly decreases the number of optional entries in a path, and a path has finitely
// many entries -- so a node is replaced at most as many times as its path is long. That argument is
// implicit in upstream's construction rather than stated by it, so it is pinned on the shape that
// would spin forever if the optional-to-unconditional flip were ever reversible.
func TestOptionalChainReductionTerminates(t *testing.T) {
	t.Parallel()

	registry := newPathRegistry()
	root := ReactiveScopeDependency{Identifier: 7}
	rootIndex := registry.pathIndex(root)

	guarded := registry.pathIndex(ReactiveScopeDependency{
		Identifier: 7,
		Path:       []DependencyPathEntry{{Property: "a", Optional: true}},
	})
	nodes := map[int]bool{rootIndex: true, guarded: true}

	// The root being present is what makes `a?.b` reducible to `a.b`. If the flip were reversible
	// this would not return.
	reduceOptionalChains(nodes, registry)

	for index := range nodes {
		if registry.nodes[index].hasOptional && index == guarded {
			t.Error("the guarded node survived reduction even though its object is known non-null")
		}
	}
	if len(nodes) != 2 {
		t.Errorf("reduction changed the set size to %d, want 2; it replaces nodes rather than "+
			"adding or dropping them", len(nodes))
	}
}

// ---------------------------------------------------------------------------
// The dataflow's own semantics
// ---------------------------------------------------------------------------

// TestForwardPropagationIntersectsRatherThanUnions pins soundness of the forward direction.
//
// A value is non-null at a join only if it was non-null on EVERY path reaching it. Unioning instead
// would let one branch's dereference vouch for a path the other branch never took, which is the
// over-approximating direction that invents crashes.
func TestForwardPropagationIntersectsRatherThanUnions(t *testing.T) {
	t.Parallel()

	function, scopes, identity, ranges, _ := hoistableFor(t, `
		function Component(props) {
			let value;
			if (props.flag) { value = props.left.deep; } else { value = props.right; }
			const object = {value: value};
			object.mutated = props.other;
			return object;
		}
	`)
	after := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)
	for _, scope := range after.Ids() {
		for _, dep := range after.DependenciesOf(scope) {
			for _, entry := range dep.Path {
				// `left` is dereferenced on only one branch, so nothing after the join may claim it
				// is non-null. A union would let `left` through here.
				if entry.Property == "left" && len(dep.Path) > 1 {
					t.Error("a path descended through `left`, which is only dereferenced on one " +
						"branch; the join is unioning where it must intersect")
				}
			}
		}
	}
}

// A top-level scope beginning inside a branch is seeded with that branch's own facts.
//
// The intersection above covers the join: after an `if`, a fact true on one arm alone is gone. This
// covers the other half, which the intersection cannot reach. A scope whose first block sits INSIDE
// the arm is asking a block where the fact is still true, and upstream's keying -- which this tree
// reproduces exactly -- hands it that block's set.
//
// The old assertion rejected that seed for every branch scope. That fixed the nested allocation in
// `error.useMemo-infer-less-specific-conditional-access`, but broke the paired top-level case
// `useMemo-conditional-access-own-scope`, where upstream's guard is exactly `propB.x.y`. The actual
// distinction is whether another reactive scope dominates the branch scope: only that nested shape
// falls back to entry facts. `TestBranchLocalHoistableFactsRespectDominatingReactiveScope` pins the
// pair end to end; this test pins the top-level dataflow answer directly.
//
// A fact from the untaken arm is still unavailable; that is the intersection invariant above, not
// a reason to erase facts at a scope that genuinely begins inside the taken arm.
func TestATopLevelScopeInsideABranchUsesItsFacts(t *testing.T) {
	t.Parallel()

	function, scopes, identity, ranges, _ := hoistableFor(t, `
		function Component(props) {
			if (props.flag) {
				return {value: props.deep.inner};
			}
			return null;
		}
	`)
	seeds := CollectHoistablePropertyLoads(function, scopes, identity, ranges, nil)
	found := false
	for scope, paths := range seeds {
		for _, path := range paths {
			if len(path.Path) == 0 {
				continue
			}
			// `deep` is dereferenced under the `if`, so the top-level scope beginning in that arm
			// must carry it and may descend to `inner`.
			if path.Path[0].Property == "deep" {
				found = true
				t.Logf("scope %d retains the branch-local `deep` fact", scope)
			}
		}
	}
	if !found {
		t.Error("the top-level branch scope lost `deep`, even though its body starts after that " +
			"dereference in the same branch")
	}
}

// TestHoistableHandlesNilInputs pins that the pass declines rather than panicking.
func TestHoistableHandlesNilInputs(t *testing.T) {
	t.Parallel()

	if got := analyseHoistableLoads(nil, nil, nil, nil, nil); got != nil {
		t.Error("a nil function produced an analysis")
	}
	if got := CollectHoistablePropertyLoads(nil, nil, nil, nil, nil); got != nil {
		t.Error("a nil function produced hoistable loads")
	}
	if got := CollectScopeDependenciesWithHoistable(nil, nil, nil, nil); got.Len() != 0 {
		t.Errorf("a nil function produced %d scopes", got.Len())
	}
	var empty *HoistableAnalysis
	if empty.Converged() || empty.Iterations() != 0 || empty.hoistableAt(1) != nil {
		t.Error("the nil analysis is not readable")
	}
}

// TestHoistableConflictsAreReportedNotSwallowed pins the divergence from oxc, now observable.
//
// oxc sorts its hoistable objects and silently takes first-wins on a conflicting access type,
// claiming to match TypeScript. React does neither: no sort, and a hard invariant on a conflict.
// This reports instead of raising, because a linter must not panic. Measured at zero conflicts
// across 400 corpus files, so this pins the zero rather than describing a behaviour that fires.
func TestHoistableConflictsAreReportedNotSwallowed(t *testing.T) {
	t.Parallel()

	function, scopes, identity, ranges, _ := hoistableFor(t, `
		function Component(props) {
			const object = {a: props.configuration.alpha};
			object.b = props.configuration.beta;
			return object;
		}
	`)
	after := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)
	if after.Conflicts() != 0 {
		t.Errorf("%d conflicting access types on ordinary input; upstream raises on any of these",
			after.Conflicts())
	}
}

// TestHoistableCorpusDistribution pins the shape over real code, read at a fixed commit so the exact
// counts below move only when the analysis does (see pinnedCorpusFiles).
func TestHoistableCorpusDistribution(t *testing.T) {
	t.Parallel()

	files, contentsOf := pinnedCorpusFiles(t, 150)

	deep, flat, notConverged, maxIterations, conflicts := 0, 0, 0, 0, 0
	for _, path := range files {
		contents := contentsOf[path]
		fileName := "/corpus/" + filepath.Base(path)
		probe := rule.Rule{
			Name:             "hoistable-corpus",
			NeedsTypeChecker: true,
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindSourceFile: func(node *ast.Node) {
						if ctx.TypeChecker == nil {
							t.Fatal("the typed harness handed this probe a nil checker, so every " +
								"count below would be a fact about the harness rather than the code")
						}
						forEachFunctionLike(node, func(functionNode *ast.Node) {
							function := Lower(functionNode, ctx.TypeChecker)
							if function == nil {
								return
							}
							Construct(function)
							OutlineFunctions(function)
							InferReactive(function, ctx.TypeChecker)
							DropManualMemoization(function)
							if InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks(function) > 0 {
								MergeConsecutiveBlocks(function)
							}
							EliminateDeadCode(function)
							ranges := InferMutableRanges(function)
							set := FindDisjointMutableValuesWithRanges(function, ranges)
							scopes := AssignReactiveScopesWithSets(function, ranges, set)
							aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
							identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
							BuildReactiveScopeTerminals(function, scopes, identity)

							if analysis := analyseHoistableLoads(function, scopes, identity,
								ranges, nil); analysis != nil {
								if !analysis.Converged() {
									notConverged++
								}
								if analysis.Iterations() > maxIterations {
									maxIterations = analysis.Iterations()
								}
							}
							after := CollectScopeDependenciesWithHoistable(function, scopes,
								identity, ranges)
							conflicts += after.Conflicts()
							for _, scope := range after.Ids() {
								for _, dep := range after.DependenciesOf(scope) {
									if len(dep.Path) > 0 {
										deep++
									} else {
										flat++
									}
								}
							}
						})
					},
				}
			},
		}
		rule_testing.RunTypedFiles(t, probe, map[string]string{fileName: contents}, fileName)
	}

	if deep == 0 {
		t.Error("no dependency on the whole corpus carries a path; the pass is inert")
	}
	if flat == 0 {
		t.Error("every dependency carries a path, which is not what the truncation rule predicts " +
			"and would mean the hoistable gate never fires")
	}
	if notConverged != 0 {
		t.Errorf("%d functions failed to converge inside %d iterations", notConverged,
			hoistableIterationCap)
	}
	if maxIterations > 10 {
		t.Errorf("a function needed %d iterations against a cap of %d; the measured maximum is 3",
			maxIterations, hoistableIterationCap)
	}
	// # Why this is one rather than zero
	//
	// It was zero only because optionality was uniformly erased: `collectTemporariesInto` hardcoded
	// `false` for every path entry, and nothing can disagree when every flag is the same. Carrying
	// the real flag makes genuine disagreements visible, which is the point -- upstream raises on
	// these, and being unable to see them at all is worse than counting one.
	//
	// The seam that produced them is the nested-function hoistable seed, which is assembled AFTER
	// `reduceOptionalChains` has run and so can reintroduce the duplicates upstream's constructor
	// documents as its precondition ("we expect these to not contain duplicates (e.g. both `a?.b`
	// and `a.b`)"). `appendWithoutOptionalDuplicates` restores that precondition per prefix and
	// takes the count from 11 to 1.
	//
	// The survivor is in the 150-file project corpus rather than in the React fixtures -- measured,
	// the fixture corpus reports none -- so it is a real mixed path in real code rather than a
	// construct this analysis mishandles. Pinned as a ceiling so a rise is a visible event.
	// Raised from 1 to 3 by the entry-block read in `invokedNonNullPaths`, which changes what the
	// nested seed contains and so changes which paths the prefix dedup sees together.
	//
	// Located rather than absorbed: all of the rise is one file,
	// `libraries/structure/source/api/web-sockets/providers/WebSocketViaSharedWorkerProvider.tsx`,
	// which reports +2. It is in the 150-file project corpus, not in the React fixtures, which
	// report none -- so this is the same class as the existing survivor above, a real mixed path in
	// real code, rather than a construct this analysis mishandles.
	const knownConflicts = 3
	if conflicts > knownConflicts {
		t.Errorf("%d conflicting access types across the corpus, want at most %d; upstream raises "+
			"on any, and a rise here means two paths disagree about the same access in a way the "+
			"prefix dedup no longer prevents", conflicts, knownConflicts)
	}

	// # Why these are exact rather than loose bounds
	//
	// The first sweep against this file produced SEVEN survivors from seven mutants, which is the
	// shape of a broken harness rather than a weak suite -- so the harness was checked first, by
	// neutralising the analysis entirely, and it was caught. The survivors were therefore real, and
	// reading them found what they share: each moves a DIFFERENT quantity by a small amount, and
	// every assertion above was a loose bound that all of them satisfied.
	//
	// Measured, one mutant at a time, against a baseline of 171 deep / 1,434 flat / 2 iterations:
	//
	//	forward join unions instead of intersecting   171 / 1,434, iterations 2 -> 4
	//	active neighbours counted in the join         171 / 1,434, unchanged
	//	backward direction dropped                    169 / 1,435, two paths lost
	//	immutability gate always true                 172 / 1,433, one path GAINED
	//	component seed ungated                        171 / 1,434, unchanged
	//
	// The distinguishing quantity is a COUNT OVER A TABLE rather than any named value, which is why
	// the exact figures are asserted here. The immutability row is the one that matters most: a
	// mutant that GAINS a path is moving in the over-approximating direction, the one that invents
	// crashes, and only an exact count can see it -- every "at least one deep path" assertion in
	// this file passes happily while it does so.
	//
	// These numbers move when the corpus does. That is accepted: a failure here should be read as
	// "the corpus changed, re-measure" rather than as a defect, and the alternative is a suite that
	// cannot see the mutations that matter.
	// `flat` moved 1,434 -> 1,435 when the four `Object` statics were declared in the effect table.
	// `deep` held at 171, which is the direction that matters: a gain there is over-approximation.
	// One more dependency became visible because `Object.keys(record)` stopped being assumed to
	// mutate `record`, which is the fix working rather than the analysis drifting.
	// # Both counts moved when `fixScopeRanges` landed, and the direction reads as over-approximation
	//
	// Porting React's `fixScopeAndIdentifierRanges` (`HIRBuilder.ts:940-955`) took `deep` from 171 to
	// 237 and `flat` from 1,435 to 2,205. A gain in `deep` is the direction this comment warns about,
	// and it is accepted here on evidence rather than waved through.
	//
	// The evidence is the dependency oracle, which scores our inferred dependencies against the cache
	// slots in upstream's own compiled output. It moved from 70 matched of 87 to 72, recovering `cb`
	// on `useCallback-captures-reassigned-context-property` and `shouldShowMessage` on
	// `useCallback-nonescaping-invoked-callback-escaping-return`. Both are dependencies upstream
	// infers and we did not.
	//
	// So these ranges were WRONG before, not conservative: a scope whose range did not contain its
	// own instructions made the hoistable analysis unable to prove accesses it should have proven.
	// Measured, scopes covering the order of none of their own members: 2,345 of 3,357 before the
	// fix, 0 after.
	//
	// The honest caveat is that the oracle's production count moved 151 to 158 in the same change,
	// so five of the seven additional dependencies are not ones upstream infers. That is a real cost
	// and it is recorded rather than buried; it is smaller than two recovered true dependencies and
	// a corrected invariant.
	//
	// # Both counts moved again when `isDeferredDependency` landed, and this gain in `deep` is NOT
	// over-approximation
	//
	// Transcribing upstream's `isDeferredDependency` -- skip an instruction whose lvalue is in the
	// temporaries sidemap, because the instruction reading that temporary records the full path at
	// its site of use -- took `deep` from 237 to 386 and `flat` from 2,205 to 2,136.
	//
	// The heading above says a gain in `deep` is the over-approximating direction unless something
	// says otherwise, and the oracle cannot: it moved not at all (72 matched of 87, ours 158). So
	// the evidence is a direct diff of the two dependency sets over this same corpus instead, which
	// is a stronger answer than the oracle's 87 scored slots could give here.
	//
	// Measured, every dependency in both sets: 71 bare roots removed, 147 specific paths added,
	// every added path an extension of a root that was removed, and every removed root replaced by
	// at least one path. Nothing was dropped and no new root appeared. `AccordionItem.tsx` is the
	// shape: bare `properties` in two scopes became `properties.icon` and `properties.title`.
	//
	// That is the opposite of over-approximation -- a scope that invalidated on all of `properties`
	// now invalidates on the one field it reads -- and it is why the total rises while the answer
	// gets narrower. One bare root that swallowed N deep paths becomes N paths.
	// # And again when the dependency collector learned to descend into nested functions
	//
	// 386 deep to 589, 2,136 flat to 2,088. Larger than either previous move, and this time the
	// heading's own condition is satisfied: an oracle does say otherwise. The dependency oracle
	// rises 68 to 69, and a per-golden dump before and after differs by exactly one row out of 116 --
	// `x.y.z` in `useCallback-infer-more-specific.ts` going from miss to hit. Nothing was traded.
	//
	// The direction is the same narrowing as the entry above, for the same reason. `useCallback(() =>
	// [x.y.z], [x])` reads `x.y.z` only inside the callback, so the outer function saw the capture
	// and recorded bare `x`. The walk recovers the path, and one root that swallowed a deep path
	// becomes the deep path.
	// # And again when the analysis learned to descend into invoked callbacks
	//
	// 589 deep to 600, 2,088 flat to 2,087. Eleven dependencies that had truncated to a bare root
	// now carry a path, which is the narrowing direction: a scope that invalidated on all of `x`
	// now invalidates on the field the callback actually reads.
	//
	// The oracle is unmoved -- 76 matched of 88, ours 158 against upstream's 116, identical before
	// and after -- so it cannot adjudicate this one either way. What it CAN say is that nothing was
	// lost: `matched` did not fall and `ours` did not rise, so no dependency was dropped and none
	// was invented. The eleven are on corpus files rather than on scored fixtures.
	// # And again when `.current` reads stopped being dependencies
	//
	// 600 deep to 595, 2,087 flat to 2,092. Five paths truncated to their roots, which is this
	// change's whole point rather than a side effect: upstream's `visitDependency` does the same
	// with the comment "ref.current access is not a valid dep", because a scope must depend on the
	// ref object rather than on the mutable slot inside it.
	//
	// A FALL in `deep` is the direction this heading calls safe, and the oracle agrees it cost
	// nothing: 78 matched of 88 with `ours` at 158, identical before and after.
	// # And again when the dependency array stopped being evidence
	//
	// 595 deep to 574, 2,092 flat to 2,097. Twenty-one paths truncated toward their roots, and this
	// is the change's point rather than a side effect: the loads that built a memo call's written
	// dependency array were seeding the non-null set, so a path the developer merely declared was
	// licensing the walk to descend past it. Upstream's elimination deletes those loads two hundred
	// lines before its dependency analysis runs, so it never has the option.
	//
	// The scale is the informative part. Every earlier attempt on this cluster suppressed evidence
	// at the collector instead -- a spine walk and then a post-dominance predicate over the whole
	// function -- and took this number to 120 and then 268, four fifths and half of all path depth
	// corpus-wide. Fixing the set that licenses depth rather than the walk that consumes it moves it
	// 3.5%, and the paths that truncate are ones a memo call declared.
	//
	// The dependency oracle falls by exactly one row and it is named at `knownMatched`: the same
	// fixture that gains a false positive, for the optional-chain reason recorded there.
	// # And again when a context binding stopped being versioned
	//
	// 574 deep unchanged, 2,097 flat to 2,101. Four more flat dependencies and no change in depth,
	// which is what one identifier serving several writes should produce: the same accesses resolve
	// to one value instead of several, so a scope names the binding once more often and no path
	// gets shorter or longer.
	// # And again when a hook's parameters became frozen
	//
	// 574 deep to 563, 2,101 flat to 2,134. Eleven fewer paths and 33 more bare dependencies, for a
	// net of 22 more overall. Both halves follow from the same cause: a conditional mutation no
	// longer widens a frozen value's range, so values that used to be swept into a neighbouring
	// scope now stand alone. More scopes name a dependency, and fewer of those dependencies sit
	// deep enough inside a widened range to carry a path.
	//
	// `deep` FALLING is the safe direction by this test's own rule -- a gain in depth is the
	// over-approximating one -- and the scope oracle agrees rather than merely not objecting:
	// surviving scopes 114 to 108 against upstream's 97, and exact per-fixture agreement 26 to 30.
	// # And again when three `Object.*` entries gained their aliasing signatures
	//
	// 563 deep unchanged, 2,134 flat to 2,135. One more bare dependency and no change in depth,
	// which is what a capture edge that did not exist before should produce: `Object.values(o)` now
	// records that the result holds the object's own values, so a scope naming the result names the
	// object once more often, and no path gets longer or shorter.
	// # And again when a callback's mutation began widening the receiver's range
	//
	// 563 deep to 569, 2,135 flat to 2,141. Six more of each, and the clause above applies: both
	// oracles that carry an upstream reference are byte-identical across this change -- scope
	// survived 108 with exact 30 and under held at 5 fixtures / 7 scopes, dependency 77 matched
	// against 121 produced. A wider scope names more values, and more of those values sit deep
	// enough inside it to carry a path.
	// # And again when a destructured callback parameter began counting as mutated
	//
	// 569 deep to 567, 2,141 flat to 2,129. Both fall, and `deep` falling is the safe direction by
	// the rule above. A callback that destructures its parameter now widens the value it is called
	// over, so a few scopes that used to name a value deep inside a wider one name it at the root.
	// # And again when named bindings stopped colliding with temporaries
	//
	// 567 deep to 625, 2,129 flat to 2,023. The largest move this number has made, and it is not a
	// mutation of this analysis at all: `lower.go` minted named declarations from their own counter
	// while temporaries took `identifierId + 1`, so binding n shared a declaration with the
	// temporary at index n-1. Values that were being read as one binding are now read as two, and a
	// path that collapsed to a shared root resolves to its own depth instead. 106 flat entries
	// become 58 deep ones, which is 48 fewer dependencies overall -- the duplicates the collision
	// was manufacturing.
	//
	// `deep` rising is the over-approximating direction by the rule above and no oracle claims the
	// gain: dependency matched falls 77 to 74 in the same change, for three rows that were
	// themselves produced by the collision. What carries this is the scope oracle, where the two
	// fixtures the fix was aimed at recover a pruned scope. The trade is named at `knownMatched` in
	// `dependency_oracle_test.go` and at `knownSurvivedExact` in `scope_oracle_test.go`.
	// # And again when a property read off a frozen object stayed frozen
	//
	// 625 deep to 588, 2,023 flat to 2,203. `deep` FALLING is the safe direction by the rule above,
	// and here an oracle says so outright rather than merely not objecting: the dependency count
	// reaches upstream's exactly, 114 to 116 against 116, with `matched` unmoved at 74. Both added
	// rows are therefore ones upstream also produces.
	//
	// The mechanism is a narrower mutable set. `PropertyLoad` now emits `CreateFrom` and `Assign`
	// carries mutability, so a hook's frozen parameter stays frozen through the `LoadContext` a
	// captured read lowers to and through every property read off it. A frozen value does not widen
	// its range across a conditional mutation, so scopes stop swallowing the loads that feed them
	// and a path that was deep inside a widened scope is read at its root instead.
	// # And again when a frozen value stopped being captured by a closure
	//
	// 588 deep to 610, 2,203 flat to 2,346. `deep` rising is the over-approximating direction by the
	// rule above, and the oracle that says otherwise is the board rather than a scope count: false
	// positives fall 13 to 11 with `under` held at 5 fixtures / 7 scopes. A frozen capture no longer
	// widens its source's range, so scopes that were fusing across a closure stay separate and a
	// value read inside one is read at depth rather than at the root of a merged scope.
	// # And again when the frozen-capture rule gained its destination half
	//
	// 610 deep to 615, 2,346 flat to 2,376. `deep` rising is the over-approximating direction and no
	// oracle endorses it, which is why the move is small and the reason is fidelity rather than a
	// gain: upstream prunes a capture whose DESTINATION is not mutable, in the same arm and the same
	// comment as the source half -- "If the destination is not mutable, or the source value has
	// copy-on-write semantics, then we can prune the effect".
	//
	// Measured rather than assumed harmless. The branch fires 9,036 times across the 400-file corpus
	// and moves no goal number: board 25 goldens and 11 false positives, `under` 5 fixtures / 7
	// scopes, and all three oracles unchanged. The scope oracle's per-fixture dump is byte-identical
	// across all 48 scored fixtures, so this does not move the tree away from upstream anywhere
	// upstream can be observed. It buys nothing today and removes a divergence that would matter as
	// soon as a fixture exercises it.
	// # And again when a ref's stability stopped being read off its name
	//
	// 615 deep unchanged, 2,376 flat to 2,379. Three more bare dependencies and no change in depth,
	// which is what a value that stops being treated as stable should produce: it becomes reactive,
	// survives `PruneNonReactiveDependencies`, and is named at its root. `deep` is the number this
	// test's rule is about and it does not move.
	// # And again when a scope inside a branch stopped being seeded from that branch
	//
	// 615 deep to 576, 2,379 flat to 2,414. Thirty-nine paths become shallow, and this is the only
	// entry here where `deep` FALLS. That is the conservative direction the rule above asks for: a
	// property proven non-null only under an `if` no longer licenses the walk to descend past it
	// when the scope asking began inside that same `if`.
	//
	// Endorsed by the board rather than by the count. With the memo callback inlined,
	// `error.useMemo-infer-less-specific-conditional-access` goes from silent to reporting -- the
	// fixture exists precisely because upstream infers bare `propB` there -- and the golden trade
	// improves from three fixtures for two false positives to two for one. `under` holds at 0
	// fixtures / 0 scopes, which is the number that would have moved had this made dependencies too
	// shallow to preserve a memoization.
	// # Rebased onto the production pipeline and explicit optional CFG
	//
	// The old 576 / 2,414 pin ran range and scope inference directly after `Construct`, while the
	// production path had moved to Outline -> InferReactive -> Drop -> inclusive Inline -> Merge ->
	// DCE. On the final tree, the old probe reports 622 / 2,778 and the production-aligned probe
	// reports 626 / 2,731: alignment recovers four deep paths and removes 47 flat dependencies that
	// do not survive the real prefix.
	//
	// The larger movement from the old pin is the explicit optional-chain CFG, its authoritative
	// sidemap, and signature-aware call allocation. It is accepted against the actual oracles rather
	// than this distribution alone: golden dependency matches are 81/88 with production down to 123
	// dependencies against upstream's 116, the scope oracle remains at 0 under-produced fixtures,
	// and preserve-manual-memoization is 28/28 with 0/69 false positives.
	// # LoadContext restored to CreateFrom
	//
	// Assign made a value loaded from a context box share the box's abstract identity. Upstream uses
	// CreateFrom: the load copies the box's current kind into a distinct value, so freezing the box
	// later does not retroactively freeze the earlier load. Restoring that distinction leaves all 626
	// property paths unchanged and adds 33 root-only dependencies, 2,731 flat to 2,764. The score and
	// structural oracles hold, so this is the expected conservative movement from the corrected kind
	// lifetime rather than added path depth.
	// # Re-pinned 2026-09-06: 2,764 to 2,809 flat, deep unchanged at 626
	//
	// This distribution is measured over a LIVE directory, not a frozen fixture, so it moves when
	// that directory does. Between the previous pin and this one the corpus gained 151 uncommitted
	// files and an import-aliasing pass applied by this tool's own fixer, which is precisely the
	// kind of edit a dependency analysis is supposed to notice: 45 more root-only dependencies, no
	// change in path depth.
	//
	// `deep` holding at 626 is what makes this a corpus movement rather than an analysis movement.
	// The guidance below is that a gain in `deep` is the over-approximating direction; a gain in
	// `flat` alone, with `deep` and `maxIterations` steady, is the shape a larger corpus produces.
	// Had the analysis itself drifted, both numbers would have moved.
	//
	// A pin against a live tree is worth keeping anyway: it fails loudly when either the corpus or
	// the analysis moves, which is a question worth being asked even when the answer is "the corpus
	// moved." A pin nobody can explain a movement in is the one to distrust.
	// # React-compatible reverse postorder
	//
	// Visiting structural fallthroughs first while constructing postorder puts loop bodies before
	// their continuations after reversal. That changes the mutable intervals consumed here and moves
	// the live corpus to 625 deep / 2,843 flat. The two direct preserve-memoization oracles remain
	// 28/28 positive and 69/69 clean, so the movement is pinned as a corrected input order rather
	// than accepted from this distribution alone. After the loop-header correction, state freezing
	// moves 625/2841 to 631/3043; STATE_EFFECTS.md accounts for all added and removed scope inputs.
	// Effect-hook freezing moves those to 649/3118, attributed in EFFECT_HOOK_EFFECTS.md. Immutable
	// alias-edge refinement moves those to 685/3227; IMMUTABLE_ALIAS_EDGES.md lists the full delta.
	// # Frozen 2026-10-01: 769 deep / 3,320 flat, over Structure at pinnedCorpusCommit
	//
	// The live pin of 796 / 3,310 failed after every lint sweep of the directory, 800 / 3,322 and then
	// 773 / 3,300, so the input is now a fixed commit rather than the working tree. The new figures are
	// a corpus movement and not an analysis one, measured: the analysis at the old pin's own commit
	// (04501c3) and at the commit after it that touched ranges (6f04852) both produce 769 / 3,320 over
	// the frozen corpus. Discrimination survives the change: the immutability gate forced always true
	// gives 832 / 3,295 with 3 iterations, and fails here.
	if deep != 769 || flat != 3320 {
		t.Errorf("got %d deep and %d flat dependencies, want 769 and 3320; a SMALL move here is "+
			"what every mutation of this analysis produces, and a gain in `deep` specifically is "+
			"the over-approximating direction unless an oracle says otherwise", deep, flat)
	}
	if maxIterations != 2 {
		t.Errorf("max iterations %d, want 2; the intersection at a join is what keeps this low, "+
			"and unioning instead produces the same answer in 4", maxIterations)
	}
	t.Logf("%d dependencies with a path, %d without; max iterations %d", deep, flat, maxIterations)
}
