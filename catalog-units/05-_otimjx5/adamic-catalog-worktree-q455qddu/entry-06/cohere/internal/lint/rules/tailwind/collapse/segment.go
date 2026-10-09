// segment, hasMathFn, and the JavaScript string primitives the data-type checks depend on.
//
// These are ports of `src/utils/segment.ts` and `src/utils/math-operators.ts` (the `hasMathFn` half
// only; the whitespace formatter in the same upstream file belongs to the printer, not to type
// inference, and is not needed here).
//
// # Why not the regexsyntax bracket scanner
//
// `internal/utilities/ecmascript/regexsyntax` has ClassEnd and SkipPatternEscape, which also walk
// brackets, and they were read before this was written. They are the wrong primitives. They scan
// JavaScript *regular-expression literal syntax*: they take RegexFlags, treat `[` as nesting only
// under the v flag, understand `\q{...}`, and decode UTF-8 runes because a regex class can contain
// them. CSS value segmentation has none of that and does have things they lack: `(` and `{` nest
// equally with `[`, quotes swallow everything to their close, a backslash skips exactly one byte
// with no escape grammar, and an unbalanced closer is ignored rather than fatal. Reusing them would
// mean passing flag values that mean nothing here and inheriting a bracket model that is close
// enough to look right and different enough to be wrong on quoted font names.
package tailwind

import (
	"strconv"
	"strings"
)

// segment splits input on a top-level occurrence of separator.
//
// Ported from `segment.ts`, whose comment explains the shape better than a restatement would: regex
// cannot recurse, so a tiny state machine tracks nesting instead. `var(--a, 0 0 1px rgb(0, 0, 0)),
// 0 0 1px rgb(0, 0, 0)` splits on exactly one comma, the top-level one.
//
// Three behaviors are load-bearing for the callers in this package:
//
//   - A final part is always pushed, so the result is never empty and `segment("", ',')` is one
//     empty string rather than nothing. isFamilyName and isLineWidth both depend on that: they
//     count parts, and an empty result would make their `count > 0` and `every` tests answer the
//     opposite of what the engine answers.
//   - A closing bracket that does not match the top of the stack is ignored rather than treated as
//     an error, so `a)b,c` still splits on its comma.
//   - The separator is compared as a single byte, matching `separator.charCodeAt(0)`, so a
//     multi-character separator would upstream match only its first character. Callers here pass
//     ' ' and ',' so the byte type makes that explicit rather than accepting a string it would
//     silently truncate.
func segment(input string, separator byte) []string {
	// The upstream shared 256-byte buffer is a single-threaded allocation optimization that would
	// be a data race here. A per-call slice costs an allocation on nested input and nothing on the
	// flat input that dominates.
	var closingBracketStack []byte
	parts := make([]string, 0, 4)
	lastPosition := 0

	for index := 0; index < len(input); index++ {
		character := input[index]

		if len(closingBracketStack) == 0 && character == separator {
			parts = append(parts, input[lastPosition:index])
			lastPosition = index + 1
			continue
		}

		switch character {
		case '\\':
			// The next byte is escaped, so skip it.
			index++
		case '\'', '"':
			// Strings are taken as-is to their close. No bracket balancing inside them, which is
			// what keeps a font name like `"Comic Sans, Bold"` from splitting on its comma.
			for index+1 < len(input) {
				index++
				next := input[index]
				if next == '\\' {
					index++
					continue
				}
				if next == character {
					break
				}
			}
		case '(':
			closingBracketStack = append(closingBracketStack, ')')
		case '[':
			closingBracketStack = append(closingBracketStack, ']')
		case '{':
			closingBracketStack = append(closingBracketStack, '}')
		case ']', '}', ')':
			if len(closingBracketStack) > 0 && character == closingBracketStack[len(closingBracketStack)-1] {
				closingBracketStack = closingBracketStack[:len(closingBracketStack)-1]
			}
		}
	}

	return append(parts, input[lastPosition:])
}

// mathFunctions is the list `hasMathFn` scans for, in upstream order.
var mathFunctions = []string{
	"calc", "min", "max", "clamp", "mod", "rem", "sin", "cos", "tan",
	"asin", "acos", "atan", "atan2", "pow", "sqrt", "hypot", "log", "exp", "round",
}

// hasMathFunction ports `hasMathFn`: does input contain `(` and any `<mathFn>(` anywhere.
//
// Anywhere, not at the start. `1px_calc(2px)` has a math function, and so does
// `url(rounded.png)` — `round` is in the list and `round(` appears inside. That last one is not a
// hypothetical: it is why `isNumber` is true for strings that are visibly not numbers, and why the
// caller's type order is what decides the answer rather than this predicate.
//
// `rem` in the list also means the substring `rem(` matches, so `[--foo:rem(1)]` and anything else
// spelling a math name followed by a paren counts.
func hasMathFunction(input string) bool {
	if !strings.ContainsRune(input, '(') {
		return false
	}
	for _, function := range mathFunctions {
		if strings.Contains(input, function+"(") {
			return true
		}
	}
	return false
}

// javaScriptSpaces is the set `\s` matches in a JavaScript regexp.
//
// Wider than ASCII whitespace: it is the Unicode White_Space property plus BOM plus the line
// terminators. Only the members that can plausibly appear in a CSS value are listed, and the
// Unicode space separators are handled by range below.
func isJavaScriptSpace(runeValue rune) bool {
	switch runeValue {
	case ' ', '\t', '\n', '\v', '\f', '\r',
		'\u00a0', // no-break space
		'\u1680', // ogham space mark
		'\u2028', // line separator
		'\u2029', // paragraph separator
		'\u202f', // narrow no-break space
		'\u205f', // medium mathematical space
		'\u3000', // ideographic space
		'\ufeff': // zero-width no-break space (BOM)
		return true
	}
	// U+2000..U+200A, the general punctuation space separators.
	return runeValue >= '\u2000' && runeValue <= '\u200a'
}

// trimLeadingJavaScriptSpace drops the leading run of `\s`-matching characters.
func trimLeadingJavaScriptSpace(value string) string {
	return strings.TrimLeftFunc(value, isJavaScriptSpace)
}

// roundTripsAsJavaScriptNumber reports whether `String(Number(value)) === value` for a value already
// known to be a run of digits with no leading zero.
//
// Only reached for 16-or-more-digit values, where float64 stops being exact and JavaScript switches
// to exponential notation at 1e21. Go's 'g'-with-shortest formatting picks the same digits
// ECMAScript's Number::toString does for the cases that matter here, and below 1e21 both print
// plain decimal, so comparing the reprint is the same test upstream runs.
func roundTripsAsJavaScriptNumber(value string) bool {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}
	// ECMAScript prints integers below 1e21 without an exponent; above it, upstream's comparison
	// against the original digit string always fails.
	if parsed >= 1e21 {
		return false
	}
	return strconv.FormatFloat(parsed, 'f', -1, 64) == value
}
