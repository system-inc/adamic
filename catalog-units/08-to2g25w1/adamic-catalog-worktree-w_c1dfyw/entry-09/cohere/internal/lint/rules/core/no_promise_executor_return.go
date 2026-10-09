package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoPromiseExecutorReturn = rule.Message{
	Id: "returnsValue",
	Description: "This returns a value from a promise executor, and nothing can ever read it. " +
		"The Promise constructor ignores whatever the executor returns, so a value here is either " +
		"dead code or, more often, a missing `resolve` call the author believed was implicit. The " +
		"usual cause is `return fetch(...)` where `resolve(fetch(...))` was meant, which leaves " +
		"the promise pending forever.",
}

var messageNoPromiseExecutorReturnPrependVoid = rule.Message{
	Id:          "prependVoid",
	Description: "Prepend `void` to the expression, saying the value is deliberately discarded.",
}

var messageNoPromiseExecutorReturnWrapBraces = rule.Message{
	Id:          "wrapBraces",
	Description: "Wrap the expression in `{}`, making the arrow body a statement rather than a return.",
}

// NoPromiseExecutorReturnOptions carries upstream's one option.
//
// `AllowVoid` accepts `void expr` as the explicit way to say the value is discarded. It defaults to
// FALSE, so the zero value is the safe direction here; the pointer is kept anyway so an absent key
// stays distinguishable from an explicit false, and so the decoder's own behaviour is testable.
type NoPromiseExecutorReturnOptions struct {
	AllowVoid *bool `json:"allowVoid"`
}

// DecodeNoPromiseExecutorReturnOptions turns the configured object into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` so empty input returns the defaults instead of an
// error the config layer turns into a nil the rule then reads as a zero value. That path is harmless
// for this rule, whose one option defaults to false, and it is written out anyway because the
// harmlessness is a property of today's default rather than of the code.
//
// The configured shape is the bare object rather than upstream's `[{...}]`: cohere's config layer
// strips the severity-and-options tuple before dispatch.
func DecodeNoPromiseExecutorReturnOptions(raw []byte) (any, error) {
	var options NoPromiseExecutorReturnOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// NoPromiseExecutorReturn flags a value returned from a `new Promise` executor.
//
//	valid:   new Promise(function (resolve, reject) { return; })
//	valid:   new Promise(function (resolve, reject) { function foo() { return 1; } })
//	valid:   let Promise; new Promise(function (resolve, reject) { return 1; })
//	invalid: new Promise(function (resolve, reject) { return 1; })
//	invalid: new Promise((resolve, reject) => 1)
//
// The Promise constructor ignores the executor's return value, so returning one is dead at best and
// a missing `resolve` at worst.
//
// # Three questions, and the middle one is the port
//
// Is this function a promise executor: it is the FIRST argument to a `new` on an identifier spelled
// `Promise` that resolves to the global. `resolvesToAGlobal` answers the last part, probed against
// all six of upstream's shadowing cases before being built on: a `let`, a `var` declared after the
// call, a parameter, a block-scoped `const`, and a named function expression each resolve to one
// declaration outside a declaration file, while the real global resolves to five inside one.
//
// Does a `return` belong to THAT function rather than to one nested inside it. Upstream maintains a
// stack through `onCodePathStart` and `onCodePathEnd`, pushing a frame per function and recording
// whether that frame is an executor. There is no code-path event here and the walk is pre-order, so
// the same question is answered by walking UP from the return statement to its nearest enclosing
// function and asking whether that one is an executor. Same decision, and it maintains no state, so
// it cannot drift out of sync with the walk the way a stack pushed and popped by two separate events
// can. The four "does not return from the promise executor" clean cases are what pin it.
//
// `enclosingFunctionOf` in `array_callback_return_callee.go` already answers that, as upstream's own
// `getUpperFunction`, and `unwrapParentheses` in `valid_typeof.go` already peels grouping. Both were
// found by a build-time name collision after writing duplicates, and both existing ones are used
// rather than respelled: `enclosingFunctionOf` additionally stops at the source file, which is
// stricter than the version this rule would have shipped.
//
// Does the arrow have a concise body, which is a return without the keyword.
//
// # Parentheses, where our parser and upstream disagree in both directions at once
//
// Espree folds a parenthesized expression away, so upstream's `node.body` for
// `new Promise(r => (1))` IS the literal `1`. typescript-go produces a real
// `KindParenthesizedExpression`. Measured against the installed eslint at 10.8.1, that difference
// lands on two surfaces that want opposite answers:
//
//	span            reports `1`, unwrapping EVERY layer: `((1))` also reports `1`
//	wrapBraces fix  writes `{(1)}`, wrapping the OUTERMOST node, parentheses included
//	prependVoid fix writes `void (1)`, treating the existing parens as satisfying its own need
//
// So the reported node is the unwrapped expression and the brace anchors are the body node as
// written. Unwrapping is a loop rather than one step because `((1))` nests, and it is written out
// rather than reaching for `ast.SkipParentheses`, which dereferences its argument.
//
// # Suggestions, not fixes
//
// Upstream sets `hasSuggestions` and no `fixable`, and the distinction is load-bearing: both repairs
// CHANGE WHAT THE CODE MEANS. Wrapping in braces discards a value that was being returned, and
// prepending `void` discards it explicitly. Neither is a spelling change, so neither may be applied
// unattended.
//
// Which suggestions are offered is itself upstream data, and it has three answers rather than two.
// `suggestions: null` on a `return` statement without `allowVoid` means nothing is offered, because
// braces cannot help a statement that already has them. An EMPTY list appears for exactly two
// inputs, an unnamed function expression and an unnamed class expression as a concise body: upstream
// considers `wrapBraces` and declines, because `{function () {}}` is a block containing a function
// DECLARATION with no name, which is a syntax error. A named one is fine and gets the suggestion.
var NoPromiseExecutorReturn = rule.Rule{
	Name:             "no-promise-executor-return",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as bare "error" is handed nil, so the type assertion yields the zero
		// value. Correct here only because the option defaults to false; written out so a later
		// default change cannot invert the rule silently.
		settings, _ := rule.OptionsAs[NoPromiseExecutorReturnOptions](options)
		allowVoid := settings.AllowVoid != nil && *settings.AllowVoid

		reportArrowBody := func(node *ast.Node) {
			// Without this the rule goes completely silent under the plain harness rather than
			// crashing, which is the more dangerous failure: every clean fixture would pass
			// vacuously. `TestNoPromiseExecutorReturnRequiresTheTypedHarness` pins it.
			//
			// A mutation removing THIS guard survives, and that is subsumption rather than a blind
			// spot: the only checker use on this path is inside `isPromiseExecutorFunction`, which
			// reaches `resolvesToAGlobal`, whose own first line returns false on a nil checker. So
			// no input can distinguish the two versions today. Widening the typed-harness test to
			// cover both listeners did not kill it, which confirmed the mechanism rather than a
			// missing case.
			//
			// Kept anyway. The verdict names the callers that exist right now and expires the moment
			// a second checker call is added to this listener, and a typed rule reading the checker
			// without guarding is the shape three shipped rules got wrong.
			if ctx.TypeChecker == nil {
				return
			}
			arrow := node.AsArrowFunction()
			body := arrow.Body
			if body == nil || body.Kind == ast.KindBlock {
				return
			}
			if !isPromiseExecutorFunction(ctx, node) {
				return
			}
			// `=> void 0` is already saying the value is discarded, which is what allowVoid asks
			// for.
			if allowVoid && unwrapParentheses(body).Kind == ast.KindVoidExpression {
				return
			}

			suggestions := []rule.Suggestion{}
			if allowVoid {
				suggestions = append(suggestions, promiseExecutorPrependVoidSuggestion(ctx, body))
			}
			if promiseExecutorCanWrapInBraces(body) {
				suggestions = append(suggestions, promiseExecutorWrapBracesSuggestion(ctx, body))
			}

			// The reported node is the UNWRAPPED expression while the brace anchors above are the
			// body as written, which is the parser difference the doc comment describes.
			ctx.ReportNodeWithSuggestions(unwrapParentheses(body), messageNoPromiseExecutorReturn,
				suggestions...)
		}

		return rule.Listeners{
			ast.KindArrowFunction: reportArrowBody,
			ast.KindReturnStatement: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				argument := node.AsReturnStatement().Expression
				// `return;` with no value is allowed, which is upstream's `node.argument` test and
				// five of its clean cases.
				if argument == nil {
					return
				}
				enclosing := enclosingFunctionOf(node)
				if enclosing == nil || !isPromiseExecutorFunction(ctx, enclosing) {
					return
				}

				if !allowVoid {
					// Upstream offers nothing here: `suggestions: null`. Braces cannot help a
					// statement that already has them, and `void` is what allowVoid is for.
					ctx.ReportNode(node, messageNoPromiseExecutorReturn)
					return
				}
				if unwrapParentheses(argument).Kind == ast.KindVoidExpression {
					return
				}
				ctx.ReportNodeWithSuggestions(node, messageNoPromiseExecutorReturn,
					promiseExecutorPrependVoidSuggestion(ctx, argument))
			},
		}
	},
}

// isPromiseExecutorFunction reports whether a function node is the executor of a `new Promise`.
//
// Upstream's four conditions, in the same order: the parent is a `new`, this node is its FIRST
// argument, the callee is a bare identifier spelled `Promise`, and that identifier resolves to the
// global rather than to something in scope.
//
// The first-argument test is what keeps `new Promise(foo, (resolve, reject) => 1)` clean, and the
// identifier test is what keeps `new Promise.foo(...)` and `new foo.Promise(...)` clean, since
// neither callee is an identifier at all.
func isPromiseExecutorFunction(ctx rule.Context, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindNewExpression {
		return false
	}
	newExpression := parent.AsNewExpression()
	// `new Promise` with no argument list at all parses with nil Arguments, so the nil test comes
	// before the index.
	if newExpression.Arguments == nil || len(newExpression.Arguments.Nodes) == 0 {
		return false
	}
	if newExpression.Arguments.Nodes[0] != node {
		return false
	}
	callee := newExpression.Expression
	if callee == nil || !ast.IsIdentifier(callee) || callee.Text() != "Promise" {
		return false
	}
	return resolvesToAGlobal(ctx, callee)
}

// promiseExecutorCanWrapInBraces reports whether `{}` around a concise body would still parse.
//
// Upstream declines for an unnamed function expression and an unnamed class expression, and its
// corpus asserts an EMPTY suggestion list for exactly those two. The reason is that `{function () {}}`
// is a block holding a function DECLARATION, which must have a name, so the rewrite would not parse.
// A named one is fine: `{function foo() {}}` is legal and upstream offers it.
//
// The test is on the body as written rather than unwrapped, matching upstream, which reads
// `node.body.type` directly. That distinction is measurable: `new Promise(r => (function () {}))`
// carries parentheses, so upstream sees a parenthesized expression rather than a function and offers
// the suggestion.
func promiseExecutorCanWrapInBraces(body *ast.Node) bool {
	switch body.Kind {
	case ast.KindFunctionExpression:
		return body.AsFunctionExpression().Name() != nil
	case ast.KindClassExpression:
		return body.AsClassExpression().Name() != nil
	}
	return true
}

// promiseExecutorWrapBracesSuggestion wraps a concise arrow body in braces.
//
// Anchored on the body AS WRITTEN rather than unwrapped, which is what makes
// `new Promise(r => (1))` produce `{(1)}` rather than `({1})`. Upstream anchors on the tokens either
// side of the arrow, arriving at the same span for a different reason.
func promiseExecutorWrapBracesSuggestion(ctx rule.Context, body *ast.Node) rule.Suggestion {
	return rule.Suggestion{
		Message: messageNoPromiseExecutorReturnWrapBraces,
		Fixes: []rule.Fix{
			ctx.InsertBefore(body, "{"),
			ctx.InsertAfter(body, "}"),
		},
	}
}

// promiseExecutorPrependVoidSuggestion prepends `void` to an expression.
//
// Two narrowings, both upstream's and both measured against the installed build rather than
// reasoned about.
//
// A leading space is needed when the expression is written adjacent to `return`, because
// `returnvoid x` is one identifier. Upstream compares the return token's end against the first
// token's start and its corpus pins it: `new Promise(r => { return(1) })` must become
// `return void (1)`. An arrow's `=>` needs no such space, which is why upstream tests for the
// `return` keyword specifically rather than for adjacency alone.
//
// Parentheses are added when the expression binds looser than `void` and is not already
// parenthesized. Measured across the operators upstream's corpus and this tree write: assignment,
// conditional and every binary operator need them, while prefix and postfix unary, `new`, a member
// access, an arrow, an array literal and an already-parenthesized expression do not.
func promiseExecutorPrependVoidSuggestion(ctx rule.Context, expression *ast.Node) rule.Suggestion {
	prefix := "void "
	if promiseExecutorIsAdjacentToReturn(ctx, expression) {
		prefix = " void "
	}

	if promiseExecutorNeedsParensUnderVoid(expression) {
		return rule.Suggestion{
			Message: messageNoPromiseExecutorReturnPrependVoid,
			Fixes: []rule.Fix{
				ctx.InsertBefore(expression, prefix+"("),
				ctx.InsertAfter(expression, ")"),
			},
		}
	}
	return rule.Suggestion{
		Message: messageNoPromiseExecutorReturnPrependVoid,
		Fixes:   []rule.Fix{ctx.InsertBefore(expression, prefix)},
	}
}

// promiseExecutorIsAdjacentToReturn reports whether an expression is written with no space after the
// `return` keyword before it.
//
// `ctx.InsertBefore` anchors on the token range, which starts past leading trivia, so the check is
// whether the source immediately before that token is the end of `return`.
func promiseExecutorIsAdjacentToReturn(ctx rule.Context, expression *ast.Node) bool {
	parent := expression.Parent
	if parent == nil || parent.Kind != ast.KindReturnStatement {
		return false
	}
	start := rule.TokenRange(ctx.SourceFile, expression).Pos()
	const keyword = "return"
	if start < len(keyword) {
		return false
	}
	return ctx.SourceFile.Text()[start-len(keyword):start] == keyword
}

// promiseExecutorNeedsParensUnderVoid reports whether prepending `void` would bind tighter than the
// expression, so the result needs parentheses to keep its meaning.
//
// `void` is a unary operator, so anything looser than unary needs wrapping. An already-parenthesized
// expression does not, which is upstream's `isParenthesised` check and which the corpus pins:
// `new Promise(r => (1 ? 2 : 3))` produces `void (1 ? 2 : 3)` with no second pair.
//
// The membership was MEASURED against the installed eslint at 10.8.1 rather than reasoned from a
// precedence table, and the arrow row is the one that reading got wrong. An arrow function binds
// looser than `void`, so `r => () => {}` must become `void (() => {})`; a neighbouring rule in this
// package narrows the same predicate to binary, conditional and yield, and copying it shipped a
// suggestion that produced `void () => {}`, which does not parse. Upstream's own corpus caught it.
//
// Measured needing parentheses: an arrow (plain and async), every binary operator including logical
// and nullish, a conditional, and an assignment.
//
// Measured NOT needing them: a function expression, a class expression, an array literal, prefix and
// postfix update, `typeof`, `new`, a tagged template, an optional member access, and unary minus.
func promiseExecutorNeedsParensUnderVoid(expression *ast.Node) bool {
	switch expression.Kind {
	case ast.KindParenthesizedExpression:
		return false
	case ast.KindBinaryExpression,
		ast.KindConditionalExpression,
		ast.KindYieldExpression,
		ast.KindArrowFunction:
		return true
	}
	return false
}
