package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageSymbolDescription = rule.Message{
	Id: "expected",
	Description: "This creates a symbol with no description, so it prints as `Symbol()` in every " +
		"stack trace, log line and debugger view, and two unrelated symbols are indistinguishable " +
		"from each other. The description costs nothing and is the only label the value will ever " +
		"carry. Pass one.",
}

// SymbolDescription flags a call to the global `Symbol` with no arguments.
//
//	valid:   Symbol("Foo");
//	valid:   var Symbol = function () {}; Symbol();
//	valid:   new Symbol();
//	invalid: Symbol();
//	invalid: let x = Symbol();
//
// # It is a scope question, not a call-shape one
//
// Four of upstream's six clean cases are shadowing, and a rule matching the name textually reports
// every one. Upstream asks `eslint-scope` for the variable named `Symbol` and proceeds only when it
// has NO definitions, which is its way of saying "this is the global one". `resolvesToAGlobal` in
// this package asks the same question through the checker: it resolves the identifier and answers
// true only when every declaration behind it lives in a declaration file.
//
// Probed against all eight of upstream's cases rather than taken from the helper's name. The four
// shadowing cases each resolve to a single non-ambient declaration and the global cases each resolve
// to four ambient ones from the standard library. The verdicts agree on all eight, and on eight more
// beyond the corpus.
//
// # Three decisions upstream makes that read as oversights
//
// All three measured against the installed build at 10.8.1, and all three reproduced rather than
// improved on.
//
// `new Symbol()` is CLEAN. It throws at run time and carries no description either, but upstream's
// `isCallee` is false for a `new` expression, so the rule never sees it. Anchoring on
// `KindCallExpression` alone reproduces that.
//
// `Symbol(undefined)` and `Symbol(...args)` are CLEAN. The check is `arguments.length === 0`, so a
// spread that expands to nothing and an explicit undefined both pass. Counting arguments rather than
// inspecting them is the decision.
//
// `globalThis.Symbol()` is CLEAN, because there the identifier is a property rather than the callee.
//
// # Parentheses
//
// Espree gives a parenthesized expression no node, so upstream's `isCallee` already sees through
// them and reports `(Symbol)()` at columns 1 to 11. typescript-go produces one, so the callee is
// skipped through before it is resolved, and the span stays on the whole call including those
// parentheses.
var SymbolDescription = rule.Rule{
	Name:             "symbol-description",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// Without this the rule goes completely silent under the plain harness rather than
				// crashing, which is the more dangerous failure: every clean fixture would pass
				// vacuously. `TestSymbolDescriptionRequiresTheTypedHarness` pins it.
				//
				// A mutation removing this guard SURVIVED the whole fixture set, and that is
				// subsumption rather than a blind spot: the only checker use in this listener is
				// inside `resolvesToAGlobal`, whose own first line returns false on a nil checker.
				// So no input can distinguish the two versions today. Kept anyway, because the
				// verdict names the callers that exist right now and expires the moment a second
				// checker call is added here, and because a typed rule that reads the checker
				// without guarding is the shape three shipped rules got wrong.
				if ctx.TypeChecker == nil {
					return
				}

				call := node.AsCallExpression()
				if call.Expression == nil {
					return
				}
				// Arguments first, because it is the cheap half and most calls in a real file fail
				// it. `Arguments` is nil for a call the parser recovered without an argument list.
				if call.Arguments != nil && len(call.Arguments.Nodes) != 0 {
					return
				}

				// `SkipParentheses` DEREFERENCES its argument, so the nil test above comes first.
				callee := ast.SkipParentheses(call.Expression)
				if !ast.IsIdentifier(callee) || callee.Text() != "Symbol" {
					return
				}
				if !resolvesToAGlobal(ctx, callee) {
					return
				}

				ctx.ReportNode(node, messageSymbolDescription)
			},
		}
	},
}
