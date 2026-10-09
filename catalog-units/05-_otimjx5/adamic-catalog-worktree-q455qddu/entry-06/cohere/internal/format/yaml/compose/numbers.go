package compose

// JavaScript's number parsing, as the schema's tags call it: parseInt, parseFloat and Number over
// strings the tags' tests have already shaped, all ASCII.

import (
	"math"
	"math/big"
	"strconv"
)

func nan() float64              { return math.NaN() }
func positiveInfinity() float64 { return math.Inf(1) }
func negativeInfinity() float64 { return math.Inf(-1) }

// digitValue is a code unit's value as a digit, as parseInt reads digits: 99 for a non-digit.
func digitValue(unit uint16) int {
	switch {
	case unit >= '0' && unit <= '9':
		return int(unit - '0')
	case unit >= 'a' && unit <= 'z':
		return int(unit-'a') + 10
	case unit >= 'A' && unit <= 'Z':
		return int(unit-'A') + 10
	}
	return 99
}

// parseInt is the global `parseInt(text, radix)`: leading whitespace and a sign skipped, a 0x prefix
// stripped for radix 16, then the longest run of the radix's digits. Radix 10 rounds the way V8 does,
// correctly (strconv), and the power-of-two radixes exactly, rounding half to even (big.Float).
func parseInt(text []uint16, radix int) float64 {
	index := 0
	for index < len(text) && isJavaScriptWhitespace(text[index]) {
		index++
	}
	sign := 1.0
	if index < len(text) && text[index] == '-' {
		sign = -1
	}
	if index < len(text) && (text[index] == '+' || text[index] == '-') {
		index++
	}
	if radix == 16 && index+1 < len(text) && text[index] == '0' && (text[index+1] == 'x' || text[index+1] == 'X') {
		index += 2
	}
	start := index
	for index < len(text) && digitValue(text[index]) < radix {
		index++
	}
	digits := asciiView(text[start:index])
	if digits == "" {
		return nan()
	}
	var magnitude float64
	if radix == 10 {
		magnitude, _ = strconv.ParseFloat(digits, 64)
	} else {
		integer, _ := new(big.Int).SetString(digits, radix)
		magnitude, _ = new(big.Float).SetInt(integer).Float64()
	}
	return sign * magnitude
}

// parseFloat is the global `parseFloat(text)`: the longest prefix that is a decimal literal, NaN when
// there is none. The tags hand it only signs, digits, dots and exponents.
func parseFloat(text []uint16) float64 {
	index := 0
	for index < len(text) && isJavaScriptWhitespace(text[index]) {
		index++
	}
	start := index
	if index < len(text) && (text[index] == '+' || text[index] == '-') {
		index++
	}
	if hasPrefixASCII(text[index:], "Infinity") {
		if text[start] == '-' {
			return negativeInfinity()
		}
		return positiveInfinity()
	}
	digits := 0
	for index < len(text) && text[index] >= '0' && text[index] <= '9' {
		index++
		digits++
	}
	if index < len(text) && text[index] == '.' {
		index++
		for index < len(text) && text[index] >= '0' && text[index] <= '9' {
			index++
			digits++
		}
	}
	if digits == 0 {
		return nan()
	}
	end := index
	if index < len(text) && (text[index] == 'e' || text[index] == 'E') {
		index++
		if index < len(text) && (text[index] == '+' || text[index] == '-') {
			index++
		}
		exponentStart := index
		for index < len(text) && text[index] >= '0' && text[index] <= '9' {
			index++
		}
		if index > exponentStart {
			end = index
		}
	}
	// A huge exponent overflows to ±Infinity and a tiny one underflows to ±0, as upstream's does;
	// strconv's range error is the same value.
	value, _ := strconv.ParseFloat(asciiView(text[start:end]), 64)
	return value
}

func hasPrefixASCII(text []uint16, prefix string) bool {
	return len(text) >= len(prefix) && equalsASCII(text[:len(prefix)], prefix)
}

// number is the global `Number(text)` for a string of digits with perhaps a dot: the empty string is 0.
func number(text []uint16) float64 {
	trimmed := trim(text)
	if len(trimmed) == 0 {
		return 0
	}
	value, err := strconv.ParseFloat(asciiView(trimmed), 64)
	if err != nil && value == 0 {
		return nan()
	}
	return value
}
