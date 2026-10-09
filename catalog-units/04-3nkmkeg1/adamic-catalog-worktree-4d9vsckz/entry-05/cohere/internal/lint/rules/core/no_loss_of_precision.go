package core

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoLossOfPrecision = rule.Message{
	Id: "noLossOfPrecision",
	Description: "This number literal does not survive being stored. A JavaScript number is a " +
		"64-bit float with 53 bits of significand, so it holds about 15 significant decimal " +
		"digits and every integer up to 9007199254740991. The digits written here exceed that, " +
		"and the value the program actually runs with is a different number than the one on the " +
		"page. Write a value the format can hold, or use a `BigInt` literal such as " +
		"`9007199254740993n` when the exact integer is the point.",
}

// NoLossOfPrecision flags a numeric literal whose written digits are not the digits the program
// gets after the literal is stored as a 64-bit float.
//
//	valid:   var x = 12345;
//	valid:   var x = 123.0000000000000000000000;
//	valid:   var x = 0x1FFF_FFFF_FFF_FFF;
//	valid:   var x = 0195;
//	valid:   var x = 9007199254740993n;
//	invalid: var x = 9007199254740993;
//	invalid: var x = 2e999;
//	invalid: var x = 0X200000_0000000_1;
//	invalid: var x = 1230000000000000000000000.0;
//
// The defect is silent. `9007199254740993 === 9007199254740992` is true, so the literal compares
// equal to a number nobody wrote and no error is raised at any point. It matters most where the
// digits came from somewhere that meant them exactly, such as an identifier out of a database, and
// the rule exists because reading the literal is the only moment the original digits still exist.
//
// # The predicate is a round-trip, not a digit count
//
// Counting digits gets `123.0000000000000000000000` wrong, which is 25 digits of which three are
// significant and which stores exactly. The question is whether the written value and the stored
// value are the same number, so the rule parses the literal, formats the stored float back to the
// same number of significant digits, and compares the two in a normalized scientific form.
//
// # Our parser normalizes literal text, so the rule reads the raw span
//
// This is the one place where a port from oxc does not transfer. Upstream reads `NumericLiteral.raw`
// and gets the source text. Our parser hands back a normalized `Text` instead: `1e1` arrives as
// `10`, `0x1FFFFFFFFFFFFF` arrives as `9007199254740991`, and `9007199254740993` arrives as
// `9007199254740992` because the parser has *already performed the rounding this rule is looking
// for*. Reading `.Text` would make the rule structurally unable to see its own subject. Measured
// with a probe rather than assumed. So the raw span comes out of the source file, and
// `TestNoLossOfPrecisionReadsRawSourceText` pins it with literals whose two spellings disagree.
//
// # BigInt is exempt because it cannot lose precision
//
// A BigInt literal is `KindBigIntLiteral`, a different node kind, so listening on
// `KindNumericLiteral` exempts `9007199254740993n` without an explicit arm. That is the same way
// upstream reaches the exemption, and it is the right answer rather than an accident: BigInt is
// arbitrary-precision, so there is no rounding to detect.
//
// # A sign is never part of the literal
//
// `-9007199254740993` parses as a unary minus applied to a positive literal, so the raw span this
// rule sees never carries a sign. Upstream trims `+` and `-` anyway; that trim is unreachable here
// and is therefore not reproduced, which is why the corpus's signed cases still report: they report
// on the unsigned literal inside them.
var NoLossOfPrecision = rule.Rule{
	Name: "no-loss-of-precision",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindNumericLiteral: func(node *ast.Node) {
				literalRange := rule.TokenRange(ctx.SourceFile, node)
				raw := ctx.SourceFile.Text()[literalRange.Pos():literalRange.End()]
				if numericLiteralLosesPrecision(raw) {
					ctx.ReportNode(node, messageNoLossOfPrecision)
				}
			},
		}
	},
}

// maximumSignificandBits is how many bits of significand a 64-bit float carries, counting the
// implicit leading one. A non-decimal literal whose value needs more than this cannot be exact.
const maximumSignificandBits = 53

// numericLiteralLosesPrecision answers the whole rule for one literal's raw source text.
func numericLiteralLosesPrecision(raw string) bool {
	if rawNumericLiteralIsBaseTen(raw) {
		return baseTenLiteralLosesPrecision(raw)
	}
	return nonBaseTenLiteralLosesPrecision(raw)
}

// rawNumericLiteralIsBaseTen says whether a literal's raw text is written in base ten.
//
// The three prefixed forms are unambiguous. The hard case is a bare leading zero, which is a legacy
// octal literal only when every remaining digit is octal: `0195`, `019.5`, and `0008` are decimal
// despite the zero, and all three are upstream pass cases that a rule treating any leading zero as
// octal reports. `0e5` is decimal for the same reason.
func rawNumericLiteralIsBaseTen(raw string) bool {
	if len(raw) < 2 || raw[0] != '0' {
		return true
	}
	switch raw[1] {
	case 'x', 'X', 'b', 'B', 'o', 'O':
		return false
	}
	for index := 1; index < len(raw); index++ {
		character := raw[index]
		if character == '_' {
			continue
		}
		if character < '0' || character > '7' {
			return true
		}
	}
	return false
}

// nonBaseTenLiteralLosesPrecision answers for a hex, octal, or binary literal.
//
// These have no exponent and no fraction, so the only question is whether the integer fits in the
// significand. The bit-width check answers that directly for everything under 53 bits and is the
// path almost every literal takes. Beyond it the value is parsed, pushed through a float64, and
// formatted back in its own radix: if the digits that come back are not the digits that went in,
// the trip through the float changed the number.
func nonBaseTenLiteralLosesPrecision(raw string) bool {
	if nonBaseTenLiteralIsExact(raw) {
		return false
	}

	digits := stripNumericSeparators(raw)
	radix, withoutPrefix := nonBaseTenRadixAndDigits(digits)
	value, err := strconv.ParseUint(withoutPrefix, radix, 64)
	if err != nil {
		// Wider than 64 bits, so the exactness check above has already established it does not
		// fit in 53 and this is unreachable for a literal the parser accepted. Reported rather
		// than ignored, because a literal this rule cannot evaluate is not a literal it has
		// cleared.
		return true
	}

	// The round trip that is the actual question: store the integer in a float and read it back.
	stored := uint64(float64(value))
	return !strings.HasSuffix(strings.ToUpper(digits), strings.ToUpper(strconv.FormatUint(stored, radix)))
}

// nonBaseTenRadixAndDigits splits a separator-free non-decimal literal into its radix and digits.
//
// The legacy octal form keeps its leading zeros rather than being trimmed. Trimming them is what
// upstream does and it changes no answer here: `ParseUint` in base eight reads `0400...1` and
// `400...1` as the same value, and the bit-width scan below skips leading zeros on its own. A
// mutation removing the trim survived every fixture, which is what a redundant call looks like, so
// it is gone rather than sitting there implying it decides something.
func nonBaseTenRadixAndDigits(raw string) (int, string) {
	if len(raw) >= 2 {
		switch raw[1] {
		case 'b', 'B':
			return 2, raw[2:]
		case 'o', 'O':
			return 8, raw[2:]
		case 'x', 'X':
			return 16, raw[2:]
		}
	}
	return 8, raw
}

// nonBaseTenLiteralIsExact says whether a non-decimal literal's value fits in the significand.
//
// It accumulates the bit width as it reads, starting at the first non-zero digit so that leading
// zeros cost nothing, and charges that first digit only the bits it actually occupies. That last
// part is what makes `0x1FFF_FFFF_FFF_FFF` exact at exactly 53 bits while `0x20000000000001` is
// not: charging every hex digit four bits would put the first at 56 and report a literal upstream
// passes.
//
// It bails to false on any character it does not recognize, which hands the literal to the
// round-trip path rather than clearing it.
func nonBaseTenLiteralIsExact(raw string) bool {
	radix, digits := nonBaseTenRadixAndDigits(raw)

	bitLength := 0
	for index := 0; index < len(digits); index++ {
		character := digits[index]
		if character == '_' {
			continue
		}

		var digit byte
		switch {
		case character >= '0' && character <= '9':
			digit = character - '0'
		case character >= 'a' && character <= 'f':
			digit = character - 'a' + 10
		case character >= 'A' && character <= 'F':
			digit = character - 'A' + 10
		default:
			// Unreachable for any literal the parser produced, measured rather than assumed: a
			// panic here survived all 150 fixtures. A prefixed literal only contains digits valid
			// for its radix, and the bare-zero form is classified as decimal by
			// `rawNumericLiteralIsBaseTen` the moment it holds anything outside `0` to `7`, so it
			// never arrives. Kept as a total switch rather than deleted, because the alternative is
			// an unhandled byte silently counting as zero bits.
			return false
		}

		if bitLength == 0 {
			// Leading zeros contribute nothing, and the first significant digit contributes only
			// the bits it occupies rather than the full width of its radix.
			if digit == 0 {
				continue
			}
			bitLength = significantBitsInLeadingDigit(radix, digit)
		} else {
			bitLength += bitsPerDigit(radix)
		}

		if bitLength > maximumSignificandBits {
			return false
		}
	}

	return true
}

// significantBitsInLeadingDigit is how many bits the first non-zero digit of a literal occupies.
func significantBitsInLeadingDigit(radix int, digit byte) int {
	switch radix {
	case 2:
		return 1
	case 8:
		switch {
		case digit == 1:
			return 1
		case digit <= 3:
			return 2
		default:
			return 3
		}
	default:
		switch {
		case digit == 1:
			return 1
		case digit <= 3:
			return 2
		case digit <= 7:
			return 3
		default:
			return 4
		}
	}
}

// bitsPerDigit is how many bits each digit after the first contributes.
func bitsPerDigit(radix int) int {
	switch radix {
	case 2:
		return 1
	case 8:
		return 3
	default:
		return 4
	}
}

// baseTenLiteralLosesPrecision answers for a decimal literal.
//
// The literal is parsed to the float the program will run with, then that float is formatted back
// to exactly as many significant digits as were written. If the written digits and the stored
// digits describe the same number in scientific form, nothing was lost.
func baseTenLiteralLosesPrecision(raw string) bool {
	withoutSeparators := stripNumericSeparators(raw)

	value, err := strconv.ParseFloat(withoutSeparators, 64)
	if err != nil {
		// `ParseFloat` reports a range error for `2e999` while still returning the infinity, and
		// that is a real finding rather than a parse failure: the literal names a value the format
		// cannot hold at all. Any other error is a literal this rule cannot evaluate.
		if !isRangeError(err) {
			return false
		}
	}
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return true
	}

	writtenDigits, writtenExponent, ok := scientificFormOfDecimal(withoutSeparators, false)
	if !ok {
		return true
	}
	if writtenDigits == "0" {
		// Every digit written was a zero, so the literal is exactly zero unless the exponent
		// carried it somewhere else, which it cannot.
		return value != 0
	}
	if len(writtenDigits) > maximumComparableSignificantDigits {
		// More significant digits than any formatter will produce, so the round trip cannot be
		// performed and the literal is far past what the format holds.
		return true
	}

	// `strconv` with 'e' and a precision of n-1 yields exactly n significant digits, which is the
	// same shape upstream builds by hand out of its own `toPrecision`.
	stored := strconv.FormatFloat(value, 'e', len(writtenDigits)-1, 64)
	storedDigits, storedExponent, ok := scientificFormOfDecimal(stored, true)
	if !ok {
		return true
	}

	return storedDigits != writtenDigits || storedExponent != writtenExponent
}

// maximumComparableSignificantDigits is the largest significant-digit count the comparison can be
// performed at. It matches the ceiling upstream applies for the same reason: past this the stored
// side cannot be rendered to the requested width.
const maximumComparableSignificantDigits = 100

// scientificFormOfDecimal reduces a decimal literal to its significant digits and an exponent, so
// that two spellings of one number compare equal.
//
// `9007199254740.993e3`, `9.007199254740993e15`, and `9007199254740993` all reduce to the same
// digits and the same exponent, which is what lets one comparison serve every spelling.
//
// `keepTrailingZeros` is the subtle parameter and it is why four upstream fail cases exist. A
// trailing zero after a decimal point is significant: `1230000000000000000000000.0` demands 25
// digits of the round trip and reports, while `123` demands three and does not. Trimming those
// zeros asks the round trip for far fewer digits than were written, and all four of
// `1230000000000000000000000.0`, `.1230000000000000000000000`,
// `123_00000000000000000000_00.0_0`, and `.1_23000000000000_00000_0000_0` pass a rule that trims
// them. Found by running the corpus, not by reading the algorithm.
//
// A raw literal containing a point sets the flag on its own; the caller passes true for the stored
// side because `strconv` pads to the requested width and those zeros must be kept to compare.
func scientificFormOfDecimal(raw string, keepTrailingZeros bool) (string, int, bool) {
	significand := raw
	exponent := 0
	if index := strings.IndexAny(significand, "eE"); index >= 0 {
		parsed, err := strconv.Atoi(significand[index+1:])
		if err != nil {
			return "", 0, false
		}
		exponent = parsed
		significand = significand[:index]
	}

	integerPart, fractionPart := significand, ""
	if index := strings.IndexByte(significand, '.'); index >= 0 {
		integerPart, fractionPart = significand[:index], significand[index+1:]
		keepTrailingZeros = true
	}

	digits := integerPart + fractionPart
	leadingZeros := 0
	for leadingZeros < len(digits) && digits[leadingZeros] == '0' {
		leadingZeros++
	}
	if leadingZeros == len(digits) {
		return "0", 0, true
	}

	// The exponent of the leading significant digit: how far the point sits from it, plus whatever
	// the written exponent contributed.
	exponent += len(integerPart) - leadingZeros - 1
	digits = digits[leadingZeros:]

	if !keepTrailingZeros {
		digits = strings.TrimRight(digits, "0")
	}

	return digits, exponent, true
}

// stripNumericSeparators removes the `_` separators a numeric literal may carry.
func stripNumericSeparators(raw string) string {
	if !strings.ContainsRune(raw, '_') {
		return raw
	}
	return strings.ReplaceAll(raw, "_", "")
}

// isRangeError says whether a `strconv` failure was an overflow rather than a malformed literal.
//
// An overflow still returns a usable infinity, which is exactly what the rule wants for `2e999`.
func isRangeError(err error) bool {
	var numError *strconv.NumError
	if !errors.As(err, &numError) {
		return false
	}
	return numError.Err == strconv.ErrRange
}
