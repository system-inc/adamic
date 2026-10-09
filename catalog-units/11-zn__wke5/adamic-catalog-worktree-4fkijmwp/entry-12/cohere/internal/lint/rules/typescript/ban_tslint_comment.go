package typescript

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// banTslintCommentDirective is upstream's ENABLE_DISABLE_REGEX, copied from the same tslint source
// it cites: https://github.com/palantir/tslint/blob/95d9d95/src/enableDisableRules.ts#L32
//
// It is anchored, so a directive has to open the comment's interior after optional whitespace. That
// anchor is what makes "a comment that merely mentions tslint" clean, which is four of upstream's
// five passing cases.
//
// The trailing group is the part a reader will misread. It requires the directive to be followed by
// a colon, whitespace, or end of interior, so `tslint:disabled` and `tslint:disable-foo` are silent
// while `tslint:disable:` and `tslint:disable extra` report. Measured against the installed
// @typescript-eslint 8.67.0 through the ESLint 10.8.1 Linter API rather than read off the pattern.
var banTslintCommentDirective = regexp.MustCompile(`^\s*tslint:(enable|disable)(?:-(line|next-line))?(:|\s|$)`)

// messageBanTslintComment is upstream's `commentDetected`, whose text interpolates the comment.
//
// `rule.Message` has no interpolation layer, so upstream's `{{ text }}` becomes concatenation here.
// The rendered comment is worth carrying rather than dropping: a file holding several directives
// produces several findings under one id, and the quoted text is the only thing separating them.
func messageBanTslintComment(text string) rule.Message {
	return rule.Message{
		Id: "commentDetected",
		Description: "This file carries the tslint directive " + text + ". tslint has been " +
			"deprecated since 2019 and nothing in this project reads its directives, so the " +
			"comment suppresses nothing and instead reads as a live suppression to anyone who " +
			"finds it. Delete it, and if the code it guarded still warrants a suppression, write " +
			"the equivalent eslint-disable comment naming the rule.",
	}
}

// BanTslintComment flags the `// tslint:<flag>` directive comments left behind by tslint.
//
//	valid:   // some other comment
//	valid:   // TODO: this is a comment that mentions tslint
//	valid:   /* another comment that mentions tslint */
//	valid:   /** tslint:disable */          a JSDoc star sits before the directive
//	valid:   // tslint:disabled             the directive must end at a colon, space, or the end
//	invalid: /* tslint:disable */
//	invalid: // tslint:disable-next-line:rule1 rule2
//	invalid: someCode(); // tslint:disable-line
//
// Ported from `@typescript-eslint/ban-tslint-comment`, reading the clone at
// `packages/eslint-plugin/src/rules/ban-tslint-comment.ts` and measuring every verdict below
// against the installed 8.67.0 build driven through the ESLint 10.8.1 Linter API.
//
// # A comment, not a node
//
// A directive is trivia, so there is no kind to anchor on. Upstream runs once per file out of a
// `Program` listener over `sourceCode.getAllComments()`; this runs out of `KindSourceFile` over
// `comments.ForFile`, the shared per-file scan.
//
// # No file gate, and that is a deliberate difference from its sibling
//
// `ban-ts-comment` in this package declines a JavaScript file, because a `@ts-ignore` there is how
// JavaScript opts into checking. This rule has no such gate and must not inherit one by analogy:
// upstream registers a plain `Program` visitor with no source-type test, and measured on the
// installed build, `// tslint:disable` reports in `a.js`, `a.mjs`, `a.ts` and `a.tsx` alike. A
// tslint directive is dead in every language it can appear in.
//
// # The subject is the comment's INTERIOR, and the span is the WHOLE comment
//
// The regex is tested against `c.value`, which is the text between the delimiters, so a JSDoc
// comment's leading star defeats the anchor and `/** tslint:disable */` is silent while
// `/* tslint:disable */` reports. The report is `node: c`, the comment itself, so the finding spans
// the delimiters that the match never saw. Both halves measured: the JSDoc case reports zero
// diagnostics, and `/* tslint:disable */` reports column 1 through end column 21 on a twenty-byte
// file.
//
// # The rendered text is normalized, not quoted
//
// `toText` trims the interior and rejoins it with single spaces around the delimiters, so the
// message never shows the comment's own spacing. Measured: `//   tslint:disable-line   ` renders as
// `// tslint:disable-line`, and `/*\n\ttslint:enable\n\t*/` renders as `/* tslint:enable */`. The
// trim is at the ends only, so an interior newline survives: `/* tslint:disable\n more */` renders
// with that newline intact. This matters because the corpus asserts the rendered text on every one
// of its eight reporting cases.
//
// # The fixer is destructive, and this port declines the destructive half
//
// Upstream removes `[start - 1, end + 1)`, taking one character before the comment and one after
// it, unconditionally except at column zero. On its own corpus that is exactly right: it removes
// the space in `someCode(); // tslint:disable-line` and the newline in a comment on its own line,
// and all eight upstream outputs reproduce byte for byte under this reading.
//
// Off the corpus it deletes code. Measured on the installed build with `cohereAndFix`:
//
//	x;// tslint:disable                silently becomes `x`      the semicolon is eaten
//	/* tslint:disable */let x = 1;     becomes `et x = 1;`       the `l` is eaten
//	\t// tslint:disable-line           becomes ``                the indent is eaten
//	x;  // tslint:disable              becomes `x; `             one of two spaces survives
//
// The second is refused by this project's parse guard. The first and third are not: both produce
// source that parses, so the guard is structurally unable to see them, and the edit engine applies
// a fix unattended. So this port narrows the range to the comment itself plus at most one character
// on each side WHEN THAT CHARACTER IS WHITESPACE, which reproduces every corpus output while
// declining to delete a character that is not blank. `banTslintCommentFixRange` holds the
// measurement and the four cases where the two readings differ.
//
// The judgment upstream makes is reproduced exactly: the same inputs report, at the same span, with
// the same rendered text. Only the repair is narrower, and it is narrower in the direction that
// cannot destroy source.
var BanTslintComment = rule.Rule{
	Name: "@typescript-eslint/ban-tslint-comment",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				reportBanTslintComments(ctx)
			},
		}
	},
}

// reportBanTslintComments is upstream's `Program` body.
func reportBanTslintComments(ctx rule.Context) {
	sourceText := ctx.SourceFile.Text()

	for _, comment := range comments.ForFile(ctx) {
		interior, readable := banTslintCommentInterior(comment)
		if !readable {
			continue
		}
		if !banTslintCommentDirective.MatchString(interior) {
			continue
		}

		ctx.ReportRangeWithFixes(comment.Range,
			messageBanTslintComment(banTslintCommentRenderedText(interior, comment.IsBlock)),
			rule.RemoveRange(banTslintCommentFixRange(comment, sourceText)))
	}
}

// banTslintCommentInterior is ESLint's `comment.value`: the text between a comment's delimiters.
//
// A line comment loses its two leading slashes and nothing else, having no closing delimiter. A
// block comment loses two characters from each end.
//
// The boolean is false for a block comment with no room for an interior. That case is reachable
// through error recovery rather than through well-formed source: the parser hands back an
// unterminated `/*` as a two-byte block comment and `/*/` as a three-byte one, and subtracting two
// from each end of those yields a start past its end, which panics on the slice. Probed on this
// parser rather than assumed; the same measurement is recorded on `commentContent` in
// `ban_ts_comment.go`, which reaches the identical shape by the identical route.
func banTslintCommentInterior(comment comments.Comment) (string, bool) {
	text := comment.Text

	if !comment.IsBlock {
		return text[2:], true
	}
	if len(text) < 4 {
		return "", false
	}
	return text[2 : len(text)-2], true
}

// banTslintCommentRenderedText is upstream's `toText`.
//
// The interior is trimmed and rejoined with a single space inside each delimiter, so the message
// shows a normalized comment rather than the one in the file. `Array.join(' ')` on two elements is
// one separator, which is why a line comment renders with exactly one space after the slashes
// however many the source had.
func banTslintCommentRenderedText(interior string, isBlock bool) string {
	trimmed := text.TrimWhitespace(interior)
	if isBlock {
		return "/* " + trimmed + " */"
	}
	return "// " + trimmed
}

// banTslintCommentFixRange is the deliberately narrowed version of upstream's removal range.
//
// Upstream computes `[getIndexFromLoc(start.column - 1), getIndexFromLoc(end.column) + 1)`, so it
// always widens by one on each side, except that a comment starting at column zero cannot widen
// left. The intent is legible and correct on the shapes it was written for: eat the space that
// separated a trailing comment from the code before it, and eat the newline that would otherwise be
// left as a blank line.
//
// The implementation does not test what those characters are. Measured against the installed build,
// which is the reason this function exists rather than a straight transcription:
//
//	x;// tslint:disable             upstream `x`            here `x;`
//	/* tslint:disable */let x = 1;  upstream `et x = 1;`    here `let x = 1;`
//	\t// tslint:disable-line        upstream ``             here `\t`
//	x;  // tslint:disable           upstream `x; `          here `x; `      (agrees)
//
// The fourth agrees because the character before the comment is a space in both readings. All eight
// of upstream's own corpus outputs agree for the same reason: every one of them widens onto
// whitespace or onto nothing.
//
// The right edge is also clamped to the end of the source. Upstream hands ESLint a range one past
// the last byte and ESLint clamps it silently; this project's fixture harness refuses an
// out-of-range fix as a rule defect, correctly, since a range past the end is a range the author did
// not mean. Clamping produces the identical output, because removing to the end and removing past
// it are the same removal.
func banTslintCommentFixRange(comment comments.Comment, sourceText string) core.TextRange {
	start := comment.Range.Pos()
	end := comment.Range.End()

	// Widen left onto a blank character, and only when the comment does not start its own line.
	//
	// Both halves of that condition are upstream's and both are load-bearing. The column test is
	// upstream's `column > 0` and it is NOT subsumed by the blank test, which is what a reader will
	// assume and what an earlier draft of this comment asserted. The character before a
	// column-zero comment is the newline that ENDED THE PREVIOUS LINE, which is blank, so a blank
	// test alone takes it. Measured on upstream's own eighth corpus case: with the column test the
	// fix writes `\nconst woah = doSomeStuff();\nconsole.log(woah);\n      `, and without it the
	// fix writes `\nconst woah = doSomeStuff();console.log(woah);\n      `, joining two statements
	// onto one line. Scored as a mutant, removing the column test fails that case.
	//
	// So the two conditions answer different questions. The column test asks whether there is
	// anything on this line before the comment; the blank test asks whether that thing is code.
	if start > 0 && comment.StartColumn > 0 && isBanTslintCommentBlank(sourceText[start-1]) {
		start--
	}

	// Widen right onto a blank character only. Upstream takes whatever is there, including the
	// first character of the statement a block comment sits in front of.
	if end < len(sourceText) && isBanTslintCommentBlank(sourceText[end]) {
		end++
	}

	return core.NewTextRange(start, end)
}

// isBanTslintCommentBlank answers whether a byte is one the fixer may absorb.
//
// Bytes rather than runes, and deliberately: the two characters this widens onto are the space that
// separated a trailing comment from its code and the newline that would be left behind as a blank
// line. Both are ASCII, and every byte of a multi-byte character has its high bit set, so no
// continuation byte can be mistaken for one of these five and no character can be split.
func isBanTslintCommentBlank(character byte) bool {
	switch character {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	}
	return false
}
