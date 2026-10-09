package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageScriptUrl = rule.Message{
	Id: "unexpectedScriptURL",
	Description: "A `javascript:` URL runs whatever follows the colon as code, so this string is " +
		"an eval that happens to be spelled as a link. Anything that can reach the value that " +
		"builds it can execute in the page, and a content security policy that blocks inline " +
		"script blocks this too, so it tends to fail in production rather than in review. A real " +
		"handler does the same work without turning a string into a program.",
}

// NoScriptUrl flags a string or template literal whose value begins `javascript:`.
//
//	valid:   'Hello World!'
//	valid:   'xjavascript:'         // the scheme has to be at the start
//	valid:   ' javascript:x'        // leading whitespace, so it does not start with it
//	valid:   foo`javaScript:`       // a tagged template, where the tag decides what the text means
//	valid:   `${foo}javascript:`    // a template with an expression, so the value is not static
//	invalid: 'javascript:void(0);'
//	invalid: `JavaScript:`
//	invalid: 'JAVASCRIPT:alert(1)'
//
// # Two surfaces, and the second one is where the decisions are
//
// Upstream registers `Literal` and `TemplateLiteral`, and both route into one check that asks
// `astUtils.getStaticStringValue`. For a string that is the value itself. For a template that
// helper returns the cooked text only when the template has no expressions and exactly one quasi,
// so a template written `${foo}javascript:` is declined for having a value that is not knowable
// statically, rather than for any judgment about what it holds.
//
// Our tree splits those cases at the parser: a template with no expressions is
// `KindNoSubstitutionTemplateLiteral` and one with expressions is `KindTemplateExpression`. So the
// zero-expression test upstream performs inside the helper is a listener key here, and the
// expression-carrying kind is simply not listened for. That is the same decision reached earlier.
//
// # The tagged-template exemption is about the DIRECT parent, and only that
//
// Upstream declines a template whose `parent` is a `TaggedTemplateExpression`, because a tag
// receives the raw strings and decides what they mean, so the text is an argument rather than a
// URL. The check is on the immediate parent and nothing wider, which has a consequence upstream's
// corpus does not write: a template nested INSIDE a tagged template's interpolation still reports.
// Measured against the installed ESLint build, a template written foo`x${`javascript:`}y` reports,
// because the inner template's parent is the outer template rather than the tagged expression.
// A port asking "is this anywhere under a tagged template" would be silent there, and every
// imported fixture would still pass.
//
// # Comparison is case-insensitive and anchored at the start
//
// Upstream lowercases the value and takes `indexOf(...) === 0`. Both halves matter and each has a
// clean case in the corpus: `JavaScript:` reports and `xjavascript:` does not. The anchor is strict
// about whitespace too, which the corpus does not cover: `' javascript:x'` is clean, measured.
//
// Case folding is ASCII only, and that is fidelity rather than an optimization. Go's
// `strings.EqualFold` folds Unicode, so it treats the long s (U+017F) as equal to `s` and would
// report a string spelling the scheme with one. JavaScript's `toLowerCase` leaves that character
// alone, so upstream is silent on it: measured against the installed build, that input is clean
// while the plain spelling beside it reports. An enumeration over every rune below U+11000 found
// the long s to be the only character on which the two folds disagree for these eleven letters, so
// the ASCII fold below is exact for this comparison rather than merely close.
//
// # One case upstream declines for a reason that does not survive the translation
//
// The `Literal` arm is guarded by `node.value && typeof node.value === "string"`, and the first
// half of that is a truthiness test, so an EMPTY string never reaches the check. That is invisible
// here because an empty string cannot start with `javascript:` anyway, and it is recorded rather
// than reproduced: there is no input on which the guard changes the answer. The second half is what
// declines a regular expression, whose value is a truthy object; measured, `/javascript:/` is clean
// upstream, and here that falls out of the listener map instead.
//
// The rule reports without a repair, matching upstream, which ships no fixer. There is no
// mechanical repair: replacing the URL means writing a handler somewhere else.
var NoScriptUrl = rule.Rule{
	Name: "no-script-url",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				if !beginsWithScriptScheme(node.Text()) {
					return
				}
				ctx.ReportNode(node, messageScriptUrl)
			},

			// A template with no expressions, which is the only template shape whose value is
			// statically knowable. One carrying expressions is a different kind and is not listened
			// for, which is upstream's zero-expression test expressed as a listener key.
			ast.KindNoSubstitutionTemplateLiteral: func(node *ast.Node) {
				// The direct parent only. A template nested inside a tagged template's
				// interpolation is not itself tagged and does report; see the doc comment.
				if node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression {
					return
				}
				if !beginsWithScriptScheme(node.Text()) {
					return
				}
				ctx.ReportNode(node, messageScriptUrl)
			},
		}
	},
}

// scriptScheme is the prefix upstream tests for, in the case it compares against.
const scriptScheme = "javascript:"

// beginsWithScriptScheme reports whether a literal's value starts with `javascript:`, ignoring case.
//
// This is upstream's `value.toLowerCase().indexOf("javascript:") === 0`. Anchored rather than a
// search, so `'xjavascript:'` is clean, and it is strict about what precedes: a single leading
// space makes the value clean, measured against the installed build.
//
// Only the prefix is examined rather than the whole string. Upstream lowercases everything and then
// looks at the first eleven characters, which is the same comparison for every input and copies the
// rest of the string for nothing.
//
// The fold is ASCII by hand rather than `strings.EqualFold`, which folds Unicode and would accept
// the long s (U+017F) where upstream's `toLowerCase` does not. See the doc comment for the
// measurement. Every byte of the scheme is a lowercase ASCII letter or a colon, so folding the
// candidate byte down and comparing is the whole of it, and a multi-byte character can never
// compare equal to any of them.
func beginsWithScriptScheme(value string) bool {
	if len(value) < len(scriptScheme) {
		return false
	}
	for index := 0; index < len(scriptScheme); index++ {
		character := value[index]
		if character >= 'A' && character <= 'Z' {
			character += 'a' - 'A'
		}
		if character != scriptScheme[index] {
			return false
		}
	}
	return true
}
