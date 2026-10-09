// Package regexpattern walks a regular expression pattern and reports each character it matches,
// with the structure enclosing it.
//
// This is deliberately not a regex AST. Two rules need character positions plus a little structure,
// and building a full parse tree for them would be a much larger thing to get right with no caller
// asking for the rest of it. What they actually need is narrow:
//
//	no-regex-spaces    which characters are spaces, and whether each sits inside a quantifier or
//	                   a character class (spaces inside either are not consecutive spaces)
//	no-control-regex   which characters are control codes, HOW each was spelled, and how many
//	                   capturing groups the pattern has (so `\1` can be read as a backreference)
//
// So the walk reports characters, and carries depth and kind alongside. A third caller wanting
// alternation or lookahead should extend this rather than find it already half-built and wrong.
package regexpattern

import (
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
)

// CharacterKind is how a character was written, which is a different question from what it matches.
//
// The distinction is the whole of `no-control-regex`: `\0` and `\u0000` are both U+0000 and only
// the second is reported, because the rule objects to spelling a control code out rather than to
// the code point being present. A walk reporting only values cannot express that.
type CharacterKind int

const (
	// KindSymbol is a character written as itself, including a raw control byte.
	KindSymbol CharacterKind = iota
	// KindSingleEscape is a named escape: `\n`, `\t`, `\r`, `\v`, `\f`, `\b`.
	KindSingleEscape
	// KindNull is `\0` with no digit following it.
	KindNull
	// KindHexadecimalEscape is `\xHH`.
	KindHexadecimalEscape
	// KindUnicodeEscape is `\uHHHH` or `\u{H...}`.
	KindUnicodeEscape
	// KindControlLetter is `\cX`.
	KindControlLetter
	// KindOctal is a legacy octal escape, `\1` through `\377`. It is one kind here rather than
	// the three upstream carries, because the only caller asks whether it is octal at all.
	KindOctal
	// KindIdentityEscape is a backslash before a character that needs no escaping, like `\-`.
	KindIdentityEscape
)

// Character is one matched character and where it sits.
//
// Start and End are byte offsets into the pattern text, not into the file, so a caller reporting a
// finding adds the pattern's own offset. Keeping them pattern-relative is what lets the same walk
// serve a regex literal and a string passed to the RegExp constructor, which sit at different
// places in a file and have different quoting.
type Character struct {
	Value uint32
	Kind  CharacterKind
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
func Walk(pattern string, flags regexsyntax.RegexFlags, callback func(Character) bool) bool {
	walker := &walker{pattern: pattern, flags: flags, callback: callback}
	return walker.run()
}

// CountCapturingGroups returns how many capturing groups a pattern declares.
//
// `no-control-regex` needs this to decide whether `\1` is a backreference or a control character,
// and it needs the total before judging any of them: a group may be referenced before it is
// defined, so a running count would call the same escape a control character early in the pattern
// and a backreference later.
func CountCapturingGroups(pattern string, flags regexsyntax.RegexFlags) int {
	count := 0
	for index := 0; index < len(pattern); {
		switch pattern[index] {
		case '\\':
			step, ok := regexsyntax.SkipPatternEscape(pattern, index, flags)
			if !ok {
				return count
			}
			index += step
		case '[':
			end, ok := regexsyntax.ClassEnd(pattern, index, flags)
			if !ok {
				return count
			}
			index = end
		case '(':
			// `(?:`, `(?=`, `(?!`, `(?<=`, `(?<!` are non-capturing. `(?<name>` captures, and it is
			// the one `(?<` form that does, which is why the character after the angle bracket has
			// to be looked at rather than the group being dismissed on `(?` alone.
			if isCapturingGroupStart(pattern, index) {
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
func isCapturingGroupStart(pattern string, index int) bool {
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
