package gitignore

import "strings"

// A glob is one pattern's wildcard part, compiled once when its ignore file is read. It is written from
// gitignore(5) and fnmatch(3)'s descriptions:
//
//   - `?` is any one byte, `*` any run of bytes, and `[...]` one byte from a set: ranges, `!` or `^` to
//     negate, the `[:name:]` classes, and `]` as a member when it comes first.
//   - A backslash makes the next byte literal, and a backslash ending the pattern makes it match nothing.
//   - Against a path (anchored patterns), none of those crosses a `/`, and `**` as a whole segment does:
//     `**/` at the start or after a `/` is zero or more directories, and a trailing `/**` everything
//     below. Any other run of asterisks is one `*`.
//   - A set that never closes, or a class name fnmatch does not define, makes the pattern match nothing.
//
// Bytes, not runes, so a pattern and a path agree exactly when their bytes do, and the classes are ASCII.
// Matching walks the text once while tracking every place in the pattern it could have reached, so no
// pattern backtracks: `*a*a*a*a*b` against a long run of `a` costs one pass.

// tokenKind is what one compiled piece of a glob matches.
type tokenKind uint8

const (
	literalToken     tokenKind = iota // one byte, itself
	anyByteToken                      // `?`
	setToken                          // `[...]`
	starToken                         // `*`: any run, within one segment against a path
	everythingToken                   // a trailing `**` segment: any run, slashes included
	directoriesToken                  // `**/` as a segment: nothing, or any run that ends in `/`
)

type token struct {
	kind    tokenKind
	literal byte
	set     *[256]bool
}

// glob is a compiled pattern. never marks one that can match nothing (a set left open, an unknown class,
// a trailing backslash); literal holds the text of one with no wildcard at all, so matching it is one
// comparison.
type glob struct {
	tokens    []token
	never     bool
	isLiteral bool
	literal   string

	// isSuffix marks a base-name glob that is one `*` and then literal text, `*.log`: matching it is a
	// suffix comparison. It is the commonest shape after a plain name.
	isSuffix bool
}

// compileGlob reads pattern. path says whether it will be matched against a path, which is what gives
// `**` its meaning; against a base name every run of asterisks is one `*`.
func compileGlob(pattern string, path bool) glob {
	compiled := glob{}
	var literal strings.Builder
	isLiteral := true
	for index := 0; index < len(pattern); index++ {
		character := pattern[index]
		switch character {
		case '\\':
			index++
			if index >= len(pattern) {
				return glob{never: true}
			}
			compiled.tokens = append(compiled.tokens, token{kind: literalToken, literal: pattern[index]})
			literal.WriteByte(pattern[index])
			continue

		case '?':
			compiled.tokens = append(compiled.tokens, token{kind: anyByteToken})

		case '[':
			set, end, ok := compileSet(pattern, index)
			if !ok {
				return glob{never: true}
			}
			compiled.tokens = append(compiled.tokens, token{kind: setToken, set: set})
			index = end

		case '*':
			runEnd := index
			for runEnd+1 < len(pattern) && pattern[runEnd+1] == '*' {
				runEnd++
			}
			kind := starToken
			if path && runEnd > index {
				segmentStart := index == 0 || pattern[index-1] == '/'
				followedBySlash := runEnd+1 < len(pattern) && pattern[runEnd+1] == '/'
				followedByEscapedSlash := runEnd+2 < len(pattern) && pattern[runEnd+1] == '\\' && pattern[runEnd+2] == '/'
				switch {
				case segmentStart && runEnd+1 == len(pattern):
					kind = everythingToken
				case segmentStart && followedBySlash:
					// The segment's slash belongs to the token: zero directories consume nothing at all.
					kind = directoriesToken
					runEnd++
				case segmentStart && followedByEscapedSlash:
					kind = directoriesToken
					runEnd += 2
				}
			}
			compiled.tokens = append(compiled.tokens, token{kind: kind})
			index = runEnd

		default:
			compiled.tokens = append(compiled.tokens, token{kind: literalToken, literal: character})
			literal.WriteByte(character)
			continue
		}
		isLiteral = false
	}
	if isLiteral {
		compiled.isLiteral = true
		compiled.literal = literal.String()
		return compiled
	}
	if !path && len(compiled.tokens) > 0 && compiled.tokens[0].kind == starToken {
		suffix := make([]byte, 0, len(compiled.tokens)-1)
		for _, rest := range compiled.tokens[1:] {
			if rest.kind != literalToken {
				return compiled
			}
			suffix = append(suffix, rest.literal)
		}
		compiled.isSuffix = true
		compiled.literal = string(suffix)
	}
	return compiled
}

// compileSet reads the bracket expression that opens at start, returning its members and the index of
// its closing `]`, or ok false for one that never closes or names an unknown class.
func compileSet(pattern string, start int) (set *[256]bool, end int, ok bool) {
	members := [256]bool{}
	index := start + 1
	negated := false
	if index < len(pattern) && (pattern[index] == '!' || pattern[index] == '^') {
		negated = true
		index++
	}

	// rangeStart is the byte a following `-` would extend, or -1 after a class or a range, which nothing
	// extends.
	rangeStart := -1
	first := true
	for {
		if index >= len(pattern) {
			return nil, 0, false
		}
		character := pattern[index]
		if character == ']' && !first {
			break
		}
		first = false

		switch {
		case character == '\\':
			index++
			if index >= len(pattern) {
				return nil, 0, false
			}
			members[pattern[index]] = true
			rangeStart = int(pattern[index])

		case character == '-' && rangeStart >= 0 && index+1 < len(pattern) && pattern[index+1] != ']':
			index++
			upper := pattern[index]
			if upper == '\\' {
				index++
				if index >= len(pattern) {
					return nil, 0, false
				}
				upper = pattern[index]
			}
			for member := rangeStart; member <= int(upper); member++ {
				members[member] = true
			}
			rangeStart = -1

		case character == '[' && index+1 < len(pattern) && pattern[index+1] == ':':
			closing := strings.IndexByte(pattern[index+2:], ']')
			if closing < 0 {
				return nil, 0, false
			}
			closing += index + 2
			if closing-1 < index+2 || pattern[closing-1] != ':' {
				// No `:]` before the next `]`, so this `[` is an ordinary member.
				members['['] = true
				rangeStart = '['
				break
			}
			if !addClass(&members, pattern[index+2:closing-1]) {
				return nil, 0, false
			}
			index = closing
			rangeStart = -1

		default:
			members[character] = true
			rangeStart = int(character)
		}
		index++
	}
	if negated {
		for member := range members {
			members[member] = !members[member]
		}
	}
	return &members, index, true
}

// addClass adds a `[:name:]` class's bytes, reporting false for a name fnmatch does not define.
func addClass(members *[256]bool, name string) bool {
	var in func(character byte) bool
	switch name {
	case "alnum":
		in = func(character byte) bool { return isLetter(character) || isDigit(character) }
	case "alpha":
		in = isLetter
	case "blank":
		in = func(character byte) bool { return character == ' ' || character == '\t' }
	case "cntrl":
		in = func(character byte) bool { return character < 0x20 || character == 0x7F }
	case "digit":
		in = isDigit
	case "graph":
		in = func(character byte) bool { return character > 0x20 && character < 0x7F }
	case "lower":
		in = func(character byte) bool { return character >= 'a' && character <= 'z' }
	case "print":
		in = func(character byte) bool { return character >= 0x20 && character < 0x7F }
	case "punct":
		in = func(character byte) bool {
			return character > 0x20 && character < 0x7F && !isLetter(character) && !isDigit(character)
		}
	case "space":
		in = func(character byte) bool {
			return character == ' ' || character == '\t' || character == '\n' || character == '\r'
		}
	case "upper":
		in = func(character byte) bool { return character >= 'A' && character <= 'Z' }
	case "xdigit":
		in = func(character byte) bool {
			return isDigit(character) || (character|0x20 >= 'a' && character|0x20 <= 'f')
		}
	default:
		return false
	}
	for member := 0; member < 256; member++ {
		if in(byte(member)) {
			members[member] = true
		}
	}
	return true
}

func isLetter(character byte) bool { return character|0x20 >= 'a' && character|0x20 <= 'z' }
func isDigit(character byte) bool  { return character >= '0' && character <= '9' }

// matches reports whether the glob matches all of text. path keeps `?`, `*` and sets out of `/`, as it
// does for a pattern matched against a path; the `**` tokens exist only in a glob compiled for a path.
func (compiled *glob) matches(text string, path bool) bool {
	if compiled.never {
		return false
	}
	if compiled.isLiteral {
		return text == compiled.literal
	}
	if compiled.isSuffix {
		return strings.HasSuffix(text, compiled.literal)
	}

	// One bit per place in the pattern, plus a second set for a directoriesToken that has consumed bytes
	// since its last `/`, which may not yet hand on to the token after it. Up to 128 places live on the
	// stack; a longer pattern allocates.
	words := (len(compiled.tokens) + 1 + 63) / 64
	var storage [4][2]uint64
	var current, inside, nextPositions, nextInside positionSet
	if words <= 2 {
		current, inside, nextPositions, nextInside = storage[0][:words], storage[1][:words], storage[2][:words], storage[3][:words]
	} else {
		current, inside = make(positionSet, words), make(positionSet, words)
		nextPositions, nextInside = make(positionSet, words), make(positionSet, words)
	}
	compiled.reach(current, 0)

	for textIndex := 0; textIndex < len(text); textIndex++ {
		character := text[textIndex]
		nextPositions.clear()
		nextInside.clear()
		crossesNothing := path && character == '/'
		advanced := false
		for position, tokenEntry := range compiled.tokens {
			reached := current.has(position)
			if !reached && !inside.has(position) {
				continue
			}
			switch tokenEntry.kind {
			case literalToken:
				if reached && character == tokenEntry.literal {
					compiled.reach(nextPositions, position+1)
					advanced = true
				}
			case anyByteToken:
				if reached && !crossesNothing {
					compiled.reach(nextPositions, position+1)
					advanced = true
				}
			case setToken:
				if reached && tokenEntry.set[character] && !crossesNothing {
					compiled.reach(nextPositions, position+1)
					advanced = true
				}
			case starToken:
				if reached && !crossesNothing {
					compiled.reach(nextPositions, position)
					advanced = true
				}
			case everythingToken:
				if reached {
					compiled.reach(nextPositions, position)
					advanced = true
				}
			case directoriesToken:
				nextInside.add(position)
				advanced = true
				if character == '/' {
					compiled.reach(nextPositions, position+1)
				}
			}
		}
		if !advanced {
			return false
		}
		current, nextPositions = nextPositions, current
		inside, nextInside = nextInside, inside
	}
	return current.has(len(compiled.tokens))
}

// reach marks position, and every position after it that a token matching nothing passes straight on to.
func (compiled *glob) reach(positions positionSet, position int) {
	for {
		positions.add(position)
		if position >= len(compiled.tokens) {
			return
		}
		switch compiled.tokens[position].kind {
		case starToken, everythingToken, directoriesToken:
			position++
		default:
			return
		}
	}
}

// positionSet is a bit set over a glob's positions.
type positionSet []uint64

func (positions positionSet) add(position int) { positions[position/64] |= 1 << (position % 64) }
func (positions positionSet) has(position int) bool {
	return positions[position/64]&(1<<(position%64)) != 0
}
func (positions positionSet) clear() { clear(positions) }
