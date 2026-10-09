package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNonoctalDecimalEscape = rule.Message{
	Id: "nonoctalDecimalEscape",
	Description: "This is a `\\8` or `\\9` escape, which the language only still accepts as a legacy " +
		"web-compatibility feature. It means the bare digit, so the backslash does nothing, and in " +
		"strict mode or a module it is a syntax error rather than a quiet no-op. Whichever was " +
		"intended, the digit or a literal backslash, saying it directly is the version that keeps " +
		"meaning the same thing everywhere.",
}

// NoNonoctalDecimalEscape flags a `\8` or `\9` escape sequence in a string literal.
//
//	valid:   '8'
//	valid:   '\\8'        // an escaped backslash, then the digit
//	valid:   '\1'         // a legacy octal escape, which no-octal covers
//	valid:   /\8/         // a regex literal, deliberately not this rule's surface
//	invalid: '\8'
//	invalid: '\\\8'       // an escaped backslash, then a decimal escape
//	invalid: '\0\8'
//
// Despite living beside the regex rules, this one has nothing to do with regular expressions.
// Upstream's corpus is explicit that `/\8/` passes: a regex literal is a different grammar where
// `\8` is a backreference question rather than a legacy escape. The surface here is string literals
// only.
//
// The whole difficulty is backslash parity, and it cannot be answered from the parsed value. A
// string literal's cooked text has already resolved its escapes, so `'\8'` arrives as `8` with the
// backslash gone, and `'\\8'` arrives as `\8`. Both readings destroy the distinction the rule
// exists to make, which is why this walks the raw source between the quotes instead.
//
// Parity comes out right by consuming the text in escape-sized steps rather than by counting
// backslashes: a backslash always takes the character after it, so a scan that skips both can never
// mistake the second backslash of `\\` for the start of an escape. `'\\8'` is an escaped backslash
// followed by a plain digit and is silent; `'\\\8'` is an escaped backslash followed by a real
// decimal escape and reports.
//
// Every finding carries suggestions rather than fixes, because the two repairs mean different
// things and only the author knows which was meant. `'\8'` may have been intended as the digit or
// as a literal backslash before a digit, and an engine that picked one unattended would silently
// change the string half the time.
var NoNonoctalDecimalEscape = rule.Rule{
	Name: "no-nonoctal-decimal-escape",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		report := func(node *ast.Node) {
			literalRange := rule.TokenRange(ctx.SourceFile, node)
			raw := ctx.SourceFile.Text()[literalRange.Pos():literalRange.End()]

			// The shortest literal that could hold one is `'\8'`: two quotes, a backslash, a digit.
			if len(raw) < 4 {
				return
			}
			if !containsNonoctalDecimalEscape(raw) {
				return
			}

			for _, found := range findNonoctalDecimalEscapes(raw) {
				escapeStart := literalRange.Pos() + found.start
				escapeEnd := literalRange.Pos() + found.end
				escapeText := raw[found.start:found.end]
				digit := raw[found.start+1 : found.end]

				suggestions := make([]rule.Suggestion, 0, 3)

				if found.precededByNullEscape {
					// `\0\8` cannot be repaired by dropping the backslash, because `\08` is a
					// legacy octal escape: the fix for one legacy escape would produce another.
					// Writing the null out in full is what keeps the string meaning what it did.
					combinedStart := literalRange.Pos() + found.start - len(`\0`)
					suggestions = append(suggestions, rule.Suggestion{
						Message: rule.Message{
							Id: messageNonoctalDecimalEscape.Id,
							Description: fmt.Sprintf(
								"Replace `\\0%s` with `\\u0000%s`, which matches the same characters "+
									"without turning the repair into a legacy octal escape.",
								escapeText, digit),
						},
						Fixes: []rule.Fix{{
							Range: core.NewTextRange(combinedStart, escapeEnd),
							Text:  `\u0000` + digit,
						}},
					})
					suggestions = append(suggestions, rule.Suggestion{
						Message: rule.Message{
							Id: messageNonoctalDecimalEscape.Id,
							Description: fmt.Sprintf(
								"Replace `%s` with `\\u003%s`, keeping the digit and leaving the "+
									"null escape before it alone.", escapeText, digit),
						},
						Fixes: []rule.Fix{{
							Range: core.NewTextRange(escapeStart, escapeEnd),
							Text:  `\u003` + digit,
						}},
					})
				} else {
					suggestions = append(suggestions, rule.Suggestion{
						Message: rule.Message{
							Id: messageNonoctalDecimalEscape.Id,
							Description: fmt.Sprintf(
								"Replace `%s` with `%s` if the digit is what was meant. The escape "+
									"already means exactly that, so this changes nothing at runtime.",
								escapeText, digit),
						},
						Fixes: []rule.Fix{{
							Range: core.NewTextRange(escapeStart, escapeEnd),
							Text:  digit,
						}},
					})
				}

				suggestions = append(suggestions, rule.Suggestion{
					Message: rule.Message{
						Id: messageNonoctalDecimalEscape.Id,
						Description: fmt.Sprintf(
							"Replace `%s` with `\\%s` if a literal backslash before the digit is "+
								"what was meant. This one does change the string.", escapeText, escapeText),
					},
					Fixes: []rule.Fix{{
						Range: core.NewTextRange(escapeStart, escapeEnd),
						Text:  `\` + escapeText,
					}},
				})

				ctx.ReportRangeWithSuggestions(
					core.NewTextRange(escapeStart, escapeEnd),
					messageNonoctalDecimalEscape,
					suggestions...,
				)
			}
		}

		return rule.Listeners{
			ast.KindStringLiteral: report,
		}
	},
}

// nonoctalDecimalEscape is one `\8` or `\9` found in a literal's raw text, as offsets into that
// text.
//
// precededByNullEscape records whether a `\0` sits immediately before it, which is the one case
// where the obvious repair would produce a different legacy escape rather than removing one.
type nonoctalDecimalEscape struct {
	start                int
	end                  int
	precededByNullEscape bool
}

// containsNonoctalDecimalEscape is the cheap gate: no backslash followed by 8 or 9 anywhere in the
// text means there is nothing here, and that is almost every string in a real tree.
//
// It deliberately over-approximates. `'\\8'` passes this test and is then found clean by the real
// scan, because deciding parity is the expensive part and doing it twice would be the same work.
func containsNonoctalDecimalEscape(raw string) bool {
	for index := 0; index+1 < len(raw); index++ {
		if raw[index] == '\\' && (raw[index+1] == '8' || raw[index+1] == '9') {
			return true
		}
	}
	return false
}

// findNonoctalDecimalEscapes returns every `\8` and `\9` in a literal's raw text that is a real
// escape rather than a digit following an escaped backslash.
//
// The scan steps over the text in escape-sized units: a backslash consumes the character after it,
// so the second backslash of `\\` is consumed as data and can never be read as opening an escape.
// That is what separates `'\\8'`, which is silent, from `'\\\8'`, which reports, without counting
// backslashes or looking behind.
func findNonoctalDecimalEscapes(raw string) []nonoctalDecimalEscape {
	var found []nonoctalDecimalEscape
	previousEscapeWasNull := false

	for index := 0; index < len(raw); {
		if raw[index] != '\\' || index+1 >= len(raw) {
			previousEscapeWasNull = false
			index++
			continue
		}

		next := raw[index+1]
		if next == '8' || next == '9' {
			found = append(found, nonoctalDecimalEscape{
				start:                index,
				end:                  index + 2,
				precededByNullEscape: previousEscapeWasNull,
			})
			previousEscapeWasNull = false
			index += 2
			continue
		}

		// A `\0` sets the null-escape flag, and nothing else does.
		//
		// `\01` needs no special case here even though it is a legacy octal escape rather than a
		// null: the walk consumes `\0` and lands on the `1`, which is not a backslash, so the flag
		// is cleared on the next step before any `\8` could see it. A mutation sweep is what
		// established that: an explicit trailing-digit guard here could be deleted with every
		// fixture still green, because the step already decides it.
		previousEscapeWasNull = next == '0'

		// The escape consumes both bytes whatever it turned out to be, which is the step that makes
		// parity fall out of the walk rather than out of a count.
		index += 2
	}

	return found
}
