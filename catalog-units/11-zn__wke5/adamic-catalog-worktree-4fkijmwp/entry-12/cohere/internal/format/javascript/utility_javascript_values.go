package javascript

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// JavaScript's own conversions, which upstream reaches for with String(number) and JSON.stringify.
// Go's nearest equivalents differ: strconv's formats never switch to an exponent at JavaScript's
// thresholds, and encoding/json escapes <, >, & and U+2028/U+2029, which JSON.stringify leaves as they are.

// jsonStringify is JSON.stringify for a string: the quotes, and escapes for the quote, the backslash
// and the control characters, with the short forms where JSON has them.
func jsonStringify(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"':
			builder.WriteString(`\"`)
		case '\\':
			builder.WriteString(`\\`)
		case '\b':
			builder.WriteString(`\b`)
		case '\f':
			builder.WriteString(`\f`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			if character < 0x20 {
				fmt.Fprintf(&builder, `\u%04x`, character)
			} else {
				builder.WriteRune(character)
			}
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

// javaScriptNumberString is String(number), Number::toString in the specification: the shortest
// digits that round-trip, written plainly from 1e-7 up to 1e21 and with an exponent outside that.
func javaScriptNumberString(value float64) string {
	switch {
	case math.IsNaN(value):
		return "NaN"
	case math.IsInf(value, 1):
		return "Infinity"
	case math.IsInf(value, -1):
		return "-Infinity"
	case value == 0:
		return "0"
	case value < 0:
		return "-" + javaScriptNumberString(-value)
	}
	// FormatFloat's 'e' with precision -1 is the shortest round-trip digits: d.ddde±x.
	scientific := strconv.FormatFloat(value, 'e', -1, 64)
	mantissa, exponentText, _ := strings.Cut(scientific, "e")
	digits := strings.Replace(mantissa, ".", "", 1)
	exponent, _ := strconv.Atoi(exponentText)
	// The specification's n: the value is 0.digits × 10^n.
	n := exponent + 1
	k := len(digits)
	switch {
	case k <= n && n <= 21:
		return digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return "0." + strings.Repeat("0", -n) + digits
	}
	sign := "+"
	if n-1 < 0 {
		sign = "-"
	}
	power := strconv.Itoa(int(math.Abs(float64(n - 1))))
	if k == 1 {
		return digits + "e" + sign + power
	}
	return digits[:1] + "." + digits[1:] + "e" + sign + power
}
