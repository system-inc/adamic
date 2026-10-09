package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageGuardForIn = rule.Message{
	Id: "wrap",
	Description: "This `for-in` loop visits inherited properties as well as the object's own. " +
		"`for-in` walks the whole prototype chain, so anything added to `Object.prototype` by a " +
		"library, a polyfill, or older code shows up as another iteration, and the body runs on a " +
		"key the object never declared. The usual symptom is a serializer emitting a phantom " +
		"field or a counter that is one too high, on a machine where some other package loaded " +
		"first. Guard the body with `if (Object.prototype.hasOwnProperty.call(o, key))`, or use " +
		"`Object.keys(o)`, which returns own enumerable keys and needs no guard.",
}

// GuardForIn flags a `for-in` loop whose body does not begin with a filter.
//
//	valid:   for (var x in o);
//	valid:   for (var x in o) {}
//	valid:   for (var x in o) if (x) f();
//	valid:   for (var x in o) { if (x) { f(); } }
//	valid:   for (var x in o) { if (x) continue; f(); }
//	valid:   for (var x in o) { if (x) { continue; } f(); }
//	invalid: for (var x in o) foo();
//	invalid: for (var x in o) { foo() }
//	invalid: for (var x in o) { if (x) f(); g(); }
//	invalid: for (var x in o) { if (x) { f(); continue; } g(); }
//
// # The predicate is structural, and deliberately shallow
//
// Upstream does not ask whether the body actually calls `hasOwnProperty`; it asks whether the body
// has the SHAPE of a guarded loop. Any `if` will do, whatever it tests. That is a real decision
// rather than an approximation, and reproducing it means not improving on it: a rule checking for a
// `hasOwnProperty` call specifically would report every loop guarded by `if (!o[k]) continue;` or
// by a type test, which upstream accepts.
//
// Five shapes exempt, and the ordering below is upstream's:
//
//	an empty statement body      nothing runs, so nothing runs unguarded
//	an `if` as the whole body    the guard is the body
//	an empty block               same as the empty statement, wearing braces
//	a block holding only an `if` the same guard, wearing braces
//	a block STARTING with an `if` whose consequent is exactly one `continue`
//
// The fifth is the early-exit idiom and it is the only one that inspects the `if`'s consequent. The
// reason it must is visible in the first two failing cases:
//
//	for (var x in o) { if (x) { f(); continue; } g(); }
//	for (var x in o) { if (x) { continue; f(); } g(); }
//
// Both hold a `continue` inside the leading `if`, and neither is the idiom, because `f()` runs on
// the unfiltered key in the first and the block is not a bare skip in the second. Upstream requires
// the consequent to be a `continue` outright, or a block whose ONLY statement is a `continue`, and
// a port testing "does the consequent contain a continue" reports neither of these two while
// upstream reports both.
//
// Note the asymmetry the corpus pins: the fourth exemption accepts a block whose only statement is
// an `if` with ANY consequent, while the fifth accepts a longer block only when the leading `if`
// skips. `{ if (x) { f(); } }` is clean and `{ if (x) { f(); } g(); }` reports, and the difference
// is entirely the trailing `g()`.
//
// No fix. The repair is a guard, and which guard depends on what the loop meant: `hasOwnProperty`,
// a key filter, or a rewrite to `Object.keys`. The rule cannot know which, and each changes what
// the loop does.
var GuardForIn = rule.Rule{
	Name: "guard-for-in",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// `for-of` is deliberately absent. It iterates values through the iteration protocol
			// and never walks the prototype chain, so it has nothing to guard. The two share
			// `ForInOrOfStatement` in our parser and a listener keyed on the wrong one would report
			// every `for-of` in the tree.
			ast.KindForInStatement: func(node *ast.Node) {
				if bodyGuardsForIn(node.AsForInOrOfStatement().Statement) {
					return
				}
				ctx.ReportNode(node, messageGuardForIn)
			},
		}
	},
}

// bodyGuardsForIn reports whether a `for-in` body has the shape upstream accepts.
//
// The arms are upstream's five exemptions in its own order. Written as a single function returning
// a verdict rather than as early returns in the listener, so the listener holds the reporting and
// this holds the whole decision.
func bodyGuardsForIn(body *ast.Node) bool {
	// No nil guard, and that is measured rather than assumed. A mutant flipping one to `return
	// false` survived the whole fixture set, and the reason is that the parser never hands this a
	// nil. Probed with six truncated inputs -- `for (var x in o)`, `for (var x in )`, `for ( in o)`,
	// `for (var x in o) else`, `for (var x in o)}` and the well-formed `for (var x in o);` -- error
	// recovery synthesizes a `KindExpressionStatement` in the body slot for all five malformed ones
	// and a `KindEmptyStatement` for the last. Never nil.
	//
	// The verdict names the callers that existed when it was taken: the one listener above, which
	// reads `AsForInOrOfStatement().Statement`. A caller reaching this from anywhere else voids it.
	switch body.Kind {
	// `for (var x in o);` runs nothing.
	case ast.KindEmptyStatement:
		return true

	// `for (var x in o) if (x) f();` is the guard written without braces.
	case ast.KindIfStatement:
		return true

	case ast.KindBlock:
		statements := body.AsBlock().Statements
		if statements == nil || len(statements.Nodes) == 0 {
			// `for (var x in o) {}` runs nothing.
			return true
		}

		first := statements.Nodes[0]
		if first.Kind != ast.KindIfStatement {
			// A block not starting with a filter runs its first statement on every key,
			// inherited ones included.
			return false
		}

		if len(statements.Nodes) == 1 {
			// A block holding only an `if`, whatever the `if` does. Upstream does not inspect the
			// consequent here, only in the longer-block arm below.
			return true
		}

		// A longer block, so the leading `if` has to be an early exit rather than merely a filter
		// over part of the work. Anything after a non-skipping `if` runs unfiltered.
		return isBareContinue(first.AsIfStatement().ThenStatement)
	}

	return false
}

// isBareContinue reports whether a statement is a `continue`, or a block whose only statement is one.
//
// Upstream's test is exactly these two spellings and nothing looser. A consequent block holding a
// `continue` alongside anything else is not a skip, in either order, which is what makes
// `{ if (x) { f(); continue; } g(); }` and `{ if (x) { continue; f(); } g(); }` both report.
func isBareContinue(consequent *ast.Node) bool {
	// No nil guard here either, for the same measured reason as `bodyGuardsForIn` above. Probed
	// with `if (x) }`, `if (x) else g();` and a bare `if (x)`: error recovery puts a synthesized
	// `KindExpressionStatement` in the then slot every time, and `if (x) ;` gives a
	// `KindEmptyStatement`. Never nil. A mutant flipping a nil guard to `return true` survived the
	// whole fixture set, which is what sent us to probe the parse rather than write the fixture.
	//
	// The verdict names its callers: the single call in `bodyGuardsForIn`, reading
	// `AsIfStatement().ThenStatement` on a node already known to be `KindIfStatement`.
	if consequent.Kind == ast.KindContinueStatement {
		return true
	}

	if consequent.Kind != ast.KindBlock {
		return false
	}

	statements := consequent.AsBlock().Statements
	return statements != nil && len(statements.Nodes) == 1 &&
		statements.Nodes[0].Kind == ast.KindContinueStatement
}
