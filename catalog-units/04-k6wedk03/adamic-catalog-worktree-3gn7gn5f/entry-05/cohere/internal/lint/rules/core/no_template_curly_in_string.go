package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageTemplateCurlyInString = rule.Message{
	Id: "unexpectedTemplateExpression",
	Description: "This ordinary string contains a template placeholder, so the placeholder is not " +
		"a placeholder: it is four literal characters that will be sent, rendered, or stored " +
		"exactly as written. The failure is silent, because the string is a perfectly valid " +
		"string and nothing at runtime objects to it. Backticks are almost always what was meant.",
}

// NoTemplateCurlyInString flags template placeholder syntax inside a single or double quoted string.
//
//	valid:   `Hello, ${name}`      // a real template literal, where the placeholder works
//	valid:   'Hello, name'
//	valid:   '$2'                  // a dollar with no brace
//	valid:   '${'                  // an opening with nothing to close it
//	valid:   '{foo}'               // braces with no dollar
//	invalid: 'Hello, ${name}'
//	invalid: "Hello, ${name}"
//	invalid: '${greeting}, ${name}'
//
// # What the rule tests, and why it is the cooked value
//
// Upstream is four lines: a `Literal` whose value is a string is matched against
// `/\$\{[^}]+\}/u`. The value rather than the raw text, which is a real distinction rather than an
// incidental one, and it runs in both directions.
//
// A dollar sign written as an escape still counts, because the escape is resolved before the match:
// measured against the installed build, `'${a}'` reports. Reading the raw source text instead
// would find `${a}`, no literal dollar-brace, and stay silent on a string that will render
// with a live-looking placeholder in it. Our parser hands the cooked value on `node.Text()`
// already, so this asks the node rather than reconstructing anything.
//
// # The pattern is narrower than it looks, in three ways worth stating
//
// `[^}]+` requires at least one character, so `'${}'` is clean; measured. It is also negated on the
// closing brace rather than lazy, so `'${{a}}'` matches `${{a}` and reports, and `'${a'` never
// finds a close and is silent. And the placeholder must be adjacent: `'$ {a}'` is clean because a
// space sits between the dollar and the brace.
//
// # One finding per literal, not one per placeholder
//
// Upstream calls `regex.test`, which answers a boolean, so a string holding two placeholders
// reports once. `'${a}${b}'` produces a single finding, measured. A port that looped over every
// match would report twice on an input upstream reports once, and no message-id fixture asserting a
// single id would notice the count until it was written down.
//
// # Only string literals, and the surface is smaller than "every Literal"
//
// A template literal is a different node in ESTree and never reaches the handler, which is the
// point: a placeholder there is doing its job. A regular expression is a `Literal` upstream and is
// declined by the `typeof value === "string"` test, so `/\${x}/` is clean; measured. Our tree has
// separate kinds for all three, so the discrimination is the listener rather than a value test.
//
// The rule reports without a repair, matching upstream, which ships no fixer. There are two
// plausible repairs and they mean different things: retyping the quotes as backticks assumes the
// names in the placeholder resolve at that point, and escaping the dollar assumes the literal text
// was intended. Choosing between them needs the author.
var NoTemplateCurlyInString = rule.Rule{
	Name: "no-template-curly-in-string",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				if !containsTemplatePlaceholder(node.Text()) {
					return
				}
				ctx.ReportNode(node, messageTemplateCurlyInString)
			},
		}
	},
}

// containsTemplatePlaceholder reports whether text holds `${`, at least one character that is not a
// closing brace, and then a closing brace.
//
// This is upstream's `/\$\{[^}]+\}/u` written out. It is a fixed pattern over a cooked string with
// no captures and no offsets wanted, so a hand scan says the same thing a compiled regex would
// while costing nothing per literal in a tree where almost every string is clean.
//
// The scan looks for `${` and then for a `}` that is not the very next byte. Rejecting the adjacent
// brace is what implements the `+` rather than a `*`, and it is the whole difference between
// `'${}'`, which upstream leaves alone, and `'${ }'`, which it reports.
//
// A `${` that never closes does not end the search, because a later one may close: `'${a${b}'`
// holds no complete placeholder starting at the first dollar and a complete one starting at the
// second, and upstream's regex engine backtracks to find it. Restarting the scan one byte past each
// failed opening is what reproduces that.
func containsTemplatePlaceholder(text string) bool {
	for searchFrom := 0; ; {
		opening := strings.Index(text[searchFrom:], "${")
		if opening < 0 {
			return false
		}
		bodyStart := searchFrom + opening + len("${")

		closing := strings.IndexByte(text[bodyStart:], '}')
		// A closing brace immediately after the opening leaves the body empty, which `[^}]+`
		// refuses. Anything further along means at least one non-brace byte precedes it, because
		// this is the first brace after the opening.
		if closing > 0 {
			return true
		}

		searchFrom = bodyStart
	}
}
