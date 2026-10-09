package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageEmptyCharacterClass = rule.Message{
	Id: "unexpectedEmptyCharacterClass",
	Description: "This character class is empty, so it matches nothing and makes the whole pattern " +
		"unmatchable. Nobody writes an unmatchable pattern on purpose, which is why this is almost " +
		"always a class whose contents were deleted or never typed. It is worth reporting because " +
		"the failure is silent: the regex compiles, the code runs, and the branch behind the match " +
		"simply never fires.",
}

// NoEmptyCharacterClass flags an empty character class in a regular expression literal.
//
//	valid:   /^abc[a-z]/
//	valid:   /^abc/
//	valid:   /[^]/            // negated, matches everything, deliberate
//	valid:   /\[]/            // the bracket is escaped, so there is no class here
//	valid:   new RegExp('[]') // not a literal, see below
//	invalid: /^abc[]/
//	invalid: /[]]/            // the class is `[]`, the trailing bracket is a literal
//	invalid: /[[]]/v          // the nested class is empty
//
// Two decisions here are worth stating because both look like gaps and neither is.
//
// **A negated empty class is not reported.** `[^]` matches any character including a newline, which
// is the shortest way to write that and the reason people reach for it. `[]` and `[^]` differ by one
// character and by everything: one matches nothing, the other matches everything. Only the first is
// a mistake.
//
// **Only literals are checked, never the RegExp constructor.** `new RegExp('[]')` is left alone,
// matching upstream. That is not symmetry with no-invalid-regexp, which checks constructors and not
// literals, and the asymmetry is on purpose: a literal with a syntax error never parses, so
// no-invalid-regexp has nothing to say about one, while an empty class parses fine and only a
// reader can call it wrong. The two rules cover the two halves rather than overlapping.
//
// Under the `v` flag classes nest, and every level is checked. `[[]]` is a non-empty outer class
// containing an empty inner one, so the finding is on the inner.
//
// A pattern the scanner cannot make sense of is skipped rather than reported. An unterminated class
// is a syntax error that the parser has already refused, so a second complaint here would be noise
// on a file that does not compile.
var NoEmptyCharacterClass = rule.Rule{
	Name: "no-empty-character-class",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				text := node.Text()

				// A regex literal's own text is `/pattern/flags`, delimiters and all, so the two
				// have to be separated before either can be used. That split is shared rather than
				// written here: it lands the same way for every regex rule, and a second copy of it
				// is a second place for `/a\/b/g` to come apart wrongly.
				pattern, flags := regexsyntax.PatternAndFlags(text)
				if pattern == "" {
					return
				}

				// The cheap gate upstream uses too: no bracket, no class, and most patterns in a
				// real tree have no bracket at all.
				if !strings.ContainsAny(pattern, "[]") {
					return
				}

				// The node's own text ends at End(), but Pos() sits before leading trivia, so the
				// start has to be derived from the end rather than read directly. Getting this
				// backwards reports the finding at whatever comment precedes the literal, on a line
				// no `-next-line` directive the author can write is able to suppress.
				literalStart := node.End() - len(text)
				patternStart := literalStart + 1

				regexFlags := regexsyntax.ParseRegexFlags(flags)
				regexsyntax.IterateRegexCharacterClasses(pattern, regexFlags, func(start, end int) {
					if !classIsEmpty(start, end) {
						return
					}
					ctx.ReportRange(
						core.NewTextRange(patternStart+start, patternStart+end),
						messageEmptyCharacterClass,
					)
				})
			},
		}
	},
}

// classIsEmpty reports whether pattern[start:end], which spans `[`..`]`, holds no elements and is
// not negated.
//
// Emptiness is decided from the span's width rather than by parsing its body, because the only
// empty class there is spells itself `[]` and is therefore exactly two bytes wide. Everything else
// is wider, including `[^]`, which is why negation needs no separate test: a negated class carries
// the `^` inside the span and can never measure two.
//
// That single comparison is the whole decision, and it is worth saying so rather than adding a
// bracket-identity check that reads as defensive. A mutation sweep is what forced the point: the
// identity check could be replaced with `return true` and every one of thirty-four fixtures stayed
// green, because the width test had already excluded everything it would have caught. Code no
// fixture can kill is code that is not deciding anything, and leaving it in implies a second guard
// exists where there is one.
func classIsEmpty(start int, end int) bool {
	return end-start == 2
}
