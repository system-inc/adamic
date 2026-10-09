// Package literal answers questions about the text a string literal holds, where the bytes on disk
// and the value the program receives are two different strings.
//
// Lifted out of `internal/rules/core/no_misleading_character_class.go`, where three rules reached
// past a package boundary for a private helper. Its second caller probed it against its own corpus
// before building on it and found a live defect, which is the argument for it having a home: a
// helper two rules deep in another rule's file is one nobody thinks to test independently.
package literal

import (
	"strings"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
)

// CookedToRaw maps each byte offset in a cooked string onto its offset in the raw source.
//
// The returned slice has one more entry than the cooked string is long, so the end offset of the
// last character is addressable. A nil return means the two texts could not be walked together,
// which happens when the cooked value is not what this scanner would produce from the raw one.
//
// Only the escapes that can appear in the patterns this rule reads are decoded. Anything else falls
// through to the single-character case, and a mismatch there returns nil rather than a wrong answer.
func CookedToRaw(raw string, cooked string) []int {
	offsets := make([]int, 0, len(cooked)+1)
	rawIndex, cookedIndex := 0, 0

	for rawIndex < len(raw) && cookedIndex < len(cooked) {
		var width int
		switch {
		case raw[rawIndex] == '\\' && rawIndex+1 < len(raw):
			width = escapeWidthInStringLiteral(raw, rawIndex)
		default:
			// A raw character carries its own UTF-8 width. Taking one byte here and comparing it
			// against a multi-byte cooked rune is what made every mapping fail: the two texts agree
			// and the comparison did not.
			_, runeWidth := utf8.DecodeRuneInString(raw[rawIndex:])
			if runeWidth == 0 {
				return nil
			}
			width = runeWidth
		}
		// Every cooked byte this raw token produced points back at the token's first byte, so a
		// span starting inside a decoded escape starts at the backslash.
		produced := cookedBytesProducedBy(raw[rawIndex:rawIndex+width], cooked[cookedIndex:])
		if produced == 0 {
			return nil
		}
		for count := 0; count < produced; count++ {
			offsets = append(offsets, rawIndex)
		}
		rawIndex += width
		cookedIndex += produced
	}
	if cookedIndex != len(cooked) {
		return nil
	}
	// Cooked finishing is necessary but not sufficient: the loop stops as soon as EITHER text runs
	// out, so a raw tail that produced no cooked characters at all is never compared against
	// anything and slips through. `'a  b\<newline>'` cooks to `a  b`, and checking only the cooked
	// side returns a complete, well formed mapping whose every offset past the run is wrong.
	//
	// Leftover raw is not wrong on its own, which is the trap and the reason this is not simply
	// `rawIndex != len(raw)`. A surrogate pair spends twelve raw bytes on one four byte rune, so
	// `[\uD83D\uDC4D]` legitimately exhausts cooked with six raw bytes still to go, and the strict
	// form rejects the case this rule exists for. What separates the two is whether the tail would
	// have produced anything: a low surrogate and a closing bracket would, a lone backslash or a
	// line continuation would not.
	//
	// Found by a mutation sweep in no-regex-spaces, which adopted this mapper and probed it against
	// its own corpus before building on it. The first form of this fix was the strict one and it
	// broke two of this rule's own fixtures, which is how the distinction got measured.
	if rawIndex < len(raw) && producesNoCookedBytes(raw[rawIndex:]) {
		return nil
	}
	offsets = append(offsets, rawIndex)
	return offsets
}

// producesNoCookedBytes reports whether a raw tail contributes nothing to the cooked string.
//
// Only two shapes do: a line continuation, which is defined to produce nothing, and a trailing lone
// backslash, which has nothing to escape. Anything else is either an ordinary character or an escape
// that decodes to one, so a tail containing any of it means the two texts genuinely disagree about
// length rather than the walk having stopped early.
func producesNoCookedBytes(tail string) bool {
	for index := 0; index < len(tail); {
		if tail[index] != '\\' {
			return false
		}
		if index+1 >= len(tail) {
			// A trailing backslash with nothing after it escapes nothing.
			return true
		}
		if tail[index+1] != '\n' && tail[index+1] != '\r' {
			return false
		}
		index += 2
		// A CRLF continuation spends one more byte.
		if tail[index-1] == '\r' && index < len(tail) && tail[index] == '\n' {
			index++
		}
	}
	return true
}

// escapeWidthInStringLiteral returns how many raw bytes a backslash escape occupies.
func escapeWidthInStringLiteral(raw string, index int) int {
	switch raw[index+1] {
	case 'u':
		if index+2 < len(raw) && raw[index+2] == '{' {
			closing := strings.IndexByte(raw[index+3:], '}')
			if closing >= 0 {
				return 3 + closing + 1
			}
			return 2
		}
		if index+5 < len(raw) && regexsyntax.AllHexDigits(raw[index+2:index+6]) {
			return 6
		}
		return 2
	case 'x':
		if index+3 < len(raw) && regexsyntax.IsHexDigit(raw[index+2]) &&
			regexsyntax.IsHexDigit(raw[index+3]) {
			return 4
		}
		return 2
	case '0', '1', '2', '3', '4', '5', '6', '7':
		// A legacy octal escape, which sloppy-mode source may still carry and which upstream's
		// corpus uses deliberately: `"[ \\ufe\60f]"` spells the `0` of a variation selector as
		// `\60` so that the escape only appears after the string is cooked. Up to three digits, and
		// a leading digit above three caps the run at two.
		width := 2
		limit := 3
		if raw[index+1] > '3' {
			limit = 2
		}
		for width < limit+1 && index+width < len(raw) && raw[index+width] >= '0' &&
			raw[index+width] <= '7' {
			width++
		}
		return width
	}
	// A backslash before anything else escapes that one character, and the character may be several
	// bytes. `\👍` is five raw bytes producing four cooked ones, and returning two here desynced the
	// walk on every case upstream wrote to exercise exactly this.
	_, runeWidth := utf8.DecodeRuneInString(raw[index+1:])
	if runeWidth == 0 {
		return 2
	}
	return 1 + runeWidth
}

// cookedBytesProducedBy returns how many bytes of the remaining cooked text one raw token produced.
//
// The comparison is by value rather than by rule: whatever the token decodes to must be a prefix of
// what is left, and if it is not, the walk has lost sync and the caller declines. Deciding it this
// way rather than by re-deriving each escape's meaning keeps this from becoming a second, subtly
// different string-literal decoder living next to the parser's.
func cookedBytesProducedBy(rawToken string, remainingCooked string) int {
	if rawToken == "" || remainingCooked == "" {
		return 0
	}
	if rawToken[0] != '\\' {
		_, width := utf8.DecodeRuneInString(remainingCooked)
		if width == 0 || width > len(rawToken) || rawToken[:width] != remainingCooked[:width] {
			return 0
		}
		return width
	}
	// A decoded escape produces one rune, whose encoded width is what it contributed.
	_, width := utf8.DecodeRuneInString(remainingCooked)
	if width == 0 {
		return 0
	}
	return width
}
