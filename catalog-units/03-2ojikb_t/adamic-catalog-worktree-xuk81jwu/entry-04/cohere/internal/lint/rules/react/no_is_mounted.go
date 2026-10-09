package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoIsMounted = rule.Message{
	Id: "noIsMounted",
	Description: "`isMounted` is called on `this`. It was removed from React and never existed " +
		"on a function component, so the call either throws or silently reads `undefined`. " +
		"Where it does still run, it is an anti-pattern: it hides a leak rather than fixing " +
		"one, because the only reason to ask whether a component is still mounted is that " +
		"something asynchronous outlived it. Cancel that work when the component unmounts " +
		"instead, so nothing is left holding a reference to ask about.",
}

// NoIsMounted flags a call to `this.isMounted()` inside a method or an object property.
//
//	valid:   class Hello extends React.Component { notIsMounted() {} render() { this.notIsMounted(); } }
//	valid:   createReactClass({ m: function() { this.someFunc = this.isMounted; } })
//	valid:   function f() { this.isMounted(); }
//	invalid: createReactClass({ m: function() { if (!this.isMounted()) { return; } } })
//	invalid: class Hello extends React.Component { m() { this.isMounted(); } }
//
// Ported from `react/no-is-mounted`, read against oxc's `no_is_mounted.rs` and against
// `eslint-plugin-react`'s own `no-is-mounted.js`. Both were also *run*, because the two disagree
// and the imported corpus covers neither disagreement.
//
// # This is not a component rule, and the name of the helper that looks right is a trap
//
// The obvious implementation asks whether the call sits inside a React component, and it is wrong.
// Upstream asks a purely syntactic question: does any ancestor happen to be an object property or
// a method definition. It never asks what the enclosing object was constructed by, and never asks
// what the enclosing class extends.
//
// So `class Whatever { m() { this.isMounted(); } }`, with no heritage clause and no relation to
// React at all, reports. Measured on oxlint rather than reasoned: that source produces a finding at
// offset 23. A plain object literal, `var o = { m: function() { this.isMounted(); } }`, reports
// too. `internal/utilities/react/EnclosingComponent` answers false for both, so a port reaching for it
// is silent on real upstream findings.
//
// The imported corpus cannot catch that mistake. All three fail cases are written inside
// `createReactClass` or a class extending `React.Component`, so a component-gated port passes every
// one of them. That gap is why the ES5-versus-ES6 question in the dispatch resolves to "neither":
// `IsEs5ComponentCall` and `IsEs6ComponentClass` are both the wrong question here, and the fixtures
// below pin the answer with cases upstream does not ship.
//
// # Which ancestors count, where our AST splits what oxc's unifies
//
// oxc accepts two ancestor kinds and our tree spells them as four, so the set is enumerated rather
// than collapsed. Each mapping was measured against oxlint, and one of them is a near miss:
//
//	oxc ObjectProperty    -> KindPropertyAssignment      { m: function() {} }
//	                      -> KindMethodDeclaration       { m() {} }         (shared, see below)
//	                      -> KindGetAccessor             { get m() {} }
//	                      -> KindSetAccessor             { set m(v) {} }
//	oxc MethodDefinition  -> KindMethodDeclaration       class { m() {} }
//	                      -> KindGetAccessor             class { get m() {} }
//
// `KindMethodDeclaration` covers a class method and an object shorthand method at once, which is
// why one arm serves two of oxc's kinds.
//
// The near miss is `KindPropertyDeclaration`, a class field: `class W { m = () => {} }`. It looks
// like it belongs with the methods and it does not. oxc calls that a `PropertyDefinition`, which is
// absent from its match, and oxlint is silent on that source. It is excluded here for that reason
// and pinned by a fixture, because nothing in the imported corpus writes a class field.
//
// # Where the two upstreams disagree, and which one this follows
//
// **A computed property name.** ESLint guards with `'name' in callee.property`, so `this["isMounted"]()`
// has no `name` and is declined. oxc calls `static_property_name()`, which answers for a string or
// template key, so it reports. Measured on both: ESLint returns 0 findings for the computed string
// and template forms, oxlint returns one for each. oxc is followed, because oxc is the gate this
// replaces. A variable key, `this[isMounted]()`, names some other property and is silent in both.
//
// **Where the finding points.** ESLint reports `callee`, spanning `this.isMounted`. oxc reports
// `call_expr.span`, spanning `this.isMounted()` including the parentheses. Measured at 16 bytes
// from offset 23 for `class Whatever { m() { this.isMounted(); } }`. oxc is followed and the span
// is asserted in the fixtures, since a message-id assertion cannot see where a finding points.
//
// # What is deliberately not caught
//
// Only `this` reaches the method. An alias (`var f = this.isMounted; f()`) and a destructured
// binding (`const { isMounted } = this; isMounted()`) are both silent upstream, confirmed by
// running oxlint on them, and both are silent here. Following the call through a local binding
// would need the checker and would be a divergence rather than a refinement, so the gap is
// reproduced rather than improved on.
//
// Reading `this.isMounted` without calling it is also silent, which upstream ships as a passing
// case: the rule is about invoking it, and the corpus pins that the bare reference is fine.
//
// A parenthesized object or callee is silent too, for the reason given on `isThisKeyword` below.
// That one is upstream being narrow rather than upstream being right, and it is reproduced anyway
// because a differential run against oxlint is the thing this port has to survive.
var NoIsMounted = rule.Rule{
	// No namespace prefix. The config writes `react/no-is-mounted` and the parity guard strips the
	// namespace on a `/` boundary, so `react-no-is-mounted` would match no inventory entry, lint no
	// files, and still pass every fixture in this package.
	Name: "react/no-is-mounted",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if !isThisIsMountedCallee(node.AsCallExpression().Expression) {
					return
				}
				if !hasMethodLikeAncestor(node) {
					return
				}
				// The whole call expression, matching oxc's `call_expr.span` rather than ESLint's
				// narrower `callee`.
				ctx.ReportNode(node, messageNoIsMounted)
			},
		}
	},
}

// isThisIsMountedCallee reports whether a callee reads `isMounted` directly off `this`.
//
// Both member spellings answer, which is oxc's `static_property_name()` and not ESLint's `name`
// check. A dotted access carries the name as an identifier; a subscript carries it as a string or
// a no-substitution template, and an identifier subscript is a *variable* holding some other name,
// so it is declined. That partition was measured on oxlint: the string and template forms report,
// the variable form does not.
//
// The object must be `this` itself. `that.isMounted()` and `this.x.isMounted()` both fail here,
// the second because its object is a property access rather than the `this` keyword.
func isThisIsMountedCallee(callee *ast.Node) bool {
	if callee == nil {
		return false
	}

	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if !isThisKeyword(access.Expression) {
			return false
		}
		name := access.Name()
		// No kind check on the name, and its absence is measured rather than an oversight. The
		// obvious guard here is `name.Kind == ast.KindIdentifier`, to keep a private name
		// (`this.#isMounted()`) out. It cannot change an answer: a `KindPrivateIdentifier` node's
		// `Text()` returns `"#isMounted"` with the hash included, so the comparison below already
		// rejects it, and a sweep dropping the kind check survived every fixture because no input
		// can distinguish the two versions. A branch that cannot change a verdict is a branch no
		// test can guard, so it is gone rather than covered by a fixture asserting nothing. The
		// private-name case stays in the fixtures, because the *behavior* is still worth pinning.
		return name != nil && name.Text() == isMountedName

	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if !isThisKeyword(access.Expression) {
			return false
		}
		argument := access.ArgumentExpression
		if argument == nil {
			return false
		}
		switch argument.Kind {
		case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
			return argument.Text() == isMountedName
		}
	}
	return false
}

// isMountedName is the one method this rule is about.
const isMountedName = "isMounted"

// isThisKeyword reports whether an expression is the `this` keyword.
//
// Parentheses are deliberately *not* skipped, which is the opposite of what the rest of this
// package does for a parenthesized object and is upstream's behavior rather than a preference.
// oxc matches `Expression::ThisExpression(_)` against `member_expr.object()` directly, and a
// parenthesized object is a `ParenthesizedExpression` node there, so it does not match. Measured on
// oxlint: `(this).isMounted()` and `(this.isMounted)()` are both silent, against a control on the
// same file shape that reports.
//
// This was written the other way first, with `SkipParentheses` on both the callee and the object,
// on the reasoning that they are the same call. They are, and upstream still does not report them,
// so the version that read better was a silent divergence from the gate this replaces. Reproduced
// rather than improved on, and pinned by fixtures in both directions.
func isThisKeyword(expression *ast.Node) bool {
	return expression != nil && expression.Kind == ast.KindThisKeyword
}

// hasMethodLikeAncestor reports whether any ancestor is an object property or a method definition.
//
// This is upstream's containment test verbatim, and the whole reason the rule stays silent on a
// bare function or a top-level call. It walks to the root because upstream does: the property or
// method need not be the immediate parent, and an arrow function nested inside an object property
// still reports, which was measured rather than assumed.
//
// The walk starts at the parent rather than the node, matching `ancestor_kinds`, which does not
// include the node itself. That cannot change an answer here, since a call expression is never one
// of these kinds, but starting at the node would be a different function than the one upstream
// runs and the difference would be invisible until some other caller reused it.
func hasMethodLikeAncestor(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		// A class field (`KindPropertyDeclaration`) is deliberately absent: oxc calls it a
		// `PropertyDefinition` and does not match it, and oxlint is silent on that source.
		case ast.KindPropertyAssignment,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor:
			return true
		}
	}
	return false
}
