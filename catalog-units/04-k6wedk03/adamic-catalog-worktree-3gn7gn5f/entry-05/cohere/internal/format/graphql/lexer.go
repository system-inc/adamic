package graphql

import (
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// graphql-js 17.0.2, language/lexer.js, with error/syntaxError.js and language/location.js's
// getLocation, which give a lexing or parsing failure its message and position.
//
// # Bytes where graphql-js has UTF-16
//
// graphql-js walks the source by UTF-16 code unit with charCodeAt. This lexer walks the UTF-8 text by
// byte: charCodeAt below is the byte (or -1 past the end), which answers every ASCII test the lexer
// makes exactly, and where graphql-js accepts a SourceCharacter (a Unicode scalar value, or a surrogate
// pair spelling a supplementary one) the Go side decodes the whole rune and steps over its bytes. A rune
// decoded from UTF-8 is always a scalar value, so the surrogate branches graphql-js needs for its
// JavaScript strings collapse into the scalar one; they are kept where upstream has them, unreachable.
//
// Text that is not valid UTF-8 decodes one U+FFFD per bad byte, a scalar value, which is what Node's
// UTF-8 decoding hands graphql-js for a file in most cases (the cooked value keeps the raw bytes, where
// Node would have U+FFFD in the string).
//
// Token offsets are bytes. Error positions are bytes too, and an error's line and column are computed
// as graphql-js's getLocation computes them, the column in UTF-16 units, so the message matches.

// SyntaxError is graphql-js's GraphQLError from syntaxError(source, position, description), as Prettier
// reports it: createParseError reads the error's first location into Prettier's createError, whose
// message appends " (line:column)".
type SyntaxError struct {
	// Description is syntaxError's description, without graphql-js's "Syntax Error: " prefix.
	Description string
	// Position is the byte offset in the text where graphql-js's error points.
	Position int
	// Line and Column are graphql-js's error.locations[0]: 1-indexed, the column in UTF-16 code units.
	Line   int
	Column int
}

// Error is Prettier's message for the error: GraphQLError's message, then the location.
func (syntaxError *SyntaxError) Error() string {
	return fmt.Sprintf("Syntax Error: %s (%d:%d)", syntaxError.Description, syntaxError.Line, syntaxError.Column)
}

// syntaxError produces a GraphQLError representing a syntax error, containing useful descriptive
// information about the syntax error's position in the source. The lexer and parser throw it, so here
// they panic with it, and Parse recovers it into its error return: threading an error through every
// recursive descent below would bury upstream's shape.
func syntaxError(body string, position int, description string) *SyntaxError {
	line, column := getLocation(body, position)
	return &SyntaxError{Description: description, Position: position, Line: line, Column: column}
}

// getLocation takes a Source and a UTF-8 character offset, and returns the corresponding line and
// column as a SourceLocation. Lines are split by /\r\n|[\n\r]/g, and the column counts UTF-16 units from
// the line's start, as graphql-js's does.
func getLocation(body string, position int) (int, int) {
	lastLineStart := 0
	line := 1
	for index := 0; index < len(body); index++ {
		if body[index] != '\n' && body[index] != '\r' {
			continue
		}
		if index >= position {
			break
		}
		length := 1
		if body[index] == '\r' && index+1 < len(body) && body[index+1] == '\n' {
			length = 2
		}
		lastLineStart = index + length
		line++
		index += length - 1
	}
	// position + 1 - lastLineStart, in code units. A line start past the position (an error pointing at
	// the "\n" of a "\r\n") gives graphql-js's 0.
	if lastLineStart > position {
		return line, 1 + position - lastLineStart
	}
	column := 1
	for _, character := range body[lastLineStart:min(position, len(body))] {
		column += utf16.RuneLen(character)
	}
	return line, column
}

// lexer is given a Source object and creates a Lexer for that source. A Lexer is a stateful stream
// generator in that every time it is advanced, it returns the next token in the Source. Assuming the
// source lexes, the final Token emitted by the lexer will be of kind EOF, after which the lexer will
// repeatedly return the same EOF token whenever called.
type lexer struct {
	// Source document used to derive error locations.
	body string

	// Most recent non-ignored token returned by the lexer.
	lastToken *token

	// Current non-ignored token at the lexer cursor.
	token *token

	// The (1-indexed) line containing the current token.
	line int

	// Character offset where the current line starts.
	lineStart int
}

func newLexer(body string) *lexer {
	startOfFileToken := newToken(tokenKindSOF, 0, 0, 0, 0)

	return &lexer{
		body:      body,
		lastToken: startOfFileToken,
		token:     startOfFileToken,
		line:      1,
		lineStart: 0,
	}
}

// advance advances the token stream to the next non-ignored token.
func (lexer *lexer) advance() *token {
	lexer.lastToken = lexer.token
	lexer.token = lexer.lookahead()
	return lexer.token
}

// lookahead looks ahead and returns the next non-ignored token, but does not change the state of Lexer.
func (lexer *lexer) lookahead() *token {
	token := lexer.token
	if token.kind != tokenKindEOF {
		for {
			if token.next != nil {
				token = token.next
			} else {
				// Read the next token and form a link in the token linked-list.
				nextToken := readNextToken(lexer, token.end)
				token.next = nextToken
				nextToken.prev = token
				token = nextToken
			}
			if token.kind != tokenKindComment {
				break
			}
		}
	}
	return token
}

func isPunctuatorTokenKind(kind tokenKind) bool {
	return kind == tokenKindBang ||
		kind == tokenKindDollar ||
		kind == tokenKindAmp ||
		kind == tokenKindParenL ||
		kind == tokenKindParenR ||
		kind == tokenKindDot ||
		kind == tokenKindSpread ||
		kind == tokenKindColon ||
		kind == tokenKindEquals ||
		kind == tokenKindAt ||
		kind == tokenKindBracketL ||
		kind == tokenKindBracketR ||
		kind == tokenKindBraceL ||
		kind == tokenKindPipe ||
		kind == tokenKindBraceR
}

// charCodeAt is body.charCodeAt(position) for the lexer's ASCII tests: the byte there, or -1 (NaN in
// graphql-js) past the end.
func charCodeAt(body string, position int) int {
	if position < 0 || position >= len(body) {
		return -1
	}
	return int(body[position])
}

// codePointAt decodes the whole character at a byte offset, with its size in bytes. graphql-js reads a
// SourceCharacter as one code unit or a surrogate pair; here it is one rune of one to four bytes.
func codePointAt(body string, position int) (int, int) {
	character, size := utf8.DecodeRuneInString(body[position:])
	return int(character), size
}

// sliceUnits is body.slice(start, start + units) where units counts UTF-16 code units, for the error
// messages that quote a span of the source. A supplementary character the span cuts in half leaves
// graphql-js's message holding a lone leading surrogate, which no UTF-8 string can hold; it becomes
// U+FFFD, which is what Node writes for it when the message leaves as UTF-8.
func sliceUnits(body string, start int, units int) string {
	end := start
	for end < len(body) {
		character, size := utf8.DecodeRuneInString(body[end:])
		if units < utf16.RuneLen(character) {
			if units > 0 {
				return body[start:end] + string(utf8.RuneError)
			}
			break
		}
		end += size
		units -= utf16.RuneLen(character)
	}
	return body[start:end]
}

// isUnicodeScalarValue: a Unicode scalar value is any Unicode code point except surrogate code points.
// In other words, the inclusive ranges of values 0x0000 to 0xD7FF and 0xE000 to 0x10FFFF.
//
// SourceCharacter ::
//   - "Any Unicode scalar value"
func isUnicodeScalarValue(code int) bool {
	return (code >= 0x0000 && code <= 0xd7ff) || (code >= 0xe000 && code <= 0x10ffff)
}

// isSupplementaryCodePoint: the GraphQL specification defines source text as a sequence of unicode
// scalar values (which Unicode defines to exclude surrogate code points). However JavaScript defines
// strings as a sequence of UTF-16 code units which may include surrogates. A surrogate pair is a valid
// source character as it encodes a supplementary code point (above U+FFFF), but unpaired surrogate
// code points are not valid source characters.
//
// UTF-8 text holds no surrogates, so this is false for every position in Go; it stays so the branches
// that call it read as upstream's.
func isSupplementaryCodePoint(body string, location int) bool {
	return isLeadingSurrogate(charCodeAt(body, location)) &&
		isTrailingSurrogate(charCodeAt(body, location+1))
}

func isLeadingSurrogate(code int) bool {
	return code >= 0xd800 && code <= 0xdbff
}

func isTrailingSurrogate(code int) bool {
	return code >= 0xdc00 && code <= 0xdfff
}

// printCodePointAt prints the code point (or end of file reference) at a given location in a source for
// use in error messages.
//
// Printable ASCII is printed quoted, while other points are printed in Unicode code point form (ie.
// U+1234).
func printCodePointAt(lexer *lexer, location int) string {
	if location >= len(lexer.body) {
		return string(tokenKindEOF)
	}
	code, _ := codePointAt(lexer.body, location)
	if code >= 0x0020 && code <= 0x007e {
		// Printable ASCII
		character := string(rune(code))
		if character == `"` {
			return `'"'`
		}
		return `"` + character + `"`
	}

	// Unicode code point
	return fmt.Sprintf("U+%04X", code)
}

// createToken creates a token with line and column location information.
func createToken(lexer *lexer, kind tokenKind, start int, end int, value ...string) *token {
	line := lexer.line
	column := 1 + start - lexer.lineStart
	return newToken(kind, start, end, line, column, value...)
}

// readNextToken gets the next token from the source starting at the given position.
//
// This skips over whitespace until it finds the next lexable token, then lexes punctuators immediately
// or calls the appropriate helper function for more complicated tokens.
func readNextToken(lexer *lexer, start int) *token {
	body := lexer.body
	bodyLength := len(body)
	position := start

	for position < bodyLength {
		code := charCodeAt(body, position)

		// SourceCharacter
		switch code {
		// Ignored ::
		//   - UnicodeBOM
		//   - WhiteSpace
		//   - LineTerminator
		//   - Comment
		//   - Comma
		//
		// UnicodeBOM :: "Byte Order Mark (U+FEFF)"
		//
		// WhiteSpace ::
		//   - "Horizontal Tab (U+0009)"
		//   - "Space (U+0020)"
		//
		// Comma :: ,
		case 0xef: // <BOM>, three bytes in UTF-8, where graphql-js's 0xfeff is one code unit
			if character, size := codePointAt(body, position); character == 0xfeff {
				position += size
				continue
			}
		case 0x0009, // \t
			0x0020, // <space>
			0x002c: // ,
			position++
			continue
		// LineTerminator ::
		//   - "New Line (U+000A)"
		//   - "Carriage Return (U+000D)" [lookahead != "New Line (U+000A)"]
		//   - "Carriage Return (U+000D)" "New Line (U+000A)"
		case 0x000a: // \n
			position++
			lexer.line++
			lexer.lineStart = position
			continue
		case 0x000d: // \r
			if charCodeAt(body, position+1) == 0x000a {
				position += 2
			} else {
				position++
			}
			lexer.line++
			lexer.lineStart = position
			continue
		// Comment
		case 0x0023: // #
			return readComment(lexer, position)
		// Token ::
		//   - Punctuator
		//   - Name
		//   - IntValue
		//   - FloatValue
		//   - StringValue
		//
		// Punctuator :: one of ! $ & ( ) ... : = @ [ ] { | }
		case 0x0021: // !
			return createToken(lexer, tokenKindBang, position, position+1)
		case 0x0024: // $
			return createToken(lexer, tokenKindDollar, position, position+1)
		case 0x0026: // &
			return createToken(lexer, tokenKindAmp, position, position+1)
		case 0x0028: // (
			return createToken(lexer, tokenKindParenL, position, position+1)
		case 0x0029: // )
			return createToken(lexer, tokenKindParenR, position, position+1)
		case 0x002e: // .
			nextCode := charCodeAt(body, position+1)
			if nextCode == 0x002e && charCodeAt(body, position+2) == 0x002e {
				return createToken(lexer, tokenKindSpread, position, position+3)
			}
			if nextCode == 0x002e {
				panic(syntaxError(body, position, `Unexpected "..", did you mean "..."?`))
			} else if isDigit(nextCode) {
				digits := body[position+1 : readDigits(lexer, position+1, nextCode)]
				panic(syntaxError(body, position, `Invalid number, expected digit before ".", did you mean "0.`+digits+`"?`))
			}
		case 0x003a: // :
			return createToken(lexer, tokenKindColon, position, position+1)
		case 0x003d: // =
			return createToken(lexer, tokenKindEquals, position, position+1)
		case 0x0040: // @
			return createToken(lexer, tokenKindAt, position, position+1)
		case 0x005b: // [
			return createToken(lexer, tokenKindBracketL, position, position+1)
		case 0x005d: // ]
			return createToken(lexer, tokenKindBracketR, position, position+1)
		case 0x007b: // {
			return createToken(lexer, tokenKindBraceL, position, position+1)
		case 0x007c: // |
			return createToken(lexer, tokenKindPipe, position, position+1)
		case 0x007d: // }
			return createToken(lexer, tokenKindBraceR, position, position+1)
		// StringValue
		case 0x0022: // "
			if charCodeAt(body, position+1) == 0x0022 &&
				charCodeAt(body, position+2) == 0x0022 {
				return readBlockString(lexer, position)
			}
			return readString(lexer, position)
		}

		// IntValue | FloatValue (Digit | -)
		if isDigit(code) || code == 0x002d {
			return readNumber(lexer, position, code)
		}

		// Name
		if isNameStart(code) {
			return readName(lexer, position)
		}

		character, _ := codePointAt(body, position)
		switch {
		case code == 0x0027:
			panic(syntaxError(body, position, `Unexpected single quote character ('), did you mean to use a double quote (")?`))
		case isUnicodeScalarValue(character) || isSupplementaryCodePoint(body, position):
			panic(syntaxError(body, position, "Unexpected character: "+printCodePointAt(lexer, position)+"."))
		default:
			panic(syntaxError(body, position, "Invalid character: "+printCodePointAt(lexer, position)+"."))
		}
	}

	return createToken(lexer, tokenKindEOF, bodyLength, bodyLength)
}

// readComment reads a comment token from the source file.
//
//	Comment :: # CommentChar* [lookahead != CommentChar]
//
//	CommentChar :: SourceCharacter but not LineTerminator
func readComment(lexer *lexer, start int) *token {
	body := lexer.body
	bodyLength := len(body)
	position := start + 1

	for position < bodyLength {
		code := charCodeAt(body, position)

		// LineTerminator (\n | \r)
		if code == 0x000a || code == 0x000d {
			break
		}

		// SourceCharacter
		if character, size := codePointAt(body, position); isUnicodeScalarValue(character) {
			position += size
		} else if isSupplementaryCodePoint(body, position) {
			position += 2
		} else {
			break
		}
	}

	return createToken(lexer, tokenKindComment, start, position, body[start+1:position])
}

// readNumber reads a number token from the source file, either a FloatValue or an IntValue depending on
// whether a FractionalPart or ExponentPart is encountered.
//
//	IntValue :: IntegerPart [lookahead != {Digit, `.`, NameStart}]
//
//	IntegerPart ::
//	  - NegativeSign? 0
//	  - NegativeSign? NonZeroDigit Digit*
//
//	NegativeSign :: -
//
//	NonZeroDigit :: Digit but not `0`
//
//	FloatValue ::
//	  - IntegerPart FractionalPart ExponentPart [lookahead != {Digit, `.`, NameStart}]
//	  - IntegerPart FractionalPart [lookahead != {Digit, `.`, NameStart}]
//	  - IntegerPart ExponentPart [lookahead != {Digit, `.`, NameStart}]
//
//	FractionalPart :: . Digit+
//
//	ExponentPart :: ExponentIndicator Sign? Digit+
//
//	ExponentIndicator :: one of `e` `E`
//
//	Sign :: one of + -
func readNumber(lexer *lexer, start int, firstCode int) *token {
	body := lexer.body
	position := start
	code := firstCode
	isFloat := false

	// NegativeSign (-)
	if code == 0x002d {
		position++
		code = charCodeAt(body, position)
	}

	// Zero (0)
	if code == 0x0030 {
		position++
		code = charCodeAt(body, position)
		if isDigit(code) {
			panic(syntaxError(body, position, "Invalid number, unexpected digit after 0: "+printCodePointAt(lexer, position)+"."))
		}
	} else {
		position = readDigits(lexer, position, code)
		code = charCodeAt(body, position)
	}

	// Full stop (.)
	if code == 0x002e {
		isFloat = true

		position++
		code = charCodeAt(body, position)
		position = readDigits(lexer, position, code)
		code = charCodeAt(body, position)
	}

	// E e
	if code == 0x0045 || code == 0x0065 {
		isFloat = true

		position++
		code = charCodeAt(body, position)
		// + -
		if code == 0x002b || code == 0x002d {
			position++
			code = charCodeAt(body, position)
		}
		position = readDigits(lexer, position, code)
		code = charCodeAt(body, position)
	}

	// Numbers cannot be followed by . or NameStart
	if code == 0x002e || isNameStart(code) {
		panic(syntaxError(body, position, "Invalid number, expected digit but got: "+printCodePointAt(lexer, position)+"."))
	}

	kind := tokenKindInt
	if isFloat {
		kind = tokenKindFloat
	}
	return createToken(lexer, kind, start, position, body[start:position])
}

// readDigits returns the new position in the source after reading one or more digits.
func readDigits(lexer *lexer, start int, firstCode int) int {
	if !isDigit(firstCode) {
		panic(syntaxError(lexer.body, start, "Invalid number, expected digit but got: "+printCodePointAt(lexer, start)+"."))
	}

	body := lexer.body
	position := start + 1 // +1 to skip first firstCode

	for isDigit(charCodeAt(body, position)) {
		position++
	}

	return position
}

// readString reads a single-quote string token from the source file.
//
//	StringValue ::
//	  - `""` [lookahead != `"`]
//	  - `"` StringCharacter+ `"`
//
//	StringCharacter ::
//	  - SourceCharacter but not `"` or `\` or LineTerminator
//	  - `\u` EscapedUnicode
//	  - `\` EscapedCharacter
//
//	EscapedUnicode ::
//	  - `{` HexDigit+ `}`
//	  - HexDigit HexDigit HexDigit HexDigit
//
//	EscapedCharacter :: one of `"` `\` `/` `b` `f` `n` `r` `t`
func readString(lexer *lexer, start int) *token {
	body := lexer.body
	bodyLength := len(body)
	position := start + 1
	chunkStart := position
	value := ""

	for position < bodyLength {
		code := charCodeAt(body, position)

		// Closing Quote (")
		if code == 0x0022 {
			value += body[chunkStart:position]
			return createToken(lexer, tokenKindString, start, position+1, value)
		}

		// Escape Sequence (\)
		if code == 0x005c {
			value += body[chunkStart:position]
			var escape escapeSequence
			if charCodeAt(body, position+1) == 0x0075 { // u
				if charCodeAt(body, position+2) == 0x007b { // {
					escape = readEscapedUnicodeVariableWidth(lexer, position)
				} else {
					escape = readEscapedUnicodeFixedWidth(lexer, position)
				}
			} else {
				escape = readEscapedCharacter(lexer, position)
			}
			value += escape.value
			position += escape.size
			chunkStart = position
			continue
		}

		// LineTerminator (\n | \r)
		if code == 0x000a || code == 0x000d {
			break
		}

		// SourceCharacter
		if character, size := codePointAt(body, position); isUnicodeScalarValue(character) {
			position += size
		} else if isSupplementaryCodePoint(body, position) {
			position += 2
		} else {
			panic(syntaxError(body, position, "Invalid character within String: "+printCodePointAt(lexer, position)+"."))
		}
	}

	panic(syntaxError(body, position, "Unterminated string."))
}

// escapeSequence is the string value and lexed size of an escape sequence. Every escape is ASCII, so
// its size is the same in bytes as in graphql-js's code units.
type escapeSequence struct {
	value string
	size  int
}

func readEscapedUnicodeVariableWidth(lexer *lexer, position int) escapeSequence {
	body := lexer.body
	// graphql-js's point is a JavaScript bitwise result, a signed 32-bit integer: the shift that pushes a
	// hex digit into the sign bit is what makes it negative.
	var point int32
	size := 3
	// Cannot be larger than 12 chars (\u{00000000}).
	for size < 12 {
		code := charCodeAt(body, position+size)
		size++
		// Closing Brace (})
		if code == 0x007d {
			// Must be at least 5 chars (\u{0}) and encode a Unicode scalar value.
			if size < 5 || !isUnicodeScalarValue(int(point)) {
				break
			}
			return escapeSequence{value: string(rune(point)), size: size}
		}
		// Append this hex digit to the code point.
		point = (point << 4) | int32(readHexDigit(code))
		if point < 0 {
			break
		}
	}

	panic(syntaxError(body, position, `Invalid Unicode escape sequence: "`+sliceUnits(body, position, size)+`".`))
}

func readEscapedUnicodeFixedWidth(lexer *lexer, position int) escapeSequence {
	body := lexer.body
	code := read16BitHexCode(body, position+2)

	if isUnicodeScalarValue(code) {
		return escapeSequence{value: string(rune(code)), size: 6}
	}

	// GraphQL allows JSON-style surrogate pair escape sequences, but only when a valid pair is formed.
	if isLeadingSurrogate(code) {
		// \u
		if charCodeAt(body, position+6) == 0x005c &&
			charCodeAt(body, position+7) == 0x0075 {
			trailingCode := read16BitHexCode(body, position+8)
			if isTrailingSurrogate(trailingCode) {
				// JavaScript defines strings as a sequence of UTF-16 code units and encodes Unicode code
				// points above U+FFFF using a surrogate pair of code units. Since this is a surrogate pair
				// escape sequence, just include both codes into the JavaScript string value. Had JavaScript
				// not been internally based on UTF-16, then this surrogate pair would be decoded to
				// retrieve the supplementary code point.
				//
				// Go's strings are UTF-8, so here it is decoded.
				return escapeSequence{value: string(utf16.DecodeRune(rune(code), rune(trailingCode))), size: 12}
			}
		}
	}

	panic(syntaxError(body, position, `Invalid Unicode escape sequence: "`+sliceUnits(body, position, 6)+`".`))
}

// read16BitHexCode reads four hexadecimal characters and returns the positive integer that 16bit
// hexadecimal string represents. For example, "000f" will return 15, and "dead" will return 57005.
//
// Returns a negative number if any char was not a valid hexadecimal digit.
func read16BitHexCode(body string, position int) int {
	// readHexDigit() returns -1 on error. ORing a negative value with any other value always produces a
	// negative value.
	return (readHexDigit(charCodeAt(body, position)) << 12) |
		(readHexDigit(charCodeAt(body, position+1)) << 8) |
		(readHexDigit(charCodeAt(body, position+2)) << 4) |
		readHexDigit(charCodeAt(body, position+3))
}

// readHexDigit reads a hexadecimal character and returns its positive integer value (0-15).
//
// '0' becomes 0, '9' becomes 9
// 'A' becomes 10, 'F' becomes 15
// 'a' becomes 10, 'f' becomes 15
//
// Returns -1 if the provided character code was not a valid hexadecimal digit.
//
// HexDigit :: one of
//   - `0` `1` `2` `3` `4` `5` `6` `7` `8` `9`
//   - `A` `B` `C` `D` `E` `F`
//   - `a` `b` `c` `d` `e` `f`
func readHexDigit(code int) int {
	switch {
	case code >= 0x0030 && code <= 0x0039: // 0-9
		return code - 0x0030
	case code >= 0x0041 && code <= 0x0046: // A-F
		return code - 0x0037
	case code >= 0x0061 && code <= 0x0066: // a-f
		return code - 0x0057
	}
	return -1
}

// readEscapedCharacter:
//
// | Escaped Character | Code Point | Character Name               |
// | ----------------- | ---------- | ---------------------------- |
// | `"`               | U+0022     | double quote                 |
// | `\`               | U+005C     | reverse solidus (back slash) |
// | `/`               | U+002F     | solidus (forward slash)      |
// | `b`               | U+0008     | backspace                    |
// | `f`               | U+000C     | form feed                    |
// | `n`               | U+000A     | line feed (new line)         |
// | `r`               | U+000D     | carriage return              |
// | `t`               | U+0009     | horizontal tab               |
func readEscapedCharacter(lexer *lexer, position int) escapeSequence {
	body := lexer.body
	code := charCodeAt(body, position+1)
	switch code {
	case 0x0022: // "
		return escapeSequence{value: "\x22", size: 2}
	case 0x005c: // \
		return escapeSequence{value: "\x5c", size: 2}
	case 0x002f: // /
		return escapeSequence{value: "/", size: 2}
	case 0x0062: // b
		return escapeSequence{value: "\b", size: 2}
	case 0x0066: // f
		return escapeSequence{value: "\f", size: 2}
	case 0x006e: // n
		return escapeSequence{value: "\n", size: 2}
	case 0x0072: // r
		return escapeSequence{value: "\r", size: 2}
	case 0x0074: // t
		return escapeSequence{value: "\t", size: 2}
	}
	panic(syntaxError(body, position, `Invalid character escape sequence: "`+sliceUnits(body, position, 2)+`".`))
}

// readBlockString reads a block string token from the source file.
//
//	StringValue ::
//	  - `"""` BlockStringCharacter* `"""`
//
//	BlockStringCharacter ::
//	  - SourceCharacter but not `"""` or `\"""`
//	  - `\"""`
func readBlockString(lexer *lexer, start int) *token {
	body := lexer.body
	bodyLength := len(body)
	lineStart := lexer.lineStart

	position := start + 3
	chunkStart := position
	currentLine := ""

	var blockLines []string
	for position < bodyLength {
		code := charCodeAt(body, position)

		// Closing Triple-Quote (""")
		if code == 0x0022 &&
			charCodeAt(body, position+1) == 0x0022 &&
			charCodeAt(body, position+2) == 0x0022 {
			currentLine += body[chunkStart:position]
			blockLines = append(blockLines, currentLine)

			token := createToken(
				lexer,
				tokenKindBlockString,
				start,
				position+3,
				// Return a string of the lines joined with U+000A.
				strings.Join(dedentBlockStringLines(blockLines), "\n"),
			)

			lexer.line += len(blockLines) - 1
			lexer.lineStart = lineStart
			return token
		}

		// Escaped Triple-Quote (\""")
		if code == 0x005c &&
			charCodeAt(body, position+1) == 0x0022 &&
			charCodeAt(body, position+2) == 0x0022 &&
			charCodeAt(body, position+3) == 0x0022 {
			currentLine += body[chunkStart:position]
			chunkStart = position + 1 // skip only slash
			position += 4
			continue
		}

		// LineTerminator
		if code == 0x000a || code == 0x000d {
			currentLine += body[chunkStart:position]
			blockLines = append(blockLines, currentLine)

			if code == 0x000d && charCodeAt(body, position+1) == 0x000a {
				position += 2
			} else {
				position++
			}

			currentLine = ""
			chunkStart = position
			lineStart = position
			continue
		}

		// SourceCharacter
		if character, size := codePointAt(body, position); isUnicodeScalarValue(character) {
			position += size
		} else if isSupplementaryCodePoint(body, position) {
			position += 2
		} else {
			panic(syntaxError(body, position, "Invalid character within String: "+printCodePointAt(lexer, position)+"."))
		}
	}

	panic(syntaxError(body, position, "Unterminated string."))
}

// readName reads an alphanumeric + underscore name from the source.
//
//	Name ::
//	  - NameStart NameContinue* [lookahead != NameContinue]
func readName(lexer *lexer, start int) *token {
	body := lexer.body
	bodyLength := len(body)
	position := start + 1

	for position < bodyLength {
		code := charCodeAt(body, position)
		if isNameContinue(code) {
			position++
		} else {
			break
		}
	}

	return createToken(lexer, tokenKindName, start, position, body[start:position])
}
