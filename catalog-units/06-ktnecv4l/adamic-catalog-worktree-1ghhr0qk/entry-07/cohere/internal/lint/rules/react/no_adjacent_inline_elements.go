package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/jsx"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// noAdjacentInlineElementsInlineNames is upstream's `inlineNames`, copied in order.
//
// It is the MDN list of inline elements, and it carries a few entries that are not obviously
// "inline text" (`script`, `object`, `map`, `input`, `select`, `textarea`, `button`) because MDN
// classes them that way. Copied whole rather than curated: a shorter list silences the rule on
// exactly the elements somebody added deliberately.
var noAdjacentInlineElementsInlineNames = map[string]bool{
	"a": true, "b": true, "big": true, "i": true, "small": true, "tt": true,
	"abbr": true, "acronym": true, "cite": true, "code": true, "dfn": true,
	"em": true, "kbd": true, "strong": true, "samp": true, "time": true, "var": true,
	"bdo": true, "br": true, "img": true, "map": true, "object": true, "q": true,
	"script": true, "span": true, "sub": true, "sup": true,
	"button": true, "input": true, "label": true, "select": true, "textarea": true,
}

var messageNoAdjacentInlineElements = rule.Message{
	Id: "inlineElement",
	Description: "These two inline elements sit directly against each other with nothing between " +
		"them, so the browser renders them as one run of text with no gap. What reads as two " +
		"separate links or labels in the source arrives as a single unbroken word on the page, " +
		"and the reader has no way to see where one ends. Put a space between them, or wrap them " +
		"in block-level elements that separate themselves.",
}

// NoAdjacentInlineElements flags two inline elements rendered with nothing between them.
//
//	valid:   <div><div></div><div></div></div>
//	valid:   <div><p></p><a></a></div>
//	valid:   <div><a></a> <a></a></div>
//	valid:   <div><a></a>&nbsp;<a></a></div>
//	valid:   React.createElement("div", undefined, [React.createElement("a"), " ", React.createElement("a")])
//	invalid: <div><a></a><a></a></div>
//	invalid: <div><a></a><span></span></div>
//	invalid: React.createElement("div", undefined, [React.createElement("a"), React.createElement("span")])
//
// Ported from `react/no-adjacent-inline-elements` in `eslint-plugin-react`, read from the clone at
// `lib/rules/no-adjacent-inline-elements.js`. No options (`schema: []`), one message, no fixer. The
// whole 17-case corpus was replayed against the installed build (7.37.5) through the ESLint Linter
// API before any code was written, and it agreed on all 17. Everything below was measured the same
// way.
//
// # The corpus's whitespace cases do not test whitespace, and that is the whole rule
//
// Six of upstream's fourteen passing cases separate two anchors with `&nbsp;` or with text, and
// they read as direct evidence that the rule checks for a gap. They are not evidence about that at
// all.
//
// `isInline` returns true for a `Literal` with no leading or trailing whitespace, and for a
// `JSXElement` whose name is on the inline list. Text written directly inside a tag is neither: its
// node type is `JSXText`, which no arm names, so it falls through to `return false`. A non-inline
// child breaks the adjacency run, so ANY text at all separates two elements, whitespace or not.
// Measured, changing the one thing the case's shape suggests matters:
//
//	<div><a></a> <a></a></div>       clean
//	<div><a></a>x<a></a></div>       clean, and this is the case that gives it away
//	<div><a></a>&nbsp;<a></a></div>  clean
//	<div><a></a>{x}<a></a></div>     clean, an expression container is not inline either
//	<div><a></a>{/* c */}<a></a></div>  clean
//	<div><a></a><a></a></div>        REPORTS
//
// So on the JSX side this rule is "two inline ELEMENTS with no node between them", and the
// whitespace regex is unreachable from it. A port that read those six cases as a whitespace test
// and implemented one would still pass all seventeen imported cases, because every fixture that
// could distinguish the two readings happens to use whitespace. The distinguishing input is a
// single non-space character, which upstream had no reason to write.
//
// # The whitespace regex is reachable only through createElement, where children are real literals
//
// In a `createElement` children array a string IS a `Literal`, so the regex decides:
//
//	[a, " ", a]     clean, the space is at both ends
//	[a, "x ", a]    clean, trailing whitespace is enough
//	[a, " x", a]    clean, leading whitespace is enough
//	[a, "x", a]     REPORTS, no whitespace at either end
//	[a, "", a]      REPORTS, an empty string has no whitespace to find
//	[1, a]          REPORTS, a number stringifies to something with no whitespace
//
// All measured. The empty-string case is the sharp one: upstream's `/(?:^\s|\s$)/` finds nothing in
// `""`, so the literal counts as inline and adjacency holds. Reproduced.
//
// # Upstream crashes here, and this port declines to
//
// `isInline`'s third arm is `astUtil.isCallExpression(node) && inlineNames.indexOf(node.arguments[0].value) > -1`.
// It tests that the node is a call and then indexes argument zero WITHOUT checking there is one and
// WITHOUT checking it is a createElement call. Any other call among the children takes it down:
//
//	React.createElement("div", undefined, [foo(), React.createElement("a")])
//	React.createElement("div", undefined, [React.createElement("a"), bar()])
//
// Both throw `Cannot read properties of undefined (reading 'value')` on the installed build,
// measured by catching around the Linter call. That is a genuine upstream defect rather than a
// judgment, and reproducing it here would be worse than useless: a panic in one rule costs every
// rule that whole file, because the walk recovers per file rather than per rule.
//
// So the arm is written with the guards upstream omits, and the resulting VERDICT is the one
// upstream would reach if it had not crashed: a call with no arguments, or whose first argument is
// not a string naming an inline element, is not inline. That is a divergence and it is stated
// rather than silent. Nothing else in the port improves on upstream.
//
// Note that the crash is unreachable from the JSX side, where a call only ever appears inside an
// expression container, which is not a child upstream inspects. Measured clean.
//
// # Which calls count as createElement, and where that is asked
//
// The rule asks twice and upstream asks DIFFERENTLY in the two places. The outer call, the one
// whose children are being examined, goes through `isCreateElement`, so the namespace must be React
// or the name must resolve to a react import. The inner children do not: `isInline` accepts ANY
// call whose first argument names an inline element. Measured, both halves:
//
//	NotReact.createElement("div", undefined, [NotReact.createElement("a"), NotReact.createElement("span")])
//
// is clean, because the OUTER call is not a createElement call, and the rule never looks inside. But
// with a React outer call and foreign inner calls the inner ones still count as inline, since
// nothing checks them. Reproduced as written: the outer gate uses this package's
// `isPragmaCreateElementCall` and the inner test looks only at the first argument.
//
// # Only an array of children is examined
//
// Upstream reads `'elements' in node.arguments[2] ? node.arguments[2].elements : undefined`, so a
// third argument that is not an array yields no children and the call is skipped. A single element
// passed directly, rather than in an array, is therefore never compared against anything. Measured
// clean.
//
// # A capitalised or dotted name is not on the list
//
// The list holds lowercase HTML names, and a component reference is a different name, so
// `<div><A></A><A></A></div>` and `<div><a.b></a.b><a.b></a.b></div>` are both clean. Measured. A
// member-named tag has no plain name to compare at all.
//
// # Only the first adjacent pair reports
//
// `validate` returns as soon as it reports, so a run of three inline elements produces ONE finding
// and a container holding two separated pairs also produces one. Both measured. A loop that kept
// going would report twice on inputs upstream reports once.
var NoAdjacentInlineElements = rule.Rule{
	Name: "react/no-adjacent-inline-elements",

	// Declared for the bare-call branch of `isPragmaCreateElementCall`, which asks the checker
	// whether `createElement` binds to a `react` import. The JSX arm needs nothing.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// isInlineJsxChild answers upstream's `isInline` for a node in a JSX child list.
		//
		// Only a JSX element counts. Text, an expression container and a fragment all fall through
		// to upstream's `return false`, which is what makes any of them break an adjacency run.
		isInlineJsxChild := func(child *ast.Node) bool {
			var opening *ast.Node
			switch child.Kind {
			case ast.KindJsxElement:
				opening = child.AsJsxElement().OpeningElement
			case ast.KindJsxSelfClosingElement:
				opening = child
			default:
				return false
			}
			if opening == nil {
				return false
			}
			tagName, _ := jsx.ElementParts(opening)
			// A member or namespaced tag has no plain name, which is what upstream reads, so
			// neither can be on the list.
			if tagName == nil || tagName.Kind != ast.KindIdentifier {
				return false
			}
			return noAdjacentInlineElementsInlineNames[tagName.Text()]
		}

		// isInlineCreateElementChild answers upstream's `isInline` for a node in a children array.
		//
		// Two arms here rather than one, because in this position a string literal is a `Literal`
		// and the whitespace test is reachable.
		isInlineCreateElementChild := func(child *ast.Node) bool {
			if noAdjacentInlineElementsIsUpstreamLiteral(child) {
				// `/(?:^\s|\s$)/` over `String(value)`. Whitespace at EITHER end makes it not
				// inline; an empty string has neither and so counts as inline.
				text := noAdjacentInlineElementsLiteralText(child)
				return !noAdjacentInlineElementsHasEdgeWhitespace(text)
			}
			if child.Kind == ast.KindCallExpression {
				// Upstream indexes `node.arguments[0].value` here with no guards at all, and any
				// call carrying no arguments takes the linter down. The guards are added; the
				// verdict is upstream's own for every input that does not crash it. See the doc
				// comment above.
				arguments := child.AsCallExpression().Arguments
				if arguments == nil || len(arguments.Nodes) == 0 {
					return false
				}
				first := arguments.Nodes[0]
				if first.Kind != ast.KindStringLiteral {
					return false
				}
				return noAdjacentInlineElementsInlineNames[first.Text()]
			}
			return false
		}

		// validate walks a child list and reports the first adjacent inline pair.
		//
		// Upstream reports on the CONTAINER rather than on either child, and returns immediately,
		// so one container yields at most one finding however many pairs it holds.
		validate := func(reported *ast.Node, children []*ast.Node, isInline func(*ast.Node) bool) {
			previousIsInline := false
			for _, child := range children {
				currentIsInline := isInline(child)
				if previousIsInline && currentIsInline {
					ctx.ReportNode(reported, messageNoAdjacentInlineElements)
					return
				}
				previousIsInline = currentIsInline
			}
		}

		return rule.Listeners{
			ast.KindJsxElement: func(node *ast.Node) {
				children := node.AsJsxElement().Children
				if children == nil {
					return
				}
				validate(node, children.Nodes, isInlineJsxChild)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if !isPragmaCreateElementCall(ctx, node) {
					return
				}
				arguments := node.AsCallExpression().Arguments
				// Upstream's guard is `arguments.length < 2 || !arguments[2]`, which needs a third
				// argument to exist even though it names the second in the length test.
				if arguments == nil || len(arguments.Nodes) < 3 {
					return
				}
				third := arguments.Nodes[2]
				// Only an array yields children. A single element passed directly is never
				// compared against anything, measured clean.
				if third.Kind != ast.KindArrayLiteralExpression {
					return
				}
				elements := third.AsArrayLiteralExpression().Elements
				if elements == nil {
					return
				}
				validate(node, elements.Nodes, isInlineCreateElementChild)
			},
		}
	},
}

// noAdjacentInlineElementsIsUpstreamLiteral reports whether a node is what upstream types as
// Literal.
//
// The same set `forbid-elements` needs, and for the same reason: upstream's `node.type === 'Literal'`
// covers the keywords and the numeric forms as well as strings, and a template literal is not one.
// Kept separate from that rule's copy rather than shared, because the two rules ask different
// questions of the answer and a shared helper would invite one of them to widen it for the other.
func noAdjacentInlineElementsIsUpstreamLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return true
	}
	return false
}

// noAdjacentInlineElementsLiteralText renders a literal the way upstream's implicit `String(value)`
// does.
//
// The keywords carry no text of their own and are spelled out. Everything else already holds its
// rendering, and none of these renderings can contain edge whitespace, which is why every non-string
// literal counts as inline.
func noAdjacentInlineElementsLiteralText(node *ast.Node) string {
	switch node.Kind {
	case ast.KindTrueKeyword:
		return "true"
	case ast.KindFalseKeyword:
		return "false"
	case ast.KindNullKeyword:
		return "null"
	}
	return node.Text()
}

// noAdjacentInlineElementsHasEdgeWhitespace is upstream's `/(?:^\s|\s$)/`.
//
// An alternation of two anchored tests rather than a trim comparison, which matters at one input:
// a string that is ENTIRELY whitespace matches the first alternative, and so does a trim test, so
// the two agree there. They differ on the empty string, where a trim comparison also finds no
// change and would answer the same. Written as the two anchor tests anyway, because that is the
// shape upstream wrote and the equivalence is an argument rather than a measurement.
//
// `text.IsWhitespace` for the class, which is JavaScript's `\s` exactly; the corpus exercises the
// space, and `&nbsp;` arrives as U+00A0, which the class accepts.
func noAdjacentInlineElementsHasEdgeWhitespace(value string) bool {
	if value == "" {
		return false
	}
	runes := []rune(value)
	return text.IsWhitespace(runes[0]) || text.IsWhitespace(runes[len(runes)-1])
}
