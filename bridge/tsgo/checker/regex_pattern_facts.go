// Raw grammar copied from the pinned cohere regexsyntax and regexpattern shelves.
// Returns scanner structure only; native Adamic owns lint judgments and repairs.
package checker

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// ---------------------------------------------------------------------------
// Public types
// ---------------------------------------------------------------------------

// RegexFlags captures the subset of ECMAScript regex flags that affects how
// patterns (and in particular character classes) are parsed.
type regexPatternRegexFlags struct {
	Unicode     bool // u flag
	UnicodeSets bool // v flag
}

// UV reports whether unicode/unicodeSets mode is active (either u or v).
func (f regexPatternRegexFlags) UV() bool { return f.Unicode || f.UnicodeSets }

// ParseRegexFlags returns a RegexFlags from a flag string (e.g. "gui").
func regexPatternParseRegexFlags(flags string) regexPatternRegexFlags {
	return regexPatternRegexFlags{
		Unicode:     strings.ContainsRune(flags, 'u'),
		UnicodeSets: strings.ContainsRune(flags, 'v'),
	}
}

// RegexCharElementKind classifies one element of a character class body.
type regexPatternRegexCharElementKind int

const (
	// RegexCharSingle is a single character element (literal, escape, …).
	regexPatternRegexCharSingle regexPatternRegexCharElementKind = iota
	// RegexCharRange is a character range `a-b` (inclusive on both ends).
	// Value is the `a` side; Max is the `b` side. Only the min/max endpoints
	// appear in the resulting character sequence , the inner characters are
	// implicit.
	regexPatternRegexCharRange
	// RegexCharBreaker is any element that interrupts a contiguous character
	// sequence within a class: `\d`, `\D`, `\w`, `\W`, `\s`, `\S`, `\b`, `\B`,
	// `\p{...}`, `\P{...}`, `\q{...}` (v-flag), nested `[...]` (v-flag), or a
	// v-flag set operator `--` / `&&`.
	regexPatternRegexCharBreaker
)

// RegexCharElement is one element of a parsed character class body.
//
// For RegexCharSingle:
//   - Value holds the element's effective value (UTF-16 code unit or the
//     astral code point when combined via `\uHHHH\uHHHH` under u/v).
//   - IsUBrace is true iff the element was written as `\u{H...}`.
//   - IsLoneSurrogate is true iff the element is a lone surrogate from a raw
//     astral character under non-u mode (where regexpp sees it as two units).
//
// For RegexCharRange:
//   - Value / IsUBrace describe the `a` endpoint.
//   - Max / MaxIsUBrace describe the `b` endpoint.
//
// For RegexCharBreaker:
//   - Value, Max, IsUBrace are unused.
//
// Start / End are byte offsets within the pattern text covering the element's
// source extent.
type regexPatternRegexCharElement struct {
	Kind            regexPatternRegexCharElementKind
	Value           uint32
	IsUBrace        bool
	IsLoneSurrogate bool

	Max         uint32
	MaxIsUBrace bool

	Start int
	End   int

	// RawStart/RawEnd optionally narrow to just the element's own source
	// extent, distinct from Start/End which always span the element. For
	// ranges this pair equals Start/End.
}

// ---------------------------------------------------------------------------
// Pattern scanner (layer 1)
// ---------------------------------------------------------------------------

// IterateRegexCharacterClasses walks a regex pattern and invokes cb once per
// top-level character class (including any v-flag nested classes, at each
// nesting level). cb receives the byte range [start, end) covering `[`..`]`.
//
// Returns false if the pattern is malformed (unterminated class, unterminated
// escape at EOF). When false, cb may have been invoked for classes encountered
// before the error.
func regexPatternIterateRegexCharacterClasses(pattern string, flags regexPatternRegexFlags, cb func(start, end int)) bool {
	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '\\':
			step, ok := regexPatternSkipPatternEscape(pattern, i, flags)
			if !ok {
				return false
			}
			i += step
		case '[':
			end, ok := regexPatternIterateClassFromLBracket(pattern, i, flags, cb)
			if !ok {
				return false
			}
			i = end
		case '(':
			// Parens don't affect character-class scanning; consume and continue.
			i++
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	return true
}

// iterateClassFromLBracket recursively scans a character class starting at
// pattern[start] (which must be `[`). Invokes cb for the class (and, under v
// flag, for any nested classes). Returns the index just past `]`, or ok=false
// on malformed input.
func regexPatternIterateClassFromLBracket(pattern string, start int, flags regexPatternRegexFlags, cb func(start, end int)) (int, bool) {
	end, ok := regexPatternClassEnd(pattern, start, flags)
	if !ok {
		return start, false
	}
	// Recurse into nested classes first (under v flag), then invoke cb on
	// the outermost.
	if flags.UnicodeSets {
		i := start + 1
		if i < end-1 && pattern[i] == '^' {
			i++
		}
		for i < end-1 {
			c := pattern[i]
			if c == '\\' {
				step, ok := regexPatternSkipPatternEscape(pattern, i, flags)
				if !ok {
					return start, false
				}
				i += step
				continue
			}
			if c == '[' {
				nestedEnd, ok := regexPatternIterateClassFromLBracket(pattern, i, flags, cb)
				if !ok {
					return start, false
				}
				i = nestedEnd
				continue
			}
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	cb(start, end)
	return end, true
}

// ClassEnd returns the byte index just past the matching `]` for a class
// starting at `[` at pattern[start]. Handles escaped `]`, v-flag nested
// classes, and `\q{...}` which contains a literal `]` inside braces.
func regexPatternClassEnd(pattern string, start int, flags regexPatternRegexFlags) (int, bool) {
	if start >= len(pattern) || pattern[start] != '[' {
		return start, false
	}
	i := start + 1
	// A leading `^` is not an element by itself.
	if i < len(pattern) && pattern[i] == '^' {
		i++
	}
	depth := 1
	for i < len(pattern) {
		c := pattern[i]
		switch {
		case c == '\\':
			step, ok := regexPatternSkipPatternEscape(pattern, i, flags)
			if !ok {
				return start, false
			}
			i += step
		case c == '[' && flags.UnicodeSets:
			depth++
			i++
		case c == ']':
			depth--
			i++
			if depth == 0 {
				return i, true
			}
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	return start, false
}

// SkipPatternEscape returns how many bytes a `\`-prefixed escape consumes at
// pattern[i] (including the leading `\`) for the purposes of class-boundary
// scanning. Returns ok=false at EOF on `\`.
func regexPatternSkipPatternEscape(pattern string, i int, flags regexPatternRegexFlags) (int, bool) {
	if i+1 >= len(pattern) {
		return 0, false
	}
	next := pattern[i+1]
	switch next {
	case 'x':
		if i+3 < len(pattern) && regexPatternIsHexDigit(pattern[i+2]) && regexPatternIsHexDigit(pattern[i+3]) {
			return 4, true
		}
		return 2, true
	case 'u':
		if i+2 < len(pattern) && pattern[i+2] == '{' {
			if !flags.UV() {
				return 2, true
			}
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel < 0 {
				return 2, true // best-effort recover
			}
			return 3 + closeRel + 1, true
		}
		if i+5 < len(pattern) && regexPatternAllHexDigits(pattern[i+2:i+6]) {
			return 6, true
		}
		return 2, true
	case 'c':
		if i+2 < len(pattern) {
			return 3, true
		}
		return 2, true
	case 'p', 'P':
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel >= 0 {
				return 3 + closeRel + 1, true
			}
		}
		return 2, true
	case 'q':
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel >= 0 {
				return 3 + closeRel + 1, true
			}
		}
		return 2, true
	}
	// Generic two-byte escape (covers identity, \d, \w, ...).
	// If the next is a multi-byte UTF-8 rune, consume its full width.
	_, w := utf8.DecodeRuneInString(pattern[i+1:])
	if w == 0 {
		return 2, true
	}
	return 1 + w, true
}

// ---------------------------------------------------------------------------
// Character class body parser (layer 2)
// ---------------------------------------------------------------------------

// ParseRegexCharacterClass parses a character class starting at pattern[start]
// (which must be `[`). It returns the flat element list (ranges expanded into
// min/max; nested classes and set operators emitted as breakers at the
// position where they appear) and the byte index just past the closing `]`.
//
// The input range [start, end) is trusted to be well-formed (e.g. from a
// prior call to IterateRegexCharacterClasses).
//
// Note: nested v-flag classes themselves are not recursed into by this
// function , they appear only as RegexCharBreaker. Callers that want to scan
// nested classes should iterate them via IterateRegexCharacterClasses and
// call ParseRegexCharacterClass on each separately.
func regexPatternParseRegexCharacterClass(pattern string, start int, flags regexPatternRegexFlags) ([]regexPatternRegexCharElement, int, bool) {
	return regexPatternParseRegexCharacterClassInternal(pattern, start, flags, 0)
}

// ParseRegexCharacterClassWithEnd is ParseRegexCharacterClass with a
// previously discovered class end. Non-v callers use the known byte span as
// an element-capacity hint to avoid repeated slice growth for large classes.
// Unicode-set classes deliberately skip that preallocation:
// each nested callback receives the outer span, whose byte length can be much
// larger than its own flat element count. The cap also bounds over-allocation
// for escape-heavy and multibyte input, where bytes can outnumber elements.
func regexPatternParseRegexCharacterClassWithEnd(pattern string, start, end int, flags regexPatternRegexFlags) ([]regexPatternRegexCharElement, int, bool) {
	if start < 0 || end <= start || end > len(pattern) || pattern[end-1] != ']' {
		return nil, start, false
	}
	elementCapacity := 0
	if !flags.UnicodeSets {
		elementCapacity = end - start - 2
		if elementCapacity < 0 {
			elementCapacity = 0
		} else if elementCapacity > regexPatternMaxRegexCharacterClassPreallocation {
			elementCapacity = regexPatternMaxRegexCharacterClassPreallocation
		}
	}
	elements, parsedEnd, ok := regexPatternParseRegexCharacterClassInternal(pattern, start, flags, elementCapacity)
	if !ok || parsedEnd != end {
		return nil, start, false
	}
	if len(elements) == 0 {
		elements = nil
	}
	return elements, parsedEnd, true
}

const regexPatternMaxRegexCharacterClassPreallocation = 16 << 10

func regexPatternParseRegexCharacterClassInternal(pattern string, start int, flags regexPatternRegexFlags, elementCapacity int) ([]regexPatternRegexCharElement, int, bool) {
	if start >= len(pattern) || pattern[start] != '[' {
		return nil, start, false
	}
	i := start + 1
	if i < len(pattern) && pattern[i] == '^' {
		i++
	}

	var out []regexPatternRegexCharElement
	if elementCapacity > 0 {
		out = make([]regexPatternRegexCharElement, 0, elementCapacity)
	}
	pendingIdx := -1 // index into `out` of the last RegexCharSingle (candidate for range start)

	for i < len(pattern) {
		c := pattern[i]
		switch {
		case c == ']':
			return out, i + 1, true

		case c == '\\':
			el, step, ok := regexPatternReadClassEscape(pattern, i, flags)
			if !ok {
				return nil, start, false
			}
			if el.Kind == regexPatternRegexCharBreaker {
				out = append(out, el)
				pendingIdx = -1
			} else {
				// RegexCharSingle (possibly with companion astral surrogate entry below)
				out = append(out, el)
				pendingIdx = len(out) - 1
				// Under non-u/v mode, a raw astral character or a `\u{H}` with H>0xFFFF
				// should emit TWO units. But class escapes only produce 1 unit per call;
				// astral raw chars are handled in the `default` case below, and `\u{H}`
				// under u/v produces the combined code point (one element). Under non-u,
				// `\u{H}` is not recognized at all (treated as identity `u`).
			}
			i += step

		case c == '[' && flags.UnicodeSets:
			// v-flag nested class , emit as breaker; caller recurses via
			// IterateRegexCharacterClasses to parse nested contents.
			nestedEnd, ok := regexPatternClassEnd(pattern, i, flags)
			if !ok {
				return nil, start, false
			}
			out = append(out, regexPatternRegexCharElement{Kind: regexPatternRegexCharBreaker, Start: i, End: nestedEnd})
			i = nestedEnd
			pendingIdx = -1

		case (c == '-' || c == '&') && flags.UnicodeSets && i+1 < len(pattern) && pattern[i+1] == c:
			// v-flag set operator `--` or `&&` , breaker
			out = append(out, regexPatternRegexCharElement{Kind: regexPatternRegexCharBreaker, Start: i, End: i + 2})
			i += 2
			pendingIdx = -1

		case c == '-' && pendingIdx >= 0 && i+1 < len(pattern) && pattern[i+1] != ']':
			// Range: combine pending with next element.
			i++
			var nextEl regexPatternRegexCharElement
			var step int
			var ok bool
			if pattern[i] == '\\' {
				nextEl, step, ok = regexPatternReadClassEscape(pattern, i, flags)
				if !ok {
					return nil, start, false
				}
			} else {
				nextEl, step, ok = regexPatternReadRawClassChar(pattern, i, flags)
				if !ok {
					return nil, start, false
				}
			}
			if nextEl.Kind != regexPatternRegexCharSingle {
				// Under u flag, `set-to-X` or `X-to-set` is a syntax error;
				// ESLint's regexpp would throw. We treat this as malformed
				// and abort class parsing (caller will skip the regex).
				return nil, start, false
			}
			minEl := out[pendingIdx]
			rangeEl := regexPatternRegexCharElement{
				Kind:        regexPatternRegexCharRange,
				Value:       minEl.Value,
				IsUBrace:    minEl.IsUBrace,
				Max:         nextEl.Value,
				MaxIsUBrace: nextEl.IsUBrace,
				Start:       minEl.Start,
				End:         nextEl.End,
			}
			out[pendingIdx] = rangeEl
			// If the min endpoint came from a raw astral in non-u mode
			// (which would logically expand to a surrogate pair), the range
			// representation is already degenerate; callers in this project
			// only inspect endpoints, so we don't split here.
			i += step
			pendingIdx = -1

		default:
			el, step, ok := regexPatternReadRawClassChar(pattern, i, flags)
			if !ok {
				return nil, start, false
			}
			if el.Kind == regexPatternRegexCharSingle {
				out = append(out, el)
				pendingIdx = len(out) - 1
				// Note: we do NOT split raw astrals into surrogate pairs here
				// , callers (e.g. the misleading-character-class rule) decide
				// per-flag whether to expand them, based on the rule's own
				// semantic needs. See IsLoneSurrogate on the element.
			}
			i += step
		}
	}
	return nil, start, false // unterminated
}

// readClassEscape parses a `\`-prefixed token inside a character class. It
// emits one RegexCharElement: either RegexCharSingle or RegexCharBreaker.
func regexPatternReadClassEscape(pattern string, i int, flags regexPatternRegexFlags) (regexPatternRegexCharElement, int, bool) {
	if i+1 >= len(pattern) {
		return regexPatternRegexCharElement{}, 0, false
	}
	next := pattern[i+1]
	switch next {
	case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B':
		// Note: inside a class, `\b` means U+0008 (backspace), not word
		// boundary. ESLint treats it as a single character. `\B` inside a
		// class is a syntax error under u mode; under non-u it's identity.
		// For our purposes, we treat `\b` as a RegexCharSingle with value 8
		// (backspace) and `\B` ... hmm.
		//
		// Actually: regexpp treats `\b` in class as Character(0x08) and `\B`
		// in class as a syntax error in u mode but identity in non-u.
		// `\d`/`\D`/`\w`/`\W`/`\s`/`\S` are CharacterSet (breaker).
		if next == 'b' {
			return regexPatternRegexCharElement{
				Kind: regexPatternRegexCharSingle, Value: 0x08, Start: i, End: i + 2,
			}, 2, true
		}
		if next == 'B' {
			if flags.UV() {
				return regexPatternRegexCharElement{}, 0, false // u mode syntax error
			}
			// Under non-u, identity escape.
			return regexPatternRegexCharElement{
				Kind: regexPatternRegexCharSingle, Value: 'B', Start: i, End: i + 2,
			}, 2, true
		}
		return regexPatternRegexCharElement{Kind: regexPatternRegexCharBreaker, Start: i, End: i + 2}, 2, true
	case 'p', 'P':
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel < 0 {
				return regexPatternRegexCharElement{}, 0, false
			}
			end := i + 3 + closeRel + 1
			return regexPatternRegexCharElement{Kind: regexPatternRegexCharBreaker, Start: i, End: end}, end - i, true
		}
		// Non-u mode: identity escape of `p`/`P`.
		return regexPatternRegexCharElement{
			Kind: regexPatternRegexCharSingle, Value: uint32(next), Start: i, End: i + 2,
		}, 2, true
	case 'q':
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel < 0 {
				return regexPatternRegexCharElement{}, 0, false
			}
			end := i + 3 + closeRel + 1
			return regexPatternRegexCharElement{Kind: regexPatternRegexCharBreaker, Start: i, End: end}, end - i, true
		}
		return regexPatternRegexCharElement{
			Kind: regexPatternRegexCharSingle, Value: 'q', Start: i, End: i + 2,
		}, 2, true
	case 'x':
		if i+3 < len(pattern) && regexPatternIsHexDigit(pattern[i+2]) && regexPatternIsHexDigit(pattern[i+3]) {
			v := regexPatternParseHexUint(pattern[i+2 : i+4])
			return regexPatternRegexCharElement{
				Kind: regexPatternRegexCharSingle, Value: v, Start: i, End: i + 4,
			}, 4, true
		}
		return regexPatternRegexCharElement{
			Kind: regexPatternRegexCharSingle, Value: 'x', Start: i, End: i + 2,
		}, 2, true
	case 'u':
		if i+2 < len(pattern) && pattern[i+2] == '{' {
			if !flags.UV() {
				// Non-u mode: treat `\u` as identity `u`.
				return regexPatternRegexCharElement{
					Kind: regexPatternRegexCharSingle, Value: 'u', Start: i, End: i + 2,
				}, 2, true
			}
			closeRel := strings.IndexByte(pattern[i+3:], '}')
			if closeRel < 0 {
				return regexPatternRegexCharElement{}, 0, false
			}
			hex := pattern[i+3 : i+3+closeRel]
			end := i + 3 + closeRel + 1
			if hex == "" || !regexPatternAllHexDigits(hex) {
				return regexPatternRegexCharElement{}, 0, false
			}
			return regexPatternRegexCharElement{
				Kind: regexPatternRegexCharSingle, Value: regexPatternParseHexUint(hex), IsUBrace: true,
				Start: i, End: end,
			}, end - i, true
		}
		if i+5 < len(pattern) && regexPatternAllHexDigits(pattern[i+2:i+6]) {
			hi := regexPatternParseHexUint(pattern[i+2 : i+6])
			// Surrogate pair `\uHHHH\uHHHH` under u/v collapses to one element
			// with an astral value.
			if flags.UV() && hi >= 0xD800 && hi <= 0xDBFF && i+11 < len(pattern) &&
				pattern[i+6] == '\\' && pattern[i+7] == 'u' && regexPatternAllHexDigits(pattern[i+8:i+12]) {
				lo := regexPatternParseHexUint(pattern[i+8 : i+12])
				if lo >= 0xDC00 && lo <= 0xDFFF {
					cp := 0x10000 + (hi-0xD800)*0x400 + (lo - 0xDC00)
					return regexPatternRegexCharElement{
						Kind: regexPatternRegexCharSingle, Value: cp, Start: i, End: i + 12,
					}, 12, true
				}
			}
			return regexPatternRegexCharElement{
				Kind: regexPatternRegexCharSingle, Value: hi, Start: i, End: i + 6,
			}, 6, true
		}
		return regexPatternRegexCharElement{
			Kind: regexPatternRegexCharSingle, Value: 'u', Start: i, End: i + 2,
		}, 2, true
	case 'n':
		return regexPatternRegexCharElement{Kind: regexPatternRegexCharSingle, Value: '\n', Start: i, End: i + 2}, 2, true
	case 't':
		return regexPatternRegexCharElement{Kind: regexPatternRegexCharSingle, Value: '\t', Start: i, End: i + 2}, 2, true
	case 'r':
		return regexPatternRegexCharElement{Kind: regexPatternRegexCharSingle, Value: '\r', Start: i, End: i + 2}, 2, true
	case 'v':
		return regexPatternRegexCharElement{Kind: regexPatternRegexCharSingle, Value: '\v', Start: i, End: i + 2}, 2, true
	case 'f':
		return regexPatternRegexCharElement{Kind: regexPatternRegexCharSingle, Value: '\f', Start: i, End: i + 2}, 2, true
	case '0':
		// In u/v mode `\0` is always NUL (no octal extension). In non-u
		// mode, if followed by a digit it's a legacy octal , but legacy
		// octals aren't meaningful for this project's detectors, so treat
		// as NUL for simplicity.
		return regexPatternRegexCharElement{Kind: regexPatternRegexCharSingle, Value: 0, Start: i, End: i + 2}, 2, true
	case 'c':
		if i+2 < len(pattern) {
			return regexPatternRegexCharElement{
				Kind: regexPatternRegexCharSingle, Value: uint32(pattern[i+2]) & 0x1F,
				Start: i, End: i + 3,
			}, 3, true
		}
		return regexPatternRegexCharElement{}, 0, false
	}
	// Identity escape , value is the (possibly multi-byte) character after `\`.
	r, w := utf8.DecodeRuneInString(pattern[i+1:])
	if w == 0 {
		return regexPatternRegexCharElement{}, 0, false
	}
	// Under u/v mode, identity escape of a letter/digit is a syntax error.
	// We relax this: let caller see the element and skip the regex if it
	// becomes problematic. (Most production code won't hit this.)
	return regexPatternRegexCharElement{
		Kind: regexPatternRegexCharSingle, Value: uint32(r), Start: i, End: i + 1 + w,
	}, 1 + w, true
}

// readRawClassChar reads a non-`\`-prefixed character inside a class body
// (i.e. a raw character at pattern[i]). Raw astral in non-u/v mode emits
// the HIGH surrogate only , callers handle the low-surrogate companion by
// inspecting IsLoneSurrogate and stepping via a subsequent call.
//
// To keep the protocol simple, ParseRegexCharacterClass calls this once and
// trusts the single element returned; we do NOT split raw astrals here. The
// misleading-character-class rule accepts this because it gets raw astrals
// only in string-literal contexts, where the layer-3 parser already does the
// UTF-16 surrogate split.
func regexPatternReadRawClassChar(pattern string, i int, flags regexPatternRegexFlags) (regexPatternRegexCharElement, int, bool) {
	r, w := utf8.DecodeRuneInString(pattern[i:])
	if w == 0 {
		return regexPatternRegexCharElement{}, 0, false
	}
	return regexPatternRegexCharElement{
		Kind: regexPatternRegexCharSingle, Value: uint32(r), Start: i, End: i + w,
	}, w, true
}

// Hex digits, which the class parser reaches for in three places: `\xNN`, `\uNNNN`, and the
// brace form `\u{NNNNN}`. They are here rather than in a shared string package because the class
// parser is their only caller, and a helper with one caller in the package that uses it is easier
// to reason about than the same helper one import away.

// IsHexDigit reports whether b is a hexadecimal digit in either case.
func regexPatternIsHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

// AllHexDigits reports whether s is non-empty and every byte in it is a hex digit.
//
// Empty is false rather than vacuously true on purpose: the callers are asking whether an escape
// has digits after it, and `\x` with nothing following is not a valid escape.
func regexPatternAllHexDigits(s string) bool {
	if s == "" {
		return false
	}
	for index := range len(s) {
		if !regexPatternIsHexDigit(s[index]) {
			return false
		}
	}
	return true
}

// ParseHexUint reads s as a hexadecimal number.
//
// A non-hex byte contributes zero rather than failing, because every caller checks AllHexDigits
// first. The pairing is the contract: ask whether it is hex, then read it.
func regexPatternParseHexUint(s string) uint32 {
	value := uint32(0)
	for index := range len(s) {
		value = (value << 4) | regexPatternHexValue(s[index])
	}
	return value
}

func regexPatternHexValue(b byte) uint32 {
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

// CharacterKind is how a character was written, which is a different question from what it matches.
//
// The distinction is the whole of `no-control-regex`: `\0` and `\u0000` are both U+0000 and only
// the second is reported, because the rule objects to spelling a control code out rather than to
// the code point being present. A walk reporting only values cannot express that.
type regexPatternCharacterKind int

const (
	// KindSymbol is a character written as itself, including a raw control byte.
	regexPatternKindSymbol regexPatternCharacterKind = iota
	// KindSingleEscape is a named escape: `\n`, `\t`, `\r`, `\v`, `\f`, `\b`.
	regexPatternKindSingleEscape
	// KindNull is `\0` with no digit following it.
	regexPatternKindNull
	// KindHexadecimalEscape is `\xHH`.
	regexPatternKindHexadecimalEscape
	// KindUnicodeEscape is `\uHHHH` or `\u{H...}`.
	regexPatternKindUnicodeEscape
	// KindControlLetter is `\cX`.
	regexPatternKindControlLetter
	// KindOctal is a legacy octal escape, `\1` through `\377`. It is one kind here rather than
	// the three upstream carries, because the only caller asks whether it is octal at all.
	regexPatternKindOctal
	// KindIdentityEscape is a backslash before a character that needs no escaping, like `\-`.
	regexPatternKindIdentityEscape
)

// Character is one matched character and where it sits.
//
// Start and End are byte offsets into the pattern text, not into the file, so a caller reporting a
// finding adds the pattern's own offset. Keeping them pattern-relative is what lets the same walk
// serve a regex literal and a string passed to the RegExp constructor, which sit at different
// places in a file and have different quoting.
type regexPatternCharacter struct {
	Value uint32
	Kind  regexPatternCharacterKind
	Start int
	End   int

	// QuantifierDepth and ClassDepth are the enclosing structure, counted separately because the
	// one caller that reads them treats them the same and a future one might not.
	//
	// `no-regex-spaces` ignores any character with either depth nonzero: `/  +/` and `/[  ]/` both
	// hold two spaces and neither is what the rule is about, since a quantifier or a class changes
	// what the repetition means.
	QuantifierDepth int
	ClassDepth      int
}

// Walk reports every character in a pattern, in source order, with its enclosing structure.
//
// The callback returning false stops the walk, which is what lets a caller counting to a decision
// avoid scanning the rest of a long pattern.
//
// A pattern the scanner cannot make sense of stops the walk and returns false. That is the same
// contract regexsyntax uses and it exists for the same reason: an unterminated construct is a
// syntax error the parser has already refused, so a second opinion here would be noise on a file
// that does not compile.
//
// # What this does NOT report, measured
//
// Four limits, each found by a caller that needed the missing thing. They are stated here rather
// than left in the rule that discovered them, because the next caller will otherwise measure them a
// fifth time:
//
//   - **A set escape never arrives.** `\d`, `\D`, `\s`, `\S`, `\w`, `\W`, `\p`, `\P`, `\k`, `\q`, and
//     `\b`/`\B` outside a class all match a set or a position rather than a single character, so
//     they have no value to report and are skipped. `no-useless-escape` has to answer about every
//     one of them.
//   - **WHICH class a character sits in is not reported.** `ClassDepth` is a counter, not a stack,
//     so a caller needing the enclosing class's own start and end cannot use this. The caret rule
//     and the hyphen rule both need exactly that: a caret means negation only as a class's first
//     character, and a hyphen opens a range only away from a class's edges.
//   - **A `v`-flag nested class is not recursed into.** `walkClass` is not re-entered, so
//     `/[[\.&]--[\.&]]/v` walks to silence where upstream reports twice. Measured directly rather
//     than read off the source.
//   - **A range's interior is not reported**, only its written endpoints. That is deliberate and
//     stated on `walkClass`: the characters between the endpoints were never written down.
//
// `no-useless-escape` needed the first three and was built on `regexsyntax.SkipPatternEscape` and
// `regexsyntax.ClassEnd` instead, which is the right call for a caller needing class structure.
// Extending this to serve it would change the contract for the two callers below, which is a larger
// change than one rule justifies.
//
// So the note at the top of this file stands with a correction: a third caller wanting more
// structure should reach for the layer underneath rather than extend this, and only a caller wanting
// *characters plus depth* belongs here.
func regexPatternWalk(pattern string, flags regexPatternRegexFlags, callback func(regexPatternCharacter) bool) bool {
	regexPatternWalker := &regexPatternWalker{pattern: pattern, flags: flags, callback: callback}
	return regexPatternWalker.run()
}

// CountCapturingGroups returns how many capturing groups a pattern declares.
//
// `no-control-regex` needs this to decide whether `\1` is a backreference or a control character,
// and it needs the total before judging any of them: a group may be referenced before it is
// defined, so a running count would call the same escape a control character early in the pattern
// and a backreference later.
func regexPatternCountCapturingGroups(pattern string, flags regexPatternRegexFlags) int {
	count := 0
	for index := 0; index < len(pattern); {
		switch pattern[index] {
		case '\\':
			step, ok := regexPatternSkipPatternEscape(pattern, index, flags)
			if !ok {
				return count
			}
			index += step
		case '[':
			end, ok := regexPatternClassEnd(pattern, index, flags)
			if !ok {
				return count
			}
			index = end
		case '(':
			// `(?:`, `(?=`, `(?!`, `(?<=`, `(?<!` are non-capturing. `(?<name>` captures, and it is
			// the one `(?<` form that does, which is why the character after the angle bracket has
			// to be looked at rather than the group being dismissed on `(?` alone.
			if regexPatternIsCapturingGroupStart(pattern, index) {
				count++
			}
			index++
		default:
			index++
		}
	}
	return count
}

// isCapturingGroupStart reports whether the `(` at index opens a capturing group.
func regexPatternIsCapturingGroupStart(pattern string, index int) bool {
	if index+1 >= len(pattern) || pattern[index+1] != '?' {
		return true
	}
	// `(?<name>` is capturing; `(?<=` and `(?<!` are lookbehind and are not.
	if index+2 < len(pattern) && pattern[index+2] == '<' {
		if index+3 >= len(pattern) {
			return false
		}
		return pattern[index+3] != '=' && pattern[index+3] != '!'
	}
	return false
}

// singleEscapeValues are the characters a named escape stands for.
//
// The backspace escape only means backspace inside a character class; outside one the same two
// characters are a word boundary, which matches a position and never reaches here. That asymmetry
// lives in readEscape rather than in this table, so the table stays a plain statement of what each
// name denotes.
var regexPatternSingleEscapeValues = map[byte]uint32{
	'n': 0x0A,
	'r': 0x0D,
	't': 0x09,
	'v': 0x0B,
	'f': 0x0C,
	'b': 0x08,
}

// escapeValue returns the code point an escape denotes, or false when the escape has no single one.
//
// The kind is passed in rather than re-derived so the two cannot disagree. A value computed against
// one reading of the spelling while the caller holds another is the kind of divergence that
// produces a correct-looking answer about the wrong character.
func regexPatternEscapeValue(text string, kind regexPatternCharacterKind) (uint32, bool) {
	if len(text) < 2 || text[0] != '\\' {
		return 0, false
	}

	switch kind {
	case regexPatternKindSingleEscape:
		value, ok := regexPatternSingleEscapeValues[text[1]]
		return value, ok

	case regexPatternKindNull:
		return 0, true

	case regexPatternKindHexadecimalEscape:
		// Exactly two hex digits. Fewer is not a hex escape at all, and the scanner would not have
		// spelled it this way, so a short one is refused rather than padded.
		if len(text) != 4 || !regexPatternAllHexDigits(text[2:]) {
			return 0, false
		}
		return regexPatternParseHexUint(text[2:]), true

	case regexPatternKindUnicodeEscape:
		return regexPatternUnicodeEscapeValue(text)

	case regexPatternKindControlLetter:
		// The control escape maps a letter onto the control code in its low five bits, so the
		// letter's case does not matter and the arithmetic is the definition rather than a trick.
		if len(text) != 3 {
			return 0, false
		}
		letter := text[2]
		if !regexPatternIsAsciiLetter(letter) {
			return 0, false
		}
		return uint32(letter % 32), true

	case regexPatternKindOctal:
		return regexPatternOctalEscapeValue(text[1:])

	case regexPatternKindIdentityEscape:
		// A backslash before a character that needs no escaping denotes that character. This is
		// what makes an escaped space a space, which matters to a caller counting spaces: it is
		// still one, and it is still written down.
		value, width := regexPatternDecodeRune(text, 1)
		if width == 0 {
			return 0, false
		}
		return value, true
	}

	return 0, false
}

// unicodeEscapeValue reads the fixed-width and braced forms of a unicode escape.
//
// The braced form is only valid under the u or v flag, and that check belongs to the scanner that
// decided the spelling rather than here: by the time a kind says unicode escape, the flags have
// already been consulted. Re-checking them at this layer would put the same decision in two places.
func regexPatternUnicodeEscapeValue(text string) (uint32, bool) {
	if len(text) < 3 {
		return 0, false
	}

	if text[2] == '{' {
		if text[len(text)-1] != '}' {
			return 0, false
		}
		digits := text[3 : len(text)-1]
		if digits == "" || !regexPatternAllHexDigits(digits) {
			return 0, false
		}
		return regexPatternParseHexUint(digits), true
	}

	if len(text) != 6 || !regexPatternAllHexDigits(text[2:]) {
		return 0, false
	}
	return regexPatternParseHexUint(text[2:]), true
}

// octalEscapeValue reads a legacy octal escape's digits.
//
// A leading digit of 8 or 9 is not octal at all, which is what separates this from the legacy
// decimal escapes another rule reports. Values above 255 are refused rather than truncated, since
// the escape does not denote a character in that case.
func regexPatternOctalEscapeValue(digits string) (uint32, bool) {
	if digits == "" {
		return 0, false
	}
	var value uint32
	for index := 0; index < len(digits); index++ {
		digit := digits[index]
		if digit < '0' || digit > '7' {
			return 0, false
		}
		value = value*8 + uint32(digit-'0')
	}
	if value > 0xFF {
		return 0, false
	}
	return value, true
}

// extendOctalEscape widens a two-byte digit escape to cover the octal digits that follow it.
//
// The language allows up to three octal digits, and stops early at a digit outside the octal range
// or at a value above 255: `\400` is `\40` followed by a literal zero, because 0400 does not name a
// character. Reproducing that rather than taking three digits blindly is what keeps the reported
// span equal to the text the escape actually covers.
//
// A backreference is left at one digit deliberately. `\1` in a pattern with capturing groups is a
// reference rather than a character, and the caller that cares decides which by counting groups; a
// span greedily widened here would take digits belonging to the text after the reference.
func regexPatternExtendOctalEscape(pattern string, start int, end int) int {
	if end-start != 2 || start+1 >= len(pattern) {
		return end
	}
	first := pattern[start+1]
	if first < '0' || first > '7' {
		return end
	}

	value := uint32(first - '0')
	scan := start + 2
	// Two bounds, and only one of them is reachable on its own: the 255 ceiling fires first for
	// every input the three-digit cap would catch, because a fourth octal digit always exceeds it.
	// A sweep confirmed the cap can be widened to nine with no fixture noticing. It stays because it
	// is the language's rule stated directly, and leaving the walk to depend on an arithmetic
	// coincidence would be a worse thing to read than one redundant condition.
	for digits := 1; digits < 3 && scan < len(pattern); digits++ {
		next := pattern[scan]
		if next < '0' || next > '7' {
			break
		}
		if value*8+uint32(next-'0') > 0xFF {
			break
		}
		value = value*8 + uint32(next-'0')
		scan++
	}
	return scan
}

// braceQuantifierEnd returns the index past a counted quantifier, or false when the brace does not
// open one.
//
// A brace that is not a quantifier is a literal brace, which is legal in a pattern, so refusing is
// the common path rather than an error path. The accepted shapes are a count, a count with a
// trailing comma, and a range.
func regexPatternBraceQuantifierEnd(pattern string, start int) (int, bool) {
	if start >= len(pattern) || pattern[start] != '{' {
		return 0, false
	}

	index := start + 1
	digitsBefore := regexPatternConsumeDigits(pattern, index)
	if digitsBefore == index {
		return 0, false
	}
	index = digitsBefore

	if index < len(pattern) && pattern[index] == ',' {
		index++
		index = regexPatternConsumeDigits(pattern, index)
	}

	if index >= len(pattern) || pattern[index] != '}' {
		return 0, false
	}
	return index + 1, true
}

// consumeDigits returns the index past a run of decimal digits, which may be empty.
func regexPatternConsumeDigits(pattern string, index int) int {
	for index < len(pattern) && pattern[index] >= '0' && pattern[index] <= '9' {
		index++
	}
	return index
}

// isAsciiLetter reports whether a byte is an unaccented letter, which is what a control escape
// accepts.
func regexPatternIsAsciiLetter(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z')
}

// walker holds the position and the enclosing structure while scanning a pattern.
//
// Depth is tracked as counters rather than a stack because no caller asks which quantifier or which
// class, only whether it is inside one. A stack would be the right shape the moment something needs
// the enclosing node itself, and this is the seam where that change goes.
type regexPatternWalker struct {
	pattern  string
	flags    regexPatternRegexFlags
	callback func(regexPatternCharacter) bool

	quantifierDepth int
	classDepth      int
	stopped         bool
}

// run scans the pattern top to bottom, returning false on malformed input or an early stop.
func (w *regexPatternWalker) run() bool {
	index := 0
	for index < len(w.pattern) {
		if w.stopped {
			return false
		}

		switch w.pattern[index] {
		case '\\':
			character, next, ok := w.readEscape(index)
			if !ok {
				return false
			}
			// A class escape like \d matches a set rather than a character, so it has no single
			// value to report and is skipped rather than invented.
			if character == nil {
				index = w.applyQuantifier(next)
				break
			}
			index = w.emitAndQuantify(*character)
			if w.stopped {
				return false
			}

		case '[':
			end, ok := regexPatternClassEnd(w.pattern, index, w.flags)
			if !ok {
				return false
			}
			// A quantifier after a class governs the whole class, so its contents carry that depth
			// too. Every character inside a class already carries ClassDepth, so this only matters
			// to a caller reading the two separately.
			next, quantified := regexPatternQuantifierAt(w.pattern, end)
			if quantified {
				w.quantifierDepth++
			}
			walked := w.walkClass(index, end)
			if quantified {
				w.quantifierDepth--
			}
			if !walked {
				return false
			}
			index = next

		case '(':
			// Group structure carries no characters of its own, and its prologue is syntax rather
			// than content: the `?:`, `?=`, `?!` and `?<name>` forms would otherwise be emitted as
			// ordinary matched characters. A probe caught that, reporting the question mark and
			// colon of a non-capturing group as things the pattern matches.
			index = regexPatternGroupPrologueEnd(w.pattern, index)

		case ')':
			index = w.applyQuantifier(index + 1)

		case '|', '^', '$':
			// Alternation and anchors match positions rather than characters.
			index++

		case '.':
			index = w.emitAndQuantify(regexPatternCharacter{Value: '.', Kind: regexPatternKindSymbol, Start: index, End: index + 1})
			if w.stopped {
				return false
			}

		default:
			value, width := regexPatternDecodeRune(w.pattern, index)
			index = w.emitAndQuantify(regexPatternCharacter{
				Value: value,
				Kind:  regexPatternKindSymbol,
				Start: index,
				End:   index + width,
			})
			if w.stopped {
				return false
			}
		}
	}
	return !w.stopped
}

// emitAndQuantify reports a character under the depth of any quantifier governing it, and returns
// the index past both.
//
// The lookahead is what makes the depth right. A quantifier contains the element it repeats, so the
// character has to be emitted while that depth is raised, not before it is discovered.
func (w *regexPatternWalker) emitAndQuantify(character regexPatternCharacter) int {
	next, quantified := regexPatternQuantifierAt(w.pattern, character.End)
	if quantified {
		w.quantifierDepth++
	}
	emitted := w.emit(character)
	if quantified {
		w.quantifierDepth--
	}
	if !emitted {
		return character.End
	}
	return next
}

// groupPrologueEnd returns the index past a group's opening syntax.
//
// The bytes that say what kind of group this is are structure rather than content, so a walk
// reporting matched characters must step over them. Everything after the prologue is the group's
// body and is scanned by the ordinary loop.
func regexPatternGroupPrologueEnd(pattern string, index int) int {
	if index+1 >= len(pattern) || pattern[index+1] != '?' {
		return index + 1
	}
	if index+2 >= len(pattern) {
		return index + 2
	}

	switch pattern[index+2] {
	case ':', '=', '!':
		return index + 3
	case '<':
		// Lookbehind is three bytes of syntax; a named group runs to its closing angle bracket, and
		// the name is not matched text either.
		if index+3 < len(pattern) && (pattern[index+3] == '=' || pattern[index+3] == '!') {
			return index + 4
		}
		for scan := index + 3; scan < len(pattern); scan++ {
			if pattern[scan] == '>' {
				return scan + 1
			}
		}
		return len(pattern)
	}
	return index + 2
}

// emit reports one character, filling in the depths the walker is currently inside.
func (w *regexPatternWalker) emit(character regexPatternCharacter) bool {
	character.QuantifierDepth = w.quantifierDepth
	character.ClassDepth = w.classDepth
	if !w.callback(character) {
		w.stopped = true
		return false
	}
	return true
}

// quantifierAt reports the index past a quantifier at this position, or false when none is there.
//
// A quantifier is looked for before the element it governs is emitted rather than consumed after,
// which is the opposite of how it reads. The reason is that a quantifier node *contains* the
// element it repeats, so the repeated character is inside it and has to carry that depth.
//
// This was wrong in the first draft and a probe is what said so: a pattern of two spaces followed
// by a plus reported both spaces at depth zero, which is exactly the case upstream passes and this
// package exists to let a caller skip. Reasoning about which came first produced a confident
// argument for the wrong answer.
func regexPatternQuantifierAt(pattern string, index int) (int, bool) {
	if index >= len(pattern) {
		return index, false
	}
	switch pattern[index] {
	case '*', '+', '?':
		index++
	case '{':
		end, ok := regexPatternBraceQuantifierEnd(pattern, index)
		if !ok {
			return index, false
		}
		index = end
	default:
		return index, false
	}
	// A lazy quantifier's question mark is part of the quantifier rather than another one.
	//
	// Consuming it here changes no output: a later scan would read the same byte as a quantifier
	// governing nothing and step past it, so a sweep finds this line deletable. It stays because
	// the alternative is a walk that is right by coincidence, and a reader asking whether lazy
	// quantifiers are handled should find the answer here rather than infer it from elsewhere.
	if index < len(pattern) && pattern[index] == '?' {
		index++
	}
	return index, true
}

// applyQuantifier consumes a quantifier following the element that just ended, and reports the
// index past it.
//
// The quantifier is consumed rather than walked into because its braces hold digits that are not
// characters the pattern matches. Missing that is what would make a two-space run followed by a
// counted quantifier read as two spaces followed by a literal digit.
func (w *regexPatternWalker) applyQuantifier(index int) int {
	if index >= len(w.pattern) {
		return index
	}
	switch w.pattern[index] {
	case '*', '+', '?':
		index++
	case '{':
		end, ok := regexPatternBraceQuantifierEnd(w.pattern, index)
		if !ok {
			// A lone brace is a literal in a pattern, and it was already emitted as one by the
			// default branch. Returning the index unchanged lets the loop move past it.
			return index
		}
		index = end
	default:
		return index
	}
	// A lazy quantifier's question mark is part of the quantifier rather than another one.
	if index < len(w.pattern) && w.pattern[index] == '?' {
		index++
	}
	return index
}

// walkClass reports the characters inside a class body, with the class depth raised.
//
// The body is re-parsed by regexsyntax rather than scanned here, so the two packages cannot
// disagree about what a class contains. Same argument as using the shared pattern-and-flags split
// rather than a second copy: a class body has enough edge cases that two implementations drift.
func (w *regexPatternWalker) walkClass(start int, end int) bool {
	w.classDepth++
	defer func() { w.classDepth-- }()

	elements, _, ok := regexPatternParseRegexCharacterClassWithEnd(w.pattern, start, end, w.flags)
	if !ok {
		return false
	}

	for _, element := range elements {
		switch element.Kind {
		case regexPatternRegexCharSingle:
			character := regexPatternCharacter{
				Value: uint32(element.Value),
				Kind:  w.kindFromSource(element.Start, element.End),
				Start: element.Start,
				End:   element.End,
			}
			if !w.emit(character) {
				return false
			}

		case regexPatternRegexCharRange:
			// A range's lower endpoint is a written character and a caller asking about spelling
			// wants it. The characters the range covers between its endpoints were never written
			// down and are not reported.
			lower := regexPatternCharacter{
				Value: uint32(element.Value),
				Kind:  w.kindFromSource(element.Start, element.End),
				Start: element.Start,
				End:   element.End,
			}
			if !w.emit(lower) {
				return false
			}
		}
	}

	return true
}

// kindFromSource reads how a character was spelled from the source text covering it.
//
// The kind cannot be derived from the value: a null character can be written four ways and the rule
// that cares about this reports three of them and not the fourth. So spelling is read from the
// bytes rather than inferred from what they mean.
func (w *regexPatternWalker) kindFromSource(start int, end int) regexPatternCharacterKind {
	if start < 0 || end > len(w.pattern) || end-start < 1 {
		return regexPatternKindSymbol
	}
	text := w.pattern[start:end]
	if text[0] != '\\' || len(text) < 2 {
		return regexPatternKindSymbol
	}

	switch text[1] {
	case 'x':
		return regexPatternKindHexadecimalEscape
	case 'u':
		return regexPatternKindUnicodeEscape
	case 'c':
		return regexPatternKindControlLetter
	case 'n', 'r', 't', 'v', 'f', 'b':
		return regexPatternKindSingleEscape
	case '0':
		// A bare zero escape is the null escape only when no digit follows. With a digit it is a
		// legacy octal escape, and the two are different kinds to the one rule that reads them.
		if len(text) == 2 {
			return regexPatternKindNull
		}
		return regexPatternKindOctal
	}
	if text[1] >= '1' && text[1] <= '9' {
		return regexPatternKindOctal
	}
	return regexPatternKindIdentityEscape
}

// readEscape reads a backslash escape outside a character class.
//
// It returns a nil character for an escape matching a set rather than a single character, and for a
// named backreference, since neither has one value to report.
func (w *regexPatternWalker) readEscape(index int) (*regexPatternCharacter, int, bool) {
	step, ok := regexPatternSkipPatternEscape(w.pattern, index, w.flags)
	if !ok {
		return nil, index, false
	}
	end := index + step
	if end > len(w.pattern) {
		return nil, index, false
	}

	// A legacy octal escape can run to three digits, and the shared scanner stops at one.
	//
	// That is correct for what it does: it decides where a character class ends, and for that
	// question a digit escape is two bytes whatever follows it. Reading the character is this
	// package's job, so the span is extended here rather than by changing a scanner four other
	// callers depend on. Without this, `\101` reports the letter A followed by two stray digits.
	end = regexPatternExtendOctalEscape(w.pattern, index, end)

	text := w.pattern[index:end]
	if len(text) < 2 {
		return nil, end, true
	}

	switch text[1] {
	case 'd', 'D', 's', 'S', 'w', 'W', 'p', 'P', 'k', 'q':
		// Set escapes and named backreferences match more than one character, or none.
		return nil, end, true
	case 'b', 'B':
		// Outside a character class both word-boundary escapes match a position rather than a
		// character. Inside one, the lowercase form means backspace instead, and that reading is
		// reached through walkClass rather than here.
		//
		// The order matters and is the whole of this case: the named-escape table also holds `b`,
		// so a kind lookup that ran first would call this a backspace and report a character the
		// pattern never matches.
		return nil, end, true
	}

	kind := w.kindFromSource(index, end)
	value, ok := regexPatternEscapeValue(text, kind)
	if !ok {
		return nil, end, true
	}

	character := regexPatternCharacter{Value: value, Kind: kind, Start: index, End: end}
	return &character, end, true
}

// decodeRune reads one rune and its width, falling back to a single byte on invalid input rather
// than stalling the walk.
func regexPatternDecodeRune(text string, index int) (uint32, int) {
	value, width := utf8.DecodeRuneInString(text[index:])
	if width == 0 {
		return uint32(text[index]), 1
	}
	return uint32(value), width
}

func (p *Program) regexPatternFacts(out *fields, question string) error {
	parts := strings.SplitN(question, "\n", 3)
	if len(parts) != 3 {
		return fmt.Errorf("regex-pattern-facts requires flags and pattern")
	}
	var units []uint16
	if parts[2] != "" {
		for _, encoded := range strings.Split(parts[2], ",") {
			value, err := strconv.ParseUint(encoded, 10, 16)
			if err != nil || strconv.FormatUint(value, 10) != encoded {
				return fmt.Errorf("invalid regex UTF-16 unit")
			}
			units = append(units, uint16(value))
		}
	}
	pattern := string(utf16.Decode(units))
	flags := regexPatternParseRegexFlags(parts[1])
	offset := func(index int) uint64 { return uint64(len(utf16.Encode([]rune(pattern[:index])))) }
	var characters []regexPatternCharacter
	walked := regexPatternWalk(pattern, flags, func(character regexPatternCharacter) bool { characters = append(characters, character); return true })
	out.yes(walked)
	out.number(uint64(len(characters)))
	for _, character := range characters {
		out.number(offset(character.Start))
		out.number(offset(character.End))
	}
	var escapes, classes []int
	for index, character := range pattern {
		if character == '\\' {
			escapes = append(escapes, index)
		}
		if character == '[' {
			classes = append(classes, index)
		}
	}
	out.number(uint64(len(escapes)))
	for _, index := range escapes {
		width, ok := regexPatternSkipPatternEscape(pattern, index, flags)
		out.number(offset(index))
		out.yes(ok && width > 0)
		end := index
		if ok && width > 0 {
			end = index + width
		}
		out.number(offset(end))
	}
	out.number(uint64(len(classes)))
	for _, index := range classes {
		end, ok := regexPatternClassEnd(pattern, index, flags)
		out.number(offset(index))
		out.yes(ok && end > index)
		if !ok || end <= index {
			end = index
		}
		out.number(offset(end))
	}
	return nil
}
