package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageJsxNoCommentTextnodes = rule.Message{
	Id: "putCommentInBraces",
	Description: "This looks like a comment but it is not one. Inside a tag's children, `//` " +
		"and `/*` are ordinary text, so React renders them to the page and the words the author " +
		"meant only for other developers become part of the interface. The failure is silent: " +
		"nothing errors, nothing warns, and the text shows up in the product. Wrap it in braces " +
		"as `{/* ... */}`, which is the form React reads as a comment and drops.",
}

// JsxNoCommentTextnodes flags text inside a tag's children that reads as a comment.
//
//	valid:   <div>{/* a real comment */}</div>
//	valid:   <div>{"// a string"}</div>
//	valid:   <div className="// not children" />
//	valid:   <pre>&#x2F;&#x2F; entity-escaped</pre>
//	valid:   <div>hello // not at a line start</div>
//	invalid: <div>// invalid</div>
//	invalid: <div>/* invalid */</div>
//	invalid: <>// invalid</>
//	invalid: <div>{'x'}
//	           // invalid
//	         </div>
//
// Ported from `react/jsx-no-comment-textnodes` in `eslint-plugin-react`, read from the clone at
// `lib/rules/jsx-no-comment-textnodes.js`. No options (`schema: []`), one message, no fixer. Every
// behaviour below was measured by driving the installed build (7.37.5) through the ESLint Linter
// API rather than reasoned from the source.
//
// # There is no file-suffix gate
//
// Three siblings in this package gate on the file being read as JSX, inherited from oxc. This rule
// was ported from the authority, which has no such gate. Measured under `.tsx`, `.jsx` and `.js`:
// all three report. A `.ts` file cannot hold JSX at all, so there is no fourth row to measure and
// the gate would be unobservable rather than wrong.
//
// # The predicate is a regex over the RAW source, and both halves of that matter
//
// Upstream tests `/^\s*\/(\/|\*)/m` against `getText(context, node)`, which is the source slice
// rather than any cooked value. Two consequences, both measured:
//
//   - The `m` flag anchors `^` at every LINE start, not only at the node's start. So a comment on
//     the third line of a multi-line text node reports, and that is the shape the corpus's own
//     failing cases are written in. A port anchoring only at the node start passes the two
//     single-line cases and goes silent on four of the seven.
//
//   - Reading the raw source is what makes entity-escaped text clean. `<pre>&#x2F;&#x2F; TODO</pre>`
//     is silent because the raw slice holds the ampersand form, and upstream ships two such cases.
//     A port reading a decoded value would report both, and no fixture written from the corpus
//     could see the difference, because the corpus asserts them as clean either way.
//
// `\s*` allows leading whitespace, so an indented comment reports. Text with `//` anywhere other
// than after a line's leading whitespace is clean: `<div>hello // not a comment</div>` is silent,
// measured.
//
// # One finding per text node, not per comment line
//
// The regex is a test rather than a scan, so a text node holding three comment lines reports once.
// Measured: `<div>\n // one\n // two\n</div>` reports once, and the same two lines separated by an
// expression container report twice, because the container splits them into two text nodes. Neither
// shape is in the corpus and the difference is invisible to a fixture asserting one id per input.
//
// # Upstream also visits every string literal, and none of them can ever report
//
// The rule registers `Literal` alongside `JSXText` and filters on the parent's type: it must
// contain "JSX" and be neither a `JSXAttribute` nor a `JSXExpressionContainer`. A string literal's
// only JSX parents are exactly those two, so the filter excludes every literal it could see. Probed
// over both positions in this parser: an attribute value's parent is `KindJsxAttribute` and a
// braced string's is `KindJsxExpression`, matching upstream's two exclusions. Measured against the
// installed rule, `<div className="// nope" />` and `<div>{"// ok"}</div>` are both silent.
//
// So the literal listener is not registered here. That is a deliberate divergence in mechanism and
// none in decision, and it is stated rather than silent: reproducing a visit that can never report
// would cost a listener over every string in the tree to reach the same answer.
//
// # Where the finding points, and why this does not use ReportNode
//
// Upstream passes the text node itself, so the span is the whole text node including its leading
// and trailing whitespace, not the comment within it.
//
// `ctx.ReportNode` cannot express that here. It routes through `rule.TokenRange`, which calls
// `GetRangeOfTokenAtPosition` to skip leading trivia, and a JSX text node whose content IS a
// comment is exactly the input that breaks: the scanner reads `// invalid` as trivia and returns
// the position of the NEXT token, past the node's own end. Probed on
// `const a = <div>// invalid</div>;` the node spans 15 to 25 while `TokenRange` answers 33 to 25,
// an INVERTED range whose start is after its end. Slicing the source with it panics.
//
// This is a defect in `TokenRange` for this node kind rather than in this rule, and it is worth
// naming because the trivia-skipping it performs is correct and load-bearing everywhere else. It
// does not reach any shipped rule today: `no-unescaped-entities` is the only other rule anchored on
// `KindJsxText` and it reports a sub-range through `ReportRange`, so it never asks `TokenRange`
// about a text node at all. This rule takes the same route for the same reason, and the second span
// fixture below pins the whole-node span so a later switch to `ReportNode` fails loudly instead of
// panicking on one shape and passing on the rest.
//
// Trivia skipping is not wanted here in any case: upstream's span deliberately includes the
// surrounding whitespace, so the node's own range is the correct answer as well as the safe one.
var JsxNoCommentTextnodes = rule.Rule{
	// No namespace prefix beyond the plugin's own. The config writes
	// `react/jsx-no-comment-textnodes` and the parity guard strips the namespace on a `/` boundary.
	Name: "react/jsx-no-comment-textnodes",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		sourceText := ctx.SourceFile.Text()

		return rule.Listeners{
			ast.KindJsxText: func(node *ast.Node) {
				// The parent test upstream performs is already satisfied by the node kind here. A
				// JsxText node's parent is a JSX element or a JSX fragment and nothing else, both
				// of which pass upstream's "contains JSX and is neither an attribute nor an
				// expression container" filter. Probed over five text positions including one in a
				// fragment; every parent came back KindJsxElement or KindJsxFragment.
				raw := sourceText[node.Pos():node.End()]
				if !jsxNoCommentTextnodesLooksLikeAComment(raw) {
					return
				}
				// `ReportRange` with the node's own span rather than `ReportNode`. See the rule
				// doc: `TokenRange` returns an inverted range for a text node whose content is a
				// comment, and upstream's span is the whole node including its whitespace anyway.
				ctx.ReportRange(core.NewTextRange(node.Pos(), node.End()), messageJsxNoCommentTextnodes)
			},
		}
	},
}

// jsxNoCommentTextnodesLooksLikeAComment reports whether any line of the raw text begins, after
// optional whitespace, with `//` or `/*`.
//
// This is upstream's `/^\s*\/(\/|\*)/m` written as a scan rather than as a regex, because the
// pattern is simple enough that a regex would be the slower and less readable of the two, and
// because `\s` and Go's `unicode.IsSpace` do not agree on the same character set. The difference is
// resolved deliberately below rather than inherited.
//
// The `m` flag is the whole reason this walks lines: `^` matches at every line start, so a comment
// on any line of a multi-line text node reports. Four of upstream's seven failing cases depend on
// it, and a version anchored only at the node start passes the other three, which is the shape of
// port that would ship green.
func jsxNoCommentTextnodesLooksLikeAComment(raw string) bool {
	atLineStart := true
	for index := 0; index < len(raw); index++ {
		character := raw[index]

		if character == '\n' {
			atLineStart = true
			continue
		}

		// JavaScript's `\s` inside a character class matches the horizontal and vertical
		// whitespace this loop skips, and it also matches `\n` itself, which is handled above so
		// that the line counter stays right. Only ASCII whitespace is skipped here.
		//
		// JavaScript's `\s` additionally matches several non-ASCII spaces, and a text node
		// beginning with one of those followed by `//` would report upstream and is silent here.
		// That divergence is stated rather than guessed at: it is not in the corpus, and it is left
		// unhandled because skipping a Unicode space would require decoding runes on every text
		// node in the tree to reach a case nobody writes. If it ever matters, the fix is to decode
		// here rather than to change the shape of the loop.
		if character == ' ' || character == '\t' || character == '\r' ||
			character == '\v' || character == '\f' {
			continue
		}

		if atLineStart && character == '/' && index+1 < len(raw) &&
			(raw[index+1] == '/' || raw[index+1] == '*') {
			return true
		}

		// Any other character means this line has begun with something that is not a comment, so
		// nothing later on this line can match. `^` in the pattern is what makes that true, and it
		// is why `<div>hello // not a comment</div>` is clean.
		atLineStart = false
	}
	return false
}
