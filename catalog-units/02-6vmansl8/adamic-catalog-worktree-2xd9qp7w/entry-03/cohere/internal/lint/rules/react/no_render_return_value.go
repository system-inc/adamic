package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoRenderReturnValue = rule.Message{
	Id: "noRenderReturnValue",
	Description: "The return value of `ReactDOM.render` is being consumed. It is a legacy escape " +
		"hatch that only ever worked for a class component rendered synchronously at the root, " +
		"and it returns null for a function component, so code holding it is reading an " +
		"instance that may not exist. React 18's concurrent root removed the guarantee entirely " +
		"and React 19 removed `ReactDOM.render` itself. Attach a ref to the thing you wanted a " +
		"handle on, or run the work in an effect, so nothing depends on a value the renderer " +
		"stopped promising.",
}

// NoRenderReturnValue flags code that consumes what `ReactDOM.render` returns.
//
//	valid:   ReactDOM.render(<div />, document.body);
//	valid:   foo(ReactDOM.render(<div />, x));
//	valid:   React.render(<div />, root);
//	invalid: var inst = ReactDOM.render(<div />, document.body);
//	invalid: function f() { return ReactDOM.render(<div />, document.body) }
//	invalid: const render = () => ReactDOM.render(<div />, document.body)
//
// Ported from `react/no-render-return-value`, read against oxc's `no_render_return_value.rs` and
// measured against the release oxlint binary on inputs the imported corpus does not contain.
//
// # This is not a use analysis, and reading it as one is the way to get it wrong
//
// The name and the message both say "return value", which invites the implementation everybody
// reaches for first: decide whether the call's result is consumed. That implementation is wrong in
// both directions and the imported corpus cannot tell you so, because upstream's eight fail cases
// are exactly the five parent kinds and nothing else.
//
// What upstream actually asks is whether the call's immediate parent is one of five node kinds. So
// these all consume the return value in any ordinary reading, and every one of them is silent,
// measured on oxlint rather than reasoned:
//
//	foo(ReactDOM.render(<div />, x));        an argument
//	ReactDOM.render(<div />, x).foo;         a property read off the result
//	if (ReactDOM.render(<div />, x)) {}      a condition
//	ReactDOM.render(<div />, x) === null;    a comparison
//	`${ReactDOM.render(<div />, x)}`;        a template interpolation
//	[ReactDOM.render(<div />, x)];           an array element
//
// A port inferring its parent set from what counts as a use reports all six, passes all seventeen
// upstream cases while doing so, and is wrong on the tree. The set is enumerated from upstream's
// match arm rather than derived, and the six inputs above are pinned as clean fixtures.
//
// # The second check, which the corpus does not exercise at all
//
// Upstream runs a *second*, independent test after the parent-kind one, and it is not a refinement
// of it. If the parent node's enclosing scope is an arrow function whose body is an expression, the
// call reports wherever it sits inside that body. Position stops mattering:
//
//	var d = () => foo(ReactDOM.render(<div />, x));    reports, though the bare form does not
//	var c = () => ReactDOM.render(<div />, x).foo;     reports, though the bare form does not
//
// No upstream input distinguishes this, so a port implementing only the parent-kind list passes the
// entire imported corpus while missing half the rule. It is pinned by invented fixtures instead.
//
// The scope is the *innermost* one, not any ancestor: a function expression inside an arrow body
// starts a new scope and stops the reach, measured silent on oxlint. Our tree has no scope table,
// so the check is spelled as a walk to the nearest function-like ancestor, which answers the same
// question by construction rather than by lookup.
//
// # Where the two checks overlap, upstream reports twice
//
// `() => () => ReactDOM.render(...)` satisfies both at once and oxc diagnoses in both branches with
// no guard against having already reported, so one input produces two identical findings. That is
// reproduced rather than deduplicated: oxlint is the gate this replaces, and collapsing it would
// read as a difference on real code. It looks like an upstream oversight and is pinned by a fixture
// that says so.
//
// # What the receiver must be
//
// `ReactDOM` exactly, as a bare identifier. `React.render` is a deliberate upstream pass, and the
// commented-out fail cases in oxc's own corpus record why: the rule targets React 15 and later,
// where the method moved off `React`. That gap is reproduced rather than improved on.
//
// # No type checker, and adding one would be a divergence rather than a refinement
//
// Every question this rule asks is syntactic, so nothing here declares `NeedsTypeChecker`. Upstream
// calls `ctx.scoping()`, which is not evidence either way: it asks that table only for scope flags
// (`is_arrow`), a property of the enclosing construct the shape of the tree already carries, rather
// than asking which declaration an identifier binds to.
//
// The stronger point is that resolution would make this port *worse*, in both directions at once,
// which was measured rather than reasoned. Upstream matches the spelling `ReactDOM` and never asks
// what it binds to:
//
//	var RD = ReactDOM; var b = RD.render(<div />, x);        silent, in a file importing react-dom
//	function f(ReactDOM) { var c = ReactDOM.render(...); }   reports, though this is a parameter
//
// The first is a real miss the checker would close, and the second is a finding on code that has
// nothing to do with react-dom, which the checker would decline. A resolution-based port therefore
// disagrees with the gate on both, and both are pinned as fixtures so a later helpful correction
// fails loudly. Needing no checker and the checker being unhelpful are separate findings, and this
// rule happens to have both.
//
// `internal/utilities/react.IsNamespacedMember` is the shelf function whose name fits this shape and it
// cannot serve here: it hardcodes its receiver to `React` alone, by documented design and confirmed
// in its body, so it declines every input this rule exists to catch. Its parenthesis-skipping is
// also a divergence here, since oxc matches `Expression::Identifier` against the object directly
// and a parenthesized receiver is a different node there.
var NoRenderReturnValue = rule.Rule{
	// No namespace prefix. The config writes `react/no-render-return-value` and the parity guard
	// strips the namespace on a `/` boundary, so `react-no-render-return-value` would match no
	// inventory entry, lint no files, and still pass every fixture in this package.
	Name: "react/no-render-return-value",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := reactDOMRenderCallee(node.AsCallExpression().Expression)
				if callee == nil {
					return
				}

				parent := node.Parent
				if parent == nil {
					return
				}

				if isReturnValueConsumingParent(parent) {
					ctx.ReportNode(callee, messageNoRenderReturnValue)
				}

				// Deliberately not an `else`. Upstream runs both checks unconditionally, so an
				// input satisfying both reports twice, and that duplicate is measured on oxlint
				// rather than assumed. See the doc comment above.
				if isInsideAnArrowExpressionBody(parent) {
					ctx.ReportNode(callee, messageNoRenderReturnValue)
				}
			},
		}
	},
}

// reactDOMRenderCallee returns the member access when a callee reads `render` off `ReactDOM`, and
// nil otherwise.
//
// The returned node is the member access rather than a boolean because it is also the report
// target: upstream's span merges the receiver identifier with the property, which is exactly this
// node's range, covering `ReactDOM.render` and stopping before the argument list.
//
// Both member spellings answer. `property.AccessedName` reads a dotted access and a subscripted one
// alike, which is what oxc's `static_property_info()` does, and `ReactDOM['render'](...)` is
// measured as a finding on oxlint with an 18-byte span.
//
// The accept set is stated rather than taken from `property.Textual`, which was the first choice
// and was silently narrow. `static_property_info` accepts a single-quasi template literal, read at
// `oxc_ast/src/ast_impl/js.rs:628`, so a template-subscripted render call reports upstream.
// Measured on oxlint: it does. A sweep widening the set to `property.Static` survived every fixture, which is
// how the gap was found rather than by reading.
//
// Reading the property text without a kind check would be a crash rather than a wrong answer here.
// `Node.Text()` panics on a `KindComputedPropertyName`, and this rule anchors on the shape most
// likely to meet one, so the read goes through the shelf function that guards it rather than
// through `Name()` directly.
func reactDOMRenderCallee(callee *ast.Node) *ast.Node {
	if callee == nil {
		return nil
	}

	var receiver *ast.Node
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = callee.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		receiver = callee.AsElementAccessExpression().Expression
	default:
		return nil
	}

	// Parentheses are deliberately not skipped. oxc matches `Expression::Identifier` against
	// `member_expr.object()` directly, so `(ReactDOM).render(...)` is a `ParenthesizedExpression`
	// there and does not match. `react.IsNamespacedMember` skips them, which is one of two reasons
	// it cannot serve this rule even setting its hardcoded receiver aside.
	if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != reactDOMReceiver {
		return nil
	}

	name, ok := property.AccessedName(callee, reactDOMRenderAcceptedKeys)
	if !ok || name != renderMethod {
		return nil
	}
	return callee
}

// The one receiver and the one method this rule is about. `React.render` is a deliberate upstream
// pass; see the doc comment on the rule.
const (
	reactDOMReceiver = "ReactDOM"
	renderMethod     = "render"
)

// reactDOMRenderAcceptedKeys is the set of key spellings that can name `render` the way upstream
// reads one.
//
// `Numeric` and `Private` are excluded because no number and no `#name` can render as `render`, so
// including them would change no answer while overstating what was checked. `Computed` is excluded
// because there are no brackets in the tree to see through here: `AccessedName` reads an element
// access's subscript directly, and a bare identifier subscript names whatever the variable holds,
// which upstream declines too. Measured on oxlint: `ReactDOM[render](...)` is silent.
//
// Widening this to `property.Static` survives every fixture, and that survivor is equivalent rather
// than a blind spot. Distinguishing the two versions needs a key that is numeric, private, or a
// computed-property-name node whose text is nonetheless `render`, and no such input exists: the
// first two spellings cannot produce those letters, and `AccessedName` never routes a
// `KindComputedPropertyName` into `Name` at all, reading a property access's name node and an
// element access's subscript expression directly. Read out of the accessor's body rather than
// inferred from the flag's name, since the flag reads as though it should apply here.
const reactDOMRenderAcceptedKeys = property.Named | property.Quoted | property.Templated

// isReturnValueConsumingParent reports whether the call's parent is one of the five node kinds
// upstream matches.
//
// The set is upstream's arm transcribed, not a judgment about what consuming a value means. Four of
// the five map straight across and one splits, so the mapping is written out:
//
//	oxc VariableDeclarator         -> KindVariableDeclaration
//	oxc ObjectProperty             -> KindPropertyAssignment
//	oxc ReturnStatement            -> KindReturnStatement
//	oxc ArrowFunctionExpression    -> KindArrowFunction
//	oxc AssignmentExpression       -> KindBinaryExpression with an assignment operator
//
// The assignment arm is the one that needs care. Our tree spells an assignment as a binary
// expression, so the kind alone would swallow every `===`, `+`, and `&&` whose operand is the call,
// and those are measured silent on oxlint. The operator is checked instead of the position: oxc
// matches the node kind without asking which side the call is on, and `+=` and `||=` both report,
// measured, so every assignment operator counts rather than `=` alone.
//
// A class field initializer looks like it belongs in the declarator arm and does not. oxc calls it
// a `PropertyDefinition`, which is absent from the match, and oxlint is silent on that source. It is
// excluded for that reason and pinned by a fixture, since nothing upstream writes one.
//
// The arrow arm needs no further test, which is measured rather than assumed; see the comment at
// that arm. A block-bodied arrow never reaches it, because the block interposes and upstream's
// check is on the immediate parent only.
func isReturnValueConsumingParent(parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindPropertyAssignment, ast.KindReturnStatement:
		return true
	case ast.KindArrowFunction:
		// No check that the call *is* the body, and its absence is measured rather than an
		// oversight. The obvious guard here is `parent.AsArrowFunction().Body == call`, to keep a
		// call sitting elsewhere under the arrow out. Nothing can reach it: a probe reporting every
		// call whose parent is an arrow found the body position to be the only one. A parameter
		// default parents to the binary expression that spells it (`(p = render()) => 1`), a
		// parenthesized body parents to the parenthesis, and a type position holds no call. So the
		// guard cannot change a verdict, a sweep dropping it survived every fixture, and a branch
		// no input can distinguish is a branch no test can guard.
		return true
	case ast.KindBinaryExpression:
		return ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind)
	}
	return false
}

// isInsideAnArrowExpressionBody reports whether a node sits inside the expression body of the
// nearest enclosing arrow function.
//
// This replaces upstream's scope-table lookup, which asks whether the parent node's scope is an
// arrow's and whether that arrow has an expression body. We have no scope table, and a walk to the
// nearest function-like ancestor answers the same question by construction: the innermost scope a
// node sits in is the innermost function-like ancestor it has, so finding that ancestor and asking
// whether it is an expression-bodied arrow is the same predicate spelled structurally.
//
// The walk stops at the first function-like node rather than continuing, and the stop is
// load-bearing rather than an optimization. A function expression inside an arrow body starts its
// own scope, so upstream's lookup returns that function rather than the arrow, and the check fails.
// Measured on oxlint: `var a = () => (function () { ReactDOM.render(<div />, x); });` is silent,
// against the same shape without the wrapper which reports. A walk continuing past the function
// would report it.
//
// The walk asks for the scope *containing* the parent, which is why it starts above the parent
// rather than at it. `parent_node.scope_id()` upstream is the scope a node sits in, and a function
// node sits in the scope outside itself rather than in its own. That distinction is the whole
// reason `var render = (a, b) => ReactDOM.render(a, b)` reports once and not twice: there the
// parent is the arrow, its enclosing scope is the module, and only the parent-kind check answers.
// Written the other way first, starting at the parent, and it double-reported two of upstream's
// eight fail cases while every clean case stayed green.
func isInsideAnArrowExpressionBody(parent *ast.Node) bool {
	start := parent
	if ast.IsFunctionLike(start) {
		start = start.Parent
	}
	for current := start; current != nil; current = current.Parent {
		if !ast.IsFunctionLike(current) {
			continue
		}
		if current.Kind != ast.KindArrowFunction {
			return false
		}
		body := current.AsArrowFunction().Body
		// An expression body is anything that is not a block. A block-bodied arrow is a statement
		// context, where upstream's `is_expression()` answers false.
		return body != nil && body.Kind != ast.KindBlock
	}
	return false
}
