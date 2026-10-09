package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoAsyncPromiseExecutor = rule.Message{
	Id: "noAsyncPromiseExecutor",
	Description: "This Promise executor is `async`, which silently discards the errors it is " +
		"there to catch. A rejection thrown inside an async executor becomes a rejection of the " +
		"executor's own promise, which nothing holds, so the constructed Promise never settles " +
		"and the error surfaces as an unhandled rejection instead of reaching the `reject` " +
		"parameter. Needing `await` here usually means the `new Promise` wrapper is unnecessary.",
}

// NoAsyncPromiseExecutor flags an async function passed as the executor of `new Promise`.
//
//	valid:   new Promise((resolve, reject) => {})
//	valid:   new Promise((resolve, reject) => {}, async function unrelated() {})
//	valid:   new Foo(async (resolve, reject) => {})
//	invalid: new Promise(async function foo(resolve, reject) {})
//	invalid: new Promise(async (resolve, reject) => {})
//	invalid: new Promise(((((async () => {})))))
//
// # Three discriminations, each of which a plausible port gets wrong
//
// The executor is argument zero and nothing else. `new Promise(fn, async function unrelated() {})`
// is upstream's own clean case, and a rule scanning every argument reports it: the second argument
// is not the executor, it is whatever the author passed next, and the Promise constructor ignores
// it. Reading only `Arguments.Nodes[0]` is the predicate, not "does any argument look like this".
//
// Both async node kinds are caught. An async arrow and an async function expression are different
// node kinds, and a rule matching one is silent on the other while looking correct against half the
// corpus. Upstream lists both as fail cases for that reason and this asks the question in a form
// that cannot answer for only one: it looks for the `async` modifier rather than for a node kind.
//
// Parentheses are transparent on both sides. `new Promise(((((async () => {})))))` is upstream's
// third fail case, and the callee side has the same exposure through `new (Promise)(...)`, which
// upstream unwraps because `is_specific_id` calls `get_inner_expression` before comparing. Measured
// here rather than assumed: without skipping parens on the callee, that form parses as a
// ParenthesizedExpression and the identifier test simply returns false.
//
// # Why the finding points at the keyword
//
// Upstream reports the five characters of `async` rather than the whole function, which is what
// makes the finding readable and what makes an `eslint-disable-next-line` land on the right line.
// The keyword is a real node, so `ReportNode` on the modifier produces that span with the leading
// trivia already trimmed. Computing it as `function.Pos() + 5` also produces the right answer for
// every corpus case and the wrong one for `new Promise(  async () => {})`, where the node position
// sits before the whitespace: that spelling was measured reporting `  asy`.
//
// # A stated divergence from oxc, in the safe direction
//
// Upstream compares the callee's identifier text to `Promise` with no scope lookup, so a locally
// shadowed `Promise` still reports. ESLint asks `isGlobalReference` and declines. This follows oxc,
// which is the port target, and matches the reasoning already shipped in `no-constant-condition`
// and `no-constant-binary-expression`: the exposure is a false positive on code that has redefined
// `Promise`, which has larger problems, and the alternative needs name resolution this rule
// otherwise has no reason to pay for.
//
// # What it does not catch, deliberately
//
// `Promise(async () => {})` without `new` is not reported, because upstream matches `NewExpression`
// only. The call form throws at runtime regardless, so the executor is not the interesting defect
// there.
var NoAsyncPromiseExecutor = rule.Rule{
	Name: "no-async-promise-executor",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindNewExpression: func(node *ast.Node) {
				newExpression := node.AsNewExpression()

				// Parens on the callee, because `new (Promise)(async () => {})` otherwise parses as
				// a ParenthesizedExpression and fails the identifier test.
				callee := ast.SkipParentheses(newExpression.Expression)
				if !ast.IsIdentifier(callee) || callee.AsIdentifier().Text != "Promise" {
					return
				}

				// `new Promise` with no argument list has a nil Arguments rather than an empty one,
				// which is a panic and not a miss if the length check is trusted to cover it.
				if newExpression.Arguments == nil || len(newExpression.Arguments.Nodes) == 0 {
					return
				}

				// Argument zero is the executor. The rest belong to whoever passed them.
				executor := ast.SkipParentheses(newExpression.Arguments.Nodes[0])
				asyncKeyword := asyncModifierOf(executor)
				if asyncKeyword == nil {
					return
				}

				ctx.ReportNode(asyncKeyword, messageNoAsyncPromiseExecutor)
			},
		}
	},
}

// asyncModifierOf returns the `async` keyword node of a function expression or arrow, or nil.
//
// Returning the node rather than a boolean is what lets the caller report the keyword's own span,
// which is the span upstream reports.
//
// It asks for the modifier rather than calling `ast.IsAsyncFunction`, and that is a measured choice
// rather than a stylistic one. `IsAsyncFunction` answers false for `async function* () {}`, whose
// `async` modifier is plainly present; oxc keeps `generator` and `r#async` as independent fields and
// reads only the second, so upstream reports that spelling and a port built on `IsAsyncFunction`
// would not. The gap is real: an async generator passed as an executor loses errors exactly the way
// an async function does.
//
// # There is no kind gate, and that is measured rather than an oversight
//
// The obvious shape here tests `node.Kind` for arrow or function expression before scanning. That
// branch survived a mutation sweep, so it was read rather than papered over with a fixture, and it
// is subsumed: the scan below already answers every case it would.
//
// Not by the hand-waving version of that argument. Walking every node of a file carrying methods,
// generators, decorators, namespaces and class members, exactly two kinds carry an `async` modifier
// while being neither an arrow nor a function expression: `KindMethodDeclaration` and
// `KindFunctionDeclaration`. So the gate is not vacuous and the language fact alone does not settle
// it. What settles it is that neither kind is an expression, so neither reaches argument zero of a
// `new` expression. Checked through error recovery too, where the assumption would break if it
// broke anywhere: `new Promise(async run() {})` parses argument zero as an Identifier rather than as
// the method it looks like.
//
// Everything else with an `async` keyword nearby carries it on a child. An object literal and a
// class expression both return a nil modifier list of their own, because the keyword belongs to the
// member, and a member is never the executor.
//
// A gate that no input can distinguish is a branch that reads as a case somebody handled, so it is
// gone rather than kept with a test that asserts nothing.
func asyncModifierOf(node *ast.Node) *ast.Node {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return nil
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindAsyncKeyword {
			return modifier
		}
	}
	return nil
}
