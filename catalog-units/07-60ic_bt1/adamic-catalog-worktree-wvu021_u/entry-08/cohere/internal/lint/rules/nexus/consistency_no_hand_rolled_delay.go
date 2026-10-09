package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const consistencyNoHandRolledDelayId = "handRolledDelay"

// consistencyNoHandRolledDelayText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-hand-rolled-delay.json`.
var consistencyNoHandRolledDelayText = policy.MessageOf("nexus/consistency-no-hand-rolled-delay", consistencyNoHandRolledDelayId)

func consistencyNoHandRolledDelayMessage() rule.Message {
	return rule.Message{Id: consistencyNoHandRolledDelayId, Description: consistencyNoHandRolledDelayText.Render(nil)}
}

// ConsistencyNoHandRolledDelay reports a promise built only to resolve after a timer, where Nexus
// `delay` says the same thing by name.
//
//	invalid: await new Promise((resolve) => setTimeout(resolve, intervalMilliseconds));
//	invalid: await new Promise(function(resolve) { setTimeout(resolve, 2000); });
//	invalid: await new Promise((resolve) => setTimeout(() => resolve(), 500));
//	valid:   await delay(intervalMilliseconds);
//	valid:   new Promise((resolve) => setTimeout(() => resolve('slow'), 50));
//	valid:   new Promise<void>((resolve) => { const timer = setTimeout(resolve, remaining); this.waiting = ...; });
//
// # Name
//
// `consistency-` because the judgment is one spelling for one act, the family the other house rules
// about shape live in, and `no-hand-rolled-delay` because the thing reported is a reimplementation of
// a named primitive: the name says what to reach for (`delay`) and that the problem is writing it out,
// not waiting. `prefer-nexus-delay` was the review's working name; Nexus rules here are named for
// what they forbid.
//
// # The shape, exactly
//
// Every condition must hold, and each one is a way the promise could be doing something `delay`
// does not:
//
//  1. `new Promise(executor)`, with optional type arguments and exactly one argument. The callee is
//     the bare identifier `Promise`.
//  2. The executor is an arrow function or a function expression (named or not) with one or two
//     plain identifier parameters. A second parameter (`reject`) is allowed only when the body never
//     mentions it, since `delay` never rejects.
//  3. The body is exactly the timer: an expression body, or a block holding a single expression
//     statement or a single `return`. Any second statement (`webSocket.close()` beside a fallback
//     timer, a handle stored for `clearTimeout`, a `this.waiting = ...` hook) means the promise
//     resolves on something besides the timer, and it is not a delay.
//  4. The timer is `setTimeout`, `globalThis.setTimeout` or `window.setTimeout`, called with exactly
//     two arguments. A third argument is passed to the callback, so `setTimeout(resolve, ms, value)`
//     resolves with a value; one argument has no duration to hand to `delay`.
//  5. The callback resolves with nothing: it is the `resolve` parameter itself, or a parameterless
//     arrow or function whose body is `resolve()` with no arguments (an expression body or a single
//     statement). `resolve('slow')` is a timed value, the test fixtures' shape, and not a delay.
//
// # The definition of `delay` is not a use of it
//
// `delay` itself is written as this shape, and it must not report, without the rule knowing its
// path or its name. The structural difference is that a definition hands the timer's duration in
// from outside: the promise is the whole body of a function (its single `return`, or an arrow's
// expression body) and the duration is a bare reference to one of that function's own parameters.
// That function is a delay primitive, and the promise inside it is what the primitive is. A
// function with a fixed duration (`return new Promise((resolve) => setTimeout(resolve, 1000))`) is
// not a definition and reports, because it is a use with a name around it.
//
// What this does not catch is a second definition of the primitive somewhere else (`function
// sleep(ms) { return new Promise(...) }`), which is a duplicate helper rather than a hand-rolled
// wait: its callers already read as a named wait. Measured on ahra at HEAD on 2026-10-01: the only
// function of that shape is `delay` in `nexus/source/coordination/Delay.ts`.
//
// # No fix
//
// The rewrite needs an import of `delay`, from a path that depends on the file's package (Nexus
// itself imports it relatively), and the replacement is an expression that resolves `undefined`
// where the original may be typed `Promise<unknown>` or used as a value. Both are easy for a person
// and not certain for a rewrite, so the finding names the import and the author writes it.
var ConsistencyNoHandRolledDelay = rule.Rule{
	Name: "nexus/consistency-no-hand-rolled-delay",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindNewExpression: func(node *ast.Node) {
				duration := consistencyNoHandRolledDelayDuration(node)
				if duration == nil {
					return
				}
				if consistencyNoHandRolledDelayIsDefinition(node, duration) {
					return
				}
				ctx.ReportNode(node, consistencyNoHandRolledDelayMessage())
			},
		}
	},
}

// consistencyNoHandRolledDelayDuration returns the timer's duration argument when the `new`
// expression is a plain timed wait, and nil otherwise.
func consistencyNoHandRolledDelayDuration(node *ast.Node) *ast.Node {
	newExpression := node.AsNewExpression()
	if !consistencyNoHandRolledDelayIsIdentifier(newExpression.Expression, "Promise") {
		return nil
	}
	if newExpression.Arguments == nil || len(newExpression.Arguments.Nodes) != 1 {
		return nil
	}
	executor := newExpression.Arguments.Nodes[0]
	if executor.Kind != ast.KindArrowFunction && executor.Kind != ast.KindFunctionExpression {
		return nil
	}

	parameters := executor.Parameters()
	if len(parameters) < 1 || len(parameters) > 2 {
		return nil
	}
	// A destructured or rest first parameter yields "", which no identifier below can match, so it
	// needs no early exit of its own (a mutant removing one survived every fixture).
	resolveName := consistencyNoHandRolledDelayParameterName(parameters[0])

	call := consistencyNoHandRolledDelaySoleExpression(executor.Body())
	if call == nil || call.Kind != ast.KindCallExpression {
		return nil
	}
	timer := call.AsCallExpression()
	if !consistencyNoHandRolledDelayIsTimer(timer.Expression) {
		return nil
	}
	if timer.Arguments == nil || len(timer.Arguments.Nodes) != 2 {
		return nil
	}
	if !consistencyNoHandRolledDelayResolvesWithNothing(timer.Arguments.Nodes[0], resolveName) {
		return nil
	}

	if len(parameters) == 2 {
		rejectName := consistencyNoHandRolledDelayParameterName(parameters[1])
		if rejectName == "" || consistencyNoHandRolledDelayMentions(executor.Body(), rejectName) {
			return nil
		}
	}
	return timer.Arguments.Nodes[1]
}

// consistencyNoHandRolledDelayIsDefinition says whether the promise is the whole body of a function
// whose own parameter is the duration, which is the definition of a delay primitive rather than a use.
func consistencyNoHandRolledDelayIsDefinition(node *ast.Node, duration *ast.Node) bool {
	if duration.Kind != ast.KindIdentifier {
		return false
	}
	var function *ast.Node
	parent := node.Parent
	switch {
	case parent.Kind == ast.KindArrowFunction && parent.Body() == node:
		function = parent
	case parent.Kind == ast.KindReturnStatement && parent.Parent.Kind == ast.KindBlock &&
		len(parent.Parent.AsBlock().Statements.Nodes) == 1 && ast.IsFunctionLike(parent.Parent.Parent):
		function = parent.Parent.Parent
	default:
		return false
	}
	durationName := duration.Text()
	for _, parameter := range function.Parameters() {
		if consistencyNoHandRolledDelayParameterName(parameter) == durationName {
			return true
		}
	}
	return false
}

// consistencyNoHandRolledDelayParameterName is a parameter's name when it is a plain identifier that
// is not a rest parameter, and "" otherwise.
//
// A default value is allowed. The executor's parameters are always supplied by `Promise`, so a default
// on `resolve` or `reject` never applies, and a default on a delay primitive's duration
// (`function delay(milliseconds = 100)`) leaves it a definition. A rest parameter binds an array, which
// is never the function the timer calls.
func consistencyNoHandRolledDelayParameterName(parameter *ast.Node) string {
	declaration := parameter.AsParameterDeclaration()
	if declaration.DotDotDotToken != nil {
		return ""
	}
	name := declaration.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.Text()
}

// consistencyNoHandRolledDelaySoleExpression is the one expression a function body consists of: the
// expression body itself, or the expression of a block's single expression statement or return.
func consistencyNoHandRolledDelaySoleExpression(body *ast.Node) *ast.Node {
	if body == nil {
		return nil
	}
	if body.Kind != ast.KindBlock {
		return ast.SkipParentheses(body)
	}
	statements := body.AsBlock().Statements
	if statements == nil || len(statements.Nodes) != 1 {
		return nil
	}
	statement := statements.Nodes[0]
	switch statement.Kind {
	case ast.KindExpressionStatement:
		return ast.SkipParentheses(statement.AsExpressionStatement().Expression)
	case ast.KindReturnStatement:
		if expression := statement.AsReturnStatement().Expression; expression != nil {
			return ast.SkipParentheses(expression)
		}
	}
	return nil
}

// consistencyNoHandRolledDelayIsTimer says whether a callee is the global `setTimeout`, bare or reached
// through `globalThis` or `window`.
func consistencyNoHandRolledDelayIsTimer(callee *ast.Node) bool {
	callee = ast.SkipParentheses(callee)
	if consistencyNoHandRolledDelayIsIdentifier(callee, "setTimeout") {
		return true
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil || access.Name().Text() != "setTimeout" {
		return false
	}
	return consistencyNoHandRolledDelayIsIdentifier(access.Expression, "globalThis") ||
		consistencyNoHandRolledDelayIsIdentifier(access.Expression, "window")
}

// consistencyNoHandRolledDelayResolvesWithNothing says whether the timer's callback resolves the
// promise with no value: `resolve` itself, or a parameterless function whose whole body is `resolve()`.
func consistencyNoHandRolledDelayResolvesWithNothing(callback *ast.Node, resolveName string) bool {
	callback = ast.SkipParentheses(callback)
	if consistencyNoHandRolledDelayIsIdentifier(callback, resolveName) {
		return true
	}
	if callback.Kind != ast.KindArrowFunction && callback.Kind != ast.KindFunctionExpression {
		return false
	}
	if len(callback.Parameters()) != 0 {
		return false
	}
	call := consistencyNoHandRolledDelaySoleExpression(callback.Body())
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	resolveCall := call.AsCallExpression()
	return consistencyNoHandRolledDelayIsIdentifier(resolveCall.Expression, resolveName) &&
		(resolveCall.Arguments == nil || len(resolveCall.Arguments.Nodes) == 0)
}

// consistencyNoHandRolledDelayMentions says whether any identifier in the subtree is spelled name.
func consistencyNoHandRolledDelayMentions(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier && node.Text() == name {
		return true
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		found = consistencyNoHandRolledDelayMentions(child, name)
		return found
	})
	return found
}

func consistencyNoHandRolledDelayIsIdentifier(node *ast.Node, name string) bool {
	return node != nil && node.Kind == ast.KindIdentifier && node.Text() == name
}
