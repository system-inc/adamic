package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageMultiStr = rule.Message{
	Id: "multilineString",
	Description: "This string is continued onto the next line with a trailing backslash, which " +
		"the language accepts and almost no reader expects. Whatever indentation the next line " +
		"carries goes into the value, so the string changes meaning when somebody reformats the " +
		"file, and the backslash has to be the last character on its line or the file stops " +
		"parsing. Concatenation or a template literal says the same thing without either hazard.",
}

// NoMultiStr flags a string literal continued across a line break with a trailing backslash.
//
//	valid:   'Line 1 Line 2'
//	valid:   'Line 1 \n Line 2'      // an escape, so no real line break is in the source
//	valid:   `Line 1
//	          Line 2`                 // a template literal, which is the supported spelling
//	invalid: 'Line 1 \
//	          Line 2'
//
// # The subject is the raw text, and the distinction it draws is sharp
//
// Upstream tests `LINEBREAK_MATCHER` against `node.raw`, not against the value. A `\n` escape and a
// real line continuation produce nearly the same value and completely different source: the escape
// puts one newline character into the string and leaves the source on one line, while the
// continuation puts nothing into the string and spreads the source across two. Only the second is
// what this rule objects to, and only the raw text can tell them apart. Reading the cooked value
// would report the escape, which is the ordinary and correct way to write a newline.
//
// # Four terminators, not one
//
// `LINEBREAK_MATCHER` matches a carriage-return-line-feed pair, or any single line feed, carriage
// return, line separator (U+2028), or paragraph separator (U+2029). All four count, and three of
// upstream's five failing cases use exactly the three that are not a line feed, which is how the
// set is pinned rather than guessed. The pair branch exists so the pattern consumes both bytes of
// a Windows break as one match; since this rule only asks whether a break is present at all, that
// branch makes no difference here and the membership test below says the same thing.
//
// # The JSX exemption is real, and it is about the PARENT
//
// Upstream declines when `node.parent.type` starts with `JSX`. The shape that reaches it is a
// string-valued JSX attribute, whose parent is a `JSXAttribute`, and the exemption is why a
// multiline attribute value is left alone. Measured against the installed ESLint build,
// `<div attr='Line 1 \` continued onto a second line is clean while the identical string in a call
// argument reports.
//
// This is not vestigial in our tree. Our parser produces exactly `KindJsxAttribute` as the parent
// of such a string, measured, so the branch is reachable and load bearing. Dropping it as
// simplification would report every multiline attribute value in every component file, a
// false-positive class the imported corpus cannot see, because the only JSX case upstream ships is
// a `<div>` whose newlines are in JSX text rather than in a string literal.
//
// Upstream's test is a prefix test over a type name, which matches every JSX node kind rather than
// the attribute alone, and the breadth is reproduced here rather than narrowed to the shapes that
// first come to mind. That is not caution, it is a measurement that a narrower version already
// failed once.
//
// Enumerating parents over ten well-formed JSX shapes gives two: `KindJsxAttribute` for a bare
// attribute value and `KindJsxExpression` for a string inside braces, whether those braces are an
// attribute value or element content. Narrowing the enumeration to exactly those two survived the
// entire fixture set. Re-running the enumeration over malformed and unusual source as well, which
// the parser recovers from rather than refusing, turned up a third: `KindJsxSpreadAttribute`, from
// a string spread directly as attributes rather than through an object literal. Upstream is silent
// on that input; measured on the installed build. So the arm is reachable, load bearing, and now
// pinned by a fixture.
//
// The remaining kinds in the enumeration are not reachable as a string literal's parent by anything
// tried, well formed or not. They are kept because upstream's question is "is this string part of
// JSX syntax" and its prefix test answers yes for all of them, so a narrower test would be a
// different rule that happens to agree on today's parse shapes. That verdict names the shapes it
// was taken over: adding a caller or a parser change can expire it.
//
// A string in an object literal spread into an element does report, because its parent is a
// `PropertyAssignment` rather than anything JSX; measured both ways.
//
// # Templates are not this rule's surface
//
// A template literal spanning lines is the repair rather than the defect, and it is a different
// node in ESTree that never reaches the handler. Measured: a multiline template is clean upstream.
// Here that falls out of the listener map.
//
// The rule reports without a repair, matching upstream, which ships no fixer and marks itself
// `frozen`. Any repair would have to choose between concatenation and a template and would have to
// decide what happens to the leading whitespace on the continued line, which is exactly the value
// the author may or may not have meant to include.
var NoMultiStr = rule.Rule{
	Name: "no-multi-str",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				if isPartOfJsxSyntax(node.Parent) {
					return
				}

				literalRange := rule.TokenRange(ctx.SourceFile, node)
				if !containsLineTerminator(ctx.SourceFile.Text()[literalRange.Pos():literalRange.End()]) {
					return
				}

				ctx.ReportNode(node, messageMultiStr)
			},
		}
	},
}

// isPartOfJsxSyntax reports whether a node belongs to the JSX grammar rather than to ordinary
// expression syntax.
//
// This is upstream's `node.type.indexOf("JSX") === 0` written as a kind test. Upstream asks it of a
// string literal's parent, and the only parent that occurs in practice is a JSX attribute, but the
// prefix test covers every JSX node so the enumeration below does too. Reproducing the breadth
// rather than the single observed shape keeps the port faithful to the question upstream asks
// instead of to the answer it happens to get on current parse trees.
func isPartOfJsxSyntax(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindJsxElement,
		ast.KindJsxSelfClosingElement,
		ast.KindJsxOpeningElement,
		ast.KindJsxClosingElement,
		ast.KindJsxFragment,
		ast.KindJsxOpeningFragment,
		ast.KindJsxClosingFragment,
		ast.KindJsxAttribute,
		ast.KindJsxAttributes,
		ast.KindJsxSpreadAttribute,
		ast.KindJsxExpression,
		ast.KindJsxText,
		ast.KindJsxNamespacedName:
		return true
	}
	return false
}

// containsLineTerminator reports whether raw source text holds one of the four characters the
// language treats as ending a line.
//
// The set is upstream's four terminators: line feed, carriage return, U+2028, and U+2029. The
// first two are single bytes; the line separator and paragraph separator are three bytes each in
// UTF-8, both beginning `0xE2 0x80`, so they are matched on their byte sequences rather than by
// decoding every rune. That keeps the scan a byte walk over text that is almost always clean, and
// it is exact rather than approximate: no other character encodes to either of those three-byte
// sequences.
func containsLineTerminator(raw string) bool {
	for index := 0; index < len(raw); index++ {
		switch raw[index] {
		case '\n', '\r':
			return true
		case 0xE2:
			// U+2028 is E2 80 A8 and U+2029 is E2 80 A9. Nothing else shares that prefix and
			// third byte, so this identifies them without decoding.
			if index+2 < len(raw) && raw[index+1] == 0x80 &&
				(raw[index+2] == 0xA8 || raw[index+2] == 0xA9) {
				return true
			}
		}
	}
	return false
}
