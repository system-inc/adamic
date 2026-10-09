package regexsyntax

// Hex digits, which the class parser reaches for in three places: `\xNN`, `\uNNNN`, and the
// brace form `\u{NNNNN}`. They are here rather than in a shared string package because the class
// parser is their only caller, and a helper with one caller in the package that uses it is easier
// to reason about than the same helper one import away.

// IsHexDigit reports whether b is a hexadecimal digit in either case.
func IsHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

// AllHexDigits reports whether s is non-empty and every byte in it is a hex digit.
//
// Empty is false rather than vacuously true on purpose: the callers are asking whether an escape
// has digits after it, and `\x` with nothing following is not a valid escape.
func AllHexDigits(s string) bool {
	if s == "" {
		return false
	}
	for index := range len(s) {
		if !IsHexDigit(s[index]) {
			return false
		}
	}
	return true
}

// ParseHexUint reads s as a hexadecimal number.
//
// A non-hex byte contributes zero rather than failing, because every caller checks AllHexDigits
// first. The pairing is the contract: ask whether it is hex, then read it.
func ParseHexUint(s string) uint32 {
	value := uint32(0)
	for index := range len(s) {
		value = (value << 4) | hexValue(s[index])
	}
	return value
}

func hexValue(b byte) uint32 {
	switch {
	case b >= '0' && b <= '9':
		return uint32(b - '0')
	case b >= 'a' && b <= 'f':
		return uint32(b-'a') + 10
	case b >= 'A' && b <= 'F':
		return uint32(b-'A') + 10
	}
	return 0
}
