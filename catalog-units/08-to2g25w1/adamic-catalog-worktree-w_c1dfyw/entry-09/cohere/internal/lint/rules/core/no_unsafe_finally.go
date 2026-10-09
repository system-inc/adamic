package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnsafeFinallyReturn = rule.Message{
	Id: "unsafeReturn",
	Description: "A `return` inside `finally` overrides whatever the `try` or `catch` was about to " +
		"do. JavaScript suspends a pending return or throw until the finally block completes, so this " +
		"statement replaces the value that was on its way out and discards a pending exception " +
		"entirely, with nothing at the call site to show an error was swallowed. Set a variable here " +
		"and return it after the try statement instead.",
}

var messageUnsafeFinallyThrow = rule.Message{
	Id: "unsafeThrow",
	Description: "A `throw` inside `finally` replaces whatever the `try` or `catch` was about to do. " +
		"A pending return never happens and a pending exception is lost, so the error that actually " +
		"caused the failure is replaced by this one and the original stack is gone. Throw after the " +
		"try statement, or attach the original as a cause.",
}

var messageUnsafeFinallyBreak = rule.Message{
	Id: "unsafeBreak",
	Description: "A `break` inside `finally` leaves the try statement before the pending return or " +
		"throw can complete, so the value is discarded and the exception never surfaces. Execution " +
		"continues after the loop as though nothing had failed.",
}

var messageUnsafeFinallyContinue = rule.Message{
	Id: "unsafeContinue",
	Description: "A `continue` inside `finally` leaves the try statement before the pending return " +
		"or throw can complete, so the value is discarded and the exception never surfaces. The loop " +
		"advances to its next iteration as though nothing had failed.",
}

// NoUnsafeFinally flags a control-flow statement in a `finally` block that can escape it.
//
//	valid:   try { return 1; } finally { console.log('done'); }
//	valid:   try {} finally { const f = () => { return 1; }; }
//	valid:   try {} finally { while (true) { break; } }
//	valid:   try {} finally { label: { break label; } }
//	invalid: try { return 1; } finally { return 3; }
//	invalid: try {} finally { throw new Error(); }
//	invalid: while (true) try {} finally { break; }
//
// A `try` or `catch` that returns or throws does not do it immediately: the action is suspended
// until the finally block finishes. If finally performs its own control flow, that action never
// resumes. A returned value is silently replaced, and a pending exception is discarded with nothing
// at the call site to show that anything failed, which is the case that turns a crash into a wrong
// answer.
//
// The question the rule actually asks is not "is this statement inside a finally block" but "can
// this statement's jump leave the finally block". Those differ, and the difference is the whole rule:
// a `break` inside a loop written in the finally block never escapes, so it is fine, while the same
// `break` written directly in the finally block escapes the try statement and is not.
//
// So each statement kind walks up its own ancestors toward a finally block and stops at the first
// construct that would absorb its jump. Those stopping points are different per kind, and getting
// them wrong in either direction is a real defect rather than a rounding error:
//
//   - `return` and `throw` are absorbed only by a function or class boundary. Nothing else catches
//     them, which is why a `return` in a callback defined inside finally is fine and a `return` in a
//     loop inside finally is not.
//   - An unlabeled `break` is absorbed by a loop or a switch as well, because either one is a legal
//     target for it.
//   - A `continue` is absorbed by a loop but not by a switch, since a switch is not a continue
//     target. That asymmetry is why ESLint's own suite carries
//     `while (true) try {} finally { switch (true) { case true: continue; } }` as a violation: the
//     switch looks like it contains the jump and does not.
//   - A labeled `break` or `continue` is absorbed by the labeled statement it names, wherever that
//     sits, so `label: { break label; }` written entirely inside finally is fine and
//     `label: try {} finally { break label; }` is not.
//
// A class boundary stops the walk alongside a function boundary because a class body's member
// initializers and static blocks are their own control-flow regions; ESLint stops there too.
//
// No fix. The repair is to hoist the value out of the finally block and act on it afterward, which
// means deciding what should happen to the pending action the statement was discarding, and that is
// the author's knowledge rather than the linter's.
var NoUnsafeFinally = rule.Rule{
	Name: "no-unsafe-finally",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindReturnStatement: func(node *ast.Node) {
				if escapesIntoFinally(node, stopsReturnOrThrow, nil) {
					ctx.ReportNode(node, messageUnsafeFinallyReturn)
				}
			},

			ast.KindThrowStatement: func(node *ast.Node) {
				if escapesIntoFinally(node, stopsReturnOrThrow, nil) {
					ctx.ReportNode(node, messageUnsafeFinallyThrow)
				}
			},

			ast.KindBreakStatement: func(node *ast.Node) {
				breakStatement := node.AsBreakStatement()
				if breakStatement == nil {
					return
				}
				// A labeled break targets its label rather than the nearest enclosing loop or
				// switch, so neither of those absorbs it and the walk uses the same stopping points
				// as `return`, plus the label itself.
				if breakStatement.Label != nil {
					if escapesIntoFinally(node, stopsReturnOrThrow, breakStatement.Label) {
						ctx.ReportNode(node, messageUnsafeFinallyBreak)
					}
					return
				}
				if escapesIntoFinally(node, stopsUnlabeledBreak, nil) {
					ctx.ReportNode(node, messageUnsafeFinallyBreak)
				}
			},

			ast.KindContinueStatement: func(node *ast.Node) {
				continueStatement := node.AsContinueStatement()
				if continueStatement == nil {
					return
				}
				// A labeled continue still has to leave through a loop, so unlike a labeled break it
				// keeps the loop stopping point and adds the label to it.
				if continueStatement.Label != nil {
					if escapesIntoFinally(node, stopsContinue, continueStatement.Label) {
						ctx.ReportNode(node, messageUnsafeFinallyContinue)
					}
					return
				}
				if escapesIntoFinally(node, stopsContinue, nil) {
					ctx.ReportNode(node, messageUnsafeFinallyContinue)
				}
			},
		}
	},
}

// stopsReturnOrThrow says whether a construct absorbs a `return` or `throw`, or a labeled `break`.
//
// Only a function or class boundary does. This is deliberately the narrowest of the three: a loop
// or a switch does not stop a return, which is why a `return` written inside a loop inside a finally
// block still escapes and still reports.
func stopsReturnOrThrow(node *ast.Node) bool {
	return ast.IsFunctionLikeDeclaration(node) || ast.IsClassLike(node)
}

// stopsUnlabeledBreak says whether a construct absorbs an unlabeled `break`.
//
// A loop or a switch is a legal target for one, so either absorbs it in addition to the function and
// class boundaries.
func stopsUnlabeledBreak(node *ast.Node) bool {
	return stopsReturnOrThrow(node) || ast.IsIterationStatement(node, false) || ast.IsSwitchStatement(node)
}

// stopsContinue says whether a construct absorbs a `continue`, labeled or not.
//
// A loop does; a switch does not, because `continue` is not legal against a switch. That single
// omission is the difference between this and stopsUnlabeledBreak, and it is what makes
// `while (true) try {} finally { switch (true) { case true: continue; } }` report.
func stopsContinue(node *ast.Node) bool {
	return stopsReturnOrThrow(node) || ast.IsIterationStatement(node, false)
}

// escapesIntoFinally says whether a jump from this statement leaves a `finally` block.
//
// Walking up from the statement, the first of two things wins. If a construct that absorbs this kind
// of jump is reached first, the jump never leaves it and the statement is fine. If a finally block is
// reached first, the jump escapes the try statement and the pending action is discarded.
//
// The label argument carries the target of a labeled break or continue. Finding that label on the way
// up means the jump lands inside the finally block rather than escaping it, so it stops the walk the
// same way a function boundary does. Matching by name rather than by symbol is what ESLint does and
// is sound here: a label may not be shadowed by another label of the same name in its own body, so
// the nearest enclosing label with a matching name is the target.
func escapesIntoFinally(node *ast.Node, stopsJump func(*ast.Node) bool, label *ast.Node) bool {
	found := ast.FindAncestorOrQuit(node.Parent, func(ancestor *ast.Node) ast.FindAncestorResult {
		if stopsJump(ancestor) {
			return ast.FindAncestorQuit
		}

		if label != nil && ast.IsLabeledStatement(ancestor) {
			labeledStatement := ancestor.AsLabeledStatement()
			if labeledStatement != nil && labeledStatement.Label != nil &&
				labeledStatement.Label.Text() == label.Text() {
				return ast.FindAncestorQuit
			}
		}

		// The finally block is a Block whose parent is the TryStatement holding it. Checking the
		// parent link rather than the kind is what distinguishes it from the try block and the catch
		// clause's block, which are the same kind and are not hazardous.
		if ancestor.Parent != nil && ast.IsTryStatement(ancestor.Parent) {
			tryStatement := ancestor.Parent.AsTryStatement()
			if tryStatement != nil && tryStatement.FinallyBlock == ancestor {
				return ast.FindAncestorTrue
			}
		}

		return ast.FindAncestorFalse
	})
	return found != nil
}
