package regexpattern

import (
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
)

// walker holds the position and the enclosing structure while scanning a pattern.
//
// Depth is tracked as counters rather than a stack because no caller asks which quantifier or which
// class, only whether it is inside one. A stack would be the right shape the moment something needs
// the enclosing node itself, and this is the seam where that change goes.
type walker struct {
	pattern  string
	flags    regexsyntax.RegexFlags
	callback func(Character) bool

	quantifierDepth int
	classDepth      int
	stopped         bool
}

// run scans the pattern top to bottom, returning false on malformed input or an early stop.
func (w *walker) run() bool {
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
			end, ok := regexsyntax.ClassEnd(w.pattern, index, w.flags)
			if !ok {
				return false
			}
			// A quantifier after a class governs the whole class, so its contents carry that depth
			// too. Every character inside a class already carries ClassDepth, so this only matters
			// to a caller reading the two separately.
			next, quantified := quantifierAt(w.pattern, end)
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
			index = groupPrologueEnd(w.pattern, index)

		case ')':
			index = w.applyQuantifier(index + 1)

		case '|', '^', '$':
			// Alternation and anchors match positions rather than characters.
			index++

		case '.':
			index = w.emitAndQuantify(Character{Value: '.', Kind: KindSymbol, Start: index, End: index + 1})
			if w.stopped {
				return false
			}

		default:
			value, width := decodeRune(w.pattern, index)
			index = w.emitAndQuantify(Character{
				Value: value,
				Kind:  KindSymbol,
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
func (w *walker) emitAndQuantify(character Character) int {
	next, quantified := quantifierAt(w.pattern, character.End)
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
func groupPrologueEnd(pattern string, index int) int {
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
func (w *walker) emit(character Character) bool {
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
func quantifierAt(pattern string, index int) (int, bool) {
	if index >= len(pattern) {
		return index, false
	}
	switch pattern[index] {
	case '*', '+', '?':
		index++
	case '{':
		end, ok := braceQuantifierEnd(pattern, index)
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
func (w *walker) applyQuantifier(index int) int {
	if index >= len(w.pattern) {
		return index
	}
	switch w.pattern[index] {
	case '*', '+', '?':
		index++
	case '{':
		end, ok := braceQuantifierEnd(w.pattern, index)
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
func (w *walker) walkClass(start int, end int) bool {
	w.classDepth++
	defer func() { w.classDepth-- }()

	elements, _, ok := regexsyntax.ParseRegexCharacterClassWithEnd(w.pattern, start, end, w.flags)
	if !ok {
		return false
	}

	for _, element := range elements {
		switch element.Kind {
		case regexsyntax.RegexCharSingle:
			character := Character{
				Value: uint32(element.Value),
				Kind:  w.kindFromSource(element.Start, element.End),
				Start: element.Start,
				End:   element.End,
			}
			if !w.emit(character) {
				return false
			}

		case regexsyntax.RegexCharRange:
			// A range's lower endpoint is a written character and a caller asking about spelling
			// wants it. The characters the range covers between its endpoints were never written
			// down and are not reported.
			lower := Character{
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
func (w *walker) kindFromSource(start int, end int) CharacterKind {
	if start < 0 || end > len(w.pattern) || end-start < 1 {
		return KindSymbol
	}
	text := w.pattern[start:end]
	if text[0] != '\\' || len(text) < 2 {
		return KindSymbol
	}

	switch text[1] {
	case 'x':
		return KindHexadecimalEscape
	case 'u':
		return KindUnicodeEscape
	case 'c':
		return KindControlLetter
	case 'n', 'r', 't', 'v', 'f', 'b':
		return KindSingleEscape
	case '0':
		// A bare zero escape is the null escape only when no digit follows. With a digit it is a
		// legacy octal escape, and the two are different kinds to the one rule that reads them.
		if len(text) == 2 {
			return KindNull
		}
		return KindOctal
	}
	if text[1] >= '1' && text[1] <= '9' {
		return KindOctal
	}
	return KindIdentityEscape
}

// readEscape reads a backslash escape outside a character class.
//
// It returns a nil character for an escape matching a set rather than a single character, and for a
// named backreference, since neither has one value to report.
func (w *walker) readEscape(index int) (*Character, int, bool) {
	step, ok := regexsyntax.SkipPatternEscape(w.pattern, index, w.flags)
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
	end = extendOctalEscape(w.pattern, index, end)

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
	value, ok := escapeValue(text, kind)
	if !ok {
		return nil, end, true
	}

	character := Character{Value: value, Kind: kind, Start: index, End: end}
	return &character, end, true
}

// decodeRune reads one rune and its width, falling back to a single byte on invalid input rather
// than stalling the walk.
func decodeRune(text string, index int) (uint32, int) {
	value, width := utf8.DecodeRuneInString(text[index:])
	if width == 0 {
		return uint32(text[index]), 1
	}
	return uint32(value), width
}
