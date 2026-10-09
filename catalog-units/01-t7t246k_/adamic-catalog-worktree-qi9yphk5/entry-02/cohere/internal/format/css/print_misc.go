package css

// src/language-css/print/misc.js, then the generic src/utilities it imports (print-string.js and its
// get-preferred-quote.js and make-string.js, print-number.js) and the one is-next-line-empty.js the
// print/ helpers import, then the JavaScript string semantics the printer leans on (\s, trim,
// toLowerCase) spelled out once.
//
// Upstream's adjustStrings and adjustNumbers run regular expressions with a backreference and lookahead,
// which Go's regexp does not have. Each is a scanner here that takes the match the backtracking engine
// would take, explained where it is written. Offsets are bytes; every character these patterns test is
// ASCII, and every non-ASCII code unit falls in the same class (U+0080 to U+FFFF) as each of its UTF-8
// bytes, so stepping by byte finds the same matches upstream finds stepping by UTF-16 code unit.

import (
	"regexp"
	"strings"

	"github.com/system-inc/cohere/internal/format/printing"
)

func printUnit(unit string) string {
	lowercased := toLowerCase(unit)
	if canonical, present := cssUnits[lowercased]; present {
		return canonical
	}
	return unit
}

// matchStringAt is STRING_REGEX, /(["'])(?:(?!\1)[^\\]|\\.)*\1/gs, anchored at start: the end of the
// quoted string that opens there, or false. The repetition stops at the first unescaped quote of the
// opening kind, and no backtracking can produce an earlier close (giving back an escape leaves the
// position on a backslash, which is not the quote), so the first unescaped quote closes it. A string
// with no close, or ending in a lone backslash, does not match.
func matchStringAt(value string, start int) (int, bool) {
	quote := value[start]
	if quote != '"' && quote != '\'' {
		return 0, false
	}
	for index := start + 1; index < len(value); index++ {
		switch value[index] {
		case quote:
			return index + 1, true
		case '\\':
			if index+1 >= len(value) {
				return 0, false
			}
			index++
		}
	}
	return 0, false
}

// adjustStrings is value.replaceAll(STRING_REGEX, (match) => printString(match, options)).
func adjustStrings(value string, options *printerOptions) string {
	var result strings.Builder
	for index := 0; index < len(value); {
		if end, matched := matchStringAt(value, index); matched {
			result.WriteString(printString(value[index:end], options))
			index = end
			continue
		}
		result.WriteByte(value[index])
		index++
	}
	return result.String()
}

// attributeFlagPattern is /^(?<value>.+?)\s+(?<flag>[a-z])$/i with JavaScript's `.` and \s.
var attributeFlagPattern = regexp.MustCompile(`(?i)^([^\n\r\x{2028}\x{2029}]+?)` + javaScriptWhitespaceClass + `+([a-z])$`)

func quoteAttributeValue(value string, options *printerOptions) string {
	quote := `"`
	if settingsOf(options).SingleQuote {
		quote = "'"
	}

	// The selector parser currently only understand `i` flag,
	// but not `s`, `S`, and `I`
	// To support future flags, we simply check if it's an alphabet letter
	// https://github.com/prettier/prettier/pull/17865#discussion_r2332698101
	flag := ""
	if match := attributeFlagPattern.FindStringSubmatch(value); match != nil {
		value, flag = match[1], match[2]
	}

	quoted := value
	if !strings.Contains(value, `"`) && !strings.Contains(value, "'") {
		quoted = quote + value + quote
	}
	if flag != "" {
		quoted += " " + flag
	}
	return quoted
}

func isDigit(character byte) bool { return character >= '0' && character <= '9' }

func isLetter(character byte) bool {
	return (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')
}

func skipDigits(value string, index int) int {
	for index < len(value) && isDigit(value[index]) {
		index++
	}
	return index
}

// matchNumberAt is NUMBER_REGEX, /(?:\d*\.\d+|\d+\.?)(?:e[+-]?\d+)?/i, anchored at start. The first
// alternative is tried first and needs a digit after the dot; backtracking its \d* cannot help, because a
// shorter run leaves a digit where the dot must be. The exponent is optional and taken when it matches.
func matchNumberAt(value string, start int) (int, bool) {
	end := -1
	integerEnd := skipDigits(value, start)
	if integerEnd < len(value) && value[integerEnd] == '.' && integerEnd+1 < len(value) && isDigit(value[integerEnd+1]) {
		end = skipDigits(value, integerEnd+1)
	} else if integerEnd > start {
		end = integerEnd
		if end < len(value) && value[end] == '.' {
			end++
		}
	} else {
		return 0, false
	}

	if end < len(value) && (value[end] == 'e' || value[end] == 'E') {
		exponent := end + 1
		if exponent < len(value) && (value[exponent] == '+' || value[exponent] == '-') {
			exponent++
		}
		if exponent < len(value) && isDigit(value[exponent]) {
			end = skipDigits(value, exponent)
		}
	}
	return end, true
}

// isWordPartStart and isWordPartCharacter are WORD_PART_REGEX's [_a-z\u0080-\uFFFF]
// and [\w\u0080-\uFFFF-] under the i flag.
func isWordPartStart(character byte) bool {
	return character == '_' || isLetter(character) || character >= 0x80
}

func isWordPartCharacter(character byte) bool {
	return isWordPartStart(character) || isDigit(character) || character == '-'
}

// adjustNumbersMatch is one match of ADJUST_NUMBERS_REGEX that is not a string: the optional word part,
// the number, and the optional unit, as byte ranges.
type adjustNumbersMatch struct {
	end         int
	hasWordPart bool
	numberStart int
	numberEnd   int
	unitStart   int
	unitEnd     int
	matched     bool
}

// matchNumberWithUnitAt is the second alternative of ADJUST_NUMBERS_REGEX, `(WORD_PART)?(NUMBER)(UNIT)?`
// with the gi flags, anchored at start. The optional word part is greedy, so it is tried first: its
// [$@]? prefix, a start character, then the longest run of word characters, giving characters back one
// at a time until a number follows (a word run swallows digits, as in `a1`, so the number is often what
// it gives back). Only when no length works is the number tried with no word part. The unit is optional
// and greedy, and nothing after it can fail.
func matchNumberWithUnitAt(value string, start int) adjustNumbersMatch {
	finish := func(hasWordPart bool, numberStart int, numberEnd int) adjustNumbersMatch {
		unitEnd := numberEnd
		for unitEnd < len(value) && isLetter(value[unitEnd]) {
			unitEnd++
		}
		return adjustNumbersMatch{
			end:         unitEnd,
			hasWordPart: hasWordPart,
			numberStart: numberStart,
			numberEnd:   numberEnd,
			unitStart:   numberEnd,
			unitEnd:     unitEnd,
			matched:     true,
		}
	}

	wordStart := start
	if value[wordStart] == '$' || value[wordStart] == '@' {
		wordStart++
	}
	if wordStart < len(value) && isWordPartStart(value[wordStart]) {
		runEnd := wordStart + 1
		for runEnd < len(value) && isWordPartCharacter(value[runEnd]) {
			runEnd++
		}
		for numberStart := runEnd; numberStart > wordStart; numberStart-- {
			if numberEnd, matched := matchNumberAt(value, numberStart); matched {
				return finish(true, numberStart, numberEnd)
			}
		}
	}

	if numberEnd, matched := matchNumberAt(value, start); matched {
		return finish(false, start, numberEnd)
	}
	return adjustNumbersMatch{}
}

func adjustNumbers(value string) string {
	var result strings.Builder
	for index := 0; index < len(value); {
		if end, matched := matchStringAt(value, index); matched {
			result.WriteString(value[index:end])
			index = end
			continue
		}
		match := matchNumberWithUnitAt(value, index)
		if !match.matched {
			result.WriteByte(value[index])
			index++
			continue
		}

		replacement := value[index:match.end]
		if !match.hasWordPart {
			number := value[match.numberStart:match.numberEnd]
			unit := toLowerCase(value[match.unitStart:match.unitEnd])

			_, isUnit := cssUnits[unit]
			if unit == "" ||
				// `2n + 1`
				unit == "n" ||
				isUnit {
				replacement = printCssNumber(number)
				if unit != "" {
					replacement += printUnit(unit)
				}
			}
		}
		result.WriteString(replacement)
		index = match.end
	}
	return result.String()
}

// printCssNumber removes a trailing `.0`: .replace(/\.0(?=$|e)/, ""), the first one only.
func printCssNumber(rawNumber string) string {
	printed := printNumber(rawNumber)
	for index := 0; index+1 < len(printed); index++ {
		if printed[index] == '.' && printed[index+1] == '0' && (index+2 == len(printed) || printed[index+2] == 'e') {
			return printed[:index] + printed[index+2:]
		}
	}
	return printed
}

func shouldPrintTrailingComma(options *printerOptions) bool {
	trailingComma := settingsOf(options).TrailingComma
	return trailingComma == "es5" || trailingComma == "all"
}

// src/utilities/get-preferred-quote.js

func getPreferredQuote(text string, preferSingleQuote bool) string {
	preferred, alternate := `"`, "'"
	if preferSingleQuote {
		preferred, alternate = "'", `"`
	}
	preferredQuoteCount := strings.Count(text, preferred)
	alternateQuoteCount := strings.Count(text, alternate)
	if preferredQuoteCount > alternateQuoteCount {
		return alternate
	}
	return preferred
}

// src/utilities/make-string.js

// makeStringPattern matches _any_ escape and unescaped quotes (both single and double).
var makeStringPattern = regexp.MustCompile(`\\(["'\\])|(["'])`)

func makeString(rawText string, enclosingQuote string) string {
	otherQuote := `"`
	if enclosingQuote == `"` {
		otherQuote = "'"
	}

	// Escape and unescape single and double quotes as needed to be able to
	// enclose `rawText` with `enclosingQuote`.
	raw := makeStringPattern.ReplaceAllStringFunc(rawText, func(match string) string {
		// If we matched an escape, and the escaped character is a quote of the
		// other type than we intend to enclose the string with, there's no need for
		// it to be escaped, so return it _without_ the backslash.
		if len(match) == 2 {
			escaped := match[1:]
			if escaped == otherQuote {
				return otherQuote
			}
			return match
		}

		// If we matched an unescaped quote and it is of the _same_ type as we
		// intend to enclose the string with, it must be escaped, so return it with
		// a backslash.
		if match == enclosingQuote {
			return `\` + match
		}
		return match
	})

	return enclosingQuote + raw + enclosingQuote
}

// src/utilities/print-string.js, for the css parser: neither a JSON parser nor an HTML attribute, so the
// preferred quote decides. Upstream asserts the raw is quoted; its slice(1, -1) and charAt(0) are
// guarded here for a raw too short to have quotes.

func printString(raw string, options *printerOptions) string {
	// `rawContent` is the string exactly like it appeared in the input source
	// code, without its enclosing quotes.
	rawContent := ""
	if len(raw) >= 2 {
		rawContent = raw[1 : len(raw)-1]
	}

	enclosingQuote := getPreferredQuote(rawContent, settingsOf(options).SingleQuote)

	originalQuote := ""
	if len(raw) > 0 {
		originalQuote = raw[:1]
	}

	if originalQuote == enclosingQuote {
		return raw
	}

	return makeString(rawContent, enclosingQuote)
}

// src/utilities/print-number.js. Go's regexp has no lookahead, so each `(?=x)` is captured and put back.

var (
	numberScientificPlusAndZeros = regexp.MustCompile(`^([+-]?[\d.]+e)(?:\+|(-))?0*(\d)`)
	numberUselessScientific      = regexp.MustCompile(`^([+-]?[\d.]+)e[+-]?0+$`)
	numberLeadingDot             = regexp.MustCompile(`^([+-])?\.`)
	numberTrailingZeros          = regexp.MustCompile(`(\.\d+?)0+(e|$)`)
	numberTrailingDot            = regexp.MustCompile(`\.(e|$)`)
)

func printNumber(rawNumber string) string {
	if len(rawNumber) == 1 {
		return rawNumber
	}
	value := strings.ToLower(rawNumber)
	// Remove unnecessary plus and zeroes from scientific notation.
	value = replaceFirst(numberScientificPlusAndZeros, value, "${1}${2}${3}")
	// Remove unnecessary scientific notation (1e0).
	value = replaceFirst(numberUselessScientific, value, "${1}")
	// Make sure numbers always start with a digit.
	value = replaceFirst(numberLeadingDot, value, "${1}0.")
	// Remove extraneous trailing decimal zeroes.
	value = replaceFirst(numberTrailingZeros, value, "${1}${2}")
	// Remove trailing dot.
	value = replaceFirst(numberTrailingDot, value, "${1}")
	return value
}

// replaceFirst is JavaScript's String.prototype.replace with a non-global pattern.
func replaceFirst(pattern *regexp.Regexp, text string, template string) string {
	match := pattern.FindStringSubmatchIndex(text)
	if match == nil {
		return text
	}
	expanded := pattern.ExpandString(nil, template, text, match)
	return text[:match[0]] + string(expanded) + text[match[1]:]
}

// src/utilities/is-next-line-empty.js. It skips JavaScript's comment syntax, exactly as upstream does for
// every language.

func isNextLineEmpty(text string, startIndex int) bool {
	// upstream's `let oldIdx = null`: a value no index can equal.
	oldIndex, hasOldIndex := 0, false
	index := startIndex
	for !hasOldIndex || index != oldIndex {
		// We need to skip all the potential trailing inline comments
		oldIndex, hasOldIndex = index, true
		index = printing.SkipToLineEnd(text, index, false)
		index = printing.SkipInlineComment(text, index)
		index = printing.SkipSpaces(text, index, false)
	}
	index = printing.SkipTrailingComment(text, index)
	index = printing.SkipNewline(text, index, false)
	// upstream's `idx !== false`: the skip functions' false is a negative position other than -1.
	return index >= -1 && printing.HasNewline(text, index, false)
}

// JavaScript's string semantics.

// javaScriptWhitespaceClass is JavaScript's \s as a Go character class: WhiteSpace and LineTerminator.
const javaScriptWhitespaceClass = `[\t\n\v\f\r \x{00A0}\x{1680}\x{2000}-\x{200A}\x{2028}\x{2029}\x{202F}\x{205F}\x{3000}\x{FEFF}]`

func isJavaScriptWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x00A0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return character >= 0x2000 && character <= 0x200A
}

// trim, trimStart and trimEnd are String.prototype.trim and its halves.
func trim(value string) string      { return strings.TrimFunc(value, isJavaScriptWhitespace) }
func trimStart(value string) string { return strings.TrimLeftFunc(value, isJavaScriptWhitespace) }
func trimEnd(value string) string   { return strings.TrimRightFunc(value, isJavaScriptWhitespace) }

// toLowerCase is String.prototype.toLowerCase. Go's ToLower maps U+0130 to a bare i where JavaScript's
// special casing gives i followed by U+0307. JavaScript's final-sigma rule (a capital sigma at the end
// of a word lowercases to U+03C2) is not ported; a CSS identifier ending in a capital sigma would differ.
func toLowerCase(value string) string {
	return strings.ToLower(strings.ReplaceAll(value, "\xc4\xb0", "i\xcc\x87"))
}
