// Splicing an immediately-invoked function expression into its caller.
//
// This is upstream's `inlineImmediatelyInvokedFunctionExpressions`
// (`Inference/InlineImmediatelyInvokedFunctionExpressions.ts`, reached from `Pipeline.ts:173`), the
// one pass in React's memoization pipeline this tree did not have. `inline_remap.go` is the half
// upstream never had to write -- ids there are minted from one environment counter, ids here are
// per-function and overlap -- and this file is the half that is upstream's, now that the copy is
// safe.
//
// # What the absence cost, stated as the chain rather than as a symptom
//
// An IIFE is a function defined and called in one place. Upstream splices its body into the caller
// before any range inference runs, so by the time mutable ranges are computed there is no nested
// function and no capture edge.
//
// Keeping the closure is not neutral. `InferMutableRanges` cannot see through a call into a body it
// does not hold, so it is conservative: a mutation inside the closure reaches back along the capture
// edge and widens the captured value's mutable range, and the widening propagates through every
// alias of that value. Scopes upstream keeps separate fuse into one.
//
// A fused scope reads exactly like a broken memo. The rule compares inferred memoization against
// the `useMemo` the programmer wrote, and when one scope spans what the programmer wrote as three,
// the dependencies disagree for all three -- so the rule fires three times on correct code. That is
// why this pass has to run BEFORE `InferMutableRanges` and not merely somewhere in the pipeline;
// running it afterwards inlines a body whose damage has already been recorded.
package high_level_intermediate_representation

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/static_single_assignment"
)

// InlineImmediatelyInvokedFunctionExpressions splices every inlinable IIFE in function into its
// caller, and reports how many it spliced.
//
// # What counts as inlinable, and every condition is upstream's
//
// A `CallExpression` qualifies when its callee is a `FunctionExpression` produced by an earlier
// instruction in a statement block, the call passes no arguments, and the nested function takes no
// parameters and is neither async nor a generator. Each condition is a case the splice has no
// answer for rather than a case it declines out of caution: an argument would need binding to a
// parameter this pass does not lower, and an async body's `await` points would need the caller's
// continuation to become a suspension point.
//
// The callee must also be a TEMPORARY -- an unnamed value. A named binding can be read again later,
// and upstream's `functions` map is keyed on that: a `FunctionExpression` stored into a name is
// never entered, and any use of a function-valued temporary that is not the callee of a call
// removes it from the map. That last part is the one a reader skips, and it is what keeps
// `const f = () => {}; g(f); f();` from inlining a function `g` may also have called.
//
// # The shape of the splice
//
// Upstream's diagram, which this reproduces:
//
//	bb0:                            bb0:
//	  t0 = Function fn0               DeclareLocal let result
//	  t1 = Call t0()          =>      Label block=entry fallthrough=continuation
//	  ...rest                       entry..exit:  (the copied body)
//	                                  result = <returned value>
//	                                  Goto continuation
//	                                continuation:
//	                                  ...rest
//
// The caller's block is cut at the call. Everything after it moves to a fresh continuation block,
// which also inherits the original terminal, so the code following the call keeps its control flow.
// The copied body sits between the two, and every `Return` inside it becomes a store to the call's
// lvalue plus a `Goto` to the continuation.
//
// The `Label` is what makes several returns expressible: it names the body as one region with a
// fallthrough, which is the structure a later reconstruction reads to print a labeled block. A body
// with exactly one exit and that exit a `Return` needs no label, so it gets a plain `Goto` and the
// return becomes a `LoadLocal` into the call's lvalue -- upstream's fast path, kept because the
// label it avoids would otherwise appear in every single-return IIFE, which is most of them.
//
// # Ordering
//
// Must run before `InferMutableRanges`. See the package comment above for the chain.
//
// Returns the number of call sites spliced, so a caller can skip the graph rebuild when nothing
// changed and a test can assert the pass did something rather than passing vacuously.
func InlineImmediatelyInvokedFunctionExpressions(function *Function) int {
	return inlineInvokedFunctions(function, false)
}

// InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks is the same pass with the memo
// exclusion lifted, which is what upstream actually does.
//
// # Why there are still two entry points
//
// `DropManualMemoization` rewrites `useMemo(callback, deps)` into a zero-argument call of the
// callback. The inclusive entry point splices that call just like upstream; production analysis and
// every structural oracle use it. The guarded entry point remains useful for callers and focused
// tests that ask specifically about programmer-written IIFEs while retaining memo markers.
//
// The distinction is explicit rather than inferred from call shape. A rewritten memo callback is
// structurally an IIFE, so only `StartMemoize`/`FinishMemoize` metadata can tell a caller which kind
// it is looking at. `memoizedResults` supplies that metadata to the guarded wrapper.
//
// # The production precondition
//
// Run this after `DropManualMemoization`, then run `MergeConsecutiveBlocks` when it reports a change.
// The splice cuts a caller block around the copied body; the merge restores straight-line regions
// before mutable ranges and reactive scopes read block structure. `AnalyzePreservedManualMemoization`
// is the canonical ordering.
//
// Earlier versions prohibited scope-producing callers because the optional CFG, marker effects,
// alias remapping, and immutable-kind propagation needed after the splice were incomplete. That
// prohibition described a measured failure, but not a property of inlining: after those layers
// landed, the production-equivalent scope oracle holds at 0 under-produced fixtures / 0 scopes,
// with 26 of 48 fixtures exact, and preserve-manual-memoization reaches 28/28 goldens and 0/69 false
// positives. The call-site census in `inline_iife_test.go` keeps new pipelines reviewed against
// those invariants.
//
// This wrapper deliberately does not perform the merge itself. Some callers need to inspect the
// splice count, and keeping graph normalization explicit matches upstream's adjacent pass ordering.
func InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks(function *Function) int {
	return inlineInvokedFunctions(function, true)
}

// inlineInvokedFunctions is the pass itself. `includeMemoCallbacks` lifts the `memoizedResults`
// exclusion; see the exported wrapper above for why that is the caller's decision.
func inlineInvokedFunctions(function *Function, includeMemoCallbacks bool) int {
	if function == nil {
		return 0
	}

	// The values that `DropManualMemoization` marked as the result of a memo call. See
	// `memoizedResults` for why they are excluded, and the wrapper above for why a caller may ask
	// for them anyway.
	memoized := map[static_single_assignment.IdentifierId]bool{}
	if !includeMemoCallbacks {
		memoized = memoizedResults(function)
	}

	inlined := 0
	// The callee temporaries of the calls that were spliced. Their defining
	// `FunctionExpression` instructions are dead afterwards and are dropped below.
	spliced := map[static_single_assignment.IdentifierId]bool{}
	// Copy the block list first. The splice appends blocks to `function.Blocks` as it runs, and a
	// range over the live slice would walk into blocks the splice just created. The continuation
	// block is pushed back on deliberately, because a second IIFE later in the same statement is
	// now in the continuation rather than in the block that was cut.
	queue := make([]*BasicBlock, len(function.Blocks))
	copy(queue, function.Blocks)
	// Upstream keeps this map for the whole function. A function expression and its call do not
	// have to share a block: lowering an argument such as an optional-chain dependency can put CFG
	// blocks between the definition and the rewritten zero-argument memo call.
	functions := map[static_single_assignment.IdentifierId]*FunctionExpression{}

	for position := 0; position < len(queue); position++ {
		block := queue[position]
		if block == nil {
			continue
		}
		// A label cannot be expressed inside an expression block, so a call there is left alone.
		// Upstream's `isStatementBlockKind` is the same guard.
		if !isStatementBlockKind(block.Kind) {
			continue
		}

		for index := 0; index < len(block.Instructions); index++ {
			instruction := function.Instructions[block.Instructions[index]]
			if instruction == nil {
				continue
			}
			switch value := instruction.Value.(type) {
			case *FunctionExpression:
				if function.IdentifierOf(instruction.LValue) != nil &&
					function.IdentifierOf(instruction.LValue).Name == "" {
					functions[instruction.LValue.Identifier] = value
				}
			case *CallExpression:
				expression, eligible := functions[value.Callee.Identifier]
				if eligible && memoized[instruction.LValue.Identifier] {
					// A memo callback wearing an IIFE's shape. See `memoizedResults`.
					forgetFunctionOperands(instruction.Value, functions)
					continue
				}
				if !eligible || len(value.Args) != 0 {
					// Not a call of a local function expression, or a call with arguments this
					// pass does not bind. Either way the callee is still a plain use of the
					// function value, so it stops being an IIFE candidate.
					forgetFunctionOperands(instruction.Value, functions)
					continue
				}
				nested := nestedFunctionOf(function, expression)
				if nested == nil || len(nested.Params) != 0 ||
					nested.IsAsync || nested.IsGenerator {
					forgetFunctionOperands(instruction.Value, functions)
					continue
				}
				if !spliceInlinedCall(function, block, index, instruction, expression, nested,
					&queue) {
					forgetFunctionOperands(instruction.Value, functions)
					continue
				}
				inlined++
				spliced[value.Callee.Identifier] = true
				// The rest of this block now lives in the continuation, which was pushed onto the
				// queue. Nothing further in this block is ours to walk.
				index = len(block.Instructions)
			default:
				// Any other use of a function-valued temporary means it is not an IIFE: the value
				// escapes into something this pass cannot see, so inlining it would leave the
				// escaped copy calling a body that is also spliced into the caller.
				forgetFunctionOperands(instruction.Value, functions)
			}
		}
	}

	if inlined > 0 {
		// The `FunctionExpression` that produced each spliced callee is now dead: the body it names
		// runs inline and nothing else read the value, which is the condition that made it an IIFE.
		// Leaving it is not cosmetic. It is a live definition of a function value, so mutable-range
		// inference still sees a closure over the captures and still widens their ranges -- which
		// is the whole chain this pass exists to break.
		dropSplicedFunctionExpressions(function, spliced)

		// Terminals changed, so blocks may have become unreachable and the reverse-postorder the
		// rest of the package relies on no longer holds.
		ReversePostorder(function)
		MarkPredecessors(function)
		MarkEvaluationOrder(function)
	}
	return inlined
}

// memoizedResults collects the values `DropManualMemoization` recorded as a memo call's result.
//
// `DropManualMemoization` rewrites a memo callback into the same zero-argument call shape as a
// programmer-written IIFE. The marker pair is the only remaining evidence that the call came from a
// memo, and its result identifies the callback the guarded entry point must leave in place.
//
// Production uses the inclusive wrapper and therefore does not consult this set. Keeping the
// guarded wrapper is still useful: focused transformation tests can assert that ordinary IIFEs are
// selected without silently selecting rewritten memos too, and callers that have not erased memo
// ownership can request that distinction explicitly.
func memoizedResults(function *Function) map[static_single_assignment.IdentifierId]bool {
	results := map[static_single_assignment.IdentifierId]bool{}
	for _, instruction := range function.Instructions {
		if instruction == nil {
			continue
		}
		if marker, isFinish := instruction.Value.(*FinishMemoize); isFinish {
			results[marker.Value.Identifier] = true
		}
	}
	return results
}

// dropSplicedFunctionExpressions removes the instructions that defined the inlined callees.
//
// Upstream's `retainWhere` over the same set. The instruction is dropped from its block rather than
// from `Function.Instructions`, because that table is indexed by id and every other block's ids
// index into it; blanking an entry would leave a hole a later `function.Instructions[id]` reads as
// nil, which several passes here do not expect.
func dropSplicedFunctionExpressions(function *Function, spliced map[static_single_assignment.IdentifierId]bool) {
	if len(spliced) == 0 {
		return
	}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		kept := block.Instructions[:0]
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction != nil && spliced[instruction.LValue.Identifier] {
				if _, isFunction := instruction.Value.(*FunctionExpression); isFunction {
					continue
				}
			}
			kept = append(kept, instructionId)
		}
		block.Instructions = kept
	}
}

// spliceInlinedCall performs the surgery for one call site, reporting whether it happened.
//
// It declines rather than half-applying: a copy that cannot complete leaves the caller's block
// untouched, because a block cut at the call with no body spliced in is a function with a dangling
// terminal, which is worse than an uninlined IIFE.
func spliceInlinedCall(function *Function, block *BasicBlock, index int,
	instruction *Instruction, expression *FunctionExpression, nested *Function,
	queue *[]*BasicBlock) bool {

	remap, copied := CopyNestedBodyInto(function, nested, expression.Captures)
	if !copied {
		return false
	}
	entry, found := function.Block(remap.Entry)
	if !found || entry == nil {
		return false
	}

	// The continuation holds everything after the call, and inherits the original terminal. Its
	// kind matches the block it was cut from, so a statement block stays a statement block and the
	// guard above keeps meaning the same thing on the next pass over the queue.
	continuation := function.NewBlock(block.Kind)
	continuation.Instructions = append(continuation.Instructions,
		block.Instructions[index+1:]...)
	continuation.Terminal = block.Terminal

	// Cut the caller's block at the call. The call instruction itself goes too: its result is now
	// produced by the stores the returns became.
	block.Instructions = block.Instructions[:index]

	result := instruction.LValue
	if singleReturnExit(function, remap) {
		// One exit and it is a return, so no label is needed: control enters the body and leaves
		// it exactly once. The return becomes a load into the call's lvalue.
		block.Terminal = &Goto{Block: remap.Entry, Variant: GotoVariantBreak}
		rewriteCopiedReturns(function, remap, continuation.Id, result, instruction.Node, true)
	} else {
		// Several exits, so the body is treated as one labeled region and every return becomes a
		// store to a declared temporary plus a jump to the fallthrough.
		block.Terminal = &Label{Block: remap.Entry, Fallthrough: continuation.Id}
		declareInlineResult(function, block, result, instruction.Node)
		rewriteCopiedReturns(function, remap, continuation.Id, result, instruction.Node, false)
	}

	// A later IIFE in the same statement now lives in the continuation.
	*queue = append(*queue, continuation)
	return true
}

// rewriteCopiedReturns routes every `Return` in the copied body to the continuation.
//
// `direct` selects between the two forms. On the single-exit path the return's value is loaded
// straight into the call's lvalue, which is a definition of that value and needs no prior
// declaration. On the labeled path the lvalue was declared before the body, so a return REASSIGNS
// it: several returns are several stores to one binding, which is precisely what a `let` plus
// reassignments expresses and what a `const` would not.
//
// Only blocks the copy produced are touched. The remap's values are that set; walking the parent's
// blocks instead would rewrite returns the caller wrote itself, which end the caller rather than
// the inlined body.
// returnedPlace is the value a copied `Return` actually yields.
//
// A `Return` terminal names `Function.Returns`, the single identifier every `return` in a function
// stores into. That identifier is never an lvalue -- nothing defines it, the stores target it -- so
// `CopyNestedBodyInto` has no remap entry for it and the copied terminal still names the nested
// function's version. Storing it produces a reassignment whose rvalue no instruction defines.
//
// Measured on `error.useMemo-infer-less-specific-conditional-access` with the memo callback inlined:
// the store read identifier 30, which is defined nowhere, while the `ObjectExpression` the callback
// actually returns sat directly above it. The store survives dead-code elimination because a
// reassignment is unconditionally live, so nothing referenced the object and the sweep pruned it
// along with the two `PropertyLoad`s feeding it. Upstream keeps all three at its own
// `DeadCodeElimination` stage.
//
// The block's last instruction is the value: a `return expr` lowers to the expression followed by a
// load into `Returns`, so the load is what the terminal means.
//
// When there is no such load the function returns undefined -- an implicit fall-through off the end
// of a body, or a bare `return;` -- and the terminal's own place is kept.
//
// That place names `Function.Returns` and so is still a value nothing defines, which is the same
// phantom this exists to remove. Emitting an explicit `undefined` there instead was built and
// measured: with the memo callback inlined it took the golden trade from three fixtures back to six,
// so the phantom is load-bearing for a valueless arm in a way the emitted primitive is not. Left as
// it is deliberately, and the next reader who sees the asymmetry should re-measure rather than
// tidy it.
func returnedPlace(function *Function, block *BasicBlock, terminal *Return) Place {
	if len(block.Instructions) > 0 {
		last := function.Instructions[block.Instructions[len(block.Instructions)-1]]
		if last != nil {
			if load, isLoad := last.Value.(*LoadLocal); isLoad && load.Place.Identifier != 0 {
				return load.Place
			}
		}
	}
	return terminal.Value
}

func rewriteCopiedReturns(function *Function, remap *InlineRemap, continuation static_single_assignment.BlockId,
	result Place, node *ast.Node, direct bool) {
	for _, blockId := range remap.Blocks {
		block, found := function.Block(blockId)
		if !found || block == nil {
			continue
		}
		terminal, isReturn := block.Terminal.(*Return)
		if !isReturn {
			continue
		}
		returned := returnedPlace(function, block, terminal)
		if direct {
			function.AddInstruction(block, &Instruction{
				LValue: result,
				Value:  &LoadLocal{Place: returned},
				Node:   node,
			})
		} else {
			function.AddInstruction(block, &Instruction{
				LValue: Place{Identifier: function.NewIdentifier("", nil, 0).Id},
				Value: &StoreLocal{
					LValue: result,
					Value:  returned,
					Kind:   InstructionKindReassign,
				},
				Node: node,
			})
		}
		block.Terminal = &Goto{Block: continuation, Variant: GotoVariantBreak}
	}
}

// declareInlineResult emits the `let` the labeled form's reassignments store into.
//
// Without it the body's stores reassign a binding nothing declared, and a later pass reading the
// declaration finds none: the value looks like it was assigned before it existed.
func declareInlineResult(function *Function, block *BasicBlock, result Place, node *ast.Node) {
	function.AddInstruction(block, &Instruction{
		LValue: Place{Identifier: function.NewIdentifier("", nil, 0).Id},
		Value: &DeclareLocal{
			LValue: result,
			Kind:   InstructionKindLet,
		},
		Node: node,
	})
}

// singleReturnExit reports whether the copied body leaves through exactly one terminal and that
// terminal is a `Return`.
//
// Upstream's `hasSingleExitReturnTerminal`, counting `Return` and `Throw` together: a body with one
// return and one throw has two exits, and the throw leaves the caller entirely rather than reaching
// the continuation, so the label is needed to give the return somewhere to land.
func singleReturnExit(function *Function, remap *InlineRemap) bool {
	exits, returns := 0, 0
	for _, blockId := range remap.Blocks {
		block, found := function.Block(blockId)
		if !found || block == nil {
			continue
		}
		switch block.Terminal.(type) {
		case *Return:
			returns++
			exits++
		case *Throw:
			exits++
		}
	}
	return exits == 1 && returns == 1
}

// nestedFunctionOf resolves a function expression's id against the parent's table.
func nestedFunctionOf(function *Function, expression *FunctionExpression) *Function {
	if int(expression.Function) >= len(function.Functions) {
		return nil
	}
	return function.Functions[expression.Function]
}

// forgetFunctionOperands drops every function-valued temporary this value reads.
//
// The candidate set is "function expressions whose only use is being called here". Any other read
// of one -- passed as an argument, stored into an object, returned -- means the value escapes, and
// a splice would then leave the escaped reference pointing at a function whose body also runs
// inline. Upstream deletes on every operand of every non-call instruction for this reason.
func forgetFunctionOperands(value InstructionValue, functions map[static_single_assignment.IdentifierId]*FunctionExpression) {
	if len(functions) == 0 {
		return
	}
	EachPlace(value, func(place Place, role PlaceRole) {
		delete(functions, place.Identifier)
	})
}

// isStatementBlockKind reports whether a block holds statements rather than an expression's parts.
//
// A `Label` terminal is a statement-level construct; putting one inside a value block would produce
// a region that later reconstruction has to print in expression position, which it cannot. Upstream
// declines the same way and for the same reason.
func isStatementBlockKind(kind BlockKind) bool {
	switch kind {
	case BlockKindValue, BlockKindSequence:
		return false
	default:
		return true
	}
}
