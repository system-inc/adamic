package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoNamespace = rule.Message{
	Id: "noNamespace",
	Description: "This element names a namespace, and React has no notion of one. The colon in a " +
		"tag like `svg:circle` is meaningful in XML and in real XHTML documents, and JSX borrows " +
		"the syntax without borrowing the meaning: React reads the whole string as a single tag " +
		"name and hands it to the host, which has no element by that name. Nothing renders and " +
		"nothing warns. Write the element without the prefix.",
}

// NoNamespace flags a React element written with an XML namespace prefix.
//
//	valid:   <testcomponent />
//	valid:   <object.TestComponent />
//	valid:   React.createElement("TestComponent")
//	invalid: <ns:testcomponent />
//	invalid: React.createElement("ns:testcomponent")
//
// Ported from `eslint-plugin-react`'s `no-namespace.js`, which is the authority. `meta.schema` is
// `[]`, so there is no option surface, checked against the installed build rather than taken from
// the inventory column. `meta.fixable` is absent, so there is no repair to port; a namespace has no
// mechanical correction, since removing the prefix and keeping the local name are different edits
// with different meanings and only the author knows which was intended.
//
// # Two arms, and the second is narrower than it looks
//
// Upstream listens on `JSXOpeningElement` and on `CallExpression`. The JSX arm renders the tag
// through `jsx-ast-utils`'s `elementType` and tests for a colon. The call arm asks
// `isCreateElement(context, node)` and then reads a string first argument.
//
// **The call arm does NOT use the same createElement predicate the rest of this package does, and
// the difference is the whole reason this rule does not call the shelf.** `react.IsCreateElementCall`
// accepts a bare `createElement(...)`, accepts any object (`Preact.createElement`), and accepts a
// computed member (`React["createElement"]`). Upstream's `util/isCreateElement.js` accepts none of
// those here: it requires a static member whose object name is exactly the configured pragma, or a
// bare call that `isDestructuredFromPragmaImport` can trace back to a React import. Measured against
// the installed `eslint-plugin-react` through the Linter API, and every line contradicts the shelf:
//
//	React.createElement("ns:x")                                     REPORTS
//	createElement("ns:x")                                           SILENT   no React import
//	import {createElement} from "react"; createElement("ns:x")      REPORTS   destructured
//	import {createElement} from "preact"; createElement("ns:x")     SILENT   wrong package
//	Foo.createElement("ns:x")                                       SILENT   object is not the pragma
//	Preact.createElement("ns:x")                                    SILENT   same
//	document.createElement("ns:x")                                  SILENT   same, not by a special case
//	React["createElement"]("ns:x")                                  SILENT   computed member declines
//	const React = 1; React.createElement("ns:x")                    REPORTS   purely syntactic
//
// The last line is the tell that no resolution happens: a local `React` holding a number still
// satisfies the arm, because the check is on the identifier's spelling. Reproducing that is
// fidelity, and reaching for the checker to "improve" it would report differently from the rule
// being replaced.
//
// **Both halves are reproduced, and neither is written here.** `isPragmaCreateElementCall` already
// exists in this package, on `checked_requires_onchange_or_readonly.go`, ported from the same
// `util/isCreateElement.js` branch for branch and carrying its own measurements: the parenthesis
// skip, the four destructured shapes that `bindsToPragmaImport` resolves through the checker, and a
// stated divergence about declaration ordering. A first draft of this rule wrote a second, narrower
// copy of it that declined the destructured half, and the build refused the redeclaration. That
// refusal was right and the existing helper is better than what it replaced, which is exactly what
// the collision is for.
//
// **The cost of reusing it is that this rule declares the type checker.** The bare-call branch asks
// the checker which binding a `createElement` identifier resolves to. That is real name resolution
// rather than a scope flag, so the declaration is earned rather than defensive. The ahra tree writes
// zero bare `createElement` calls today, measured, so the checker buys nothing here right now and is
// declared because the rule's correctness rests on it rather than because the tree exercises it. The
// alternative was a second helper differing from the first in one direction, which is the shape the
// house rules refuse.
//
// # The colon test is on the whole rendered name, not on a namespace node
//
// Upstream never asks whether the parser produced a namespaced name. It renders the tag to a string
// and calls `indexOf(':') === -1`. That distinction is invisible in the JSX arm, where our parser
// hands back a `KindJsxNamespacedName` for exactly the inputs that carry a colon, and it is
// load-bearing in the call arm, where the argument is an ordinary string literal whose contents the
// parser has no opinion about. `React.createElement(":")` and `React.createElement("a:b:c")` both
// report, measured, which a namespace-shaped test would decline.
//
// # What our parser does with a namespaced tag, probed rather than assumed
//
//	<ns:testcomponent />      KindJsxNamespacedName, ns="ns" name="testcomponent", Text "ns:testcomponent"
//	<object.testcomponent />  KindPropertyAccessExpression
//	<a.b.c />                 nested KindPropertyAccessExpression
//	<TestComponent />         KindIdentifier
//	<this.Foo />              KindPropertyAccessExpression whose object is KindThisKeyword
//
// **`node.Text()` PANICS on a property-access tag name.** `ast.Node.Text` has no case for
// `*ast.PropertyAccessExpression` and reaches its unhandled-case panic, which the walk recovers per
// FILE rather than per rule, so one dotted tag name would cost every rule in this package every
// finding in that file. The first version of this probe hit it directly. That is why the JSX arm
// reads the namespaced kind and never calls `Text()` on an arbitrary tag: the dotted forms are all
// clean upstream anyway, so there is nothing to gain from rendering them and a whole file to lose.
//
// # Both element kinds, because our parser splits what upstream's joins
//
// Every reporting JSX case upstream ships is self-closing. Our parser gives a self-closing element
// its own `KindJsxSelfClosingElement` with no opening element inside it, so a rule listening only on
// `KindJsxOpeningElement` would be silent on all eight of them while every non-JSX fixture stayed
// green. `jsx_props_no_spread_multi.go` established this here first.
var NoNamespace = rule.Rule{
	Name: "react/no-namespace",
	// Declared for the bare-call branch of `isPragmaCreateElementCall`, which asks the checker
	// which binding `createElement` resolves to. The JSX arm needs nothing from it.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		reportElement := func(node *ast.Node) {
			// `jsx.ElementParts` rather than a switch over the two kinds. A first draft wrote the
			// switch by hand and `TestRulePackagesDoNotReachPastWrappedAccessors` refused it,
			// correctly: this is exactly the question that helper decides, and two spellings of one
			// decision is how the two drift.
			tagName, _ := jsx.ElementParts(node)

			// The kind test stands in for upstream's colon test on the rendered name, and it is a
			// narrowing that costs nothing: our parser produces this kind for exactly the tags
			// carrying a colon. Reading Text() on any other kind would panic on a dotted tag.
			if tagName == nil || tagName.Kind != ast.KindJsxNamespacedName {
				return
			}
			ctx.ReportNode(node, messageNoNamespace)
		}

		return rule.Listeners{
			ast.KindJsxOpeningElement:     reportElement,
			ast.KindJsxSelfClosingElement: reportElement,
			ast.KindCallExpression: func(node *ast.Node) {
				// Guarded because a typed rule handed a nil checker goes silent rather than
				// crashing on `GetSymbolAtLocation`, and a vacuous green is the more dangerous of
				// the two failures. The JSX arm above needs no checker and deliberately sits
				// outside this guard, so a nil checker costs only the bare-call spelling.
				if ctx.TypeChecker == nil {
					return
				}
				call := node.AsCallExpression()
				if !isPragmaCreateElementCall(ctx, node) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				// A string literal specifically. `React.createElement(null)`, `(true)` and `({})`
				// are three of upstream's own clean cases, and a template literal is silent too,
				// measured: upstream tests `type === 'Literal'`, which a template is not.
				first := call.Arguments.Nodes[0]
				if first.Kind != ast.KindStringLiteral {
					return
				}
				if !strings.Contains(first.Text(), ":") {
					return
				}
				ctx.ReportNode(node, messageNoNamespace)
			},
		}
	},
}
