package high_level_intermediate_representation

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// terminalsFor lowers one source, runs the full scope pipeline, and builds the terminals.
//
// Typed for the reason `rangesFor` and `scopesFor` are: a checker-less lowering never emits
// `StoreContext`, so several assertions below would pass vacuously on a nil checker.
func terminalsFor(t *testing.T, source string) (*Function, *ReactiveScopes, ScopeIdentity,
	ScopeTerminals) {
	t.Helper()
	function, scopes := scopesFor(t, source)
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
	return function, scopes, identity, BuildReactiveScopeTerminals(function, scopes, identity)
}

// scopeTerminalsIn returns every Scope terminal in the function, in block order.
func scopeTerminalsIn(function *Function) []*Scope {
	var found []*Scope
	for _, block := range function.Blocks {
		if terminal, ok := block.Terminal.(*Scope); ok {
			found = append(found, terminal)
		}
	}
	return found
}

// ---------------------------------------------------------------------------
// The variant is CONSTRUCTED, which is the whole reason it was added
// ---------------------------------------------------------------------------

// TestScopeTerminalsAreBuilt is the guard against this variant becoming a second `Optional`.
//
// `Optional` is declared in this package's terminal set and is never constructed: on `props?.a?.b`
// the lowering produces zero `Optional` terminals and two `PropertyLoad` with `Optional: true`. A
// terminal nothing builds is worse than no terminal, because a consumer writes an arm for it, the
// arm never fires, and the dead branch reads as coverage.
//
// Four stages declined to add `Scope` early for exactly that reason. This test is what makes the
// decision reversible: if the producer is ever removed or silently stops firing, this fails rather
// than leaving a hole in every switch.
func TestScopeTerminalsAreBuilt(t *testing.T) {
	t.Parallel()

	function, _, _, result := terminalsFor(t, `
		function Component(props) {
			const items = [];
			items.push(props.a);
			return <div>{items}</div>;
		}
	`)
	if result.Built == 0 {
		t.Fatal("no Scope terminal was constructed; the variant would be a second Optional")
	}
	built := scopeTerminalsIn(function)
	if len(built) != result.Built {
		t.Errorf("reported %d terminals built but the graph holds %d", result.Built, len(built))
	}
	for _, terminal := range built {
		if terminal.Scope == 0 {
			t.Error("a Scope terminal names scope zero, which is the absent value")
		}
		if !HasBlock(terminal.Block) {
			t.Error("a Scope terminal names no body block")
		}
		if !HasBlock(terminal.Fallthrough) {
			t.Error("a Scope terminal names no fallthrough")
		}
	}
}

// TestScopeTerminalIsReachableThroughFallthrough pins that the new variant is wired into the two
// switches that have a `default` arm and would otherwise have accepted it silently.
//
// `Fallthrough` and `terminalTestIsReactive` both default rather than going red, so the compiler
// could not have found them. They were found by enumerating every switch over `Terminal` before the
// variant was added, which is the method this test exists to keep honest.
func TestScopeTerminalIsReachableThroughFallthrough(t *testing.T) {
	t.Parallel()

	terminal := &Scope{Scope: 7, Block: 3, Fallthrough: 9, Order: 4}

	block, ok := Fallthrough(terminal)
	if !ok || block != 9 {
		t.Errorf("Fallthrough gave (%d, %v), want (9, true): the default arm swallowed the variant",
			block, ok)
	}
	if order := TerminalOrder(terminal); order != 4 {
		t.Errorf("TerminalOrder gave %d, want 4", order)
	}

	// The body is a real edge; the fallthrough is not.
	var successors []static_single_assignment.BlockId
	EachSuccessor(terminal, func(id static_single_assignment.BlockId) { successors = append(successors, id) })
	if len(successors) != 1 || successors[0] != 3 {
		t.Errorf("EachSuccessor gave %v, want [3]: the fallthrough must not be an edge", successors)
	}
	var withFallthrough []static_single_assignment.BlockId
	EachSuccessorAndFallthrough(terminal, func(id static_single_assignment.BlockId) {
		withFallthrough = append(withFallthrough, id)
	})
	if len(withFallthrough) != 2 {
		t.Errorf("EachSuccessorAndFallthrough gave %v, want the body and the fallthrough",
			withFallthrough)
	}

	// A Scope terminal reads no places, exactly like Goto.
	places := 0
	EachTerminalPlace(terminal, func(Place, PlaceRole) { places++ })
	if places != 0 {
		t.Errorf("a Scope terminal reported %d places, want 0", places)
	}
}

// TestScopeTerminalPrints pins the printer arm, which has no default and would have gone red, but
// whose OUTPUT nothing else asserts.
func TestScopeTerminalPrints(t *testing.T) {
	t.Parallel()

	function := NewFunction(nil, "probe", FunctionKindOther)
	got := printTerminal(function, &Scope{Scope: 2, Block: 5, Fallthrough: 8})
	want := "Scope @2 block=bb5 fallthrough=bb8"
	if got != want {
		t.Errorf("printTerminal gave %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Identity: the defect that built more terminals than there are scopes
// ---------------------------------------------------------------------------

// TestScopeTerminalsKeyOnTheMergedGroup pins the fix for the one real defect this pass shipped and
// then corrected.
//
// `ReactiveScopes.ScopeOf` returns the PRE-MERGE id. Keying on it while reading the MERGED range
// produces a coherent-looking wrong graph: every range is plausible, nothing is empty or inverted,
// and the nesting precondition still passes. Measured on the corpus it built 3,286 terminals for
// 2,874 scopes, 412 of them duplicates wrapping a range another terminal already owned.
//
// The tell was that the built count EXCEEDED the control, which is why the corpus test below asserts
// against the merged count rather than merely asserting the count is non-zero.
func TestScopeTerminalsKeyOnTheMergedGroup(t *testing.T) {
	t.Parallel()

	function, scopes, identity, result := terminalsFor(t, `
		function Component(props) {
			const a = {};
			const b = {};
			a.x = props.x;
			b.y = a;
			return <div>{b}</div>;
		}
	`)
	if result.Built == 0 {
		t.Skip("this source produced no scopes, so there is nothing to key")
	}

	// No two terminals may name the same surviving scope.
	seen := map[ScopeId]int{}
	for _, terminal := range scopeTerminalsIn(function) {
		seen[terminal.Scope]++
	}
	for scope, count := range seen {
		if count > 1 {
			t.Errorf("scope %d has %d terminals; duplicates mean the pass keyed on the pre-merge id",
				scope, count)
		}
	}

	// Every terminal must name a SURVIVING scope: one that is its own group.
	for scope := range seen {
		if group := identity.GroupOf(scope); group != scope {
			t.Errorf("terminal names scope %d, which merged into %d", scope, group)
		}
	}
	_ = scopes
}

// ---------------------------------------------------------------------------
// The precondition
// ---------------------------------------------------------------------------

// TestScopeTerminalsDeclineOnUnnestedScopes pins that the pass refuses rather than corrupting a
// graph whose scopes upstream's own invariant would reject.
//
// Upstream raises `CompilerError.invariant('Invalid nesting in program blocks or scopes')` and
// stops. A linter declines the function instead, for the reason `ValidateScopes` gives: a linter
// that stops on a graph it dislikes is worse than one that declines a function.
func TestScopeTerminalsDeclineOnUnnestedScopes(t *testing.T) {
	t.Parallel()

	function, scopes := scopesFor(t, `
		function Component(props) {
			const a = {};
			a.x = props.x;
			return <div>{a}</div>;
		}
	`)
	if scopes.Len() == 0 {
		t.Skip("no scopes in this source")
	}

	// Force a violation: two scopes that overlap without nesting.
	broken := overlappingIdentity{scopes: scopes}
	if ScopeTerminalsPrecondition(function, scopes, broken) == 0 {
		t.Skip("this source has too few scopes to construct an overlap")
	}
	before := len(scopeTerminalsIn(function))
	result := BuildReactiveScopeTerminals(function, scopes, broken)
	if result.Built != 0 {
		t.Errorf("built %d terminals on a graph violating the nesting precondition", result.Built)
	}
	if after := len(scopeTerminalsIn(function)); after != before {
		t.Errorf("the graph was mutated on a declined function: %d terminals became %d", before, after)
	}
}

// overlappingIdentity forces every scope into a deliberately non-nested arrangement, so the
// precondition has something to reject.
type overlappingIdentity struct{ scopes *ReactiveScopes }

func (o overlappingIdentity) GroupOf(scope ScopeId) ScopeId { return scope }
func (o overlappingIdentity) RangeOf(scope ScopeId) mutation_aliasing.MutableRange {
	// Staircase: each scope starts one later and ends far past the previous end, so consecutive
	// scopes overlap and neither contains the other.
	return mutation_aliasing.MutableRange{Start: static_single_assignment.EvaluationOrder(scope), End: static_single_assignment.EvaluationOrder(scope) + 10}
}

// ---------------------------------------------------------------------------
// Termination
// ---------------------------------------------------------------------------

// TestScopeTerminalsAreASinglePass pins the termination argument empirically.
//
// There is no worklist and no fixpoint: the scope traversal enters each scope once and exits it
// once, and the block sweep visits each block once and each instruction position within it once,
// consuming from a queue that only shrinks. This counts the visits and asserts they are bounded by
// exactly that, so an accidental re-entry becomes a failure rather than a slowdown.
func TestScopeTerminalsAreASinglePass(t *testing.T) {
	t.Parallel()

	function, scopes := scopesFor(t, `
		function Component(props) {
			const list = [];
			for (const item of props.items) {
				list.push(item);
			}
			const other = {};
			other.value = props.value;
			return <div>{list}{other}</div>;
		}
	`)
	if scopes.Len() == 0 {
		t.Skip("no scopes in this source")
	}
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}

	items := scopeItemsInNestingOrder(function, scopes, identity)
	rewrites := queueScopeRewrites(function, items)

	// Exactly one start and one end per scope: 2n rewrites, never more.
	if len(rewrites) != 2*len(items) {
		t.Errorf("queued %d rewrites for %d scopes, want exactly %d",
			len(rewrites), len(items), 2*len(items))
	}
	starts, ends := 0, 0
	perScope := map[ScopeId]int{}
	for _, rewrite := range rewrites {
		if rewrite.isStart {
			starts++
			perScope[rewrite.scope]++
		} else {
			ends++
		}
	}
	if starts != len(items) || ends != len(items) {
		t.Errorf("got %d starts and %d ends for %d scopes", starts, ends, len(items))
	}
	for scope, count := range perScope {
		if count != 1 {
			t.Errorf("scope %d was entered %d times, want once", scope, count)
		}
	}
}

// TestScopeRewritesAreQueuedInTraversalOrderNotSortedByPosition pins the ordering property, and it
// exists because getting this wrong is what this pass actually got wrong.
//
// The queue is the traversal's own output, REVERSED so the sweep can pop from the end. It is
// deliberately not sorted. An earlier draft sorted by position with a start-versus-end tiebreak,
// which reads as a harmless normalisation; two mutants of that comparator survived the sweep, and
// the fixture written to kill one of them FAILED against the unmutated code. That failure was the
// finding: the comparator was deciding something the traversal had already decided, and position
// alone cannot reconstruct nesting.
//
// What must hold is that popping from the end yields non-decreasing positions, which is what makes
// the single forward sweep over instructions correct. That is a consequence of the traversal, not of
// a sort, and asserting it here means a future reintroduction of a sort has to keep it true.
func TestScopeRewritesAreQueuedInTraversalOrderNotSortedByPosition(t *testing.T) {
	t.Parallel()

	function, scopes := scopesFor(t, `
		function Component(props) {
			const a = {};
			a.x = props.x;
			const b = {};
			b.y = props.y;
			return <div>{a}{b}</div>;
		}
	`)
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
	items := scopeItemsInNestingOrder(function, scopes, identity)
	rewrites := queueScopeRewrites(function, items)
	if len(rewrites) < 4 {
		t.Fatalf("this source produced only %d rewrites; it can no longer discriminate", len(rewrites))
	}

	// Consumed back to front, positions must never decrease: the sweep walks instructions forward
	// and applies every rewrite at or before the current position, so a later pop at an earlier
	// position would be applied in the wrong block.
	previous := rewrites[len(rewrites)-1].order
	for index := len(rewrites) - 2; index >= 0; index-- {
		if rewrites[index].order < previous {
			t.Errorf("popping yields position %d after %d; the sweep applies rewrites in one "+
				"forward pass and cannot revisit an earlier position",
				rewrites[index].order, previous)
		}
		previous = rewrites[index].order
	}
}

// TestAScopeEndingWhereAnotherBeginsClosesFirst pins the behaviour at a position collision.
//
// Two scopes can meet at a point: one ends at position 7, the next begins at 7. They are DISJOINT --
// `curr.start >= parent.end` -- so the traversal exits the first before entering the second, and the
// end rewrite is queued before the start with no comparison between them.
//
// This is the case a sort with a tiebreak gets wrong in whichever direction the tiebreak is written,
// which is why both mutants of it survived: the fixtures could not see a comparator that should not
// have been consulted. Measured on this source, the collision is at position 7.
func TestAScopeEndingWhereAnotherBeginsClosesFirst(t *testing.T) {
	t.Parallel()

	function, scopes := scopesFor(t, `
		function Component(props) {
			const a = {};
			a.x = props.x;
			const b = {};
			b.y = props.y;
			return <div>{a}{b}</div>;
		}
	`)
	aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
	identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
	items := scopeItemsInNestingOrder(function, scopes, identity)
	rewrites := queueScopeRewrites(function, items)

	collisions := 0
	for index := len(rewrites) - 1; index > 0; index-- {
		applied, next := rewrites[index], rewrites[index-1]
		if applied.order != next.order {
			continue
		}
		collisions++
		if applied.isStart && !next.isStart {
			t.Errorf("at position %d a start is applied before an end at the same position; the "+
				"traversal exits a disjoint scope before entering the next and must queue the end "+
				"first", applied.order)
		}
	}
	if collisions == 0 {
		t.Fatal("this source no longer produces two rewrites at one position, so it cannot see " +
			"the collision case; find a new one by logging rewrite positions")
	}
}

// ---------------------------------------------------------------------------
// What the rewrite must not break
// ---------------------------------------------------------------------------

// TestScopeTerminalsPreserveSingleAssignmentForm pins that block-splitting leaves the graph in the
// state every pass built on it assumes.
//
// Splitting invalidates block order, predecessor edges, evaluation order, and phi operand keys. The
// first three are re-established by `Finalize`; the fourth is repaired by this pass, and getting it
// wrong is silent -- a phi reading from a block that no longer flows there produces wrong values and
// no crash.
func TestScopeTerminalsPreserveSingleAssignmentForm(t *testing.T) {
	t.Parallel()

	function, _, _, result := terminalsFor(t, `
		function Component(props) {
			let total = 0;
			for (const item of props.items) {
				total = total + item;
			}
			const wrapper = {};
			wrapper.total = total;
			return <div>{wrapper}</div>;
		}
	`)
	if result.Built == 0 {
		t.Skip("no terminals built for this source")
	}
	if violations := VerifySSA(function); len(violations) != 0 {
		t.Errorf("the rewrite left %d single-assignment violations: %v", len(violations), violations)
	}

	byId := map[static_single_assignment.BlockId]*BasicBlock{}
	for _, block := range function.Blocks {
		if byId[block.Id] != nil {
			t.Fatalf("duplicate block id %d after the rewrite", block.Id)
		}
		byId[block.Id] = block
	}
	for _, block := range function.Blocks {
		if block.Terminal == nil {
			t.Errorf("block %d has no terminal after the rewrite", block.Id)
			continue
		}
		EachSuccessorAndFallthrough(block.Terminal, func(id static_single_assignment.BlockId) {
			if byId[id] == nil {
				t.Errorf("block %d names block %d, which is not in the graph", block.Id, id)
			}
		})
		for _, phi := range block.Phis {
			for _, predecessor := range PhiOperandsInOrder(phi) {
				if byId[predecessor] == nil {
					t.Errorf("a phi in block %d names predecessor %d, which is gone",
						block.Id, predecessor)
				}
			}
		}
	}
}

// TestScopeEndJumpsWithBreak pins the goto variant on the terminal that closes a scope.
//
// React writes `variant: GotoVariant.Break` (bundle 26326) and oxc writes `GotoVariant::Break`
// (`build_reactive_scope_terminals_hir.rs:140`), independently. `GotoVariant` records WHY a jump
// exists so a later pass reconstructing structure can tell a loop's back edge from an exit, and
// `Continue` on a scope's closing jump would read as the head of a loop that does not exist.
//
// A mutant flipping it to `Continue` survived the first sweep: nothing in the graph's shape changes,
// only the annotation, so single-assignment form, edges, and ordering all stay valid. This is the
// "fixtures assert the wrong layer" category -- every fixture covering these blocks asserted their
// structure and none asserted what the jump claimed to be.
func TestScopeEndJumpsWithBreak(t *testing.T) {
	t.Parallel()

	function, _, _, result := terminalsFor(t, `
		function Component(props) {
			const a = {};
			a.x = props.x;
			return <div>{a}</div>;
		}
	`)
	if result.Built == 0 {
		t.Skip("no terminals built for this source")
	}

	// Every fallthrough named by a Scope terminal must be reached by a Goto with variant Break.
	fallthroughs := map[static_single_assignment.BlockId]bool{}
	for _, terminal := range scopeTerminalsIn(function) {
		fallthroughs[terminal.Fallthrough] = true
	}
	if len(fallthroughs) == 0 {
		t.Fatal("no scope fallthrough to check")
	}
	closing := 0
	for _, block := range function.Blocks {
		jump, isGoto := block.Terminal.(*Goto)
		if !isGoto || !fallthroughs[jump.Block] {
			continue
		}
		closing++
		if jump.Variant != GotoVariantBreak {
			t.Errorf("the jump closing a scope has variant %v, want GotoVariantBreak: a scope exit "+
				"is not a loop back edge", jump.Variant)
		}
	}
	if closing == 0 {
		t.Error("no Goto reaches a scope fallthrough, so this test asserts nothing")
	}
}

// TestScopeTerminalsRenumberTheGraph pins that the rewrite re-establishes ALL THREE invariants
// `Finalize` maintains, not just the predecessor edges.
//
// Splitting blocks appends new ones to the end of `Function.Blocks` with no evaluation order at all.
// `Finalize` is `ReversePostorder` then `MarkPredecessors` then `MarkEvaluationOrder`, and a mutant
// replacing it with `MarkPredecessors` alone SURVIVED the first sweep: the graph still verified in
// single-assignment form and every edge still resolved, because predecessors were the only thing
// those assertions read.
//
// The distinguishing output, measured: with `Finalize` the corpus source below yields zero
// non-monotone order steps; with the mutant it yields six, and the new blocks carry terminal order
// ZERO -- which `MarkEvaluationOrder` documents as meaning the finalizer never reached them. Any
// pass reading evaluation order after this one would silently compare against zero.
func TestScopeTerminalsRenumberTheGraph(t *testing.T) {
	t.Parallel()

	function, _, _, result := terminalsFor(t, `
		function Component(props) {
			const a = {};
			a.x = props.x;
			const b = {};
			b.y = props.y;
			return <div>{a}{b}</div>;
		}
	`)
	if result.Built == 0 {
		t.Skip("no terminals built for this source")
	}

	// Every instruction and terminal must carry a real position, and positions must increase across
	// the block slice, which is what reverse postorder plus renumbering guarantees together.
	var previous static_single_assignment.EvaluationOrder
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			order := function.Instructions[instructionId].Order
			if order == 0 {
				t.Errorf("an instruction in block %d has evaluation order zero, which means the "+
					"finalizer never reached it", block.Id)
			}
			if order <= previous {
				t.Errorf("evaluation order went from %d to %d across the block slice; the blocks "+
					"are not in reverse postorder or were not renumbered", previous, order)
			}
			previous = order
		}
		order := TerminalOrder(block.Terminal)
		if order == 0 {
			t.Errorf("block %d's terminal has evaluation order zero after the rewrite", block.Id)
		}
		if order <= previous {
			t.Errorf("terminal order went from %d to %d across the block slice", previous, order)
		}
		previous = order
	}
}

// TestScopeTerminalsGapsAreDeclared pins the gap list, so closing one is a visible event.
func TestScopeTerminalsGapsAreDeclared(t *testing.T) {
	t.Parallel()

	gaps := ScopeTerminalsGaps()
	if len(gaps) != 1 || gaps[0] != ScopeTerminalsGapPrunedScope {
		t.Errorf("gaps are %v, want exactly [ScopeTerminalsGapPrunedScope]", gaps)
	}
}

// TestPrunedScopeIsNotDeclared pins the decision NOT to add the second variant.
//
// `PrunedScope` is constructed only by `pruneUnusedScopes`, `flattenReactiveLoopsHIR`,
// `flattenScopesWithHooksOrUseHIR` and `pruneAlwaysInvalidatingScopes`. All four are ported and
// record the decision as `ReactiveScopeBlock.Pruned` instead, so adding the variant would create the
// exact `Optional` shape this package already carries as a warning.
func TestPrunedScopeIsNotDeclared(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("terminal.go")
	if err != nil {
		t.Fatalf("reading terminal.go: %v", err)
	}
	if strings.Contains(string(contents), "type PrunedScope struct") {
		t.Error("PrunedScope is declared but nothing in this tree constructs it, which is the " +
			"Optional failure this package already carries once")
	}
}

// ---------------------------------------------------------------------------
// The corpus: the measurement this stage exists to make
// ---------------------------------------------------------------------------

// TestScopeTerminalsRecoverEveryScopeFromTheGraph is the headline.
//
// `propagate_scope_dependencies_hir` does not read the scope side table. It recovers scopes by
// walking blocks for a Scope terminal -- React's `keyByScopeId` is six lines of
// `if (block.terminal.kind === 'scope')`. Before this pass, that walk found ZERO scopes on every
// function in the corpus, so a dependency pass ported onto this graph would have produced empty
// dependencies by construction while looking correct.
//
// This asserts the walk now recovers every scope, with the merged scope count as its own control.
// The control is what makes the number meaningful: an earlier version of this pass built 3,286
// terminals against 2,874 scopes, and the excess was 412 duplicates. A bare non-zero count could not
// have seen that.
func TestScopeTerminalsRecoverEveryScopeFromTheGraph(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)

	var files []string
	err := filepath.Walk(corpusRoot(t), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the corpus: %v", err)
	}
	sort.Strings(files)
	if len(files) > 400 {
		files = files[:400]
	}
	if len(files) < 10 {
		t.Fatalf("the corpus holds only %d files; the path is probably wrong", len(files))
	}

	functions, mergedTotal, recovered, built := 0, 0, 0, 0
	declined, ssaViolations, emptyBodies := 0, 0, 0

	for _, path := range files {
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		fileName := "/corpus/" + filepath.Base(path)
		probe := rule.Rule{
			Name:             "scope-terminals-corpus",
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
							functions++

							ranges := InferMutableRanges(function)
							set := FindDisjointMutableValuesWithRanges(function, ranges)
							scopes := AssignReactiveScopesWithSets(function, ranges, set)
							aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
							identity := MergedScopeIdentity{Aligned: aligned, Merged: merged}
							mergedTotal += merged.Len()

							if ScopeTerminalsPrecondition(function, scopes, identity) != 0 {
								declined++
								return
							}
							result := BuildReactiveScopeTerminals(function, scopes, identity)
							built += result.Built

							seen := map[ScopeId]bool{}
							byId := map[static_single_assignment.BlockId]*BasicBlock{}
							for _, block := range function.Blocks {
								byId[block.Id] = block
							}
							for _, block := range function.Blocks {
								terminal, ok := block.Terminal.(*Scope)
								if !ok {
									continue
								}
								seen[terminal.Scope] = true
								if body := byId[terminal.Block]; body == nil {
									emptyBodies++
								}
							}
							recovered += len(seen)
							ssaViolations += len(VerifySSA(function))
						})
					},
				}
			},
		}
		rule_testing.RunTypedFiles(t, probe, map[string]string{fileName: string(contents)}, fileName)
	}

	t.Logf("functions                             %d", functions)
	t.Logf("scopes after align-then-merge         %d", mergedTotal)
	t.Logf("Scope terminals built                 %d", built)
	t.Logf("scopes recovered from the graph       %d", recovered)

	if functions < 100 {
		t.Fatalf("only %d functions lowered; the corpus is probably wrong", functions)
	}
	if declined != 0 {
		t.Errorf("%d functions were declined on the nesting precondition, want 0 after alignment "+
			"and merging", declined)
	}
	if recovered == 0 {
		t.Fatal("no scope is recoverable from a CFG terminal, which is the zero this pass exists " +
			"to move")
	}
	if recovered != mergedTotal {
		t.Errorf("recovered %d scopes from the graph but the merged table holds %d; a mismatch "+
			"means the pass built duplicates or dropped scopes", recovered, mergedTotal)
	}
	if built != mergedTotal {
		t.Errorf("built %d terminals for %d merged scopes", built, mergedTotal)
	}
	if ssaViolations != 0 {
		t.Errorf("the rewrite left %d single-assignment violations across the corpus", ssaViolations)
	}
	if emptyBodies != 0 {
		t.Errorf("%d Scope terminals name a body block that is not in the graph", emptyBodies)
	}
}
