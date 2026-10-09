package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoUselessReturn = rule.Message{
	Id: "unnecessaryReturn",
	Description: "This `return` ends the function at a point where the function was going to end " +
		"anyway, so it changes nothing and only adds a line the next reader has to check for a " +
		"reason. Delete it.",
}

// noUselessReturnEvent is what the control-flow graph records at each statement.
//
// The two flags are computed at the statement's position rather than looked up later, because the
// graph hands blocks back without a path from an event to its syntactic context.
type noUselessReturnEvent struct {
	node *ast.Node

	// enclosingTry is the `try` statement whose TRY BLOCK contains this node, or nil. See the
	// rule doc comment: it is the boundary that decides whether a later statement counts.
	enclosingTry *ast.Node

	// exempt marks a return the rule never reports whatever follows it, because it is in a loop
	// or in a `finally`.
	exempt bool
}

// NoUselessReturn flags a `return;` that ends a function where it was ending anyway.
//
//	valid:   function foo() { return 5; }
//	valid:   function foo() { if (bar) { doSomething(); return; } else { doSomethingElse(); } qux(); }
//	valid:   function foo() { for (var i = 0; i < 10; i++) { return; } }
//	valid:   function foo() { try { return 5; } finally { return; } }
//	valid:   function foo() { return; doSomething(); }
//	invalid: function foo() { doSomething(); return; }
//	invalid: function foo() { if (condition) { bar(); return; } else { baz(); } }
//
// # The judgment, and why it needs a control-flow graph
//
// A bare `return` is useless when nothing executes after it. That is not a syntactic question:
// `if (bar) { return; } else { baz(); }` reports while the same shape followed by `qux();` does
// not, and the difference is whether control reaches a statement once the return is removed.
//
// Upstream answers it with ESLint code paths, pushing candidate returns onto per-segment lists and
// removing them again when any statement is later reached on that segment. This port asks the same
// question of `control_flow_graph`: from the return's block, can the walk reach a statement that
// would have executed? The two formulations agree on all forty nine of upstream's cases, which was
// measured before this file was written rather than after.
//
// # What counts as a statement that rescues a return
//
// Upstream registers its "mark as used" handler on every statement kind EXCEPT `BlockStatement`,
// `FunctionDeclaration` and `BreakStatement`, and separately on a `ReturnStatement` only when that
// return carries a value. Both halves of that last distinction are load-bearing and each moves
// cases in a different direction:
//
//	function foo() { return; return; }             reports ONCE, so a bare return rescues nothing
//	function foo() { if (bar) return; return baz; } clean, so a valued return rescues
//
// A block rescues nothing because it is not itself executable, a function declaration because it is
// hoisted and runs whether the return is there or not, and a `break` because it merely moves on.
//
// # Two exemptions taken from the syntax rather than from the graph
//
// A return inside a LOOP is never reported: it exits the loop early, which is a real effect
// whatever follows. A return inside a `finally` is never reported either, because it overrides the
// value the `try` was going to return, which upstream's own corpus writes with a comment saying so.
//
// # Three properties of the walk, each found by a corpus case rather than by reading
//
// **An unreachable tail is followed only when a return created it.** Upstream recurses into an
// unreachable segment's predecessors filtered by `isReturned`. The shape that pins it is
// `case 1: if (a) { doSomething(); return; } break;`, where the break's block has an unreachable
// successor leading to the next case. That successor is the fall-through that would exist without
// the break, and following it finds the next case's code and wrongly rescues the return.
//
// **Once on the tail, keep following it.** `try { bar(); return; } finally { baz(); } qux();` is
// clean, and the tail runs return, then the finally's copy, then `qux()`. The finally copy ends in
// an expression statement rather than a return, so a test applied at every hop stops one block
// short of the `qux()` that makes the return legitimate.
//
// **A return inside a `try` is rescued only by code after the WHOLE statement.** Measured against
// the installed build at 10.8.1, both directions:
//
//	try { foo(); return; } catch (e) { qux(); }        REPORTS, though qux() is ordinary code
//	try { bar(); return; } finally { baz(); } qux();   clean
//
// Reaching the catch means the try threw, so that path was never this return's continuation, and
// the finally runs on every exit including this return's own. The first case rules out asking
// whether a statement merely follows; the second rules out asking whether it is contained in the
// try. What answers both is comparing the statement's position against the try statement's end.
var NoUselessReturn = rule.Rule{
	Name:             "no-useless-return",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// One code path per root, as upstream runs one per function. A nested function's
				// returns belong to the nested path: `try { return 5; } finally { function bar() {
				// return; } }` reports the INNER return and not the outer one.
				//
				// Only a root whose own code holds a bare `return;` can report, so only those are
				// built. On a cold ahra run that is 4.4% of roots and 13% of the blocks every root
				// used to cost (#tmn9n27).
				for _, root := range codePathRoots(ctx, node) {
					if root.BareReturn {
						noUselessReturnCheckRoot(ctx, root.Node)
					} else if CheckCodePathGates {
						noUselessReturnCheckSkipped(root.Node)
					}
				}
			},
		}
	},
}

// noUselessReturnCheckRoot runs the judgment over one code-path root.
func noUselessReturnCheckRoot(ctx rule.Context, root *ast.Node) {
	graph := control_flow_graph.Build(root, control_flow_graph.Hooks[noUselessReturnEvent]{
		Statement: func(builder *control_flow_graph.Builder[noUselessReturnEvent], node *ast.Node) {
			builder.Emit(noUselessReturnEvent{
				node:         node,
				enclosingTry: noUselessReturnEnclosingTryBlock(node, root),
				exempt: noUselessReturnInLoop(node, root) ||
					noUselessReturnInFinally(node, root),
			})
		},
	})

	for _, block := range graph.Blocks {
		// A return in unreachable code is not useless, it is dead, and deleting it is a different
		// rule's business. `function foo() { return; doSomething(); }` is clean upstream.
		if !block.Reachable {
			continue
		}
		for index, event := range block.Events {
			candidate := event.node
			if candidate.Kind != ast.KindReturnStatement ||
				candidate.AsReturnStatement().Expression != nil {
				continue
			}
			if event.exempt {
				continue
			}
			if noUselessReturnSomethingRunsAfter(block, index, event.enclosingTry) {
				continue
			}
			if noUselessReturnTheCompilerRequiresIt(ctx, root, candidate) {
				continue
			}

			if fix, canFix := noUselessReturnFix(ctx, candidate); canFix {
				ctx.ReportNodeWithFixes(candidate, messageNoUselessReturn, fix)
				continue
			}
			ctx.ReportNode(candidate, messageNoUselessReturn)
		}
	}
}

// noUselessReturnCheckSkipped is the gate's check under CheckCodePathGates: a root skipped for
// holding no bare return must lay out none.
func noUselessReturnCheckSkipped(root *ast.Node) {
	control_flow_graph.Build(root, control_flow_graph.Hooks[struct{}]{
		Statement: func(builder *control_flow_graph.Builder[struct{}], node *ast.Node) {
			if node.Kind == ast.KindReturnStatement && node.AsReturnStatement().Expression == nil {
				panic("no-useless-return skipped a root whose graph holds a bare return")
			}
		},
	})
}

// noUselessReturnSomethingRunsAfter answers whether any statement executes after the candidate.
func noUselessReturnSomethingRunsAfter(block *control_flow_graph.Block[noUselessReturnEvent],
	candidateIndex int, enclosingTry *ast.Node) bool {
	for _, later := range block.Events[candidateIndex+1:] {
		if noUselessReturnExecutes(later.node) && !noUselessReturnInsideTry(later.node, enclosingTry) {
			return true
		}
	}
	return noUselessReturnWalk(block,
		map[*control_flow_graph.Block[noUselessReturnEvent]]bool{}, enclosingTry)
}

// noUselessReturnWalk follows the successors, including the dead tail the return itself created.
func noUselessReturnWalk(from *control_flow_graph.Block[noUselessReturnEvent],
	seen map[*control_flow_graph.Block[noUselessReturnEvent]]bool, enclosingTry *ast.Node) bool {
	for _, successor := range from.Successors {
		if seen[successor] {
			continue
		}
		seen[successor] = true

		// An unreachable successor is entered only from a block that ended in a return, which is
		// upstream's `isReturned` filter. Once inside the tail, with `from` unreachable too, keep
		// going, because the tail can pass through blocks that do not end in a return.
		//
		// Both halves were found by a corpus case rather than derived, and neither is obvious from
		// the graph alone.
		//
		// The first: in `case 1: if (a) { doSomething(); return; } break;` the break's block has an
		// unreachable successor leading into the NEXT case. That successor is the fall-through that
		// would exist if the break were not written, so it is not this return's continuation.
		// Following it finds the next case's code, rescues the return, and goes silent on an input
		// upstream reports.
		//
		// The second: in `try { bar(); return; } finally { baz(); } qux();` the tail runs return,
		// then the finally's copy, then `qux()`. That copy ends in an expression statement rather
		// than a return, so re-asking "did this block end in a return" at every hop stops one block
		// short of the `qux()` that makes the return legitimate, and reports a clean input.
		if !successor.Reachable && from.Reachable && !noUselessReturnEndsInReturn(from) {
			continue
		}

		for _, event := range successor.Events {
			if !noUselessReturnExecutes(event.node) {
				continue
			}
			if noUselessReturnInsideTry(event.node, enclosingTry) {
				continue
			}
			return true
		}
		if noUselessReturnWalk(successor, seen, enclosingTry) {
			return true
		}
	}
	return false
}

// noUselessReturnInsideTry answers whether a statement lies in the CATCH or FINALLY of the try
// statement the candidate return sits in, and therefore cannot rescue it.
//
// The boundary is the try BLOCK's end rather than the try statement's, and getting that wrong is a
// real false positive rather than a detail: an early exit followed by more work inside the same
// `try` is the ordinary shape of guarded code, and it appeared 51 times on this tree before this
// was corrected. Measured against the installed build at 10.8.1, all six directions:
//
//	try { if (x) { return; } more(); } catch (e) {}          clean, later code in the TRY rescues
//	try { foo(); return; } catch (e) { qux(); }              REPORTS, the catch does not
//	try { return; } finally { bar(); }                       REPORTS, the finally does not
//	try { bar(); return; } finally { baz(); } qux();         clean, code after the whole statement
//	try { bar(); return; } catch (e) {} baz();               clean, same
//	try { bar(); return; } catch (e) {}                      REPORTS, nothing follows
//
// Reaching the catch means the try threw, so that path was never this return's continuation, and
// the finally runs on every exit including this return's own. Statements still in the try block are
// the return's actual continuation and do rescue it.
func noUselessReturnInsideTry(node *ast.Node, enclosingTry *ast.Node) bool {
	if enclosingTry == nil {
		return false
	}
	tryBlock := enclosingTry.AsTryStatement().TryBlock
	if tryBlock == nil {
		return false
	}
	// Inside the try block: a real continuation, so it rescues.
	if node.Pos() < tryBlock.End() {
		return false
	}
	// Between the try block's end and the whole statement's end: the catch or the finally.
	return node.Pos() < enclosingTry.End()
}

// noUselessReturnExecutes answers whether reaching this statement rescues an earlier return.
func noUselessReturnExecutes(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindBlock, ast.KindFunctionDeclaration, ast.KindBreakStatement:
		// A block is not itself executable, a hoisted function runs either way, and a break
		// merely moves on. Upstream omits all three from its handler list.
		return false
	case ast.KindReturnStatement:
		// A valued return marks earlier returns as used; a bare one does not.
		return node.AsReturnStatement().Expression != nil
	}
	return true
}

// noUselessReturnEndsInReturn answers upstream's `isReturned` for a block.
func noUselessReturnEndsInReturn(block *control_flow_graph.Block[noUselessReturnEvent]) bool {
	for index := len(block.Events) - 1; index >= 0; index-- {
		switch block.Events[index].node.Kind {
		case ast.KindReturnStatement:
			return true
		case ast.KindBlock:
			// A block event is bookkeeping rather than a statement, so it is looked past.
			continue
		default:
			return false
		}
	}
	// A block carrying no statements of its own is part of a tail already being walked.
	return true
}

// noUselessReturnInLoop answers whether the node sits inside a loop belonging to this root.
func noUselessReturnInLoop(node *ast.Node, root *ast.Node) bool {
	for parent := node.Parent; parent != nil && parent != root; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindWhileStatement, ast.KindDoStatement:
			return true
		}
	}
	return false
}

// noUselessReturnInFinally answers whether the node sits inside a `finally` block.
//
// A return there overrides whatever the `try` was going to return, so it is never useless. Upstream
// writes the case with a comment saying exactly that.
func noUselessReturnInFinally(node *ast.Node, root *ast.Node) bool {
	for current := node; current != nil && current.Parent != nil && current != root; current = current.Parent {
		if current.Parent.Kind == ast.KindTryStatement &&
			current.Parent.AsTryStatement().FinallyBlock == current {
			return true
		}
	}
	return false
}

// noUselessReturnEnclosingTryBlock returns the innermost try statement whose TRY BLOCK contains the
// node, or nil.
func noUselessReturnEnclosingTryBlock(node *ast.Node, root *ast.Node) *ast.Node {
	for current := node; current != nil && current.Parent != nil && current != root; current = current.Parent {
		if current.Parent.Kind == ast.KindTryStatement &&
			current.Parent.AsTryStatement().TryBlock == current {
			return current.Parent
		}
	}
	return nil
}

// noUselessReturnTheCompilerRequiresIt reports whether deleting this return stops the file compiling.
//
// A `return;` that ends a function where it was ending anyway changes nothing at run time, but it can
// still be the only thing satisfying the compiler. TypeScript requires an explicit return statement
// from a function whose declared return type is neither `void`-including, `any`, nor exactly
// `undefined` (TS2355, `checkAllCodePathsInNonVoidFunctionReturnOrThrow` in the vendored checker),
// and from every `get` accessor whatever its type (TS2378). Upstream is not type-aware and reports
// it anyway; its fixer then breaks the build.
//
// Found on api-phi-health's `GoogleAdsEnhancedConversionsService.ts`, a method declared
// `DictionaryType<unknown> | undefined` whose body was `return;`. Every row below was compiled after
// deleting the return rather than predicted:
//
//	DictionaryType<unknown> | undefined   TS2355     unknown                TS2355
//	Promise<string | undefined>, async    TS2355     get g(): string | ...  TS2378
//	undefined, void, Promise<void>, any   compiles   unannotated            compiles
//	Nothing | string, Nothing = void      compiles   another return remains compiles
//
// The alias row is why this asks the checker rather than reading the annotation: `void` can sit
// behind a name. The last row is why it matters that the candidate is the only return: TS2355 is
// the compiler's "no explicit return at all" check, so one other return of any kind satisfies it.
//
// A generator with an annotated return type declines without a type question. The checker unwraps
// its return type through the generator's own type arguments, which is not ported, and staying
// silent costs a missed report where guessing costs a broken build.
func noUselessReturnTheCompilerRequiresIt(ctx rule.Context, root *ast.Node, candidate *ast.Node) bool {
	if !ast.IsFunctionLikeDeclaration(root) || !noUselessReturnIsTheOnlyReturn(root, candidate) {
		return false
	}
	if root.Kind == ast.KindGetAccessor {
		return true
	}
	if root.Type() == nil {
		return false
	}
	flags := ast.GetFunctionFlags(root)
	if flags&ast.FunctionFlagsGenerator != 0 {
		return true
	}
	declared := ctx.TypeChecker.GetTypeFromTypeNode(root.Type())
	if flags&ast.FunctionFlagsAsync != 0 {
		declared = ctx.TypeChecker.GetPromisedTypeOfPromise(declared)
	}
	if declared == nil {
		return true
	}
	return !noUselessReturnMaybeTypeOfKind(declared, checker.TypeFlagsVoid) &&
		declared.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUndefined) == 0
}

// noUselessReturnIsTheOnlyReturn reports whether the candidate is the only return statement in its
// function, which is the binder's `HasExplicitReturn` with the candidate removed. A nested function
// owns its own returns.
func noUselessReturnIsTheOnlyReturn(root *ast.Node, candidate *ast.Node) bool {
	only := true
	var visit func(*ast.Node) bool
	visit = func(current *ast.Node) bool {
		if current == nil || !only {
			return true
		}
		if current != root && (ast.IsFunctionLike(current) || current.Kind == ast.KindClassStaticBlockDeclaration) {
			return false
		}
		if current.Kind == ast.KindReturnStatement && current != candidate {
			only = false
			return true
		}
		current.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
	return only
}

// noUselessReturnMaybeTypeOfKind is the checker's `maybeTypeOfKind`: the type has the flag, or a
// union or intersection member does.
func noUselessReturnMaybeTypeOfKind(t *checker.Type, kind checker.TypeFlags) bool {
	if t.Flags()&kind != 0 {
		return true
	}
	if t.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		for _, member := range t.Types() {
			if noUselessReturnMaybeTypeOfKind(member, kind) {
				return true
			}
		}
	}
	return false
}

// noUselessReturnFix proposes deleting the return statement, or declines.
//
// # Two declines, both upstream's and both reproduced
//
// **A statement that is not a member of a statement list.** `if (foo) return;` has the return as
// the `if`'s consequent, so deleting it leaves `if (foo)` with nothing to run and does not parse.
// Upstream's `isRemovable` asks the same question through `STATEMENT_LIST_PARENTS`, and the case is
// one of its three `output: null` entries.
//
// **A comment inside the statement.** `return/**/;` and `return//\n;` are the other two, and this
// is the guard that matters most in this file: without it the repair silently deletes a comment
// somebody wrote, and the deletion is applied unattended.
//
// The boundary is narrower than "the statement has a comment near it", and that narrowness is the
// part nothing upstream pins. `getCommentsInside` asks only about comments the statement's own span
// covers, so a comment beside the return is untouched by the guard and the repair still runs.
// Measured against the installed build at 10.8.1, both directions:
//
//	function foo() { bar(); return/**/; }        reports, NO repair
//	function foo() { bar(); /*keep*/ return; }   reports, repair runs, comment survives
//	function foo() { bar(); return; /*keep*/ }   reports, repair runs, comment survives
//
// Both halves are load-bearing and they fail in opposite directions. Losing the first eats a
// comment unattended. Losing the second stops the rule repairing ordinary code because a comment
// happens to sit next to it, which is the quieter failure and the one a reader would not report as
// a bug. Upstream's corpus covers the first two rows and writes nothing for the adjacency, so the
// fixtures for it here are hand-written rather than imported.
//
// The span is the statement's own text via `RemoveNode`, which trims to the token and so leaves the
// surrounding whitespace alone. That matters for a deletion fixer in this tree specifically: a span
// built from `Pos()` would eat the preceding trivia, and a span built from a neighbouring node's
// offset is how a sibling rule stranded TypeScript type annotations across eight files.
func noUselessReturnFix(ctx rule.Context, node *ast.Node) (rule.Fix, bool) {
	if !noUselessReturnIsRemovable(node) {
		return rule.Fix{}, false
	}

	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	if noUselessReturnHasCommentInside(ctx.SourceFile.Text(), nodeRange.Pos(), nodeRange.End()) {
		return rule.Fix{}, false
	}

	return ctx.RemoveNode(node), true
}

// noUselessReturnHasCommentInside scans the statement's own text for a comment.
//
// The shelf's `comments.ForFile` is the obvious tool and it is the wrong one here, which was
// measured rather than assumed. It collects LEADING and TRAILING trivia around nodes, so a comment
// between the `return` keyword and its semicolon belongs to no node's trivia and it reports zero:
//
//	function foo() { bar(); return/**/; }        comments.All finds 0
//	function foo() { bar(); return /* why */ ; } comments.All finds 0
//	function foo() { bar(); /*keep*/ return; }   comments.All finds 1
//
// The helper is right for what it is for; the first two shapes are exactly the ones this guard
// exists to catch, and they are upstream's own `output: null` cases. So the span is scanned
// directly. A string or a regular expression cannot appear between `return` and `;` in a statement
// this rule reports on, since a bare return has no expression at all, so a scan that only knows
// about comment delimiters is complete here rather than approximate.
func noUselessReturnHasCommentInside(text string, start int, end int) bool {
	if start < 0 || end > len(text) {
		return false
	}
	for position := start; position+1 < end; position++ {
		if text[position] != '/' {
			continue
		}
		switch text[position+1] {
		case '/':
			return true
		case '*':
			return true
		}
	}
	return false
}

// noUselessReturnIsRemovable answers upstream's `isRemovable`: is this statement a member of a
// statement list, so that removing it leaves a well-formed program?
func noUselessReturnIsRemovable(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock,
		ast.KindCaseClause, ast.KindDefaultClause:
		return true
	}
	return false
}
