package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoFindDOMNode = rule.Message{
	Id: "noFindDOMNode",
	Description: "`findDOMNode` was deprecated in 2018 and removed in React 19, so this call " +
		"throws on any current React. Even where it still runs it reaches around the " +
		"component that owns the node: the parent reads a child's DOM element the child " +
		"never agreed to expose, so the child cannot change what it renders without " +
		"breaking a caller it does not know about. Pass a ref to the element instead, so " +
		"the component decides what it hands out.",
}

// NoFindDOMNode flags a call to `findDOMNode`, bare or through a React namespace.
//
//	valid:   var Hello = function() {};
//	valid:   this.someFunc = React.findDOMNode;
//	valid:   SomeModule.findDOMNode(this).scrollIntoView();
//	valid:   React.someFunc(this);
//	invalid: findDOMNode(this).scrollIntoView();
//	invalid: ReactDOM.findDOMNode(this);
//
// Ported from `react/no-find-dom-node`, read against oxc's `no_find_dom_node.rs`. There is no
// readable `eslint-plugin-react` source for this one in the tree to compare against, so oxc is the
// only implementation consulted, and every claim below that is not visible in the eleven imported
// cases was measured by running the release oxlint binary on a probe file rather than reasoned out.
//
// # This is a syntactic rule and it resolves nothing
//
// The obvious reading of the name is that the rule finds calls to *React's* `findDOMNode`, which
// would need to know where the callee came from and therefore need the checker. It does not ask
// that, and declaring the checker here would be wrong on cost and wrong on behavior.
//
// oxc's `run` reads only `call_expr.callee`. It never touches `ctx.scoping()`, never resolves a
// reference to a declaration, and never looks at the imports. Two measurements pin that this is
// upstream's behavior and not an omission in the reading:
//
//	function f(findDOMNode) { findDOMNode(this); }        reports, though it is a parameter
//	import { findDOMNode as f } from 'react-dom'; f(this) silent, though it IS React's function
//
// The first is a local binding that has nothing to do with React and it reports; the second is
// genuinely React's `findDOMNode` under another name and it is silent. So the rule is a check on
// the *spelling at the call site*, and any port that resolves the identifier would disagree with
// the gate in both directions. The one imported case with a `demo.ts` filename and a real
// `import ReactDOM from 'react-dom'` is therefore testing nothing about the import: the same source
// with the import line deleted reports identically.
//
// # Two shapes, and the receiver set is closed
//
// A bare identifier callee reports on the name alone, which is oxc's
// `callee.get_identifier_reference()` arm. Note that arm's unconditional `return`: once the callee
// is an identifier, a non-matching name falls out of the rule entirely rather than continuing to
// the member check. That cannot change an answer, since an identifier is not a member expression,
// but it is why the code below is a switch rather than two sequential checks.
//
// A member callee reports only when the object is one of exactly three identifiers, `React`,
// `ReactDOM` or `ReactDom`. This is the discrimination `react.IsNamespacedMember` looks built for
// and cannot serve: that helper hardcodes the receiver as `React` alone, by design and with a
// documented reason, so it answers no for `ReactDOM.findDOMNode` and would drop two thirds of the
// imported fail cases. The shelf was read before this was written rather than after; using it here
// would have been a name match rather than a semantic one. Casing is exact, measured:
// `reactdom.findDOMNode(this)` is silent.
//
// The object must be the receiver *directly*. `a.ReactDOM.findDOMNode(this)` and
// `getReactDOM().findDOMNode(this)` are both silent on the release binary, because oxc's
// `is_specific_id` asks whether the object node is that identifier rather than whether its text
// ends in it.
//
// # Where the finding points, including the case that is easy to get wrong
//
// The span is the *name*, never the call and never the member expression. That is `ident.span` in
// the identifier arm and `static_property_info()`'s span in the member arm, and the imported
// snapshot confirms it at eleven columns wide on all six findings.
//
// The subscript form is the trap. `ReactDOM["findDOMNode"](this)` reports on `"findDOMNode"` *with
// the quotes*, thirteen bytes rather than eleven, because `static_property_info` hands back the
// literal node's own span rather than the name inside it. Same for a template subscript and its
// backticks. Nothing in the imported corpus writes a subscript at all, so this is pinned by an
// invented span assertion below; a port reporting the cooked name would be off by one on each side
// and every message-id fixture would still pass.
//
// # What is skipped and what is not
//
// Parentheses are skipped on both the callee and the object, which is the opposite of the sibling
// `no-is-mounted` in this package and is upstream's behavior rather than a preference here. oxc
// reaches the callee through `get_identifier_reference()` and `get_member_expr()`, both of which
// unwrap parenthesized expressions, and `is_specific_id` unwraps the object. Measured on all three:
// `(findDOMNode)(this)`, `(ReactDOM).findDOMNode(this)` and `((ReactDOM)).findDOMNode(this)` all
// report, and each finding still points at the bare name.
//
// Optional chaining reports in both positions, `ReactDOM?.findDOMNode(this)` and
// `findDOMNode?.(this)`, since neither changes the callee's kind in our tree.
//
// A `new findDOMNode(this)` is silent, because upstream anchors on `CallExpression` and a `new` is a
// different node. That is upstream being narrow rather than upstream being right, and it is
// reproduced rather than improved on, because a differential run against oxlint is what this port
// has to survive.
var NoFindDOMNode = rule.Rule{
	// No namespace prefix. The config writes `react/no-find-dom-node` and the parity guard strips
	// the namespace on a `/` boundary, so `react-no-find-dom-node` would match no inventory entry,
	// lint no files, and still pass every fixture in this package.
	Name: "react/no-find-dom-node",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				name := findDOMNodeCalleeName(node.AsCallExpression().Expression)
				if name == nil {
					return
				}
				// The name node alone, matching `ident.span` and `static_property_info`'s span
				// rather than the whole call. For a subscript this is the string or template
				// literal node, so the quotes are included, which is upstream's span and not a
				// slip.
				ctx.ReportNode(name, messageNoFindDOMNode)
			},
		}
	},
}

// findDOMNodeCalleeName returns the node to underline when a callee names `findDOMNode`, else nil.
//
// Returning the node rather than a bool is what keeps the span correct at the one call site: the
// thing being reported is never the node the listener received, and a signature answering only
// "does this match" would push the job of re-deriving the name node back onto the caller, which is
// how a rule ends up pointing at its anchor.
//
// Parentheses are skipped here, matching oxc's accessors. See the note on the rule above for the
// three measured cases, and note that this differs from `no-is-mounted` in this same package, which
// deliberately does not skip them because *its* upstream does not.
func findDOMNodeCalleeName(callee *ast.Node) *ast.Node {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return nil
	}

	switch callee.Kind {
	case ast.KindIdentifier:
		if callee.Text() == findDOMNodeName {
			return callee
		}
		// oxc returns unconditionally from its identifier arm rather than falling through to the
		// member check, and this arm reproduces that by answering for the whole switch.
		return nil

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if !isReactDOMReceiver(access.Expression) {
			return nil
		}
		name := access.Name()
		// A private name (`ReactDOM.#findDOMNode()`) cannot reach the comparison, since a
		// `KindPrivateIdentifier` node's `Text()` carries its `#` and so never equals the bare
		// name. No kind check is written for it, because a branch that cannot change a verdict is a
		// branch no fixture can guard.
		if name != nil && name.Text() == findDOMNodeName {
			return name
		}

	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if !isReactDOMReceiver(access.Expression) {
			return nil
		}
		argument := access.ArgumentExpression
		if argument == nil {
			return nil
		}
		// A string or a no-substitution template is a static property name and oxc's
		// `static_property_info` answers for both. An identifier subscript, `ReactDOM[k](this)`,
		// names whatever `k` holds and is silent on the release binary.
		switch argument.Kind {
		case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
			if argument.Text() == findDOMNodeName {
				// The literal node, so the reported span includes its quotes or backticks. That is
				// thirteen bytes for `"findDOMNode"` and it is what upstream underlines.
				return argument
			}
		}
	}
	return nil
}

// findDOMNodeName is the one function this rule is about.
const findDOMNodeName = "findDOMNode"

// isReactDOMReceiver reports whether an expression is one of the three accepted namespace objects.
//
// The set is closed and the casing is exact, reproducing oxc's three `is_specific_id` calls.
// `ReactDom` is in there because upstream ships a fail case for it, and `reactdom` is not, which was
// measured rather than inferred from the absence of a case.
//
// Written as a helper rather than inline because both member arms above ask it, and because a
// mutation to the accepted set should have one place to land.
func isReactDOMReceiver(expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	if expression == nil || expression.Kind != ast.KindIdentifier {
		return false
	}
	switch expression.Text() {
	case "React", "ReactDOM", "ReactDom":
		return true
	}
	return false
}
