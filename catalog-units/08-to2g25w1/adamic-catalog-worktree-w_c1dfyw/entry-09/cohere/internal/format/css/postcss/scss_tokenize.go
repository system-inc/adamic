package postcss

// postcss-scss 4.0.9, lib/scss-tokenize.js: postcss's tokenizer with the SCSS PATCH blocks (commas as
// words, #{} interpolation, `//` comments, nested parentheses in url(), strings that can hold
// interpolation).
//
// Upstream is a closure over one set of variables (code, next, n, escaped, pos, ...) that nextToken and
// interpolation both write; the port keeps them as the struct's fields so interpolation moves `next` for
// its caller exactly as upstream's does. Offsets are UTF-16 units over input.units, as in tokenize.go.
//
// options.ignoreErrors and nextToken's opts.ignoreUnclosed are never set on this path (ScssParser calls
// nextToken() with no options and Prettier passes none), so the branches that read them are false.

// The SCSS PATCH constants.
const (
	comma = ','
	hash  = '#'
)

// isScssWordEnd is the scss RE_WORD_END's class, /[,\t\n\f\r !"#'():;@[\\\]{}]/: postcss's with the
// comma; the regex's other branch, /\/(?=\*)/, is checked by the caller.
func isScssWordEnd(unit uint16) bool {
	return unit == ',' || isWordEnd(unit)
}

// isNewLine is RE_NEW_LINE's class, /[\n\f\r]/.
func isNewLine(unit uint16) bool {
	return unit == '\n' || unit == '\f' || unit == '\r'
}

type scssTokenizer struct {
	input *input
	css   []uint16

	code, next, quote int
	content           string
	escape            bool
	escaped           bool
	prev              string
	n                 int
	currentToken      token

	length   int
	pos      int
	buffer   []token
	returned []token

	brackets int // SCSS PATCH
}

// newScssTokenizer is scssTokenize(input).
func newScssTokenizer(in *input) *scssTokenizer {
	return &scssTokenizer{
		input:    in,
		css:      in.units,
		length:   len(in.units),
		pos:      0,
		buffer:   []token{},
		returned: []token{},
	}
}

// charCodeAt is css.charCodeAt(index), -1 standing for NaN past either end.
func (tokens *scssTokenizer) charCodeAt(index int) int {
	if index < 0 || index >= tokens.length {
		return -1
	}
	return int(tokens.css[index])
}

// indexOf is css.indexOf(search, from) for an ASCII search string.
func (tokens *scssTokenizer) indexOf(search string, from int) int {
	from = max(from, 0)
	for index := from; index+len(search) <= tokens.length; index++ {
		matched := true
		for offset := 0; offset < len(search); offset++ {
			if tokens.css[index+offset] != uint16(search[offset]) {
				matched = false
				break
			}
		}
		if matched {
			return index
		}
	}
	return -1
}

func (tokens *scssTokenizer) position() int {
	return tokens.pos
}

func (tokens *scssTokenizer) unclosed(what string) {
	panic(tokens.input.errorAt("Unclosed "+what, tokens.pos))
}

func (tokens *scssTokenizer) endOfFile() bool {
	return len(tokens.returned) == 0 && tokens.pos >= tokens.length
}

// SCSS PATCH {
func (tokens *scssTokenizer) interpolation() {
	deep := 1
	// stringQuote is false or a quote's char code; 0 stands for false, no quote being NUL.
	stringQuote := 0
	stringEscaped := false
	for deep > 0 {
		tokens.next += 1
		if tokens.length <= tokens.next {
			tokens.unclosed("interpolation")
		}

		tokens.code = tokens.charCodeAt(tokens.next)
		tokens.n = tokens.charCodeAt(tokens.next + 1)

		if stringQuote != 0 {
			if !stringEscaped && tokens.code == stringQuote {
				stringQuote = 0
				stringEscaped = false
			} else if tokens.code == backslash {
				stringEscaped = !stringEscaped
			} else if stringEscaped {
				stringEscaped = false
			}
		} else if tokens.code == singleQuote || tokens.code == doubleQuote {
			stringQuote = tokens.code
		} else if tokens.code == closeCurly {
			deep -= 1
		} else if tokens.code == hash && tokens.n == openCurly {
			deep += 1
		}
	}
}

// } SCSS PATCH

// nextToken returns the next token, and false where upstream returns undefined.
func (tokens *scssTokenizer) nextToken() (token, bool) {
	if len(tokens.returned) > 0 {
		last := tokens.returned[len(tokens.returned)-1]
		tokens.returned = tokens.returned[:len(tokens.returned)-1]
		return last, true
	}
	if tokens.pos >= tokens.length {
		return token{}, false
	}

	tokens.code = tokens.charCodeAt(tokens.pos)

	switch tokens.code {
	case newline, space, tab, carriageReturn, feed:
		tokens.next = tokens.pos
		for {
			tokens.next += 1
			tokens.code = tokens.charCodeAt(tokens.next)
			if !(tokens.code == space || tokens.code == newline || tokens.code == tab || tokens.code == carriageReturn || tokens.code == feed) {
				break
			}
		}

		tokens.currentToken = token{kind: "space", value: tokens.input.slice(tokens.pos, tokens.next), start: -1, end: -1}
		tokens.pos = tokens.next - 1

	case openSquare, closeSquare, openCurly, closeCurly, colon, semicolon, closeParentheses:
		controlChar := string(rune(tokens.code))
		tokens.currentToken = token{kind: controlChar, value: controlChar, start: tokens.pos, end: -1}

	// SCSS PATCH {
	case comma:
		tokens.currentToken = token{kind: "word", value: ",", start: tokens.pos, end: tokens.pos + 1}
	// } SCSS PATCH

	case openParentheses:
		tokens.prev = ""
		if len(tokens.buffer) > 0 {
			tokens.prev = tokens.buffer[len(tokens.buffer)-1].value
			tokens.buffer = tokens.buffer[:len(tokens.buffer)-1]
		}
		tokens.n = tokens.charCodeAt(tokens.pos + 1)

		// SCSS PATCH {
		if tokens.prev == "url" && tokens.n != singleQuote && tokens.n != doubleQuote {
			tokens.brackets = 1
			tokens.escaped = false
			tokens.next = tokens.pos + 1
			for tokens.next <= tokens.length-1 {
				tokens.n = tokens.charCodeAt(tokens.next)
				if tokens.n == backslash {
					tokens.escaped = !tokens.escaped
				} else if tokens.n == openParentheses {
					tokens.brackets += 1
				} else if tokens.n == closeParentheses {
					tokens.brackets -= 1
					if tokens.brackets == 0 {
						break
					}
				}
				tokens.next += 1
			}

			tokens.content = tokens.input.slice(tokens.pos, tokens.next+1)
			tokens.currentToken = token{kind: "brackets", value: tokens.content, start: tokens.pos, end: tokens.next}
			tokens.pos = tokens.next
			// } SCSS PATCH
		} else {
			tokens.next = tokens.indexOf(")", tokens.pos+1)
			tokens.content = tokens.input.slice(tokens.pos, tokens.next+1)

			if tokens.next == -1 || hasBadBracket(tokens.css[tokens.pos:tokens.next+1]) {
				tokens.currentToken = token{kind: "(", value: "(", start: tokens.pos, end: -1}
			} else {
				tokens.currentToken = token{kind: "brackets", value: tokens.content, start: tokens.pos, end: tokens.next}
				tokens.pos = tokens.next
			}
		}

	case singleQuote, doubleQuote:
		// SCSS PATCH {
		tokens.quote = tokens.code
		tokens.next = tokens.pos

		tokens.escaped = false
		for tokens.next < tokens.length {
			tokens.next++
			if tokens.next == tokens.length {
				tokens.unclosed("string")
			}

			tokens.code = tokens.charCodeAt(tokens.next)
			tokens.n = tokens.charCodeAt(tokens.next + 1)

			if !tokens.escaped && tokens.code == tokens.quote {
				break
			} else if tokens.code == backslash {
				tokens.escaped = !tokens.escaped
			} else if tokens.escaped {
				tokens.escaped = false
			} else if tokens.code == hash && tokens.n == openCurly {
				tokens.interpolation()
			}
		}
		// } SCSS PATCH

		tokens.currentToken = token{kind: "string", value: tokens.input.slice(tokens.pos, tokens.next+1), start: tokens.pos, end: tokens.next}
		tokens.pos = tokens.next

	case at:
		// RE_AT_END.lastIndex = pos + 1; RE_AT_END.test(css): a match at index m leaves lastIndex m + 1,
		// and no match leaves it 0.
		match := -1
		for index := tokens.pos + 1; index < tokens.length; index++ {
			if isAtEnd(tokens.css[index]) {
				match = index
				break
			}
		}
		if match == -1 {
			tokens.next = tokens.length - 1
		} else {
			tokens.next = match - 1
		}

		tokens.currentToken = token{kind: "at-word", value: tokens.input.slice(tokens.pos, tokens.next+1), start: tokens.pos, end: tokens.next}

		tokens.pos = tokens.next

	case backslash:
		tokens.next = tokens.pos
		tokens.escape = true
		for tokens.charCodeAt(tokens.next+1) == backslash {
			tokens.next += 1
			tokens.escape = !tokens.escape
		}
		tokens.code = tokens.charCodeAt(tokens.next + 1)
		if tokens.escape &&
			tokens.code != slash &&
			tokens.code != space &&
			tokens.code != newline &&
			tokens.code != tab &&
			tokens.code != carriageReturn &&
			tokens.code != feed {
			tokens.next += 1
			if isHexEscape(tokens.charCodeAt(tokens.next)) {
				for isHexEscape(tokens.charCodeAt(tokens.next + 1)) {
					tokens.next += 1
				}
				if tokens.charCodeAt(tokens.next+1) == space {
					tokens.next += 1
				}
			}
		}

		tokens.currentToken = token{kind: "word", value: tokens.input.slice(tokens.pos, tokens.next+1), start: tokens.pos, end: tokens.next}

		tokens.pos = tokens.next

	default:
		// SCSS PATCH {
		tokens.n = tokens.charCodeAt(tokens.pos + 1)

		if tokens.code == hash && tokens.n == openCurly {
			tokens.next = tokens.pos
			tokens.interpolation()
			tokens.content = tokens.input.slice(tokens.pos, tokens.next+1)
			tokens.currentToken = token{kind: "word", value: tokens.content, start: tokens.pos, end: tokens.next}
			tokens.pos = tokens.next
		} else if tokens.code == slash && tokens.n == asterisk {
			// } SCSS PATCH
			tokens.next = tokens.indexOf("*/", tokens.pos+2) + 1
			if tokens.next == 0 {
				// ignore and ignoreUnclosed are always false here.
				tokens.unclosed("comment")
			}

			tokens.currentToken = token{kind: "comment", value: tokens.input.slice(tokens.pos, tokens.next+1), start: tokens.pos, end: tokens.next}
			tokens.pos = tokens.next

			// SCSS PATCH {
		} else if tokens.code == slash && tokens.n == slash {
			// RE_NEW_LINE.lastIndex = pos + 1; RE_NEW_LINE.test(css).
			match := -1
			for index := tokens.pos + 1; index < tokens.length; index++ {
				if isNewLine(tokens.css[index]) {
					match = index
					break
				}
			}
			if match == -1 {
				tokens.next = tokens.length - 1
			} else {
				tokens.next = match - 1
			}

			tokens.content = tokens.input.slice(tokens.pos, tokens.next+1)
			tokens.currentToken = token{kind: "comment", value: tokens.content, start: tokens.pos, end: tokens.next, inline: true}

			tokens.pos = tokens.next
			// } SCSS PATCH
		} else {
			// RE_WORD_END.lastIndex = pos + 1; RE_WORD_END.test(css).
			match := -1
			for index := tokens.pos + 1; index < tokens.length; index++ {
				unit := tokens.css[index]
				if isScssWordEnd(unit) || (unit == slash && tokens.charCodeAt(index+1) == asterisk) {
					match = index
					break
				}
			}
			if match == -1 {
				tokens.next = tokens.length - 1
			} else {
				tokens.next = match - 1
			}

			tokens.currentToken = token{kind: "word", value: tokens.input.slice(tokens.pos, tokens.next+1), start: tokens.pos, end: tokens.next}
			tokens.buffer = append(tokens.buffer, tokens.currentToken)
			tokens.pos = tokens.next
		}
	}

	tokens.pos++
	return tokens.currentToken, true
}

func (tokens *scssTokenizer) back(each token) {
	tokens.returned = append(tokens.returned, each)
}
