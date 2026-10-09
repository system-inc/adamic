// Data types: what a CSS value is, read the way Tailwind reads it.
//
// This is a port of Tailwind 4.3.3's `inferDataType` and `isColor` (`src/utils/infer-data-type.ts`
// and `src/utils/is-color.ts`), plus the two helpers they lean on, `segment` and `hasMathFn`.
//
// It is the most load-bearing piece of the Go port because the reading of a utility root is a pure
// function of the inferred type. `bg-[#fff]` is a color utility and `bg-[10px]` is a size utility
// because this file said so; nothing downstream re-derives it. An inexact answer here does not
// produce a visible error, it produces a different utility, silently, and every ordering decision
// built on top inherits the mistake.
//
// # Two details decide correctness
//
// A value starting with `var(` returns no type at all, before any check runs. Not "unknown, try the
// checks anyway" — the short-circuit is the whole behavior, because `var(--x)` could hold anything
// and Tailwind declines to guess. `IsColor` on its own would happily say `var(--color-red)` is not a
// color and `isLength` would say it is not a length, but the short-circuit means callers never get
// far enough to ask.
//
// The first matching type in the caller's list wins, so the caller's type order is part of the
// contract, not a detail of iteration. This is not theoretical: with the full type list in
// declaration order the engine reads `1 2 3` as `line-width` rather than `vector`, and `serif` as
// `family-name` rather than `generic-name`, purely because the earlier type also matches. Callers
// pass the order Tailwind's utility definitions pass, and this function preserves it.
//
// # Why the checks are byte scanners rather than regexps
//
// The upstream predicates are regexps built by string concatenation, e.g. `IS_LENGTH` is
// `^<number>(cm|mm|Q|...)$`. Go's regexp package would match these, but the number grammar
// `[+-]?\d*\.?\d+(?:[eE][+-]?\d+)?` recurs in six of them and the units are matched by alternation
// with no anchoring between number and unit, which is exactly where a transcribed regexp goes
// subtly wrong. Scanning the number once and then testing the tail against a unit set is the same
// grammar expressed so that the two halves cannot drift apart, and it is checked against the real
// engine rather than against this reasoning.
package tailwind

import "strings"

// DataType is one of the CSS data types Tailwind can infer.
//
// Spelled as the engine spells them, hyphens and all, because these strings cross into generated
// descriptor tables where a Go-flavored renaming would have to be undone on every comparison.
type DataType string

const (
	DataTypeColor          DataType = "color"
	DataTypeLength         DataType = "length"
	DataTypePercentage     DataType = "percentage"
	DataTypeRatio          DataType = "ratio"
	DataTypeNumber         DataType = "number"
	DataTypeInteger        DataType = "integer"
	DataTypeURL            DataType = "url"
	DataTypePosition       DataType = "position"
	DataTypeBackgroundSize DataType = "bg-size"
	DataTypeLineWidth      DataType = "line-width"
	DataTypeImage          DataType = "image"
	DataTypeFamilyName     DataType = "family-name"
	DataTypeGenericName    DataType = "generic-name"
	DataTypeAbsoluteSize   DataType = "absolute-size"
	DataTypeRelativeSize   DataType = "relative-size"
	DataTypeAngle          DataType = "angle"
	DataTypeVector         DataType = "vector"
)

// AllDataTypes is every type, in the order the engine's own `checks` table declares them.
//
// Callers with no opinion about order should use this rather than building their own list, because
// first-match-wins makes an ad-hoc order a silent behavior change. It is a function returning a
// fresh slice rather than a package-level variable so that a caller sorting or truncating it cannot
// reach back and change what every other caller sees.
func AllDataTypes() []DataType {
	return []DataType{
		DataTypeColor,
		DataTypeLength,
		DataTypePercentage,
		DataTypeRatio,
		DataTypeNumber,
		DataTypeInteger,
		DataTypeURL,
		DataTypePosition,
		DataTypeBackgroundSize,
		DataTypeLineWidth,
		DataTypeImage,
		DataTypeFamilyName,
		DataTypeGenericName,
		DataTypeAbsoluteSize,
		DataTypeRelativeSize,
		DataTypeAngle,
		DataTypeVector,
	}
}

// InferDataType returns the first type in types that value satisfies, or empty when none do.
//
// The empty DataType stands for the engine's `null`. A separate boolean would be more idiomatic Go,
// but every caller of this in the port immediately compares the result against a descriptor's
// declared type, and a two-value return turns each of those comparisons into a two-line dance
// around a value that is already unambiguous: no type is a valid answer, not an error.
func InferDataType(value string, types []DataType) DataType {
	// Before any check runs. See the package comment: this is behavior, not an optimization.
	if strings.HasPrefix(value, "var(") {
		return ""
	}
	for _, dataType := range types {
		if matchesDataType(value, dataType) {
			return dataType
		}
	}
	return ""
}

// matchesDataType is the engine's `checks` table.
//
// An unknown type matches nothing, mirroring `checks[type]?.(value)`, where an absent entry makes
// the optional call evaluate to undefined and the branch fall through. That matters because the
// generated descriptor tables carry type names from Tailwind's source, and a Tailwind version that
// adds a type would otherwise panic here instead of declining to match.
func matchesDataType(value string, dataType DataType) bool {
	switch dataType {
	case DataTypeColor:
		return IsColor(value)
	case DataTypeLength:
		return IsLength(value)
	case DataTypePercentage:
		return isPercentage(value)
	case DataTypeRatio:
		return isFraction(value)
	case DataTypeNumber:
		return isNumber(value)
	case DataTypeInteger:
		return IsPositiveInteger(value)
	case DataTypeURL:
		return isURL(value)
	case DataTypePosition:
		return isBackgroundPosition(value)
	case DataTypeBackgroundSize:
		return isBackgroundSize(value)
	case DataTypeLineWidth:
		return isLineWidth(value)
	case DataTypeImage:
		return isImage(value)
	case DataTypeFamilyName:
		return isFamilyName(value)
	case DataTypeGenericName:
		return isGenericName(value)
	case DataTypeAbsoluteSize:
		return isAbsoluteSize(value)
	case DataTypeRelativeSize:
		return isRelativeSize(value)
	case DataTypeAngle:
		return isAngle(value)
	case DataTypeVector:
		return isVector(value)
	}
	return false
}

/* -------------------------------------------------------------------------- */

// isURL ports `IS_URL = /^url\(.*\)$/`.
//
// `.` in JavaScript regexps without the `s` flag does not match a newline, so a value containing one
// between the parens is not a url to the engine even though the string plainly starts with `url(`
// and ends with `)`. Reproduced rather than tidied, because "the obvious reading" is how ports drift.
func isURL(value string) bool {
	if !strings.HasPrefix(value, "url(") || !strings.HasSuffix(value, ")") {
		return false
	}
	// Needs at least `url(` and `)` without overlapping.
	if len(value) < len("url()") {
		return false
	}
	return !strings.ContainsAny(value[len("url("):len(value)-1], "\n\r  ")
}

/* -------------------------------------------------------------------------- */

// isLineWidth reports whether every space-separated part is a length, a number, or a width keyword.
//
// Note that `every` on an empty list is true in JavaScript, and segment never returns an empty list
// (it always pushes a final part), so the empty string reaches this as one empty part and fails on
// the part rather than vacuously passing.
func isLineWidth(value string) bool {
	for _, part := range segment(value, ' ') {
		if IsLength(part) || isNumber(part) {
			continue
		}
		if part == "thin" || part == "medium" || part == "thick" {
			continue
		}
		return false
	}
	return true
}

/* -------------------------------------------------------------------------- */

// imageFunctions ports `IS_IMAGE_FN = /^(?:element|image|cross-fade|image-set)\(/`.
var imageFunctions = []string{"element(", "image(", "cross-fade(", "image-set("}

// gradientFunctions ports `IS_GRADIENT_FN = /^(repeating-)?(conic|linear|radial)-gradient\(/`.
//
// Written out rather than assembled from an optional prefix, because the alternation is small and
// six literal strings are easier to check against the source than a loop that builds them.
var gradientFunctions = []string{
	"conic-gradient(",
	"linear-gradient(",
	"radial-gradient(",
	"repeating-conic-gradient(",
	"repeating-linear-gradient(",
	"repeating-radial-gradient(",
}

// isImage reports whether every comma-separated part is a url, a gradient, or an image function.
//
// `var(` parts are skipped without counting, so a value made only of `var()` parts has count zero
// and is not an image. The count is what makes that distinction: `every`-style iteration alone would
// return true for a list of nothing but skips.
//
// Both regexps are case-sensitive and unanchored at the end, so this is a prefix test.
func isImage(value string) bool {
	count := 0
	for _, part := range segment(value, ',') {
		if strings.HasPrefix(part, "var(") {
			continue
		}
		if isURL(part) {
			count++
			continue
		}
		if hasAnyPrefix(part, gradientFunctions) {
			count++
			continue
		}
		if hasAnyPrefix(part, imageFunctions) {
			count++
			continue
		}
		return false
	}
	return count > 0
}

// hasAnyPrefix reports whether value starts with any of the prefixes.
func hasAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

/* -------------------------------------------------------------------------- */

// genericNames is the `<generic-family>` keyword set, matched exactly and case-sensitively.
var genericNames = map[string]bool{
	"serif": true, "sans-serif": true, "monospace": true, "cursive": true,
	"fantasy": true, "system-ui": true, "ui-serif": true, "ui-sans-serif": true,
	"ui-monospace": true, "ui-rounded": true, "math": true, "emoji": true,
	"fangsong": true,
}

func isGenericName(value string) bool {
	return genericNames[value]
}

// isFamilyName is deliberately permissive: anything not starting with a digit counts.
//
// This is why `[foo_bar]` and `larger` both read as family names when `family-name` appears earlier
// in the caller's type list than the type the author meant. It is not a bug to fix here; it is the
// reason type order is part of the contract.
//
// The digit test reads the first byte of each part, including the empty part produced by a leading
// or doubled comma, where JavaScript's `charCodeAt(0)` yields NaN and fails the range test. An empty
// part therefore counts rather than rejecting, which is why `,serif` is a family name.
func isFamilyName(value string) bool {
	count := 0
	for _, part := range segment(value, ',') {
		if len(part) > 0 && part[0] >= '0' && part[0] <= '9' {
			return false
		}
		if strings.HasPrefix(part, "var(") {
			continue
		}
		count++
	}
	return count > 0
}

var absoluteSizes = map[string]bool{
	"xx-small": true, "x-small": true, "small": true, "medium": true,
	"large": true, "x-large": true, "xx-large": true, "xxx-large": true,
}

func isAbsoluteSize(value string) bool {
	return absoluteSizes[value]
}

func isRelativeSize(value string) bool {
	return value == "larger" || value == "smaller"
}

/* -------------------------------------------------------------------------- */

// scanNumber ports the shared `HAS_NUMBER = /[+-]?\d*\.?\d+(?:[eE][+-]?\d+)?/` grammar, anchored at
// the start of value, and returns how many bytes it consumed.
//
// Returns 0 when no number is there. The grammar is fussier than it looks and each clause earns its
// place: the digits after the optional `.` are required (`\d+`), the digits before it are not
// (`\d*`), so `.5` is a number and `5.` is not. The exponent is all-or-nothing, so `1e` is not a
// number but `1` followed by the leftover `e` is, which is what makes `1em` a length rather than a
// malformed number.
func scanNumber(value string) int {
	index := 0
	if index < len(value) && (value[index] == '+' || value[index] == '-') {
		index++
	}
	// `\d*`
	for index < len(value) && value[index] >= '0' && value[index] <= '9' {
		index++
	}
	// `\.?\d+` — the fractional digits are required, whether or not a dot appeared.
	if index < len(value) && value[index] == '.' {
		index++
		digitsStart := index
		for index < len(value) && value[index] >= '0' && value[index] <= '9' {
			index++
		}
		if index == digitsStart {
			return 0
		}
	} else if index == 0 || !isDigit(value[index-1]) {
		// No dot, so the `\d+` must have been satisfied by the `\d*` run above. It was not if we
		// consumed nothing, or consumed only a sign.
		return 0
	}
	// `(?:[eE][+-]?\d+)?`
	if index < len(value) && (value[index] == 'e' || value[index] == 'E') {
		exponent := index + 1
		if exponent < len(value) && (value[exponent] == '+' || value[exponent] == '-') {
			exponent++
		}
		digitsStart := exponent
		for exponent < len(value) && value[exponent] >= '0' && value[exponent] <= '9' {
			exponent++
		}
		if exponent > digitsStart {
			index = exponent
		}
	}
	return index
}

func isDigit(character byte) bool {
	return character >= '0' && character <= '9'
}

// numberWithSuffix reports whether value is exactly a number followed by suffix.
//
// This is the `^<number><something>$` shape shared by percentage, length, and angle. Anchoring both
// ends here rather than at each call site is the point: the upstream regexps are assembled by string
// concatenation and an unanchored one would be a whole class of false positives.
func numberWithSuffix(value string, suffixes []string) bool {
	consumed := scanNumber(value)
	if consumed == 0 {
		return false
	}
	tail := value[consumed:]
	for _, suffix := range suffixes {
		if tail == suffix {
			return true
		}
	}
	return false
}

/* -------------------------------------------------------------------------- */

// isNumber ports `IS_NUMBER.test(value) || hasMathFn(value)`.
//
// The math-function fallback is why `calc(1px+2px)` reads as a number, a percentage, a fraction and
// a length all at once, and therefore why the caller's type order decides which one it becomes.
func isNumber(value string) bool {
	if consumed := scanNumber(value); consumed == len(value) && consumed > 0 {
		return true
	}
	return hasMathFunction(value)
}

func isPercentage(value string) bool {
	return numberWithSuffix(value, percentSuffix) || hasMathFunction(value)
}

var percentSuffix = []string{"%"}

// isFraction ports `^<number>\s*/\s*<number>$`.
//
// `\s` in a JavaScript regexp is a wider set than a plain space: it includes tab, newline, form
// feed, carriage return, vertical tab, non-breaking space, BOM, and the Unicode space separators.
// Reproduced through isJavaScriptSpace rather than narrowed to ASCII, because narrowing would make
// a value with a non-breaking space around the slash read as not-a-ratio here and as a ratio in the
// engine.
func isFraction(value string) bool {
	if hasMathFunction(value) {
		return true
	}
	consumed := scanNumber(value)
	if consumed == 0 {
		return false
	}
	rest := value[consumed:]
	rest = trimLeadingJavaScriptSpace(rest)
	if !strings.HasPrefix(rest, "/") {
		return false
	}
	rest = trimLeadingJavaScriptSpace(rest[1:])
	denominator := scanNumber(rest)
	return denominator > 0 && denominator == len(rest)
}

/* -------------------------------------------------------------------------- */

// lengthUnits is the CSS length unit list, matched case-sensitively and exactly as the engine's
// alternation matches it. `Q` is uppercase on purpose.
var lengthUnits = []string{
	"cm", "mm", "Q", "in", "pc", "pt", "px",
	"em", "ex", "ch", "rem", "lh", "rlh",
	"vw", "vh", "vmin", "vmax", "vb", "vi",
	"svw", "svh", "lvw", "lvh", "dvw", "dvh",
	"cqw", "cqh", "cqi", "cqb", "cqmin", "cqmax",
}

// IsLength ports `IS_LENGTH.test(value) || IS_LENGTH_FN.test(value) || hasMathFn(value)`.
//
// `IS_LENGTH_FN = /^(--spacing)\(/i` is case-insensitive, which only matters for the letters, so
// `--SPACING(` is a length to the engine.
//
// Exported because the utility evaluator and the value parser both need to ask this directly rather
// than through a one-element type list.
func IsLength(value string) bool {
	if numberWithSuffix(value, lengthUnits) {
		return true
	}
	if len(value) >= len("--spacing(") && strings.EqualFold(value[:len("--spacing(")], "--spacing(") {
		return true
	}
	return hasMathFunction(value)
}

/* -------------------------------------------------------------------------- */

var backgroundPositionKeywords = map[string]bool{
	"center": true, "top": true, "right": true, "bottom": true, "left": true,
}

// isBackgroundPosition reports whether every space-separated part is a keyword, a length, or a
// percentage, with `var(` parts skipped and at least one part counted.
func isBackgroundPosition(value string) bool {
	count := 0
	for _, part := range segment(value, ' ') {
		if backgroundPositionKeywords[part] {
			count++
			continue
		}
		if strings.HasPrefix(part, "var(") {
			continue
		}
		if IsLength(part) || isPercentage(part) {
			count++
			continue
		}
		return false
	}
	return count > 0
}

/* -------------------------------------------------------------------------- */

// isBackgroundSize ports the `<bg-size>#` grammar.
//
// The control flow here is worth reading twice, because it is not the same shape as its neighbors.
// A comma part that has one or two space-separated values but fails the `auto | length | percentage`
// test does not return false, it falls out of the loop body without counting. So
// `cover,nonsense` is still a background size, on the strength of `cover` alone, while
// `cover,a b c` is not, because the three-value part takes the explicit return. Upstream reads as
// though it might be an oversight; it is reproduced rather than corrected, since the whole point of
// this file is to answer what the engine answers.
func isBackgroundSize(value string) bool {
	count := 0
	for _, size := range segment(value, ',') {
		if size == "cover" || size == "contain" {
			count++
			continue
		}
		values := segment(size, ' ')
		if len(values) != 1 && len(values) != 2 {
			return false
		}
		allValid := true
		for _, part := range values {
			if part == "auto" || IsLength(part) || isPercentage(part) {
				continue
			}
			allValid = false
			break
		}
		if allValid {
			count++
		}
	}
	return count > 0
}

/* -------------------------------------------------------------------------- */

var angleUnits = []string{"deg", "rad", "grad", "turn"}

// isAngle ports `^<number>(deg|rad|grad|turn)$`. No math-function fallback, unlike length.
func isAngle(value string) bool {
	return numberWithSuffix(value, angleUnits)
}

/* -------------------------------------------------------------------------- */

// isVector ports `^<number> +<number> +<number>$`.
//
// The separator is `+` on a literal space, not `\s+`, so a tab between components is not a vector
// even though the fraction check would accept a tab around its slash. The two upstream regexps
// genuinely differ here.
func isVector(value string) bool {
	rest := value
	for component := range 3 {
		consumed := scanNumber(rest)
		if consumed == 0 {
			return false
		}
		rest = rest[consumed:]
		if component == 2 {
			break
		}
		spaces := 0
		for spaces < len(rest) && rest[spaces] == ' ' {
			spaces++
		}
		if spaces == 0 {
			return false
		}
		rest = rest[spaces:]
	}
	return rest == ""
}

/* -------------------------------------------------------------------------- */

// IsPositiveInteger ports `Number.isInteger(num) && num >= 0 && String(num) === String(value)`.
//
// The round-trip through `String(Number(value))` is the entire specification and the reason this is
// not `strconv.Atoi`. It rejects every spelling JavaScript's number parser accepts but does not
// print back identically: `+1`, `1.0`, `01`, ` 1 `, `0x10`, `1e3`, and the empty string, which
// `Number(”)` reads as 0 while `String(0)` is `"0"`.
//
// So the rule reduces to: a run of ASCII digits with no leading zero unless the value is exactly
// `0`. That equivalence holds for every string, including ones too large to be an exact float,
// because JavaScript prints those in exponential form and the round-trip catches it. The bound is
// applied explicitly below rather than left implicit.
func IsPositiveInteger(value string) bool {
	if value == "" {
		return false
	}
	for index := range len(value) {
		if !isDigit(value[index]) {
			return false
		}
	}
	if len(value) > 1 && value[0] == '0' {
		return false
	}
	// Past 2^53 the round-trip fails upstream: `String(Number('9007199254740993'))` is
	// `'9007199254740992'`. Digit-length is a sufficient guard because any 16-digit value is below
	// 2^53 and this only needs to reject what JavaScript would reprint differently.
	if len(value) > 15 {
		return roundTripsAsJavaScriptNumber(value)
	}
	return true
}

// IsStrictPositiveInteger is IsPositiveInteger with zero excluded.
func IsStrictPositiveInteger(value string) bool {
	return IsPositiveInteger(value) && value != "0"
}
