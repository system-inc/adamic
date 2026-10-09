package postcss

// postcss 8.5.16, lib/tokenize.js.
//
// The tokenizer reads input.units, the text's UTF-16 code units, so its offsets are unit offsets like
// upstream's. A token's text is sliced from the Go string through input.slice.
//
// Prettier passes no options, so options.ignoreErrors is false; ignoreUnclosed is never passed by the
// css parser either (only postcss-safe-parser and the scss and less parsers pass it), and the branches
// that read them are kept as false.

const (
	singleQuote      = '\''
	doubleQuote      = '"'
	backslash        = '\\'
	slash            = '/'
	newline          = '\n'
	space            = ' '
	feed             = '\f'
	tab              = '\t'
	carriageReturn   = '\r'
	openSquare       = '['
	closeSquare      = ']'
	openParentheses  = '('
	closeParentheses = ')'
	openCurly        = '{'
	closeCurly       = '}'
	semicolon        = ';'
	asterisk         = '*'
	colon            = ':'
	at               = '@'
)

// token is postcss's token array [type, value, start, end]. start and end are unit offsets, end being
// the offset of the token's last unit; a space token has neither and a control character has no end,
// which is -1 here where upstream's array is shorter.
//
// inline is the fifth element postcss-scss's tokenizer gives a `//` comment, token[4] === 'inline'
// (scss_tokenize.go); postcss's own tokenizer never sets it.
type token struct {
	kind   string
	value  string
	start  int
	end    int
	inline bool
}

// tokenStream is the object a tokenizer function returns, { back, endOfFile, nextToken, position }: the
// parser holds whichever createTokenizer made, postcss's (tokenizer) or postcss-scss's (scssTokenizer).
type tokenStream interface {
	back(each token)
	endOfFile() bool
	nextToken() (token, bool)
	position() int
}

// endOrStart is `token[3] || token[2]`: 0 is falsy there too, so a position at offset 0 falls through,
// and the result is -1 when neither is set.
func (each token) endOrStart() int {
	if each.end > 0 {
		return each.end
	}
	return each.start
}

// isAtEnd is RE_AT_END's class, /[\t\n\f\r "#'()/;[\\\]{}]/.
func isAtEnd(unit uint16) bool {
	switch unit {
	case '\t', '\n', '\f', '\r', ' ', '"', '#', '\'', '(', ')', '/', ';', '[', '\\', ']', '{', '}':
		return true
	}
	return false
}

// isWordEnd is RE_WORD_END's class, /[\t\n\f\r !"#'():;@[\\\]{}]/; the regex's other branch,
// /\/(?=\*)/, is checked by the caller.
func isWordEnd(unit uint16) bool {
	switch unit {
	case '\t', '\n', '\f', '\r', ' ', '!', '"', '#', '\'', '(', ')', ':', ';', '@', '[', '\\', ']', '{', '}':
		return true
	}
	return false
}

// isHexEscape is RE_HEX_ESCAPE, /[\da-f]/i, on one character.
func isHexEscape(code int) bool {
	return (code >= '0' && code <= '9') || (code >= 'a' && code <= 'f') || (code >= 'A' && code <= 'F')
}

// hasBadBracket is RE_BAD_BRACKET.test(content), /.[\r\n"'(/\\]/: some unit that is not a line
// terminator (what `.` will not match) followed by one of the class.
func hasBadBracket(units []uint16) bool {
	for index := 0; index+1 < len(units); index++ {
		switch units[index] {
		case '\n', '\r', 0x2028, 0x2029:
			continue
		}
		switch units[index+1] {
		case '\r', '\n', '"', '\'', '(', '/', '\\':
			return true
		}
	}
	return false
}

type tokenizer struct {
	input *input
	css   []uint16

	length       int
	pos          int
	buffer       []token
	returned     []token
	lastBadParen int
}

func newTokenizer(in *input) *tokenizer {
	return &tokenizer{
		input:        in,
		css:          in.units,
		length:       len(in.units),
		pos:          0,
		buffer:       []token{},
		returned:     []token{},
		lastBadParen: -1,
	}
}

// charCodeAt is css.charCodeAt(index), -1 standing for NaN past either end.
func (tokens *tokenizer) charCodeAt(index int) int {
	if index < 0 || index >= tokens.length {
		return -1
	}
	return int(tokens.css[index])
}

// indexOf is css.indexOf(search, from) for an ASCII search string.
func (tokens *tokenizer) indexOf(search string, from int) int {
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

func (tokens *tokenizer) position() int {
	return tokens.pos
}

func (tokens *tokenizer) unclosed(what string) {
	panic(tokens.input.errorAt("Unclosed "+what, tokens.pos))
}

func (tokens *tokenizer) endOfFile() bool {
	return len(tokens.returned) == 0 && tokens.pos >= tokens.length
}

// nextToken returns the next token, and false where upstream returns undefined.
func (tokens *tokenizer) nextToken() (token, bool) {
	if len(tokens.returned) > 0 {
		last := tokens.returned[len(tokens.returned)-1]
		tokens.returned = tokens.returned[:len(tokens.returned)-1]
		return last, true
	}
	if tokens.pos >= tokens.length {
		return token{}, false
	}

	var currentToken token
	var next int
	pos := tokens.pos
	code := tokens.charCodeAt(pos)

	switch code {
	case newline, space, tab, carriageReturn, feed:
		next = pos
		for {
			next += 1
			code = tokens.charCodeAt(next)
			if !(code == space || code == newline || code == tab || code == carriageReturn || code == feed) {
				break
			}
		}

		currentToken = token{kind: "space", value: tokens.input.slice(pos, next), start: -1, end: -1}
		pos = next - 1

	case openSquare, closeSquare, openCurly, closeCurly, colon, semicolon, closeParentheses:
		controlChar := string(rune(code))
		currentToken = token{kind: controlChar, value: controlChar, start: pos, end: -1}

	case openParentheses:
		prev := ""
		if len(tokens.buffer) > 0 {
			prev = tokens.buffer[len(tokens.buffer)-1].value
			tokens.buffer = tokens.buffer[:len(tokens.buffer)-1]
		}
		n := tokens.charCodeAt(pos + 1)
		if prev == "url" &&
			n != singleQuote &&
			n != doubleQuote &&
			n != space &&
			n != newline &&
			n != tab &&
			n != feed &&
			n != carriageReturn {
			next = pos
			for {
				escaped := false
				next = tokens.indexOf(")", next+1)
				if next == -1 {
					// ignore and ignoreUnclosed are always false here.
					tokens.pos = pos
					tokens.unclosed("bracket")
				}
				escapePos := next
				for tokens.charCodeAt(escapePos-1) == backslash {
					escapePos -= 1
					escaped = !escaped
				}
				if !escaped {
					break
				}
			}

			currentToken = token{kind: "brackets", value: tokens.input.slice(pos, next+1), start: pos, end: next}

			pos = next
		} else if pos <= tokens.lastBadParen {
			currentToken = token{kind: "(", value: "(", start: pos, end: -1}
		} else {
			next = tokens.indexOf(")", pos+1)
			var contentUnits []uint16
			if next != -1 {
				contentUnits = tokens.css[pos : next+1]
			}

			if next == -1 || hasBadBracket(contentUnits) {
				if next == -1 {
					tokens.lastBadParen = tokens.length
				} else {
					tokens.lastBadParen = next
				}
				currentToken = token{kind: "(", value: "(", start: pos, end: -1}
			} else {
				currentToken = token{kind: "brackets", value: tokens.input.slice(pos, next+1), start: pos, end: next}
				pos = next
			}
		}

	case singleQuote, doubleQuote:
		quote := "'"
		if code == doubleQuote {
			quote = "\""
		}
		next = pos
		for {
			escaped := false
			next = tokens.indexOf(quote, next+1)
			if next == -1 {
				// ignore and ignoreUnclosed are always false here.
				tokens.pos = pos
				tokens.unclosed("string")
			}
			escapePos := next
			for tokens.charCodeAt(escapePos-1) == backslash {
				escapePos -= 1
				escaped = !escaped
			}
			if !escaped {
				break
			}
		}

		currentToken = token{kind: "string", value: tokens.input.slice(pos, next+1), start: pos, end: next}
		pos = next

	case at:
		// RE_AT_END.lastIndex = pos + 1; RE_AT_END.test(css): a match at index m leaves lastIndex m + 1,
		// and no match leaves it 0.
		match := -1
		for index := pos + 1; index < tokens.length; index++ {
			if isAtEnd(tokens.css[index]) {
				match = index
				break
			}
		}
		if match == -1 {
			next = tokens.length - 1
		} else {
			next = match - 1
		}

		currentToken = token{kind: "at-word", value: tokens.input.slice(pos, next+1), start: pos, end: next}

		pos = next

	case backslash:
		next = pos
		escape := true
		for tokens.charCodeAt(next+1) == backslash {
			next += 1
			escape = !escape
		}
		code = tokens.charCodeAt(next + 1)
		if escape &&
			code != slash &&
			code != space &&
			code != newline &&
			code != tab &&
			code != carriageReturn &&
			code != feed {
			next += 1
			if isHexEscape(tokens.charCodeAt(next)) {
				for isHexEscape(tokens.charCodeAt(next + 1)) {
					next += 1
				}
				if tokens.charCodeAt(next+1) == space {
					next += 1
				}
			}
		}

		currentToken = token{kind: "word", value: tokens.input.slice(pos, next+1), start: pos, end: next}

		pos = next

	default:
		if code == slash && tokens.charCodeAt(pos+1) == asterisk {
			next = tokens.indexOf("*/", pos+2) + 1
			if next == 0 {
				// ignore and ignoreUnclosed are always false here.
				tokens.pos = pos
				tokens.unclosed("comment")
			}

			currentToken = token{kind: "comment", value: tokens.input.slice(pos, next+1), start: pos, end: next}
			pos = next
		} else {
			// RE_WORD_END.lastIndex = pos + 1; RE_WORD_END.test(css).
			match := -1
			for index := pos + 1; index < tokens.length; index++ {
				unit := tokens.css[index]
				if isWordEnd(unit) || (unit == slash && tokens.charCodeAt(index+1) == asterisk) {
					match = index
					break
				}
			}
			if match == -1 {
				next = tokens.length - 1
			} else {
				next = match - 1
			}

			currentToken = token{kind: "word", value: tokens.input.slice(pos, next+1), start: pos, end: next}
			tokens.buffer = append(tokens.buffer, currentToken)
			pos = next
		}
	}

	pos++
	tokens.pos = pos
	return currentToken, true
}

func (tokens *tokenizer) back(each token) {
	tokens.returned = append(tokens.returned, each)
}
