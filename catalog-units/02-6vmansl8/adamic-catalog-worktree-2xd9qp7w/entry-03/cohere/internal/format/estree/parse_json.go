package estree

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/printing"
)

// ParseJSON is Prettier's json parsers, src/language-json/parse/json.js: Babel's parseExpression over
// the text, checked by assertJsonNode, wrapped in a JsonRoot. allowComments false is json-stringify,
// which refuses a comment.
//
// Babel parses any JavaScript expression and assertJsonNode then refuses everything that is not JSON
// (with JSON5's identifiers and signs). This parses only that subset directly, and refuses whatever
// else it meets, which is the same set of accepted inputs: any construct beyond the subset would
// have been refused by assertJsonNode after Babel parsed it.
//
// The nodes are Babel's shapes, because that is what Prettier prints for JSON: ObjectProperty,
// StringLiteral, NumericLiteral, BooleanLiteral, NullLiteral. Babel keeps a literal's source text in
// extra.raw; here it is the raw property, which is where the printer's getRaw reads it for every
// tree. Comments are Line and Block nodes, the typescript-estree spelling, where Babel says
// CommentLine and CommentBlock; the printer treats the two spellings identically everywhere.
//
// allowEmpty (jsonc) is not supported: no parser our files resolve to sets it.
//
// Every refusal is marked printing.Syntax, so the format phase skips a file that is not JSON rather
// than reporting a broken formatter.
func ParseJSON(text string, allowComments bool) (*Node, []*Node, error) {
	root, comments, err := parseJSON(text, allowComments)
	return root, comments, printing.Syntax(err)
}

func parseJSON(text string, allowComments bool) (*Node, []*Node, error) {
	reader := &jsonReader{text: text}
	reader.skipTrivia()
	if reader.position >= len(text) {
		return nil, nil, fmt.Errorf("json: empty input")
	}
	expression, err := reader.parseValue()
	if err != nil {
		return nil, nil, err
	}
	reader.skipTrivia()
	if reader.err != nil {
		return nil, nil, reader.err
	}
	if reader.position < len(text) {
		return nil, nil, reader.errorf("unexpected %q", reader.text[reader.position:min(reader.position+10, len(text))])
	}
	if !allowComments && len(reader.comments) > 0 {
		return nil, nil, fmt.Errorf("json: Comment is not allowed in JSON")
	}
	root := New("JsonRoot", 0, len(text), "node", expression)
	return root, reader.comments, nil
}

type jsonReader struct {
	text     string
	position int
	comments []*Node
	err      error
}

func (reader *jsonReader) errorf(format string, arguments ...any) error {
	line := strings.Count(reader.text[:reader.position], "\n") + 1
	return fmt.Errorf("json: line %d: %s", line, fmt.Sprintf(format, arguments...))
}

// skipTrivia skips whitespace and comments, collecting the comments.
func (reader *jsonReader) skipTrivia() {
	text := reader.text
	for reader.position < len(text) {
		start := reader.position
		switch {
		case strings.HasPrefix(text[start:], "//"):
			end := start + 2
			for end < len(text) && lineTerminatorSize(text, end) == 0 {
				end++
			}
			reader.comments = append(reader.comments, New("Line", start, end, "value", text[start+2:end]))
			reader.position = end
		case strings.HasPrefix(text[start:], "/*"):
			closing := strings.Index(text[start+2:], "*/")
			if closing < 0 {
				reader.err = reader.errorf("unterminated comment")
				reader.position = len(text)
				return
			}
			end := start + 2 + closing + 2
			reader.comments = append(reader.comments, New("Block", start, end, "value", text[start+2:end-2]))
			reader.position = end
		default:
			size := whitespaceSize(text, start)
			if size == 0 {
				return
			}
			reader.position += size
		}
	}
}

// expect skips trivia and consumes one punctuator.
func (reader *jsonReader) expect(punctuator byte) error {
	reader.skipTrivia()
	if reader.err != nil {
		return reader.err
	}
	if reader.position >= len(reader.text) || reader.text[reader.position] != punctuator {
		return reader.errorf("expected %q", punctuator)
	}
	reader.position++
	return nil
}

// peek skips trivia and returns the next byte, or 0 at the end.
func (reader *jsonReader) peek() byte {
	reader.skipTrivia()
	if reader.position >= len(reader.text) {
		return 0
	}
	return reader.text[reader.position]
}

// parseValue reads one value after trivia: assertJsonNode's accepted node types.
func (reader *jsonReader) parseValue() (*Node, error) {
	character := reader.peek()
	if reader.err != nil {
		return nil, reader.err
	}
	switch {
	case character == '{':
		return reader.parseObject()
	case character == '[':
		return reader.parseArray()
	case character == '"' || character == '\'':
		return reader.parseString()
	case character == '`':
		return reader.parseTemplate()
	case character == '+' || character == '-':
		start := reader.position
		reader.position++
		argument, err := reader.parseValue()
		if err != nil {
			return nil, err
		}
		// assertJsonNode: a sign only before a number, Infinity or NaN.
		if !argument.Is("NumericLiteral") &&
			!(argument.Is("Identifier") && (argument.String("name") == "Infinity" || argument.String("name") == "NaN")) {
			return nil, reader.errorf("Operator '%c' before '%s' is not allowed in JSON", character, argument.Type())
		}
		return New("UnaryExpression", start, argument.End(), "operator", string(character), "prefix", true,
			"argument", argument), nil
	case character == '.' || character >= '0' && character <= '9':
		return reader.parseNumber()
	}
	start := reader.position
	name := reader.readIdentifierName()
	switch name {
	case "":
		return nil, reader.errorf("unexpected %q", reader.text[start:min(start+10, len(reader.text))])
	case "true", "false":
		return New("BooleanLiteral", start, reader.position, "value", name == "true"), nil
	case "null":
		return New("NullLiteral", start, reader.position), nil
	case "Infinity", "NaN", "undefined":
		return New("Identifier", start, reader.position, "name", name), nil
	}
	return nil, reader.errorf("Identifier '%s' is not allowed in JSON", name)
}

func (reader *jsonReader) parseObject() (*Node, error) {
	start := reader.position
	reader.position++
	properties := []*Node{}
	for {
		character := reader.peek()
		if reader.err != nil {
			return nil, reader.err
		}
		if character == '}' {
			reader.position++
			break
		}
		property, err := reader.parseProperty()
		if err != nil {
			return nil, err
		}
		properties = append(properties, property)
		if reader.peek() == ',' {
			reader.position++
			continue
		}
		if err := reader.expect('}'); err != nil {
			return nil, err
		}
		break
	}
	return New("ObjectExpression", start, reader.position, "properties", properties), nil
}

// parseProperty reads `key: value`. A computed, shorthand or method property is refused, as
// assertJsonNode refuses them.
func (reader *jsonReader) parseProperty() (*Node, error) {
	character := reader.peek()
	start := reader.position
	var key *Node
	var err error
	switch {
	case character == '"' || character == '\'':
		key, err = reader.parseString()
	case character == '.' || character >= '0' && character <= '9':
		key, err = reader.parseNumber()
	case character == '[':
		return nil, reader.errorf("Computed key is not allowed in JSON")
	default:
		name := reader.readIdentifierName()
		if name == "" {
			return nil, reader.errorf("unexpected %q", reader.text[start:min(start+10, len(reader.text))])
		}
		key = New("Identifier", start, reader.position, "name", name)
	}
	if err != nil {
		return nil, err
	}
	if reader.peek() != ':' {
		return nil, reader.errorf("Shorthand property is not allowed in JSON")
	}
	reader.position++
	value, err := reader.parseValue()
	if err != nil {
		return nil, err
	}
	return New("ObjectProperty", start, value.End(), "method", false, "key", key, "computed", false,
		"shorthand", false, "value", value), nil
}

func (reader *jsonReader) parseArray() (*Node, error) {
	start := reader.position
	reader.position++
	elements := []*Node{}
	for {
		character := reader.peek()
		if reader.err != nil {
			return nil, reader.err
		}
		if character == ']' {
			reader.position++
			break
		}
		// A hole, `[1, , 2]`: Babel's null element, which assertJsonNode skips.
		if character == ',' {
			reader.position++
			elements = append(elements, nil)
			continue
		}
		element, err := reader.parseValue()
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
		if reader.peek() == ',' {
			reader.position++
			continue
		}
		if err := reader.expect(']'); err != nil {
			return nil, err
		}
		break
	}
	return New("ArrayExpression", start, reader.position, "elements", elements), nil
}

// parseString reads a single or double quoted string, keeping its raw text and decoding its value.
func (reader *jsonReader) parseString() (*Node, error) {
	text := reader.text
	start := reader.position
	quote := text[start]
	position := start + 1
	for position < len(text) && text[position] != quote {
		if text[position] == '\\' {
			position++
		} else if text[position] == '\n' || text[position] == '\r' {
			return nil, reader.errorf("unterminated string")
		}
		position++
	}
	if position >= len(text) {
		return nil, reader.errorf("unterminated string")
	}
	reader.position = position + 1
	raw := text[start:reader.position]
	value, err := decodeJavaScriptString(raw[1 : len(raw)-1])
	if err != nil {
		return nil, reader.errorf("%v", err)
	}
	return New("StringLiteral", start, reader.position, "value", value, "raw", raw), nil
}

// parseTemplate reads a template literal with no expressions, the only kind assertJsonNode accepts.
func (reader *jsonReader) parseTemplate() (*Node, error) {
	text := reader.text
	start := reader.position
	position := start + 1
	for position < len(text) && text[position] != '`' {
		if text[position] == '\\' {
			position++
		} else if strings.HasPrefix(text[position:], "${") {
			return nil, reader.errorf("'TemplateLiteral' with expression is not allowed in JSON")
		}
		position++
	}
	if position >= len(text) {
		return nil, reader.errorf("unterminated template")
	}
	reader.position = position + 1
	raw := strings.ReplaceAll(strings.ReplaceAll(text[start+1:position], "\r\n", "\n"), "\r", "\n")
	cooked, err := decodeJavaScriptString(raw)
	if err != nil {
		return nil, reader.errorf("%v", err)
	}
	element := New("TemplateElement", start+1, position, "value", &TemplateValue{Cooked: &cooked, Raw: raw}, "tail", true)
	return New("TemplateLiteral", start, reader.position, "expressions", []*Node{}, "quasis", []*Node{element}), nil
}

// parseNumber reads a numeric literal: decimal with fraction and exponent, or hexadecimal, octal and
// binary with their prefixes, numeric separators allowed. Legacy octal (`017`) and bigint are refused.
func (reader *jsonReader) parseNumber() (*Node, error) {
	text := reader.text
	start := reader.position
	position := start
	isDigit := func(character byte) bool { return character >= '0' && character <= '9' || character == '_' }
	var value float64
	if position+1 < len(text) && text[position] == '0' && strings.ContainsRune("xXoObB", rune(text[position+1])) {
		base := map[byte]int{'x': 16, 'o': 8, 'b': 2}[text[position+1]|0x20]
		position += 2
		digitsStart := position
		for position < len(text) && (text[position] == '_' || digitValue(text[position]) < base) {
			position++
		}
		digits := strings.ReplaceAll(text[digitsStart:position], "_", "")
		integer, ok := new(big.Int).SetString(digits, base)
		if !ok {
			return nil, reader.errorf("invalid number %q", text[start:position])
		}
		value, _ = new(big.Float).SetInt(integer).Float64()
	} else {
		for position < len(text) && isDigit(text[position]) {
			position++
		}
		if position < len(text) && text[position] == '.' {
			position++
			for position < len(text) && isDigit(text[position]) {
				position++
			}
		}
		if position < len(text) && (text[position] == 'e' || text[position] == 'E') {
			position++
			if position < len(text) && (text[position] == '+' || text[position] == '-') {
				position++
			}
			for position < len(text) && isDigit(text[position]) {
				position++
			}
		}
		literal := strings.ReplaceAll(text[start:position], "_", "")
		if len(literal) > 1 && literal[0] == '0' && literal[1] >= '0' && literal[1] <= '9' {
			return nil, reader.errorf("legacy octal literal %q", literal)
		}
		parsed, err := strconv.ParseFloat(literal, 64)
		if err != nil && !isRangeError(err) {
			return nil, reader.errorf("invalid number %q", literal)
		}
		value = parsed
	}
	if position < len(text) && (text[position] == 'n' || isIdentifierPart(text, position)) {
		return nil, reader.errorf("invalid number %q", text[start:position+1])
	}
	reader.position = position
	return New("NumericLiteral", start, position, "value", value, "raw", text[start:position]), nil
}

func isRangeError(err error) bool {
	numberError, isNumberError := err.(*strconv.NumError)
	return isNumberError && numberError.Err == strconv.ErrRange
}

func digitValue(character byte) int {
	switch {
	case character >= '0' && character <= '9':
		return int(character - '0')
	case character >= 'a' && character <= 'f':
		return int(character-'a') + 10
	case character >= 'A' && character <= 'F':
		return int(character-'A') + 10
	}
	return math.MaxInt
}

// readIdentifierName reads an IdentifierName at the position, or returns "" without moving.
func (reader *jsonReader) readIdentifierName() string {
	start := reader.position
	for reader.position < len(reader.text) {
		character, size := utf8.DecodeRuneInString(reader.text[reader.position:])
		isStart := character == '$' || character == '_' || unicode.IsLetter(character) || unicode.Is(unicode.Nl, character)
		if reader.position == start && !isStart {
			break
		}
		if !isStart && !unicode.IsDigit(character) && !unicode.Is(unicode.Mn, character) &&
			!unicode.Is(unicode.Mc, character) && !unicode.Is(unicode.Pc, character) &&
			character != 0x200C && character != 0x200D {
			break
		}
		reader.position += size
	}
	return reader.text[start:reader.position]
}

func isIdentifierPart(text string, position int) bool {
	character, _ := utf8.DecodeRuneInString(text[position:])
	return character == '$' || character == '_' || unicode.IsLetter(character) || unicode.IsDigit(character)
}

// decodeJavaScriptString is a string literal's value from its raw content: JavaScript's escapes.
func decodeJavaScriptString(content string) (string, error) {
	if !strings.Contains(content, "\\") {
		return content, nil
	}
	var builder strings.Builder
	for index := 0; index < len(content); index++ {
		character := content[index]
		if character != '\\' {
			builder.WriteByte(character)
			continue
		}
		index++
		if index >= len(content) {
			return "", fmt.Errorf("a string ends in a backslash")
		}
		switch escaped := content[index]; escaped {
		case 'n':
			builder.WriteByte('\n')
		case 't':
			builder.WriteByte('\t')
		case 'r':
			builder.WriteByte('\r')
		case 'b':
			builder.WriteByte('\b')
		case 'f':
			builder.WriteByte('\f')
		case 'v':
			builder.WriteByte('\v')
		case '0':
			if index+1 < len(content) && content[index+1] >= '0' && content[index+1] <= '9' {
				return "", fmt.Errorf("an octal escape in a string")
			}
			builder.WriteByte(0)
		case '\r':
			// A line continuation, \r\n counting as one.
			if index+1 < len(content) && content[index+1] == '\n' {
				index++
			}
		case '\n':
		case 'x':
			if index+2 >= len(content) {
				return "", fmt.Errorf("an incomplete \\x escape")
			}
			code, err := strconv.ParseUint(content[index+1:index+3], 16, 8)
			if err != nil {
				return "", fmt.Errorf("an invalid \\x escape")
			}
			builder.WriteRune(rune(code))
			index += 2
		case 'u':
			code, length, err := readUnicodeEscape(content[index+1:])
			if err != nil {
				return "", err
			}
			index += length
			// A surrogate pair spelled as two escapes is one character.
			if code >= 0xD800 && code <= 0xDBFF && strings.HasPrefix(content[index+1:], "\\u") {
				if low, lowLength, err := readUnicodeEscape(content[index+3:]); err == nil && low >= 0xDC00 && low <= 0xDFFF {
					code = (code-0xD800)<<10 + (low - 0xDC00) + 0x10000
					index += 2 + lowLength
				}
			}
			builder.WriteRune(rune(code))
		default:
			if escaped >= '1' && escaped <= '9' {
				return "", fmt.Errorf("an octal escape in a string")
			}
			// Any other character escapes to itself, including a multi-byte one.
			_, size := utf8.DecodeRuneInString(content[index:])
			if strings.HasPrefix(content[index:], "\u2028") || strings.HasPrefix(content[index:], "\u2029") {
				index += size - 1
				continue
			}
			builder.WriteString(content[index : index+size])
			index += size - 1
		}
	}
	return builder.String(), nil
}

// readUnicodeEscape reads the hex after `\u`: four digits or a braced code point. It returns the code
// and how many bytes it consumed.
func readUnicodeEscape(text string) (int, int, error) {
	if strings.HasPrefix(text, "{") {
		closing := strings.IndexByte(text, '}')
		if closing < 0 {
			return 0, 0, fmt.Errorf("an incomplete \\u{} escape")
		}
		code, err := strconv.ParseUint(text[1:closing], 16, 32)
		if err != nil || code > 0x10FFFF {
			return 0, 0, fmt.Errorf("an invalid \\u{} escape")
		}
		return int(code), closing + 1, nil
	}
	if len(text) < 4 {
		return 0, 0, fmt.Errorf("an incomplete \\u escape")
	}
	code, err := strconv.ParseUint(text[:4], 16, 16)
	if err != nil {
		return 0, 0, fmt.Errorf("an invalid \\u escape")
	}
	return int(code), 4, nil
}
