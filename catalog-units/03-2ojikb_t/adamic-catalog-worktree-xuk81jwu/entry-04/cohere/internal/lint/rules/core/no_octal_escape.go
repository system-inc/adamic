package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageOctalEscape = rule.Message{
	Id: "octalEscapeSequence",
	Description: "This is a legacy octal escape, a numbering the language kept only for old web " +
		"pages and forbids outright in strict mode and in modules, where it is a syntax error " +
		"rather than a character. It also reads as decimal to almost everybody: the escape here " +
		"is not the number it looks like. A unicode escape says the same character in a " +
		"numbering that is unambiguous and legal everywhere.",
}

// NoOctalEscape flags a legacy octal escape sequence in a string literal.
//
//	valid:   '\x51'
//	valid:   '\0'          // the null character, which is legal when no digit follows
//	valid:   '\8'          // a decimal escape, which no-nonoctal-decimal-escape covers
//	valid:   '\\251'       // an escaped backslash, then three plain digits
//	valid:   '01'          // digits with no escape at all
//	invalid: '\01'
//	invalid: '\251'
//	invalid: '\1'
//	invalid: '\08'         // reports the `\0`, because a digit follows it
//
// # The subject is the raw text, and it has to be
//
// A string literal's cooked value has already resolved its escapes, so `'\251'` arrives as the
// single character it denotes and `'\\251'` arrives as a backslash and three digits. The first is
// the violation and the second is clean, and both readings of the cooked value destroy the
// distinction. So this walks the raw source between the quotes, the same way
// no-nonoctal-decimal-escape does and for the same reason.
//
// # What counts as an octal escape, which is narrower than "a backslash and a digit"
//
// Upstream matches `\([0-3][0-7]{1,2}|[4-7][0-7]|0(?=[89])|[1-7])` and the four alternatives are
// tried in that order, which is the whole of the arithmetic:
//
//   - `[0-3][0-7]{1,2}` is greedy and takes up to three digits, because a value starting 0 through
//     3 fits in a byte with three digits. `\377` reports `377`; `\378` reports `37`, since `8` is
//     not an octal digit; `\37a` likewise reports `37`.
//   - `[4-7][0-7]` takes exactly two, because a value starting 4 through 7 would overflow a byte
//     with three. `\400` reports `40` and leaves the final `0` as a plain digit. That is upstream's
//     own corpus, not an inference.
//   - `0(?=[89])` is the one that surprises. A `\0` followed by `8` or `9` reports, capturing just
//     the `0`, because `\08` is the legacy production rather than a null character followed by a
//     digit. `\08` and `\09` both report; a porter told otherwise would ship a false negative on
//     the same banned form.
//   - `[1-7]` is the single-digit fallback. `\1` through `\7` report on their own.
//
// A bare `\0` with no digit after it is the null character and is legal, which is why it is absent
// from the list and present in upstream's clean cases.
//
// # Backslash parity, and why the walk is in escape-sized steps
//
// The prefix upstream matches is `(?:[^\\]|\\.)*?`: either a byte that is not a backslash, or a
// backslash and whatever follows it. That is a walk in escape-sized units rather than a search for
// a backslash, and it is what separates `'\1'`, which reports, from `'\\1'`, which does not, since
// the second backslash of `\\` is consumed as data and can never open an escape. Counting
// backslashes and testing parity gets the same answer on these inputs and is harder to be sure of;
// stepping is what upstream does.
//
// # At most one finding per literal, and it is the first
//
// The prefix is lazy, so the match is the earliest escape rather than the longest run, and `test`
// answers once. `'\01\02'` reports `01` and says nothing about `02`; `'\02\01'` reports `02`. Both
// are in the corpus, which is how the direction is pinned rather than assumed. A port reporting
// every escape would double the findings on those two inputs and stay green on every other case.
//
// # The finding names the sequence, so the message is interpolated
//
// Upstream's message embeds `{{sequence}}`, and the captured text is not always the digits a reader
// would guess: `\378` names `37` and `\08` names `0`. That makes the rendered text a real output of
// this rule rather than a constant, and it is asserted as one in the tests rather than through a
// message id, which cannot see anything the format string does.
//
// The rule reports without a repair, matching upstream, which ships no fixer. The obvious repair is
// mechanical (write `A` for `\101`) but it is only correct when the escape is the whole story:
// in `\400` the trailing `0` is a plain digit that the rewritten escape must keep, and in `\08` the
// `8` is likewise data. A fixer getting that wrong changes the string unattended.
var NoOctalEscape = rule.Rule{
	Name: "no-octal-escape",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				literalRange := rule.TokenRange(ctx.SourceFile, node)
				raw := ctx.SourceFile.Text()[literalRange.Pos():literalRange.End()]

				sequence, found := firstOctalEscapeSequence(raw)
				if !found {
					return
				}

				ctx.ReportNode(node, rule.Message{
					Id: messageOctalEscape.Id,
					Description: fmt.Sprintf("Do not use the octal escape `\\%s`. %s",
						sequence, messageOctalEscape.Description),
				})
			},
		}
	},
}

// firstOctalEscapeSequence returns the digits of the earliest legacy octal escape in a literal's
// raw text, or reports that there is none.
//
// The walk consumes the text in escape-sized units: a backslash takes the byte after it, anything
// else takes one byte. That is upstream's `(?:[^\\]|\\.)*?` prefix, and it is what makes parity
// fall out of the step rather than out of a count, so the second backslash of `\\` is never read as
// opening an escape.
//
// At each backslash the four productions are tried in upstream's order, and the first that matches
// wins. Order is load bearing rather than a tidy way to write four independent tests: `0` is in
// range for the first production, so `\08` would be taken as a one-digit `[0-3][0-7]{1,2}` match if
// that production could accept a single digit. It cannot, since `{1,2}` demands at least one digit
// AFTER the leading one, which is why `\0` alone is clean and why the `0(?=[89])` production has to
// exist separately.
func firstOctalEscapeSequence(raw string) (string, bool) {
	for index := 0; index < len(raw); {
		if raw[index] != '\\' {
			index++
			continue
		}
		// A trailing backslash has nothing to escape. The literal's closing quote means this cannot
		// happen in well-formed source, but the parser recovers from source that is not.
		if index+1 >= len(raw) {
			return "", false
		}

		if digits, matched := octalEscapeAt(raw, index+1); matched {
			return digits, true
		}

		// The escape consumes both bytes whatever it turned out to be. This is the step that makes
		// `'\\1'` clean without looking behind.
		index += 2
	}
	return "", false
}

// octalEscapeAt tries upstream's four productions at the byte after a backslash and returns the
// digits the winning one captures.
//
// Returning the captured text rather than a length is deliberate: the capture is what the message
// renders, and for `0(?=[89])` the capture is shorter than what the escape consumes, since the
// lookahead matches the `8` without taking it. Anything measuring a width would have to special
// case that one production.
func octalEscapeAt(raw string, at int) (string, bool) {
	if at >= len(raw) {
		return "", false
	}
	first := raw[at]

	// `[0-3][0-7]{1,2}` is greedy, up to three digits total, because a value starting 0 through 3
	// still fits in a byte with three of them.
	if first >= '0' && first <= '3' {
		width := 1
		for width < 3 && at+width < len(raw) && isOctalDigit(raw[at+width]) {
			width++
		}
		if width >= 2 {
			return raw[at : at+width], true
		}
	}

	// `[4-7][0-7]` takes exactly two, because a third digit would overflow a byte. `\400` captures
	// `40` and leaves the final `0` as data, which is upstream's own corpus.
	if first >= '4' && first <= '7' {
		if at+1 < len(raw) && isOctalDigit(raw[at+1]) {
			return raw[at : at+2], true
		}
	}

	// `0(?=[89])` is the surprising one: a null escape followed by a decimal digit is the legacy production, and the
	// capture is the `0` alone because the lookahead does not consume. This is the case a reader
	// expects to be clean and upstream reports; measured against the installed build, `'\08'` and
	// `'\09'` both report with sequence `0`.
	if first == '0' && at+1 < len(raw) && (raw[at+1] == '8' || raw[at+1] == '9') {
		return raw[at : at+1], true
	}

	// `[1-7]` is the single-digit fallback. `0` is absent, which is what makes a bare `\0` the legal
	// null character rather than an octal escape.
	if first >= '1' && first <= '7' {
		return raw[at : at+1], true
	}

	return "", false
}

// isOctalDigit reports whether a byte is one of the eight digits an octal escape may continue with.
//
// `8` and `9` are excluded, which is why `\378` captures `37` rather than `378` and why `\08`
// reaches the lookahead production instead of the greedy one.
func isOctalDigit(character byte) bool {
	return character >= '0' && character <= '7'
}
