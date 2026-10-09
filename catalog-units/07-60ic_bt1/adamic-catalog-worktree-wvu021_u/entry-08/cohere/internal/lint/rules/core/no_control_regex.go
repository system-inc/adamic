package core

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexpattern"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoControlRegex flags a control character spelled out inside a regular expression.
//
//	valid:   /x1f/
//	valid:   /\t/
//	valid:   /\0/
//	invalid: /\x1f/
//	invalid: /\u{1F}/u
//
// A control character in a pattern is almost always a typo for the text that spells it: somebody
// meant to match the two characters and wrote an escape instead. Matching a control code on purpose
// is rare enough that saying so explicitly is cheap.
//
// # The rule is about spelling, not about the code point
//
// A raw control byte and its escape are the same character, and only the escape reports. `\t` is
// U+0009, a control character by value, and is silent. What is refused is writing a control code out
// by its number, so the predicate is how the character was spelled rather than what it matches.
//
// That is why this uses `regexpattern.Walk`, whose `CharacterKind` records the spelling. The shelf
// doc names this rule as the reason that field exists: a walk reporting only values cannot express
// the distinction and would flag every `\n` in the tree.
//
// A raw control byte pasted into a pattern is silent too, matching upstream. That looks like an
// omission and is theirs; reproducing it keeps the differential clean, and arguing to change it is a
// separate question from porting it.
//
// # One finding per pattern, not per character
//
// A pattern holding three control characters reports once and names all three. Upstream batches the
// same way, which is why its 57 failing inputs produce 34 diagnostics. A port reporting per
// character disagrees with the snapshot on every multi-character case while passing any fixture
// written from the input list alone.
var NoControlRegex = rule.Rule{
	Name: "no-control-regex",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				text := node.Text()

				pattern, flags := regexsyntax.PatternAndFlags(text)
				if pattern == "" {
					return
				}

				var spelled []string
				regexpattern.Walk(pattern, regexsyntax.ParseRegexFlags(flags),
					func(character regexpattern.Character) bool {
						if isSpelledControlCharacter(pattern, character) {
							spelled = append(spelled, pattern[character.Start:character.End])
						}
						return true
					})

				if len(spelled) == 0 {
					return
				}

				// Pos() sits before leading trivia, so the literal's start is derived from its end.
				// Reading Pos() directly puts the finding on whatever comment precedes the literal,
				// which is a line a suppressing directive cannot reach.
				literalStart := node.End() - len(text)

				// The whole literal rather than one character, because the finding is about the
				// pattern and the characters it names may be far apart inside it.
				ctx.ReportRange(core.NewTextRange(literalStart, node.End()), rule.Message{
					Id: "noControlRegex",
					Description: fmt.Sprintf(
						"This regular expression spells out the control character %s. A control code "+
							"written by number is almost always a typo for the text that spells it, "+
							"and matching one on purpose is rare enough to be worth saying "+
							"explicitly.", strings.Join(spelled, ", ")),
				})
			},
		}
	},
}

// isSpelledControlCharacter reports whether a character is a control code written out by number.
//
// Three conditions, each excluding a real case rather than being defensive. The value has to be in
// the control range. The spelling has to be a hexadecimal or unicode escape, so a named escape like
// `\t` is silent even though it is a control character by value. And the null character written as
// `\0` is excluded by kind, since that is the character written as itself rather than as a number.
func isSpelledControlCharacter(pattern string, character regexpattern.Character) bool {
	if character.Value > 0x1F {
		return false
	}

	// The spelling decides, and it subsumes the null case rather than needing an arm for it: `\0`
	// trims to `0`, which is neither prefix, so the character written as itself is excluded by the
	// same test that excludes `\t`. A `KindNull` check here compiled and changed no fixture, which
	// is how it was found to be unreachable rather than by reading.
	trimmed := strings.TrimLeft(pattern[character.Start:character.End], "\\")
	return strings.HasPrefix(trimmed, "x") || strings.HasPrefix(trimmed, "u")
}
