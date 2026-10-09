// Per-file memoization of lowering, so the rules that share a function share the work.
package high_level_intermediate_representation

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ForFunction is Lower followed by Construct, computed once per function per file and shared by
// every rule that asks.
//
// # Why this exists
//
// Three react rules lower the same functions independently, each over the whole source file:
// set-state-in-render, static-components, and set-state-in-effect. A component holding a `useMemo`,
// a state setter, and a JSX tag was lowered three times and converted to single-assignment form
// three times. Measured with `cohere --timing` before this existed: set-state-in-effect 829ms,
// set-state-in-render 535ms, static-components 494ms, against every syntactic react rule under
// 20ms. The author of the most expensive one named the cause exactly — "a property of lowering, not
// this rule" — and a shared per-file cache as the fix.
//
// # Why Construct is inside the cached value rather than left to the caller
//
// This is the part that makes a shared lowering safe, and getting it wrong would have been silent.
// `Construct` mutates in place: it renames identifiers and appends phis. It is NOT idempotent.
// Measured on `let y = 2; if (y > 1) { y = 1; } else { y = 2; } let x = y; return x;`, a second
// Construct over the same graph moved it from 1 phi and 5 named values to 2 phis and 6 named
// values.
//
// The dangerous part is that the doubly-constructed graph is still WELL-FORMED — `VerifySSA`
// reports zero violations on it. So a cache that handed out a raw lowering and let each of the
// three callers run its own Construct would have given the second and third caller a graph carrying
// phis that correspond to no branch in the source, with no verifier, no panic, and no test
// complaining. It would have surfaced later as a wrong finding in whichever rule happened to run
// second, and it would have looked like a bug in that rule rather than in this cache.
//
// So the unit stored here is the finished graph, lowered and constructed, and Construct runs
// exactly once per function. That also answers whether construction wants a cache of its own: it
// does not, because it is not separable from the lowering it mutates. Measured on the probe
// component in cache_test.go, construction is about 18% of the combined cost (23µs against 102µs
// for lowering), and all of it is now paid once rather than three times. Callers read it and must not mutate it; the three callers today only
// read, deriving their own maps (`UnconditionalBlocks` builds a fresh map, the taint passes build
// their own).
//
// # Why the function node alone is the key
//
// `Lower` takes a node and a type checker, so a key naming only the node is wrong if the checker
// can differ between two calls within one file. It cannot. `internal/program/walk.go` builds one
// `fileChecker` per file and closes it into every rule's Context alongside the one `FileCache`
// created on the line above, so the checker and this cache have exactly the same lifetime. A
// differing checker would mean a differing cache, and there is no path that produces one.
//
// # Why this is safe for the on-disk findings cache
//
// It is not run-scoped state and it reaches nothing outside the file. A lowering is derived purely
// from a function node in the file being linted plus that file's checker, and the cache is created
// and discarded per file by the walk. So no rule using this acquires a `ProgramReads` declaration,
// and the findings cache stays correct: nothing here can make one file's result depend on another
// file's contents beyond what the checker already answers, which is the failure `ProgramReads`
// exists to declare.
func ForFunction(ctx rule.Context, node *ast.Node) *Function {
	if node == nil {
		return nil
	}
	// Declining a checker-less file here rather than in each caller keeps the answer independent of
	// which rule asked first. `Lower`'s own comment is that a lowering built without a checker
	// resolves every reference as a global, making the result well-formed and meaningless, and two
	// of the three callers already refused to run in that state. Deciding it once means a cached
	// entry cannot depend on rule order.
	if ctx.TypeChecker == nil {
		return nil
	}
	return rule.Cached(ctx.FileCache, cacheKeyFor(node), func() *Function {
		lowered := Lower(node, ctx.TypeChecker)
		if lowered == nil {
			return nil
		}
		Construct(lowered)
		return lowered
	})
}

// AsCompilationUnit returns the lowering a rule should analyze a component or hook through: the
// function lowered on its own, through the rule's own entry (ForFunction or
// ForFunctionWithoutManualMemoization).
//
// The compiler rules lower the outermost function on each branch and, when that is not a component
// or hook, look for one among the functions nested in its lowering. A unit found that way was then
// analyzed as a NESTED function, with the function around it as its enclosing builder, so a variable
// that function declared became a capture: a local of the component, as far as the analysis could
// tell. React Compiler compiles such a component as a unit of its own, and its HIR builder resolves
// any binding outside the unit as it resolves a module-level one (`resolveIdentifier` asks the unit's
// parent scope, and Babel's `getBinding` walks every enclosing scope).
//
// Measured on TanStack Query's tests, a `Page` declared inside an `it(...)` callback: React Compiler
// compiles it, and writing `isRefetch = true` from its onClick handler or `result = query` from its
// effect, where both variables belong to the test, raises nothing. cohere reported 21 immutability
// findings on that shape, every one of them reading the test's variable as the component's own
// (#zx5xvtg item 8).
//
// For a unit already lowered on its own this is the same cached lowering, so a top-level component or
// hook is unaffected. A component nested in a unit is not reached here at all: it is part of its
// parent's graph, as upstream lowers it.
func AsCompilationUnit(
	ctx rule.Context,
	function *Function,
	lower func(rule.Context, *ast.Node) *Function,
) *Function {
	if function == nil || function.Node == nil {
		return function
	}
	if standalone := lower(ctx, function.Node); standalone != nil {
		return standalone
	}
	return function
}

// ForFunctionWithoutManualMemoization is ForFunction with `useMemo` and `useCallback` erased and
// the resulting immediately-invoked calls inlined, which is the graph upstream's validators read.
//
// # Why this is a second cache entry rather than a step inside the first
//
// Upstream erases manual memoization at `Pipeline.ts:169`, before any validator runs, so a rule
// ported from one of those validators is only faithful if its input has been through the same
// erasure. `set-state-in-effect` is one: `const f = useCallback(() => setS(1), []); useEffect(() =>
// f())` is reported by React 7.1.1 and was silent here, because the wrapper broke the alias chain
// the rule follows. The rewrite `useCallback(fn, deps)` -> `LoadLocal fn` restores it with no
// change to the rule.
//
// Putting the pass in `ForFunction` was tried first and is wrong, which the suite caught rather
// than review: three other rules read that same cached graph and need the memo call intact.
// `set-state-in-render` distinguishes a setter called inside a `useMemo` callback from one called
// during render, and with the call erased the first shape becomes the second, so two fixtures
// reported twice. `refs` misnamed a finding on `error.maybe-mutable-ref-not-preserved`, and
// `immutability` failed too. The pass mutates in place, so sharing it silently changes what every
// later rule sees — the same hazard the Construct note above describes, in a form no verifier
// catches.
//
// So the erasure is one rule family's view of the graph, not the representation's, and it gets its
// own entry under its own key.
//
// # What the second lowering costs, measured twice because the first measurement was wrong
//
// Ungated, it costs about 100ms of a 2.1s lint phase. Gated by `mentionsManualMemoization` below it
// costs nothing measurable: over Kirk's tree, twenty-five interleaved runs of each binary, median
// 2.160s with the erasure against 2.170s without it, and the minimum is 1.950s against 1.960s. Two
// statistics that disagree in opposite directions by ten milliseconds is what "no difference" looks
// like. The gate is why: 12,986 of 13,171 functions skip the second lowering entirely, so only the
// 185 that actually memoize pay for one.
//
// The inlining pass added alongside the erasure is free for the same reason: it only runs on the
// functions the gate already admitted. Ten interleaved runs each, before and after wiring it in:
// mean 2.038s against 2.059s, minimum 1.970s against 1.980s.
//
// An earlier version of this comment claimed 570ms, and that number was measured wrong in a way
// worth recording because the trap is easy to fall into twice. It came from runs under `--timing`,
// which wraps every listener call in a `time.Now` pair; `cmd/cohere/timing.go` says so on the line
// that prints the total, in as many words -- compare rules to each other, never these totals to a
// normal run. It also came from three samples against a machine whose run-to-run spread is larger
// than the effect being measured. The honest procedure is what produced the numbers above:
// interleave the binaries so drift hits both, take twenty-plus samples, and read the median and the
// minimum rather than the mean, since a single descheduled run moves a mean and cannot move a
// minimum.
//
// A future `ForFunction` could keep the memo call and record the rewrite alongside it, letting both
// families read one graph and removing even the gated cost. That is a change to the representation
// and wants its own commit, its own measurement, and the six rules re-run against it. It is not
// worth doing for the time: it is worth doing only if a third rule family ever wants the erased
// form, at which point the gate stops being enough.
//
// Running the pass after Construct rather than before it, as upstream's line order has it, is safe
// for the reason the pass's own header gives: it creates and deletes no instruction, block, or
// terminal, so the control-flow graph is bit-identical and single-assignment form is undisturbed.
// Only one instruction value per memo call changes, toward fewer operands.
// `AnalyzePreservedManualMemoization` already sequences it post-Construct for the same reason.
func ForFunctionWithoutManualMemoization(ctx rule.Context, node *ast.Node) *Function {
	if node == nil || ctx.TypeChecker == nil {
		return nil
	}
	// A function with no memo call anywhere in its text lowers to a graph the erasure cannot
	// change, so it shares the entry every other rule already paid for. This is the whole
	// optimisation and it is worth stating why it is safe rather than merely fast:
	// `DropManualMemoization` only ever rewrites a `CallExpression` or `MethodCall` whose callee
	// resolves to one of the two names, so a function whose source cannot spell either name, plainly
	// or through an escape, has nothing for it to find. The check is over the function's own span
	// including its nested functions, which is the same subtree the lowering covers.
	//
	// The numbers this buys are in the header above rather than repeated here, so there is one
	// place to correct when they go stale.
	if !mentionsManualMemoization(node) {
		return ForFunction(ctx, node)
	}
	return rule.Cached(ctx.FileCache, cacheKeyFor(node)+":no-manual-memo", func() *Function {
		lowered := Lower(node, ctx.TypeChecker)
		if lowered == nil {
			return nil
		}
		Construct(lowered)
		DropManualMemoization(lowered)
		// Upstream runs the inlining pass one line after the erasure, and the two together are
		// what make a memoized callback reachable: the erasure turns `useMemo(fn, deps)` into
		// `fn()`, and this turns `fn()` into the closure itself. The memo-inclusive entry point is
		// deliberate and its own comment says why the other caller wants the guard this one lifts.
		//
		// `Construct` again because the splice copies a body that arrives carrying its own phis,
		// and only when something was actually spliced, which is most files never.
		if InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks(lowered) > 0 {
			Construct(lowered)
		}
		return lowered
	})
}

// mentionsManualMemoization reports whether this function could hold a memo call the erasure
// recognises, which is `MayNameManualMemoization` over the function's own span.
//
// What it must never do is answer "no" for a function the erasure would have changed, because that
// function is then handed the memo-intact graph and the rules reading it go silent with no error
// anywhere. A plain substring search was not enough for that: an escaped spelling such as
// `use\u0043allback` reaches the erasure through the cooked identifier text without the name's letters
// appearing in a row, and `MayNameManualMemoization`'s own comment has the measurement. Answering
// "yes" for a function that merely mentions the name in a comment or a string costs one extra
// lowering and nothing else, which is the direction that is allowed to be wrong, and an unreadable
// span answers "yes" for the same reason.
//
// The one spelling this does not catch is an import renamed to something else entirely
// (`import {useCallback as memo}` called as `memo(...)`), and that is already outside what the pass
// recognises: its sidemap keys on the name at the call site, matching upstream's syntactic
// recognition, so a call spelled `memo(...)` is not erased by either implementation.
func mentionsManualMemoization(node *ast.Node) bool {
	sourceFile := ast.GetSourceFileOfNode(node)
	if sourceFile == nil {
		return true
	}
	text := sourceFile.Text()
	start, end := node.Pos(), node.End()
	if start < 0 || end > len(text) || start >= end {
		return true
	}
	return MayNameManualMemoization(node, text[start:end])
}

// cacheKeyFor names one function node within one file.
//
// The node's source position is its identity here. That is sufficient precisely because the cache
// is per-file: two function nodes in one file cannot start at the same offset, and a node in
// another file is another cache. The kind is included so the key stays unambiguous if a future
// caller ever asks for a different node that happens to share a position with this one.
//
// A pointer address would be the obvious alternative and is worse: it is not printable in a way
// `Fills` can report usefully, and it would silently key two runs of the same file
// differently.
func cacheKeyFor(node *ast.Node) string {
	return "hir.Function:" + strconv.Itoa(int(node.Kind)) + ":" + strconv.Itoa(node.Pos())
}
