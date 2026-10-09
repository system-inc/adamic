package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoDirectMutationState = rule.Message{
	Id: "noDirectMutationState",
	Description: "`this.state` is written to directly. React does not see the write, so nothing " +
		"re-renders, and the next `setState` merges into the state object React still believes " +
		"is current and silently discards what was written here. The bug surfaces later and " +
		"somewhere else, as a value that was set and then was not. Call `setState` with the new " +
		"value instead, and treat `this.state` as read-only everywhere but the constructor.",
}

// NoDirectMutationState flags a write whose target is rooted at `this.state` inside a React class
// component.
//
//	valid:   class H extends React.Component { constructor() { this.state = {}; } }
//	valid:   class H extends React.Component { render() { return <p>{this.state.x}</p>; } }
//	valid:   class Plain { m() { this.state.x = 1; } }              (not a component)
//	valid:   createReactClass({ m() { var o = {state:{}}; o.state.x = 1; } })
//	invalid: class H extends React.Component { componentDidMount() { this.state.x = 1; } }
//	invalid: class H extends React.Component { m() { this.state.a.b = 1; } }
//	invalid: class H extends React.Component { m() { this.state.x++; } }
//	invalid: class H extends React.Component { constructor() { go(() => { this.state = 1; }); } }
//
// Ported from `react/no-direct-mutation-state`, read against oxc's
// `no_direct_mutation_state.rs` and the three helpers it calls: `get_outer_member_expression`
// (`oxc_linter/src/ast_util.rs:939`), `is_state_member_expression` and `is_es5_component` /
// `is_es6_component` (`oxc_linter/src/utils/react.rs:830`, `:556`, `:577`). ESLint's
// `eslint-plugin-react` source is in this tree and was read as a second opinion; it carries no
// `meta.schema` at all, which is the authoritative statement that this rule takes no options, and
// oxc likewise deserializes nothing. Where the two differ oxc wins, and they differ: ESLint routes
// everything through its `Components` detector and tracks a `mutateSetState` flag per component,
// which oxc does not reproduce and neither does this.
//
// # What counts as a write, measured rather than assumed
//
// Upstream anchors on exactly two node kinds, an assignment expression and an update expression,
// and nothing else. That is narrower than the rule's name suggests and the narrowness is the
// interesting part. Run against the release oxlint binary, inside a real component method:
//
//	this.state = 1              reports    assignment
//	this.state.x = 1            reports    nested property write
//	this.state.a.b = 1          reports    any depth
//	this.state["x"] = 1         reports    computed subscript ON state
//	this.state.x++              reports    update, prefix and postfix alike
//	this.state.x += 1           reports    every assignment operator, seven measured
//	delete this.state.x         SILENT     a unary expression, not an assignment
//	[this.state.x] = [1]        SILENT     a destructuring target is not a simple target
//	({y: this.state.x} = {})    SILENT     the same
//	this.state.push(1)          SILENT     a call is not a write to upstream
//	Object.assign(this.state,{}) SILENT    the same
//	const s = this.state; s.x=1 SILENT     no alias tracking of any kind
//	this["state"].x = 1         SILENT     see the static-member requirement below
//
// The last three lines are why this rule is much weaker than it reads, and reproducing that
// weakness is the point rather than a shortcoming of the port: a differential run against oxlint is
// what this has to survive. `delete` and the destructuring targets in particular are real mutations
// of state that upstream simply does not see.
//
// # The root of the chain must be a dotted `this.state`
//
// oxc's `get_outer_member_expression` is named for the outer expression and walks to the INNERMOST
// one, descending through `object()` while the object is itself a member expression, then requires
// the result to be a `StaticMemberExpression`. So the chain is reduced to its root before anything
// is compared, and the root has to be dot access rather than subscript.
//
// That produces the asymmetry above. `this.state["x"] = 1` reports, because its root is still the
// dotted `this.state` and only the last hop is computed. `this["state"].x = 1` is silent, because
// its root is an element access and the static-member requirement rejects it before the name is
// ever compared. Both measured on the release binary; neither is in the corpus, and a port that
// treated computed access uniformly would get exactly one of the two wrong with nothing to tell it.
//
// The ownership check is real here, which is worth stating because three sibling rules in this
// family turned out to have none. `is_state_member_expression` requires the root's object to be a
// `ThisExpression` and its property to be named `state`, so a plain object carrying a `state` key
// is clean, and upstream ships that as its second passing case: `var obj = {state: {}};
// obj.state.name = 'foo';`. An object literal with a state-shaped property nested anywhere does not
// report.
//
// # Parentheses, which are inconsistent within this one rule
//
// Measured on the release binary because no imported fixture in either corpus writes a
// parenthesized anything, so guessing is invisible in both directions:
//
//	(this.state.foo) = 1        REPORTS    span is the inner `this.state.foo`, without the parens
//	(this.state.foo)++          REPORTS    span is the whole `(this.state.foo)++`, with them
//	(this).state.foo = 1        SILENT
//	(this.state).foo = 1        SILENT
//	((this).state).foo = 1      SILENT
//
// The split is not a decision upstream made, it is where the parentheses land in oxc's tree. A
// parenthesized assignment target is stored unwrapped there, so the outermost parens vanish before
// the rule sees anything; parentheses anywhere INSIDE the chain survive as a real node, and the
// `object()` walk does not skip them, so the descent stops at a `ParenthesizedExpression` that is
// not a member expression and the root is never reached.
//
// Our tree keeps the outer parentheses as a real `KindParenthesizedExpression` on the assignment's
// left. So reproducing upstream needs an explicit `SkipParentheses` at the target and none inside
// the chain, which reads backwards from the usual advice that paren-skipping is a free correctness
// improvement. Adding it to the descent would report three cases oxlint is silent on; leaving it
// off the target would miss one it reports and would put the span two characters wide of upstream
// on that one. All five forms are pinned by fixtures below.
//
// # The constructor exemption, and the call-expression escape hatch inside it
//
// `should_ignore_component` ignores a write when `is_constructor && !is_call_expression`, and also
// whenever no component was found. The second half is the ordinary component gate. The first half
// is this rule's characteristic behavior and it is where a port most often diverges, because
// "constructor is exempt" is only half of what upstream does:
//
//	constructor() { this.state = {}; }                 SILENT   the exemption
//	constructor() { this.state.foo = 1; }              SILENT   nested too
//	constructor() { go(() => { this.state = 1; }); }   REPORTS  a call intervenes
//	constructor() { foo(this.state.x = 1); }           REPORTS  and it need not be a callback
//
// The third is upstream's own failing case and reads as "an async callback runs later, so the
// exemption should not cover it". The fourth shows the implementation does not actually ask that.
// The latch is set by any `CallExpression` ancestor at all, so a write in an ARGUMENT position is
// enough, and `foo(this.state.x = 1)` reports even though it runs synchronously during construction
// exactly like the exempt form. That is upstream being crude rather than upstream being right, and
// it is reproduced rather than improved on. Measured on the release binary; no fixture upstream
// distinguishes it, so the case below is invented and says so.
//
// # The ancestor walk stops at the first class
//
// The loop breaks on `AstKind::Class`, after running the body for that node. So a plain class
// nested inside a real component hides its own writes: the walk reaches the inner class, finds no
// component, breaks, and answers "not a component". Upstream pins the mirror image of this with a
// nested class inside a component constructor as a passing case. But the break is on a class only,
// so an ES5 component created inside a real component's method still reports, because a call
// expression does not stop the walk. Both measured.
//
// # The ES5 factory name set is narrower than the shelf
//
// `internal/utilities/react.IsEs5ComponentCall` accepts `createClass` and `React.createClass` as well
// as the `createReactClass` spellings. oxc keys both arms of `is_es5_component` on one constant,
// `CREATE_CLASS = "createReactClass"`. Measured: `createClass({...})` and `React.createClass({...})`
// containing a state write are both clean on the release binary, while `createReactClass` and
// `React.createReactClass` both report. This calls `isEs5ComponentCallStrict` in
// `no_did_mount_set_state.go`, which is that narrowing and was measured there for the same reason.
// Following the shelf helper's name instead of its body would report on clean upstream code, and no
// imported fixture could see it.
//
// # This is not the question `reference.WritesToBinding` answers
//
// That helper is what eight rules in this tree ask, and it is the wrong shelf here rather than a
// near miss. It asks whether an IDENTIFIER writes the BINDING it names, and its own test file pins
// `this.a = 1` and `o.a = 1` as answering FALSE, deliberately, because a rule watching a binding
// called `a` must not report a property write that merely shares the name. This rule asks the
// opposite question: not which binding is written, but whether the member chain being written is
// rooted at `this.state`. There is no binding involved at any point. Read the body, not the name.
//
// # Where the finding points
//
// Two different spans, which is the one thing a message-id assertion cannot see. An assignment
// reports `assignment_expr.left.span()`, the left-hand side alone without the operator or the right
// side. An update expression reports `update_expr.span`, the whole thing including `++`. The
// upstream snapshot underlines `this.state.foo` for the first and `this.state.foo++` for the
// second, and every fixture below asserts the sliced text rather than the id alone.
var NoDirectMutationState = rule.Rule{
	// No namespace prefix. The config writes `react/no-direct-mutation-state` and the parity guard
	// strips the namespace on a `/` boundary, so `react-no-direct-mutation-state` would match no
	// inventory entry, lint no files, and still pass every fixture in this package.
	Name: "react/no-direct-mutation-state",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// No file gate. oxc gated this on `source_type().is_jsx()` and that gate came along with the
		// port from oxc, but the authority here is eslint-plugin-react, which does not gate on the
		// file name at all. React code in a `.ts` file is ordinary and legal, and the gate made this
		// rule silent across more `.ts` files than the `.tsx` files it could see.

		return rule.Listeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				assignment := node.AsBinaryExpression()
				// `AssignmentExpression` in oxc is a distinct node covering every assignment
				// operator; ours is a binary expression, so the operator set is the discriminator.
				// `ast.IsAssignmentOperator` was probed against all seven assignment operators
				// oxlint reports on and the three comparison operators it is silent on, and matched
				// on all ten.
				if !ast.IsAssignmentOperator(assignment.OperatorToken.Kind) {
					return
				}

				// The outer parentheses only. See the rule doc: oxc's tree has already dropped
				// these by the time the rule runs, so skipping here reproduces it, while skipping
				// inside the chain would not.
				target := ast.SkipParentheses(assignment.Left)
				if !isStateRootedMemberChain(target) {
					return
				}
				if shouldIgnoreComponent(node) {
					return
				}

				// The left side alone, not the whole assignment. Upstream underlines
				// `this.state.foo` and stops before the `=`.
				ctx.ReportNode(target, messageNoDirectMutationState)
			},

			ast.KindPostfixUnaryExpression: func(node *ast.Node) {
				operand := node.AsPostfixUnaryExpression()
				// oxc's `UpdateExpression` is `++` and `--` and nothing else, and a postfix node in
				// our tree is only ever one of those two, so no operator check is needed on this
				// arm. The prefix arm below does need one.
				reportUpdate(ctx, node, operand.Operand)
			},

			ast.KindPrefixUnaryExpression: func(node *ast.Node) {
				operand := node.AsPrefixUnaryExpression()
				// Our prefix node also carries `!`, `-`, `+`, `~` and `void`, none of which is an
				// update expression to oxc. `delete this.state.x` is a different kind again and is
				// silent at both tools; see the rule doc.
				switch operand.Operator {
				case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
					reportUpdate(ctx, node, operand.Operand)
				}
			},
		}
	},
}

// reportUpdate is the shared body of the two update-expression arms.
//
// The span is the whole update expression rather than its operand, which is upstream's
// `update_expr.span` and is the one place this rule points somewhere other than at the target. The
// snapshot underlines `this.state.foo++` including the operator, and `(this.state.foo)++` including
// the parentheses, which is the visible consequence of the operand being paren-skipped for the
// match while the reported node is the expression around it.
func reportUpdate(ctx rule.Context, updateExpression *ast.Node, operand *ast.Node) {
	if !isStateRootedMemberChain(ast.SkipParentheses(operand)) {
		return
	}
	if shouldIgnoreComponent(updateExpression) {
		return
	}
	ctx.ReportNode(updateExpression, messageNoDirectMutationState)
}

// isStateRootedMemberChain reports whether a write target's member chain is rooted at `this.state`.
//
// This is oxc's `get_outer_member_expression` composed with `is_state_member_expression`. The walk
// descends through the object of each member expression while that object is itself a member
// expression, so any depth of chain reduces to its root, then the root must be a DOTTED access
// whose object is the `this` keyword and whose name is `state`.
//
// Three properties of that, each of which a fixture pins:
//
//	this.state.a.b = 1      the descent, arbitrary depth
//	this.state["x"] = 1     a computed hop above the root is fine, the root is still dotted
//	this["state"].x = 1     a computed ROOT is not, so this is silent
//
// Parentheses are not skipped inside the descent, matching oxc, whose `object()` accessor returns
// the node as written. `(this.state).foo = 1` therefore stops at a parenthesized expression that is
// not a member expression and answers false, which was measured on the release binary rather than
// reasoned about. The caller skips the parentheses around the whole target, and only there.
//
// Nothing here reads `Text()` on a member expression. A `KindPropertyAccessExpression` panics
// outright on `Text()` in our shim, and every node this function touches is one, so the names are
// taken off the access's own name node after checking that node's kind.
func isStateRootedMemberChain(target *ast.Node) bool {
	if target == nil {
		return false
	}

	// Descend to the root of the chain. `this.state.person.name.first` arrives here as the
	// outermost access and leaves as `this.state`.
	root := target
	for {
		var object *ast.Node
		switch root.Kind {
		case ast.KindPropertyAccessExpression:
			object = root.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			object = root.AsElementAccessExpression().Expression
		default:
			return false
		}
		if object == nil {
			return false
		}
		if object.Kind != ast.KindPropertyAccessExpression && object.Kind != ast.KindElementAccessExpression {
			break
		}
		root = object
	}

	// The root must be dotted. oxc's match arm keeps only `StaticMemberExpression`, so an element
	// access here is rejected before its subscript is ever looked at, even when that subscript is
	// the literal string "state".
	if root.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := root.AsPropertyAccessExpression()
	if access.Expression == nil || access.Expression.Kind != ast.KindThisKeyword {
		return false
	}
	name := access.Name()
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "state"
}

// shouldIgnoreComponent is oxc's `should_ignore_component` at our AST's spelling.
//
// It answers true for a write that should NOT be reported, which is upstream's polarity and is kept
// rather than inverted so the two read the same way side by side. The condition is
// `(is_constructor && !is_call_expression) || !is_component`, and the three latches are gathered in
// one climb that stops at the first class.
//
// # The three latches
//
// A constructor ancestor sets the first. oxc matches `MethodDefinition` with
// `MethodDefinitionKind::Constructor`; our tree has a dedicated `KindConstructor` node, so the kind
// alone is the test and no name comparison is involved.
//
// Any call expression ancestor sets the second, and this is the crude part reproduced deliberately.
// It is not "is this inside a callback": an argument position is enough, so `foo(this.state.x = 1)`
// inside a constructor reports while the bare form does not. See the rule doc for the measurement.
//
// A component ancestor sets the third, either an ES5 factory call or an ES6 class extending a React
// base. The ES5 arm calls `isEs5ComponentCallStrict` rather than the shelf helper, because the
// shelf accepts two factory names oxc does not; see that function's own doc.
//
// # Why the climb stops at a class and why the checks run before the break
//
// Upstream breaks on `AstKind::Class` at the END of the loop body, so the class that stops the walk
// is itself tested first. That ordering is load-bearing rather than incidental: for a write inside
// a component's own method, the component class IS the node the loop breaks on, and testing it
// before breaking is the only reason the whole ES6 half of this rule reports at all. Reversing the
// two lines silences every class-component finding while leaving all eight passing cases green,
// which is the shape of defect no imported fixture can catch.
//
// Both class kinds are checked because `IsEs6ComponentClass` accepts a declaration and an
// expression alike, and a component assigned to a variable is the second.
func shouldIgnoreComponent(node *ast.Node) bool {
	isConstructor := false
	isCallExpression := false
	isComponent := false

	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindConstructor {
			isConstructor = true
		}

		if current.Kind == ast.KindCallExpression {
			isCallExpression = true
		}

		if utilsreact.IsEs6ComponentClass(current) || isEs5ComponentCallStrict(current) {
			isComponent = true
		}

		// After the checks, never before. See the doc above.
		if current.Kind == ast.KindClassDeclaration || current.Kind == ast.KindClassExpression {
			break
		}
	}

	return (isConstructor && !isCallExpression) || !isComponent
}
