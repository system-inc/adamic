package high_level_intermediate_representation

import (
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The relationship between a memo block's markers and the scopes that carry its dependencies,
// counted across the whole corpus.
//
// # Why this exists
//
// `preserve-manual-memoization` compares what the compiler inferred as a scope's dependencies
// against what the developer wrote in the memo block enclosing it. That comparison can only run on a
// scope the walk reaches while the block is still open. A scope carrying dependencies that closes
// OUTSIDE the marker pair is invisible to the rule no matter how correct the comparison is.
//
// Tracing one fixture by hand says whether that happened once. It cannot say whether the shape is
// the common case or an idiosyncrasy of three fixtures, and that distinction decides whether the
// scope boundary is a systematic defect worth its own repair or a curiosity. This counts it.
//
// # What it measures, and why the counts are shaped this way
//
// The walk here reproduces `manualMemoValidator.walk` exactly: same descent, same marker pairing by
// `ManualMemoId`, same "a scope is inside the block if the walk is between its markers when the
// scope CLOSES" rule. It has to, because the number is only meaningful as a statement about what
// the validator can see. A structurally similar walk that differed in when it considers a scope
// closed would produce a number about nothing.
//
// Four populations, per memo block:
//
//	scopesInside          scopes closing between the markers
//	carryingInside        those with at least one inferred dependency  <- the rule can compare these
//	emptyInside           those with none                              <- reachable but nothing to say
//	carryingOutsideAfter  dependency-carrying scopes closing after `FinishMemoize`
//
// `carryingOutsideAfter` is deliberately NOT a defect count on its own, which took a wrong draft to
// learn. The JSX consuming a memoized value legitimately closes after the block and legitimately
// depends on it: `<button onClick={onClick} />` is a scope carrying `onClick`, outside the block,
// and correct. Counting those as misplaced reported 3 of 4 fixtures broken when none of them were.
//
// So the readable signal is `emptyInside` against `carryingInside`. A memo block whose interior
// scopes all carry nothing is a block the rule cannot fire on for any reason at all.
func TestMemoBlockScopeRelationshipAcrossCorpus(t *testing.T) {
	t.Parallel()

	fixtures, err := react_conformance.Load("../../rules/react/conformance/testdata/fixtures")
	if err != nil {
		t.Fatalf("loading the vendored corpus: %v", err)
	}

	var (
		fixturesScored     int
		flowExcluded       int
		blocksTotal        int
		blocksAllEmpty     int
		blocksWithCarrying int
		scopesInside       int
		carryingInside     int
		emptyInside        int
		carryingAfter      int
		capturingInside    int

		blocksAllEmptyWithCapture  int
		blocksAllEmptyContradicted int
		emptyWithCaptureNames      []string
	)
	for _, fixture := range fixtures {
		if !strings.Contains(fixture.Source, "validatePreserveExistingMemoizationGuarantees") {
			continue
		}
		if fixture.RequiresFlow() {
			// Same exclusion the dependency oracle states: a Flow fixture lowers to nothing here, so
			// every memo block in it would count as zero-scope and read as a defect.
			flowExcluded++
			continue
		}
		observations, ok := memoBlockObservations(t, fixture.Source)
		if !ok {
			continue
		}
		fixturesScored++
		for _, observation := range observations {
			blocksTotal++
			scopesInside += observation.scopesInside
			carryingInside += observation.carryingInside
			emptyInside += observation.emptyInside
			carryingAfter += observation.carryingAfter
			capturingInside += observation.capturingInside
			if observation.carryingInside > 0 {
				blocksWithCarrying++
				continue
			}
			if observation.scopesInside == 0 {
				continue
			}
			blocksAllEmpty++
			if len(goldenDependencies(t, fixture.ExpectPath)) > 0 {
				// Upstream's compiled output emits a cache slot for this fixture, so it inferred a
				// dependency where every scope inside our block carries none. This is the answer
				// key rather than our own shape: a disagreement visible per fixture.
				blocksAllEmptyContradicted++
			}
			if observation.capturingInside > 0 {
				// The actionable subset: nothing to compare against, and a closure inside the block
				// that demonstrably closes over something. The value the block reads is real and
				// lives in a nested function the dependency collector does not descend into.
				blocksAllEmptyWithCapture++
				emptyWithCaptureNames = append(emptyWithCaptureNames, fixture.Name)
			}
		}
	}

	if blocksTotal == 0 {
		t.Fatal("no memo block was observed across the corpus, so this test asserts nothing; " +
			"marker construction or the walk changed rather than the corpus being empty")
	}

	t.Logf("fixtures=%d flowExcluded=%d blocks=%d blocksWithCarrying=%d blocksAllEmpty=%d",
		fixturesScored, flowExcluded, blocksTotal, blocksWithCarrying, blocksAllEmpty)
	t.Logf("scopesInside=%d carryingInside=%d emptyInside=%d capturingInside=%d carryingAfter=%d",
		scopesInside, carryingInside, emptyInside, capturingInside, carryingAfter)
	t.Logf("blocksAllEmptyWithCapture=%d blocksAllEmptyContradictedByGolden=%d of %d all-empty blocks",
		blocksAllEmptyWithCapture, blocksAllEmptyContradicted, blocksAllEmpty)
	sort.Strings(emptyWithCaptureNames)
	for _, name := range emptyWithCaptureNames {
		t.Logf("  empty-with-capture: %s", name)
	}

	// Measured floors. `blocksWithCarrying` is the population the rule's third condition can fire
	// on at all, so it is the number that must not fall; `blocksAllEmpty` is the defect count and
	// must not rise.
	//
	// Both are pinned rather than logged because the asymmetry that nearly shipped a regression as
	// an improvement in the score test applies here identically: a change that empties more blocks
	// while emptying fewer elsewhere would read as neutral on a single total.
	// # Moved when React's stable built-ins stopped being dependencies
	//
	// 74 carrying to 69, 15 all-empty to 20. Five memo blocks became empty because the only
	// dependency they carried was a value React guarantees is identity-stable -- a `useTransition`
	// start function, a `useState` setter, a `useOptimistic` setter.
	//
	// An empty block is the RIGHT answer for those, and upstream's own output says so: both
	// `preserve-use-memo-transition.ts` and `preserve-use-callback-stable-built-ins.ts` compile to
	// `if ($[0] === Symbol.for("react.memo_cache_sentinel"))`, which is the marker for a value
	// memoized once with no dependencies at all. Neither expects an error.
	//
	// So this count moving down is not a regression here, and the direction of `blocksWithCarrying`
	// is not always the direction of correctness. The ceiling on `blocksAllEmptyWithCapture` below
	// is what still guards the defect this oracle was built for.
	// Lowered from 69 to 67 by defining a context binding once in SSA rather than versioning every
	// write (see `defineIn` in `ssa.go`). Two blocks lose an interior carrying scope because the
	// declaration and the reassignment now share one class rather than forming two, so what was two
	// scopes carrying dependencies is one.
	//
	// The paragraph above is the reason this is not read as a regression: `blocksAllEmpty` holds at
	// 20 and `blocksAllEmptyWithCapture` at 14, which are the counts that guard the defect this
	// oracle exists for. And the change it records fires both
	// `error.invalid-useCallback-captures-reassigned-context` fixtures, which is the third
	// condition reaching programs it could not reach before rather than fewer.
	// # Lowered again when a callback's mutation began widening the receiver's range
	//
	// 67 to 66, and the block that stops carrying is named rather than assumed: logging the fixture
	// per carrying block and diffing the two configurations gives exactly
	// `error.validate-object-values-mutation`, with nothing gained.
	//
	// Upstream carries zero scopes inside that memo block. Read at the same stage, its output is
	// two scopes -- `SCOPE 1 decls=[52]` and `SCOPE 3 decls=[65]` -- with the memo markers sitting
	// INSIDE scope 1 rather than wrapping a scope of their own, so there is no scope between them
	// to carry anything. Our widening produces that same shape.
	//
	// So this is the third lowering for the reason the paragraph above already records twice: the
	// direction of `blocksWithCarrying` is not always the direction of correctness. Unlike the
	// earlier two, this one is checked against upstream's own output rather than argued from ours.
	// # Lowered once more, and the block is the sibling of the last one
	//
	// 66 to 65 on exactly `error.validate-object-entries-mutation`, named the same way and with
	// nothing gained. It differs from `error.validate-object-values-mutation` only in destructuring
	// its callback parameter, and upstream's output for it has the same shape: two scopes,
	// `SCOPE 1 decls=[54]` and `SCOPE 3 decls=[69]`, with the memo markers inside scope 1 rather
	// than wrapping one, so there is no scope between them to carry anything.
	//
	// Read from upstream rather than inferred from the sibling, because the two fixtures differ in
	// the callback and the symmetry was worth checking.
	const knownBlocksWithCarrying = 65
	//
	// 22 with the frozen-capture rule in `ranges.go`. Both added blocks are
	// `useMemo-constant-prop`, whose two memo blocks memoize over a constant.
	const knownBlocksAllEmpty = 22
	if blocksWithCarrying < knownBlocksWithCarrying {
		t.Errorf("blocks with a dependency-carrying interior scope = %d, want at least %d; the "+
			"rule's third condition can no longer reach programs it could reach before",
			blocksWithCarrying, knownBlocksWithCarrying)
	}
	if blocksAllEmpty > knownBlocksAllEmpty {
		t.Errorf("memo blocks whose every interior scope carries no dependency = %d, want at most "+
			"%d; more blocks became invisible to the comparison", blocksAllEmpty, knownBlocksAllEmpty)
	}
	// The capture subset is pinned separately because it is the one with a known cause. It should
	// fall to zero when the dependency collector descends into nested functions, and a rise means
	// more blocks lost their dependencies to a closure boundary.
	// Raised from 10 by the same change. The four added are the stable-built-in fixtures above,
	// whose blocks hold a capturing function expression and correctly carry no dependency.
	//
	// 16 with the frozen-capture rule, the same two blocks as above.
	const knownBlocksAllEmptyWithCapture = 16
	if blocksAllEmptyWithCapture > knownBlocksAllEmptyWithCapture {
		t.Errorf("all-empty memo blocks holding a capturing closure = %d, want at most %d; the "+
			"closure boundary is swallowing more dependencies than before",
			blocksAllEmptyWithCapture, knownBlocksAllEmptyWithCapture)
	}
	// The strongest of the three, because it is upstream's answer rather than our own shape: a
	// fixture whose compiled output emits a cache slot while every scope in our block carries
	// nothing. Zero is the target and any rise is a straightforward regression.
	// Raised from 4 by the same change, and this one deserves the most scrutiny of the four: it
	// counts blocks where upstream emits a cache slot and we carry nothing. The added fixture is
	// `preserve-use-callback-stable-built-ins.ts`, whose slot is the sentinel rather than a
	// dependency comparison -- `_c(1)` with `$[0] === Symbol.for("react.memo_cache_sentinel")`.
	// Counting a sentinel as a contradicted dependency is a property of this oracle's slot pattern
	// rather than of our output, and is left recorded rather than silently filtered.
	//
	// 7 with the frozen-capture rule, and this rise was checked against the fixture rather than
	// accepted, because zero is this number's target. Both added blocks are `useMemo-constant-prop`,
	// and its slots are `Symbol.for("react.memo_cache_sentinel")` rather than dependency
	// comparisons -- the fixture holds exactly one `!==` guard and it is on the returned array, not
	// on either memo block. So upstream infers no dependency in either block either, and this is the
	// same sentinel-counting property of the slot pattern already recorded above rather than a
	// disagreement about dependencies.
	//
	// The rule's own answer on that fixture went from wrong to right in the same change: it is one
	// of the two false positives that close at `knownFalsePositives`.
	const knownBlocksAllEmptyContradicted = 7
	if blocksAllEmptyContradicted > knownBlocksAllEmptyContradicted {
		t.Errorf("all-empty memo blocks contradicted by upstream's compiled output = %d, want at "+
			"most %d; upstream infers a dependency in a block where we now infer none",
			blocksAllEmptyContradicted, knownBlocksAllEmptyContradicted)
	}
}

// memoBlockObservation is one memo block's scope population.
type memoBlockObservation struct {
	scopesInside    int
	carryingInside  int
	emptyInside     int
	carryingAfter   int
	capturingInside int
}

// memoBlockObservations walks one fixture the way the validator does and counts per memo block.
func memoBlockObservations(t *testing.T, source string) ([]memoBlockObservation, bool) {
	t.Helper()

	var all []memoBlockObservation
	lowered := false
	probe := rule.Rule{
		Name:             "memo-block-scope-oracle",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						return
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						lowered = true
						all = append(all, observeMemoBlocks(function, ctx.TypeChecker)...)
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")

	return all, lowered
}

// observeMemoBlocks runs the phase-6 pipeline and counts the scope population of each memo block.
//
// The pipeline is `pipelineFindings`' pipeline, for the same reason that one states: the tree the
// validator walks is the tree several pruning passes produced, and observing a differently-built
// tree would measure a world the rule never sees.
func observeMemoBlocks(function *Function, checker *shimchecker.Checker) []memoBlockObservation {
	InferReactive(function, checker)
	DropManualMemoization(function)

	ranges := InferMutableRanges(function)
	set := FindDisjointMutableValuesWithRanges(function, ranges)
	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
	BuildReactiveScopeTerminals(function, scopes, identity)
	dependencies := CollectScopeDependenciesWithHoistable(function, scopes, identity, ranges)

	tree, _ := BuildReactiveFunction(function)
	if tree == nil {
		return nil
	}

	MergeReactiveScopesThatInvalidateTogether(tree, function, dependencies, checker)
	PruneNonEscapingScopesWithScopes(tree, function, dependencies, scopes, checker)
	PruneUnusedScopes(tree, dependencies)
	PruneAlwaysInvalidatingScopes(tree, function, dependencies)
	PruneNonReactiveDependencies(tree, function, dependencies)

	observer := memoBlockObserver{function: function, dependencies: dependencies}
	observer.walk(tree.Body)
	return observer.observations()
}

// memoBlockObserver reproduces `manualMemoValidator.walk`'s descent and marker pairing.
type memoBlockObserver struct {
	function     *Function
	dependencies *ScopeDependencies
	// open are the `ManualMemoId`s of blocks opened and not yet closed, mirroring the validator's
	// `openMemoBlocks` rather than a stack, because memo blocks do not reliably nest once the
	// callbacks are inlined.
	open   map[int]bool
	closed []int
	byId   map[int]*memoBlockObservation
	order  []int
}

func (o *memoBlockObserver) ensure(id int) *memoBlockObservation {
	if o.byId == nil {
		o.byId = map[int]*memoBlockObservation{}
	}
	if existing, ok := o.byId[id]; ok {
		return existing
	}
	fresh := &memoBlockObservation{}
	o.byId[id] = fresh
	o.order = append(o.order, id)
	return fresh
}

func (o *memoBlockObserver) observations() []memoBlockObservation {
	out := make([]memoBlockObservation, 0, len(o.order))
	for _, id := range o.order {
		out = append(out, *o.byId[id])
	}
	return out
}

func (o *memoBlockObserver) walk(block ReactiveBlock) {
	for _, statement := range block {
		switch shape := statement.(type) {
		case *ReactiveInstructionStatement:
			o.visitInstruction(shape.Instruction)

		case *ReactiveScopeBlock:
			o.walk(shape.Instructions)
			// Counted on scope EXIT, which is where the validator runs its comparison. A scope that
			// opened inside a block and closes after it is not comparable, so open-time membership
			// would count a population the rule cannot act on.
			carrying := len(o.dependencies.DependenciesOf(shape.Scope)) > 0
			if len(o.open) > 0 {
				for id := range o.open {
					observation := o.ensure(id)
					observation.scopesInside++
					if carrying {
						observation.carryingInside++
					} else {
						observation.emptyInside++
					}
					if o.scopeCaptures(shape) {
						observation.capturingInside++
					}
				}
				continue
			}
			if carrying {
				// Outside every block. Attributed to the most recently closed one, which is the
				// block such a scope would belong to if the boundary were drawn differently.
				if len(o.closed) > 0 {
					o.ensure(o.closed[len(o.closed)-1]).carryingAfter++
				}
			}

		case *ReactiveTerminalStatement:
			o.walkTerminal(shape)
		}
	}
}

// scopeCaptures reports whether the scope holds a `FunctionExpression` that closes over anything.
//
// Recorded alongside the dependency counts because a scope carrying no dependency but holding a
// capturing closure is a distinguishable population: the value it reads is real and lives in a
// nested function, rather than the scope genuinely reading nothing.
func (o *memoBlockObserver) scopeCaptures(scope *ReactiveScopeBlock) bool {
	for _, statement := range scope.Instructions {
		instruction, ok := statement.(*ReactiveInstructionStatement)
		if !ok || instruction.Instruction == nil || instruction.Instruction.Value == nil {
			continue
		}
		plain, isPlain := instruction.Instruction.Value.(*ReactiveInstructionValue)
		if !isPlain || plain.Value == nil {
			continue
		}
		if expression, isFunction := plain.Value.(*FunctionExpression); isFunction {
			if len(expression.Captures) > 0 {
				return true
			}
		}
	}
	return false
}

func (o *memoBlockObserver) visitInstruction(instruction *ReactiveInstruction) {
	if instruction == nil || instruction.Value == nil {
		return
	}
	if sequence, isSequence := instruction.Value.(*ReactiveSequenceValue); isSequence {
		for _, nested := range sequence.Instructions {
			o.visitInstruction(nested)
		}
	}
	plain, isPlain := instruction.Value.(*ReactiveInstructionValue)
	if !isPlain || plain.Value == nil {
		return
	}
	switch marker := plain.Value.(type) {
	case *StartMemoize:
		if o.open == nil {
			o.open = map[int]bool{}
		}
		o.open[marker.ManualMemoId] = true
		o.ensure(marker.ManualMemoId)
	case *FinishMemoize:
		if !o.open[marker.ManualMemoId] {
			return
		}
		delete(o.open, marker.ManualMemoId)
		o.closed = append(o.closed, marker.ManualMemoId)
	}
}

func (o *memoBlockObserver) walkTerminal(statement *ReactiveTerminalStatement) {
	switch shape := statement.Terminal.(type) {
	case *ReactiveIf:
		o.walk(shape.Consequent)
		if shape.Alternate != nil {
			o.walk(*shape.Alternate)
		}
	case *ReactiveSwitch:
		for index := range shape.Cases {
			if shape.Cases[index].Block != nil {
				o.walk(*shape.Cases[index].Block)
			}
		}
	case *ReactiveFor:
		o.walk(shape.Loop)
	case *ReactiveForOf:
		o.walk(shape.Loop)
	case *ReactiveForIn:
		o.walk(shape.Loop)
	case *ReactiveWhile:
		o.walk(shape.Loop)
	case *ReactiveDoWhile:
		o.walk(shape.Loop)
	case *ReactiveLabelTerminal:
		o.walk(shape.Block)
	case *ReactiveTry:
		o.walk(shape.Block)
		o.walk(shape.Handler)
	}
}
