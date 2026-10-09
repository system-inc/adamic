package cst

// Ported from eemeli/yaml 2.9.0, dist/parse/lexer.js.

import "iter"

/*
START -> stream

stream
  directive -> line-end -> stream
  indent + line-end -> stream
  [else] -> line-start

line-end
  comment -> line-end
  newline -> .
  input-end -> END

line-start
  doc-start -> doc
  doc-end -> stream
  [else] -> indent -> block-start

block-start
  seq-item-start -> block-start
  explicit-key-start -> block-start
  map-value-start -> block-start
  [else] -> doc

doc
  line-end -> line-start
  spaces -> doc
  anchor -> doc
  tag -> doc
  flow-start -> flow -> doc
  flow-end -> error -> doc
  seq-item-start -> error -> doc
  explicit-key-start -> error -> doc
  map-value-start -> doc
  alias -> doc
  quote-start -> quoted-scalar -> doc
  block-scalar-header -> line-end -> block-scalar(min) -> line-start
  [else] -> plain-scalar(false, min) -> doc

flow
  line-end -> flow
  spaces -> flow
  anchor -> flow
  tag -> flow
  flow-start -> flow -> flow
  flow-end -> .
  seq-item-start -> error -> flow
  explicit-key-start -> flow
  map-value-start -> flow
  alias -> flow
  quote-start -> quoted-scalar -> flow
  comma -> flow
  [else] -> plain-scalar(true, 0) -> flow

quoted-scalar
  quote-end -> .
  [else] -> quoted-scalar

block-scalar(min)
  newline + peek(indent < min) -> .
  [else] -> block-scalar(min)

plain-scalar(is-flow, min)
  scalar-end(is-flow) -> .
  peek(newline + (indent < min)) -> .
  [else] -> plain-scalar(min)
*/

// isEmpty: undefined (-1), space, newline, carriage return or tab.
func isEmpty(ch int) bool {
	switch ch {
	case -1, ' ', '\n', '\r', '\t':
		return true
	default:
		return false
	}
}

// inSet is upstream's `set.has(ch)` over a set of ASCII characters; undefined is in no set.
func inSet(set string, ch int) bool {
	if ch < 0 || ch >= 0x80 {
		return false
	}
	for index := 0; index < len(set); index++ {
		if int(set[index]) == ch {
			return true
		}
	}
	return false
}

const hexDigits = "0123456789ABCDEFabcdef"
const tagChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-#;/?:@&=+$_.!~*'()"
const flowIndicatorChars = ",[]{}"
const invalidAnchorChars = " ,[]{}\n\r\t"

// isNotAnchorChar is `!ch || invalidAnchorChars.has(ch)`.
func isNotAnchorChar(ch int) bool { return ch == -1 || inSet(invalidAnchorChars, ch) }

// Lexer splits an input string into lexical tokens, i.e. smaller strings that are easily identifiable by
// TokenType.
//
// Lexing starts always in a "stream" context. Incomplete input may be buffered until a complete token can
// be emitted.
//
// In addition to slices of the original input, the following control characters may also be emitted:
//
//   - Document (Start of Text): A document starts with the next token
//   - FlowEnd (Cancel): Unexpected end of flow-mode (indicates an error)
//   - Scalar (Unit Separator): Next token is a scalar value
//   - BOM (Byte order mark): Emitted separately outside documents
type Lexer struct {
	// atEnd: flag indicating whether the end of the current buffer marks the end of all input.
	atEnd bool
	// blockScalarIndent: explicit indent set in block scalar header, as an offset from the current minimum
	// indent, so e.g. set to 1 from a header `|2+`. Set to -1 if not explicitly set.
	blockScalarIndent int
	// blockScalarKeep: block scalars that include a + (keep) chomping indicator in their header include
	// trailing empty lines, which are otherwise excluded from the scalar's contents.
	blockScalarKeep bool
	// buffer: current input.
	buffer []uint16
	// flowKey: flag noting whether the map value indicator : can immediately follow this node within a
	// flow context.
	flowKey bool
	// flowLevel: count of surrounding flow collection levels.
	flowLevel int
	// indentNext: minimum level of indentation required for next lines to be parsed as a part of the
	// current scalar value.
	indentNext int
	// indentValue: indentation level of the current line.
	indentValue int
	// lineEndPos: position of the next \n character; hasLineEndPos false is upstream's null.
	lineEndPos    int
	hasLineEndPos bool
	// next stores the state of the lexer if reaching the end of incomplete input; "" is null.
	next string
	// pos: a pointer to buffer; the current position of the lexer.
	pos int

	// yield is the consumer of the running Lex iteration; stopped is set once it returns false.
	yield   func([]uint16) bool
	stopped bool
}

// NewLexer is upstream's `new Lexer()`.
func NewLexer() *Lexer {
	return &Lexer{blockScalarIndent: -1}
}

// emit is upstream's `yield source`.
func (lexer *Lexer) emit(source []uint16) {
	if lexer.stopped {
		return
	}
	if !lexer.yield(source) {
		lexer.stopped = true
	}
}

// Lex generates YAML tokens from the source string. If incomplete, a part of the last line may be left as
// a buffer for the next call.
func (lexer *Lexer) Lex(source []uint16, incomplete bool) iter.Seq[[]uint16] {
	return func(yield func([]uint16) bool) {
		lexer.yield = yield
		lexer.stopped = false
		// `if (source)`: an empty string is falsy and leaves the buffer as it is.
		if len(source) > 0 {
			if len(lexer.buffer) > 0 {
				buffer := make([]uint16, 0, len(lexer.buffer)+len(source))
				buffer = append(buffer, lexer.buffer...)
				lexer.buffer = append(buffer, source...)
			} else {
				lexer.buffer = source[:len(source):len(source)]
			}
			lexer.hasLineEndPos = false
		}
		lexer.atEnd = !incomplete
		next := lexer.next
		if next == "" {
			next = "stream"
		}
		for next != "" && (incomplete || lexer.hasChars(1)) && !lexer.stopped {
			next = lexer.parseNext(next)
		}
	}
}

func (lexer *Lexer) atLineEnd() bool {
	i := lexer.pos
	ch := characterAt(lexer.buffer, i)
	for ch == ' ' || ch == '\t' {
		i++
		ch = characterAt(lexer.buffer, i)
	}
	if ch == -1 || ch == '#' || ch == '\n' {
		return true
	}
	if ch == '\r' {
		return characterAt(lexer.buffer, i+1) == '\n'
	}
	return false
}

func (lexer *Lexer) charAt(n int) int {
	return characterAt(lexer.buffer, lexer.pos+n)
}

func (lexer *Lexer) continueScalar(offset int) int {
	ch := characterAt(lexer.buffer, offset)
	if lexer.indentNext > 0 {
		indent := 0
		for ch == ' ' {
			indent++
			ch = characterAt(lexer.buffer, indent+offset)
		}
		if ch == '\r' {
			next := characterAt(lexer.buffer, indent+offset+1)
			if next == '\n' || (next == -1 && !lexer.atEnd) {
				return offset + indent + 1
			}
		}
		if ch == '\n' || indent >= lexer.indentNext || (ch == -1 && !lexer.atEnd) {
			return offset + indent
		}
		return -1
	}
	if ch == '-' || ch == '.' {
		dt := substr(lexer.buffer, offset, 3)
		if (equals(dt, "---") || equals(dt, "...")) && isEmpty(characterAt(lexer.buffer, offset+3)) {
			return -1
		}
	}
	return offset
}

// getLine returns false for upstream's null: no complete line is buffered yet.
func (lexer *Lexer) getLine() ([]uint16, bool) {
	end := lexer.lineEndPos
	// `typeof end !== 'number'` is the null lineEndPos.
	if !lexer.hasLineEndPos || (end != -1 && end < lexer.pos) {
		end = indexOf(lexer.buffer, '\n', lexer.pos)
		lexer.lineEndPos = end
		lexer.hasLineEndPos = true
	}
	if end == -1 {
		if lexer.atEnd {
			return substring(lexer.buffer, lexer.pos, len(lexer.buffer)), true
		}
		return nil, false
	}
	if characterAt(lexer.buffer, end-1) == '\r' {
		end--
	}
	return substring(lexer.buffer, lexer.pos, end), true
}

func (lexer *Lexer) hasChars(n int) bool {
	return lexer.pos+n <= len(lexer.buffer)
}

// setNext returns upstream's null, the state that ends the lex loop.
func (lexer *Lexer) setNext(state string) string {
	lexer.buffer = substring(lexer.buffer, lexer.pos, len(lexer.buffer))
	lexer.pos = 0
	lexer.hasLineEndPos = false
	lexer.next = state
	return ""
}

func (lexer *Lexer) peek(n int) []uint16 {
	return substr(lexer.buffer, lexer.pos, n)
}

func (lexer *Lexer) parseNext(next string) string {
	switch next {
	case "stream":
		return lexer.parseStream()
	case "line-start":
		return lexer.parseLineStart()
	case "block-start":
		return lexer.parseBlockStart()
	case "doc":
		return lexer.parseDocument()
	case "flow":
		return lexer.parseFlowCollection()
	case "quoted-scalar":
		return lexer.parseQuotedScalar()
	case "block-scalar":
		return lexer.parseBlockScalar()
	case "plain-scalar":
		return lexer.parsePlainScalar()
	}
	return ""
}

func (lexer *Lexer) parseStream() string {
	line, ok := lexer.getLine()
	if !ok {
		return lexer.setNext("stream")
	}
	if characterAt(line, 0) == int(BOM) {
		lexer.pushCount(1)
		line = substring(line, 1, len(line))
	}
	if characterAt(line, 0) == '%' {
		dirEnd := len(line)
		cs := indexOf(line, '#', 0)
		for cs != -1 {
			ch := characterAt(line, cs-1)
			if ch == ' ' || ch == '\t' {
				dirEnd = cs - 1
				break
			} else {
				cs = indexOf(line, '#', cs+1)
			}
		}
		for {
			ch := characterAt(line, dirEnd-1)
			if ch == ' ' || ch == '\t' {
				dirEnd--
			} else {
				break
			}
		}
		n := lexer.pushCount(dirEnd) + lexer.pushSpaces(true)
		lexer.pushCount(len(line) - n) // possible comment
		// Upstream calls `this.pushNewline()` here without `yield*`, which creates a generator and never
		// runs it: nothing is pushed and pos stays put. The next "stream" state emits the newline.
		return "stream"
	}
	if lexer.atLineEnd() {
		sp := lexer.pushSpaces(true)
		lexer.pushCount(len(line) - sp)
		lexer.pushNewline()
		return "stream"
	}
	lexer.emit([]uint16{Document})
	return lexer.parseLineStart()
}

func (lexer *Lexer) parseLineStart() string {
	ch := lexer.charAt(0)
	if ch == -1 && !lexer.atEnd {
		return lexer.setNext("line-start")
	}
	if ch == '-' || ch == '.' {
		if !lexer.atEnd && !lexer.hasChars(4) {
			return lexer.setNext("line-start")
		}
		s := lexer.peek(3)
		if (equals(s, "---") || equals(s, "...")) && isEmpty(lexer.charAt(3)) {
			lexer.pushCount(3)
			lexer.indentValue = 0
			lexer.indentNext = 0
			if equals(s, "---") {
				return "doc"
			}
			return "stream"
		}
	}
	lexer.indentValue = lexer.pushSpaces(false)
	if lexer.indentNext > lexer.indentValue && !isEmpty(lexer.charAt(1)) {
		lexer.indentNext = lexer.indentValue
	}
	return lexer.parseBlockStart()
}

func (lexer *Lexer) parseBlockStart() string {
	// `const [ch0, ch1] = this.peek(2)` destructures by code point, not code unit: when the two units
	// are a surrogate pair, ch0 is the whole astral character and ch1 is undefined.
	ch0, ch1 := -1, -1
	if peeked := lexer.peek(2); len(peeked) == 2 && isHighSurrogate(peeked[0]) && isLowSurrogate(peeked[1]) {
		ch0 = 0x10000 + (int(peeked[0])-0xD800)<<10 + int(peeked[1]) - 0xDC00
	} else {
		ch0 = characterAt(peeked, 0)
		ch1 = characterAt(peeked, 1)
	}
	if ch1 == -1 && !lexer.atEnd {
		return lexer.setNext("block-start")
	}
	if (ch0 == '-' || ch0 == '?' || ch0 == ':') && isEmpty(ch1) {
		n := lexer.pushCount(1) + lexer.pushSpaces(true)
		lexer.indentNext = lexer.indentValue + 1
		lexer.indentValue += n
		return "block-start"
	}
	return "doc"
}

func isHighSurrogate(unit uint16) bool { return unit >= 0xD800 && unit <= 0xDBFF }
func isLowSurrogate(unit uint16) bool  { return unit >= 0xDC00 && unit <= 0xDFFF }

func (lexer *Lexer) parseDocument() string {
	lexer.pushSpaces(true)
	line, ok := lexer.getLine()
	if !ok {
		return lexer.setNext("doc")
	}
	n := lexer.pushIndicators()
	switch characterAt(line, n) {
	case '#':
		lexer.pushCount(len(line) - n)
		fallthrough
	case -1:
		lexer.pushNewline()
		return lexer.parseLineStart()
	case '{', '[':
		lexer.pushCount(1)
		lexer.flowKey = false
		lexer.flowLevel = 1
		return "flow"
	case '}', ']':
		// this is an error
		lexer.pushCount(1)
		return "doc"
	case '*':
		lexer.pushUntil(isNotAnchorChar)
		return "doc"
	case '"', '\'':
		return lexer.parseQuotedScalar()
	case '|', '>':
		n += lexer.parseBlockScalarHeader()
		n += lexer.pushSpaces(true)
		lexer.pushCount(len(line) - n)
		lexer.pushNewline()
		return lexer.parseBlockScalar()
	default:
		return lexer.parsePlainScalar()
	}
}

func (lexer *Lexer) parseFlowCollection() string {
	var nl, sp int
	indent := -1
	for {
		nl = lexer.pushNewline()
		if nl > 0 {
			sp = lexer.pushSpaces(false)
			indent = sp
			lexer.indentValue = sp
		} else {
			sp = 0
		}
		sp += lexer.pushSpaces(true)
		if nl+sp <= 0 {
			break
		}
	}
	line, ok := lexer.getLine()
	if !ok {
		return lexer.setNext("flow")
	}
	if (indent != -1 && indent < lexer.indentNext && characterAt(line, 0) != '#') ||
		(indent == 0 &&
			(startsWith(line, "---") || startsWith(line, "...")) &&
			isEmpty(characterAt(line, 3))) {
		// Allowing for the terminal ] or } at the same (rather than greater)
		// indent level as the initial [ or { is technically invalid, but
		// failing here would be surprising to users.
		atFlowEndMarker := indent == lexer.indentNext-1 &&
			lexer.flowLevel == 1 &&
			(characterAt(line, 0) == ']' || characterAt(line, 0) == '}')
		if !atFlowEndMarker {
			// this is an error
			lexer.flowLevel = 0
			lexer.emit([]uint16{FlowEnd})
			return lexer.parseLineStart()
		}
	}
	n := 0
	for characterAt(line, n) == ',' {
		n += lexer.pushCount(1)
		n += lexer.pushSpaces(true)
		lexer.flowKey = false
	}
	n += lexer.pushIndicators()
	switch characterAt(line, n) {
	case -1:
		return "flow"
	case '#':
		lexer.pushCount(len(line) - n)
		return "flow"
	case '{', '[':
		lexer.pushCount(1)
		lexer.flowKey = false
		lexer.flowLevel++
		return "flow"
	case '}', ']':
		lexer.pushCount(1)
		lexer.flowKey = true
		lexer.flowLevel--
		if lexer.flowLevel != 0 {
			return "flow"
		}
		return "doc"
	case '*':
		lexer.pushUntil(isNotAnchorChar)
		return "flow"
	case '"', '\'':
		lexer.flowKey = true
		return lexer.parseQuotedScalar()
	case ':':
		next := lexer.charAt(1)
		if lexer.flowKey || isEmpty(next) || next == ',' {
			lexer.flowKey = false
			lexer.pushCount(1)
			lexer.pushSpaces(true)
			return "flow"
		}
		// fallthrough
		lexer.flowKey = false
		return lexer.parsePlainScalar()
	default:
		lexer.flowKey = false
		return lexer.parsePlainScalar()
	}
}

func (lexer *Lexer) parseQuotedScalar() string {
	quote := lexer.charAt(0)
	end := -1
	if quote != -1 {
		end = indexOf(lexer.buffer, uint16(quote), lexer.pos+1)
	}
	if quote == '\'' {
		for end != -1 && characterAt(lexer.buffer, end+1) == '\'' {
			end = indexOf(lexer.buffer, '\'', end+2)
		}
	} else {
		// double-quote
		for end != -1 {
			n := 0
			for characterAt(lexer.buffer, end-1-n) == '\\' {
				n++
			}
			if n%2 == 0 {
				break
			}
			end = indexOf(lexer.buffer, '"', end+1)
		}
	}
	// Only looking for newlines within the quotes. substring(0, -1) is the empty string, so an
	// unterminated quote looks for none.
	qb := substring(lexer.buffer, 0, end)
	nl := indexOf(qb, '\n', lexer.pos)
	if nl != -1 {
		for nl != -1 {
			cs := lexer.continueScalar(nl + 1)
			if cs == -1 {
				break
			}
			nl = indexOf(qb, '\n', cs)
		}
		if nl != -1 {
			// this is an error caused by an unexpected unindent
			if characterAt(qb, nl-1) == '\r' {
				end = nl - 2
			} else {
				end = nl - 1
			}
		}
	}
	if end == -1 {
		if !lexer.atEnd {
			return lexer.setNext("quoted-scalar")
		}
		end = len(lexer.buffer)
	}
	lexer.pushToIndex(end+1, false)
	if lexer.flowLevel != 0 {
		return "flow"
	}
	return "doc"
}

func (lexer *Lexer) parseBlockScalarHeader() int {
	lexer.blockScalarIndent = -1
	lexer.blockScalarKeep = false
	i := lexer.pos
	for {
		i++
		ch := characterAt(lexer.buffer, i)
		if ch == '+' {
			lexer.blockScalarKeep = true
		} else if ch > '0' && ch <= '9' {
			lexer.blockScalarIndent = ch - '0' - 1
		} else if ch != '-' {
			break
		}
	}
	return lexer.pushUntil(func(ch int) bool { return isEmpty(ch) || ch == '#' })
}

func (lexer *Lexer) parseBlockScalar() string {
	nl := lexer.pos - 1 // may be -1 if this.pos === 0
	indent := 0
	var ch int
loop:
	for i := lexer.pos; ; i++ {
		ch = characterAt(lexer.buffer, i)
		if ch == -1 {
			break
		}
		switch ch {
		case ' ':
			indent++
		case '\n':
			nl = i
			indent = 0
		case '\r':
			next := characterAt(lexer.buffer, i+1)
			if next == -1 && !lexer.atEnd {
				return lexer.setNext("block-scalar")
			}
			if next == '\n' {
				continue
			}
			// fallthrough
			break loop
		default:
			break loop
		}
	}
	if ch == -1 && !lexer.atEnd {
		return lexer.setNext("block-scalar")
	}
	if indent >= lexer.indentNext {
		if lexer.blockScalarIndent == -1 {
			lexer.indentNext = indent
		} else {
			base := lexer.indentNext
			if base == 0 {
				base = 1
			}
			lexer.indentNext = lexer.blockScalarIndent + base
		}
		for {
			cs := lexer.continueScalar(nl + 1)
			if cs == -1 {
				break
			}
			nl = indexOf(lexer.buffer, '\n', cs)
			if nl == -1 {
				break
			}
		}
		if nl == -1 {
			if !lexer.atEnd {
				return lexer.setNext("block-scalar")
			}
			nl = len(lexer.buffer)
		}
	}
	// Trailing insufficiently indented tabs are invalid.
	// To catch that during parsing, we include them in the block scalar value.
	i := nl + 1
	ch = characterAt(lexer.buffer, i)
	for ch == ' ' {
		i++
		ch = characterAt(lexer.buffer, i)
	}
	if ch == '\t' {
		for ch == '\t' || ch == ' ' || ch == '\r' || ch == '\n' {
			i++
			ch = characterAt(lexer.buffer, i)
		}
		nl = i - 1
	} else if !lexer.blockScalarKeep {
		for {
			i := nl - 1
			ch := characterAt(lexer.buffer, i)
			if ch == '\r' {
				i--
				ch = characterAt(lexer.buffer, i)
			}
			lastChar := i // Drop the line if last char not more indented
			for ch == ' ' {
				i--
				ch = characterAt(lexer.buffer, i)
			}
			if ch == '\n' && i >= lexer.pos && i+1+indent > lastChar {
				nl = i
			} else {
				break
			}
		}
	}
	lexer.emit([]uint16{Scalar})
	lexer.pushToIndex(nl+1, true)
	return lexer.parseLineStart()
}

func (lexer *Lexer) parsePlainScalar() string {
	inFlow := lexer.flowLevel > 0
	end := lexer.pos - 1
	i := lexer.pos - 1
	var ch int
	for {
		i++
		ch = characterAt(lexer.buffer, i)
		if ch == -1 {
			break
		}
		if ch == ':' {
			next := characterAt(lexer.buffer, i+1)
			if isEmpty(next) || (inFlow && inSet(flowIndicatorChars, next)) {
				break
			}
			end = i
		} else if isEmpty(ch) {
			next := characterAt(lexer.buffer, i+1)
			if ch == '\r' {
				if next == '\n' {
					i++
					ch = '\n'
					next = characterAt(lexer.buffer, i+1)
				} else {
					end = i
				}
			}
			if next == '#' || (inFlow && inSet(flowIndicatorChars, next)) {
				break
			}
			if ch == '\n' {
				cs := lexer.continueScalar(i + 1)
				if cs == -1 {
					break
				}
				i = max(i, cs-2) // to advance, but still account for ' #'
			}
		} else {
			if inFlow && inSet(flowIndicatorChars, ch) {
				break
			}
			end = i
		}
	}
	if ch == -1 && !lexer.atEnd {
		return lexer.setNext("plain-scalar")
	}
	lexer.emit([]uint16{Scalar})
	lexer.pushToIndex(end+1, true)
	if inFlow {
		return "flow"
	}
	return "doc"
}

// pushCount yields `this.buffer.substr(this.pos, n)` and advances pos by n, even when substr was cut
// short by the end of the buffer.
func (lexer *Lexer) pushCount(n int) int {
	if n > 0 {
		lexer.emit(substr(lexer.buffer, lexer.pos, n))
		lexer.pos += n
		return n
	}
	return 0
}

func (lexer *Lexer) pushToIndex(i int, allowEmpty bool) int {
	s := slice(lexer.buffer, lexer.pos, i)
	if len(s) > 0 {
		lexer.emit(s)
		lexer.pos += len(s)
		return len(s)
	} else if allowEmpty {
		lexer.emit([]uint16{})
	}
	return 0
}

func (lexer *Lexer) pushIndicators() int {
	n := 0
loop:
	for {
		switch lexer.charAt(0) {
		case '!':
			n += lexer.pushTag()
			n += lexer.pushSpaces(true)
			continue loop
		case '&':
			n += lexer.pushUntil(isNotAnchorChar)
			n += lexer.pushSpaces(true)
			continue loop
		case '-', // this is an error
			'?', // this is an error outside flow collections
			':':
			inFlow := lexer.flowLevel > 0
			ch1 := lexer.charAt(1)
			if isEmpty(ch1) || (inFlow && inSet(flowIndicatorChars, ch1)) {
				if !inFlow {
					lexer.indentNext = lexer.indentValue + 1
				} else if lexer.flowKey {
					lexer.flowKey = false
				}
				n += lexer.pushCount(1)
				n += lexer.pushSpaces(true)
				continue loop
			}
		}
		break loop
	}
	return n
}

func (lexer *Lexer) pushTag() int {
	if lexer.charAt(1) == '<' {
		i := lexer.pos + 2
		ch := characterAt(lexer.buffer, i)
		for !isEmpty(ch) && ch != '>' {
			i++
			ch = characterAt(lexer.buffer, i)
		}
		if ch == '>' {
			return lexer.pushToIndex(i+1, false)
		}
		return lexer.pushToIndex(i, false)
	}
	i := lexer.pos + 1
	ch := characterAt(lexer.buffer, i)
	for ch != -1 {
		if inSet(tagChars, ch) {
			i++
			ch = characterAt(lexer.buffer, i)
		} else if ch == '%' &&
			inSet(hexDigits, characterAt(lexer.buffer, i+1)) &&
			inSet(hexDigits, characterAt(lexer.buffer, i+2)) {
			i += 3
			ch = characterAt(lexer.buffer, i)
		} else {
			break
		}
	}
	return lexer.pushToIndex(i, false)
}

func (lexer *Lexer) pushNewline() int {
	ch := characterAt(lexer.buffer, lexer.pos)
	if ch == '\n' {
		return lexer.pushCount(1)
	} else if ch == '\r' && lexer.charAt(1) == '\n' {
		return lexer.pushCount(2)
	}
	return 0
}

func (lexer *Lexer) pushSpaces(allowTabs bool) int {
	i := lexer.pos - 1
	var ch int
	for {
		i++
		ch = characterAt(lexer.buffer, i)
		if !(ch == ' ' || (allowTabs && ch == '\t')) {
			break
		}
	}
	n := i - lexer.pos
	if n > 0 {
		lexer.emit(substr(lexer.buffer, lexer.pos, n))
		lexer.pos = i
	}
	return n
}

func (lexer *Lexer) pushUntil(test func(ch int) bool) int {
	i := lexer.pos
	ch := characterAt(lexer.buffer, i)
	for !test(ch) {
		i++
		ch = characterAt(lexer.buffer, i)
	}
	return lexer.pushToIndex(i, false)
}
