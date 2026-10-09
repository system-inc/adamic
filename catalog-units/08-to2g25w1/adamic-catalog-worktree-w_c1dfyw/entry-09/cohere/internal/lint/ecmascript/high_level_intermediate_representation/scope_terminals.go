// Reactive scope terminals: turning the scope side table into control-flow structure.
//
// This is React's `buildReactiveScopeTerminalsHIR` (development bundle line 26224), run over this
// graph. oxc transcribes it at
// `oxc_react_compiler/src/react_compiler_inference/build_reactive_scope_terminals_hir.rs`; where the
// two disagree React wins, and each disagreement is recorded at the line that resolves it.
//
// # What this pass does, and why it is a rewrite rather than an annotation
//
// Every pass before this one records scopes in a SIDE TABLE: `scopes.go` assigns them,
// `align_scopes.go` widens them to block boundaries, `merge_scopes.go` unions the overlapping ones.
// After all three, a scope is still only a pair of evaluation-order positions with a member list.
// Nothing in the control-flow graph knows a scope exists.
//
// The consumer that needs them does not read the table. `propagate_scope_dependencies_hir` recovers
// scopes by WALKING BLOCKS FOR A SCOPE TERMINAL -- React's `keyByScopeId` is six lines of
// `if (block.terminal.kind === 'scope')`, and oxc's entrypoint matches `Terminal::Scope { .. }`
// identically. Measured before this file was written, over 400 corpus files and 677 functions:
//
//	scopes in the side table                     3,286
//	scopes after align-then-merge                2,874
//	scopes reachable from a CFG terminal             0
//
// That zero is the denominator, not a gap to widen later. A dependency pass ported onto a graph with
// no scope terminals would run, terminate, pass its fixtures, and produce zero dependencies BY
// CONSTRUCTION, because its visitors are gated on a non-empty scope stack that nothing ever pushes.
// This pass is what moves that zero.
//
// # The rewrite, which splits blocks
//
// For each scope, two positions matter: `range.start`, where the scope begins, and `range.end`,
// where it ends. At each, the block containing that position is SPLIT, and a terminal is inserted:
//
//	at range.start   a Scope terminal, whose Block is the scope body and whose Fallthrough is
//	                 where control resumes after the scope completes
//	at range.end     a Goto with variant Break, jumping to that same fallthrough
//
// So one block becomes up to three, and the instructions are distributed between them by position.
// A block containing several scope boundaries splits several times; the rewrites are applied in one
// left-to-right sweep per block, each carving off the instructions before it.
//
// # Nesting is a PRECONDITION, and it is enforced upstream by an assertion
//
// `recursivelyTraverseItems` -- which this pass uses to visit scopes in nesting order -- carries a
// hard `CompilerError.invariant(disjoint || nested)`. It is not a filter; upstream stops. React
// calls `assertValidBlockNesting(hir)` on the line IMMEDIATELY BEFORE this pass (bundle 50213/50214)
// and oxc does the same (`pipeline.rs:293/295`), independently, which is upstream telling you the
// precondition rather than leaving it to be discovered.
//
// Three stages exist because of that assertion. Measured on this corpus, scope-versus-scope
// non-nesting was 40 before `merge_scopes.go` and 0 after; the FULL assertion, which also compares
// every block's fallthrough subtree, was 249 before `align_scopes.go` and 8 after on the current
// live corpus. All eight are scope-involving residues in plain helper functions; components and
// hooks are at zero. The former block-against-block residue disappeared when reverse postorder began
// matching React's traversal. `ScopeTerminalsPrecondition` reports whether a function satisfies the
// scope-nesting portion, so a caller declines rather than producing a corrupt graph.
//
// # Termination: two sweeps over finite slices, no fixpoint
//
// There is no worklist and no iteration to convergence. `recursivelyTraverseItems` sorts the scopes
// once and walks them once, maintaining a stack that only shrinks when an item is disjoint, so each
// scope is entered exactly once and exited exactly once -- 2n rewrites for n scopes. The block sweep
// then visits each block once and, within it, each instruction position once, consuming rewrites
// from a queue that only shrinks. Work is linear in blocks plus instructions plus scopes, and every
// loop is over a slice whose length is fixed before it begins.
//
// This is deliberately a weaker claim than a fixpoint argument, because there is no fixpoint here to
// argue about. `TestScopeTerminalsAreASinglePass` pins it empirically by counting visits.
//
// # What must re-run afterwards, and why getting it wrong is silent
//
// Splitting blocks invalidates three things the graph maintains: block order, predecessor edges, and
// evaluation order. React re-runs `reversePostorderBlocks`, `markPredecessors`, and
// `markInstructionIds`; `Finalize` here is exactly that trio in that order, and this pass calls it.
//
// Phi operands are the fourth, and they are the one a reader does not expect. A phi names its
// operands BY PREDECESSOR BLOCK ID. When a block is split, the predecessor that reaches a successor
// is no longer the original block but the LAST block carved out of it, so every phi operand keyed by
// the original id has to be rekeyed to the final one. React does this explicitly after swapping the
// block table, and omitting it produces a phi reading from a block that no longer flows there --
// wrong values, no crash, and nothing in a well-formedness check that would notice.
//
// `Construct` is NOT re-run. It is not idempotent: running it twice on the same function moved 1 phi
// to 2 while `VerifySSA` reported zero violations. This pass therefore runs AFTER single-assignment
// construction and repairs the phis in place rather than rebuilding them, which is also what
// upstream does.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"sort"
)

// ScopeIdentity is how this pass resolves a value's scope to the scope that actually survives.
//
// # Why this exists, and the defect it was written to fix
//
// A value's scope id comes from `ReactiveScopes.ScopeOf`, which is the PRE-MERGE table.
// `merge_scopes.go` then unions overlapping scopes, so several pre-merge ids can name one surviving
// scope, and `align_scopes.go` widens the ranges under both. A pass that keys on the pre-merge id
// while reading the merged RANGE gets a coherent-looking wrong answer: every scope has a plausible
// range, so nothing is empty, nothing is inverted, and the nesting precondition still passes.
//
// Measured on this corpus, that mistake builds 3,286 terminals for 2,874 scopes -- 412 duplicates,
// each a second terminal wrapping a range another terminal already owns. `propagate_scope_dependencies`
// keys its output by scope id, so the duplicates would collide and the surviving entry would depend
// on map order. The tell was that the built count EXCEEDED the control, which is why the control is
// measured alongside the headline rather than after it.
//
// Threading the resolver rather than taking `*MergedScopes` directly keeps this pass testable
// against a hand-built table and lets a caller that has not merged pass `UnmergedScopes`.
type ScopeIdentity interface {
	// GroupOf maps a pre-merge scope id to the id of the scope that survives, or zero for a value
	// with no scope. An unmerged scope maps to itself.
	GroupOf(scope ScopeId) ScopeId
	// RangeOf gives the surviving scope's final range, after alignment and merging.
	RangeOf(scope ScopeId) mutation_aliasing.MutableRange
}

// MergedScopeIdentity resolves scope identity through an align-then-merge result.
//
// This is what a caller holding the output of `AlignThenMergeReactiveScopes` passes.
type MergedScopeIdentity struct {
	Aligned *AlignedScopes
	Merged  *MergedScopes
}

// GroupOf returns the merged group a pre-merge scope belongs to, or the scope itself when the merge
// left it alone.
func (m MergedScopeIdentity) GroupOf(scope ScopeId) ScopeId {
	if scope == 0 {
		return 0
	}
	if m.Merged != nil {
		if group := m.Merged.GroupOf(scope); group != 0 {
			return group
		}
	}
	return scope
}

// RangeOf returns the surviving scope's range, preferring the merged table and falling back to the
// aligned one for a scope the merge did not touch.
func (m MergedScopeIdentity) RangeOf(scope ScopeId) mutation_aliasing.MutableRange {
	if m.Merged != nil {
		if bounds := m.Merged.RangeOf(scope); bounds.Start != bounds.End {
			return bounds
		}
	}
	if m.Aligned != nil {
		return m.Aligned.RangeOf(scope)
	}
	return mutation_aliasing.MutableRange{}
}

// ScopeTerminalsGap names a rule this pass cannot apply, for a caller that needs to know.
//
// Modelled on `RangeGap`, `DisjointGap`, `ScopeGap`, `MergeGap` and `AlignGap` deliberately, and for
// the same reason: a gap answerable through the API is one a consumer can decline on, while a gap
// living only in a comment is one the next reader inherits by accident.
type ScopeTerminalsGap uint8

const (
	// ScopeTerminalsGapPrunedScope is upstream's second reactive-scope terminal, not built here.
	//
	// `PrunedScope` is constructed by four passes -- `pruneUnusedScopes` (bundle 41515),
	// `flattenReactiveLoopsHIR` (43404), `flattenScopesWithHooksOrUseHIR` (43487), and
	// `pruneAlwaysInvalidatingScopes` (43563). All four are ported, and none of them builds the
	// variant: the two that run over this graph hand the builder the ids they pruned, the two that
	// run over the reactive tree set `ReactiveScopeBlock.Pruned`, and the label case of the hook
	// flatten rewrites to the `Label` this set already has. Adding the variant would put a terminal
	// in the set that nothing constructs, which is strictly worse than leaving it out: a consumer
	// writes an arm for it, the arm never fires, and the dead branch reads as coverage. `Optional` is
	// already that shape here and is the standing warning.
	ScopeTerminalsGapPrunedScope ScopeTerminalsGap = iota
)

// ScopeTerminalsGaps are the rules BuildReactiveScopeTerminals does not apply.
//
// Returned as a value rather than documented alone so a test can assert on it, which makes closing a
// gap a visible event rather than a silent improvement.
func ScopeTerminalsGaps() []ScopeTerminalsGap {
	return []ScopeTerminalsGap{ScopeTerminalsGapPrunedScope}
}

// ScopeTerminals reports what the rewrite did, for measurement and for a caller that needs to know
// whether anything changed.
//
// The zero value is the correct answer for a function with no scopes: nothing built, nothing split.
type ScopeTerminals struct {
	// Built is how many Scope terminals were inserted, which equals the number of scopes rewritten.
	Built int
	// BlocksSplit is how many original blocks were divided into two or more.
	BlocksSplit int
	// PhiOperandsRekeyed is how many phi operands named a block that split and were repointed at the
	// final block carved out of it.
	PhiOperandsRekeyed int
}

// scopeRewrite is one queued edit: begin a scope here, or end one here.
//
// Upstream models this as a two-variant discriminated union (`TerminalRewriteInfo` in oxc,
// `{kind: 'StartScope' | 'EndScope'}` in React). One struct with a boolean is the same information
// and avoids an interface allocation per rewrite in a linter that runs this per function.
type scopeRewrite struct {
	isStart bool
	// order is the evaluation position this rewrite lands at: the scope's range start or end.
	order static_single_assignment.EvaluationOrder
	// block is the scope body's block id. Meaningful only for a start.
	block static_single_assignment.BlockId
	// fallthrough is where control resumes after the scope. Shared by the start and its end.
	fallthroughBlock static_single_assignment.BlockId
	// scope is the scope being begun. Meaningful only for a start.
	scope ScopeId
}

// ScopeTerminalsPrecondition reports whether function's scopes satisfy the nesting invariant this
// pass requires, returning the number of violations.
//
// Zero means the rewrite is safe. A non-zero result means `recursivelyTraverseItems` would hit its
// invariant upstream, and this pass declines rather than building a graph whose scope structure
// contradicts its block structure. See the package comment for why the three passes upstream of this
// one exist and what each moved.
//
// This compares scopes to each other only. Upstream's `assertValidBlockNesting` additionally folds
// in every block's fallthrough subtree; that check belongs with the pass that aligns them and lives
// in `align_scopes.go`.
func ScopeTerminalsPrecondition(function *Function, scopes *ReactiveScopes,
	identity ScopeIdentity) int {
	items := scopeItemsInNestingOrder(function, scopes, identity)
	violations := 0
	var active []scopeItem
	for _, curr := range items {
		for index := len(active) - 1; index >= 0; index-- {
			parent := active[index]
			disjoint := curr.start >= parent.end
			nested := curr.end <= parent.end
			if !disjoint && !nested {
				violations++
			}
			if disjoint {
				active = active[:index]
			} else {
				break
			}
		}
		active = append(active, curr)
	}
	return violations
}

// scopeItem is one scope reduced to the two positions the traversal orders it by.
type scopeItem struct {
	scope      ScopeId
	start, end static_single_assignment.EvaluationOrder
}

// scopeItemsInNestingOrder is React's `getScopes` followed by its `rangePreOrderComparator`.
//
// `getScopes` collects the scope of every place reachable from an instruction lvalue, an instruction
// operand, or a terminal operand, and drops any whose range is empty. That reachability filter is
// not incidental: a scope no place mentions has no values to memoize, and upstream would emit a
// terminal around nothing. Measured on this corpus all 2,874 post-merge scopes are reachable from a
// place and none has an empty range, so the filter removes nothing today -- which is a fact about
// this corpus and not a reason to drop the filter, since a scope whose members are all eliminated by
// a future pass would produce exactly the shape it exists to catch.
//
// The comparator is start ASCENDING then end DESCENDING, which puts a parent immediately before the
// children it contains and is what makes the single-stack traversal correct.
//
// # The end-descending clause is INERT on this corpus, and is kept anyway
//
// That second clause only decides when two scopes share a start position. Measured over 400 files
// and 677 functions, ZERO of the 2,874 post-merge scopes share a start with another, and zero share
// one in the pre-merge table either. The control alongside that zero is shared END positions, which
// is 1, so the measurement can see a collision when one exists.
//
// A mutant flipping the clause to ascending therefore survives the sweep, and it is neither a
// fixture blind spot nor an equivalent mutation: the branch is reachable in the code and
// unreachable through this input. It is kept rather than deleted because it is upstream's
// `rangePreOrderComparator` verbatim, because merging is what removes shared starts and a future
// pass that widens scopes could reintroduce them, and because a parent sorted AFTER its child would
// invert the nesting the traversal depends on -- a silent corruption rather than a crash.
//
// The verdict names its inputs: it holds for scopes produced by `AlignThenMergeReactiveScopes` on
// this corpus. It EXPIRES if a pass lands that can give two scopes the same start.
func scopeItemsInNestingOrder(function *Function, scopes *ReactiveScopes,
	identity ScopeIdentity) []scopeItem {
	if function == nil || scopes == nil {
		return nil
	}
	seen := map[ScopeId]bool{}
	var items []scopeItem
	visit := func(place Place, _ PlaceRole) {
		scope := identity.GroupOf(scopes.ScopeOf(place.Identifier))
		if scope == 0 {
			return
		}
		if seen[scope] {
			return
		}
		bounds := identity.RangeOf(scope)
		if bounds.Start == bounds.End {
			return
		}
		seen[scope] = true
		items = append(items, scopeItem{scope: scope, start: bounds.Start, end: bounds.End})
	}
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			EachInstructionPlace(function.Instructions[instructionId], visit)
		}
		EachTerminalPlace(block.Terminal, visit)
	}
	sort.SliceStable(items, func(a, b int) bool {
		if items[a].start != items[b].start {
			return items[a].start < items[b].start
		}
		return items[a].end > items[b].end
	})
	return items
}

// BuildReactiveScopeTerminals rewrites function's control-flow graph so every reactive scope is
// carried by a Scope terminal rather than only by the side table.
//
// Requires scopes that have been aligned and merged: pass the output of
// `AlignThenMergeReactiveScopes`, and read ranges through `rangeOf` so this reads the FINAL range
// rather than the pre-alignment one. Declines and returns a zero result when the nesting
// precondition is violated; see `ScopeTerminalsPrecondition`.
//
// Mutates function in place and re-establishes every invariant `Finalize` maintains, plus the phi
// operand keys that block-splitting invalidates. See the package comment.
func BuildReactiveScopeTerminals(function *Function, scopes *ReactiveScopes,
	identity ScopeIdentity) ScopeTerminals {
	if function == nil || scopes == nil || scopes.Len() == 0 {
		return ScopeTerminals{}
	}
	items := scopeItemsInNestingOrder(function, scopes, identity)
	if len(items) == 0 {
		return ScopeTerminals{}
	}
	if ScopeTerminalsPrecondition(function, scopes, identity) != 0 {
		return ScopeTerminals{}
	}

	rewrites := queueScopeRewrites(function, items)
	if len(rewrites) == 0 {
		return ScopeTerminals{}
	}
	result := applyScopeRewrites(function, rewrites)

	// Step 5, and it is the fifth thing a split invalidates. See `fixScopeAndIdentifierRanges`.
	fixScopeRanges(function, scopes, identity)
	return result
}

// fixScopeRanges restates every scope range in the numbering the split produced.
//
// This is React's `fixScopeAndIdentifierRanges` (`HIR/HIRBuilder.ts:940-955`), called as its Step 5
// with the comment that says why: "the renumbering instructions invalidates scope and identifier
// ranges, so we fix them in the next step."
//
// # The range is read off the terminal, not recomputed from members
//
// A scope's range starts at its own `scope` terminal and ends at the first instruction of its
// fallthrough block. That is exact by construction, because this pass just built those terminals
// from the pre-renumber range: the terminal marks where the scope opens and the fallthrough marks
// where it closes, and both carry their new numbering already.
//
// Three other mappings were built and measured wrong before this one was read at the source, and
// each failed for a reason worth keeping.
//
// Taking the span from the minimum to the maximum member order WIDENS a merged scope, because a
// merge folds scattered scopes into one survivor and the span then covers everything between. It
// produced a nesting-containment violation and moved the dependency production count from 151 to
// 158.
//
// Translating each endpoint through a member whose old order matches it cannot work, and the
// measurement says why: over 2,946 merged scopes only 2,728 starts and 1,759 ends coincide with any
// member's order at all.
//
// Interpolating an endpoint over the old-to-new map is sound -- that map is strictly monotonic, zero
// inversions over 42,500 pairs, since a split only inserts terminals -- and still wrong in effect:
// it moved the hoistable analysis from 171 deep to 238, which is the over-approximating direction
// that test's own comment warns about.
//
// The terminal already knows the answer. None of the three had to be invented.
//
// # Scale of what this repairs
//
// Measured over 677 corpus functions and 3,357 scopes, counting scopes whose range contains the
// order of none of their own members: 0 before the renumber, 2,345 after. The zero is the control.
func fixScopeRanges(function *Function, scopes *ReactiveScopes, identity ScopeIdentity) {
	if function == nil {
		return
	}
	updated := map[ScopeId]mutation_aliasing.MutableRange{}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		terminal, isScope := block.Terminal.(*Scope)
		if !isScope {
			continue
		}
		fallthroughBlock, found := function.Block(terminal.Fallthrough)
		if !found || fallthroughBlock == nil {
			continue
		}
		end := TerminalOrder(fallthroughBlock.Terminal)
		if len(fallthroughBlock.Instructions) > 0 {
			if first := function.Instructions[fallthroughBlock.Instructions[0]]; first != nil {
				end = first.Order
			}
		}
		updated[terminal.Scope] = mutation_aliasing.MutableRange{Start: terminal.Order, End: end}
	}
	if len(updated) == 0 {
		return
	}

	if scopes != nil {
		for scope, bounds := range updated {
			if _, present := scopes.ranges[scope]; present {
				scopes.ranges[scope] = bounds
			}
		}
	}
	// The merged table is what every consumer reads, through `MergedScopeIdentity`.
	if merged, ok := identity.(MergedScopeIdentity); ok && merged.Merged != nil {
		for scope, bounds := range updated {
			if _, present := merged.Merged.ranges[scope]; present {
				merged.Merged.ranges[scope] = bounds
			}
		}
	}
}

// queueScopeRewrites is React's `recursivelyTraverseItems` with `pushStartScopeTerminal` and
// `pushEndScopeTerminal` as its enter and exit callbacks.
//
// The traversal maintains a stack of active scopes. A scope is entered when reached and exited when
// a later scope proves it finished, which is the moment a candidate's start is at or past the
// active scope's end. Everything still active at the walk's end is exited in reverse.
//
// Blocks for the body and the fallthrough are minted at ENTER, so a nested scope's blocks are minted
// after its parent's, which keeps ids ascending with nesting depth. The end rewrite reuses the
// fallthrough its start recorded -- that pairing is the whole reason the stack exists.
func queueScopeRewrites(function *Function, items []scopeItem) []scopeRewrite {
	var rewrites []scopeRewrite
	fallthroughs := map[ScopeId]static_single_assignment.BlockId{}

	enter := func(item scopeItem) {
		body := function.NewBlock(BlockKindBlock)
		after := function.NewBlock(BlockKindBlock)
		fallthroughs[item.scope] = after.Id
		rewrites = append(rewrites, scopeRewrite{
			isStart:          true,
			order:            item.start,
			block:            body.Id,
			fallthroughBlock: after.Id,
			scope:            item.scope,
		})
	}
	exit := func(item scopeItem) {
		after, ok := fallthroughs[item.scope]
		if !ok {
			// Upstream raises `CompilerError.invariant('Expected scope to exist')` here. A linter
			// declines instead: the traversal below only ever exits a scope it entered, so this is
			// unreachable today, and the enumeration is every call site of `exit` in this function.
			return
		}
		rewrites = append(rewrites, scopeRewrite{
			isStart:          false,
			order:            item.end,
			fallthroughBlock: after,
		})
	}

	var active []scopeItem
	for _, curr := range items {
		for index := len(active) - 1; index >= 0; index-- {
			if curr.start >= active[index].end {
				exit(active[index])
				active = active[:index]
				continue
			}
			break
		}
		enter(curr)
		active = append(active, curr)
	}
	for index := len(active) - 1; index >= 0; index-- {
		exit(active[index])
	}

	// Reverse, so the block sweep can consume from the END of the slice. This is upstream's
	// `queuedRewrites.reverse()` followed by its `.at(-1)` loop, and it is deliberately NOT a sort.
	//
	// An earlier draft here sorted by position with a start-versus-end tiebreak, which looked like a
	// safe normalisation and is a real divergence. The traversal ALREADY emits rewrites in the order
	// the sweep needs, because the stack discipline that produces them is the same one that
	// establishes the nesting: a scope is exited before any scope disjoint from it is entered, and a
	// nested scope is entered while its parent is still open. Position alone cannot reconstruct
	// that. Two scopes meeting at a point -- one ending at 7, the next starting at 7 -- are DISJOINT,
	// so the traversal emits the end before the start with no tiebreak involved, and a comparator
	// choosing between them is deciding something the traversal had already decided correctly.
	//
	// Found by writing a fixture for a surviving mutant, watching the fixture fail against the
	// unmutated code, and re-reading upstream rather than adjusting the fixture. The sort was the
	// defect; the mutants that survived were mutating code that should not have existed.
	for left, right := 0, len(rewrites)-1; left < right; left, right = left+1, right-1 {
		rewrites[left], rewrites[right] = rewrites[right], rewrites[left]
	}
	return rewrites
}

// applyScopeRewrites performs the block splitting and rebuilds the graph.
//
// One sweep over the original blocks. Within each, the instruction positions are walked in order and
// every rewrite landing at or before the current position is applied, carving off the instructions
// seen so far into a new block terminated by the rewrite's terminal. Whatever remains becomes the
// final block, which keeps the original block's own terminal.
func applyScopeRewrites(function *Function, rewrites []scopeRewrite) ScopeTerminals {
	result := ScopeTerminals{}
	original := function.Blocks
	// The blocks minted by `queueScopeRewrites` were appended to Function.Blocks; the sweep below
	// rebuilds the slice from scratch, so take the originals only.
	originalCount := 0
	for _, block := range original {
		if len(block.Instructions) > 0 || block.Terminal != nil {
			originalCount++
		}
	}

	var rebuilt []*BasicBlock
	// finalBlockOf maps an original block id to the id of the last block carved out of it, which is
	// the block that now reaches whatever the original reached. Phi operands are rekeyed through it.
	finalBlockOf := map[static_single_assignment.BlockId]static_single_assignment.BlockId{}

	for _, block := range original {
		if block.Terminal == nil {
			// A block minted for a scope body or fallthrough, not yet filled. Emitted by the sweep
			// below when its contents are known.
			continue
		}

		nextBlockId := block.Id
		sliceStart := 0
		var carved []*BasicBlock

		for position := 0; position <= len(block.Instructions); position++ {
			var order static_single_assignment.EvaluationOrder
			if position < len(block.Instructions) {
				order = function.Instructions[block.Instructions[position]].Order
			} else {
				order = TerminalOrder(block.Terminal)
			}
			for len(rewrites) > 0 && rewrites[len(rewrites)-1].order <= order {
				rewrite := rewrites[len(rewrites)-1]
				rewrites = rewrites[:len(rewrites)-1]

				var terminal Terminal
				if rewrite.isStart {
					terminal = &Scope{
						Scope:       rewrite.scope,
						Block:       rewrite.block,
						Fallthrough: rewrite.fallthroughBlock,
					}
				} else {
					terminal = &Goto{Block: rewrite.fallthroughBlock, Variant: GotoVariantBreak}
				}
				carved = append(carved, &BasicBlock{
					Id:           nextBlockId,
					Kind:         block.Kind,
					Instructions: append([]InstructionId{}, block.Instructions[sliceStart:position]...),
					Terminal:     terminal,
					Phis:         phisFor(block, len(carved) == 0),
				})
				sliceStart = position
				if rewrite.isStart {
					nextBlockId = rewrite.block
				} else {
					nextBlockId = rewrite.fallthroughBlock
				}
			}
		}

		if len(carved) == 0 {
			rebuilt = append(rebuilt, block)
			continue
		}

		final := &BasicBlock{
			Id:           nextBlockId,
			Kind:         block.Kind,
			Instructions: append([]InstructionId{}, block.Instructions[sliceStart:]...),
			Terminal:     block.Terminal,
			Phis:         nil,
		}
		carved = append(carved, final)
		rebuilt = append(rebuilt, carved...)
		finalBlockOf[block.Id] = final.Id
		result.BlocksSplit++
		for _, candidate := range carved {
			if _, isScope := candidate.Terminal.(*Scope); isScope {
				result.Built++
			}
		}
	}
	_ = originalCount

	function.Blocks = rebuilt
	// `blocksById` maps an id to a POINTER, and the carved blocks are new pointers reusing the
	// original ids. Repoint every entry before Finalize, which resolves successors through this
	// index: a stale pointer would leave the walk reading the pre-split block and silently rebuild
	// the graph it was meant to replace.
	function.reindexBlocks()

	// Phi operands name predecessors by block id. A split block no longer reaches what it reached;
	// the last block carved out of it does. Rekey before Finalize, because MarkPredecessors reads
	// the graph and not the phis, so a stale operand would survive it silently.
	for _, block := range function.Blocks {
		for _, phi := range block.Phis {
			for _, predecessor := range PhiOperandsInOrder(phi) {
				replacement, split := finalBlockOf[predecessor]
				if !split || replacement == predecessor {
					continue
				}
				value := phi.Operands.At(predecessor)
				phi.Operands.Delete(predecessor)
				phi.Operands.Set(replacement, value)
				result.PhiOperandsRekeyed++
			}
		}
	}

	Finalize(function)
	return result
}

// reindexBlocks rebuilds the id-to-block index from Function.Blocks.
//
// Lives here rather than in high_level_intermediate_representation.go because this is the only pass that replaces the block slice
// wholesale; lowering appends through `NewBlock`, which maintains the index as it goes.
func (f *Function) reindexBlocks() {
	f.blocksById = make(map[static_single_assignment.BlockId]*BasicBlock, len(f.Blocks))
	for _, block := range f.Blocks {
		f.blocksById[block.Id] = block
	}
}

// phisFor gives the first block carved out of an original its phis and every later one none.
//
// A phi belongs at the point control MERGES, which is the top of the original block. The blocks
// carved after it have exactly one predecessor -- the carve before them -- so a phi there would name
// a single operand and mean nothing. Upstream does the same with
// `phis: context.rewrites.length === 0 ? context.source.phis : new Set()`.
func phisFor(block *BasicBlock, isFirst bool) []*Phi {
	if !isFirst {
		return nil
	}
	return block.Phis
}
