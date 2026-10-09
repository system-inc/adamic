// Port of postcss-values-parser 2.0.1 lib/parser.js, with lib/errors/ParserError.js.
//
// Upstream throws; this port panics with an *Error and Parse recovers it, so the control flow reads
// like the source. Upstream also throws TypeErrors where it reads a property of undefined (a value
// that is only whitespace, an operator with nothing after it); typeError reproduces those with V8's
// message, since Prettier catches every throw alike.

package values

import (
	"strings"
	"unicode/utf16"

	"github.com/system-inc/cohere/internal/format/estree"
)

type parser struct {
	// cache needs to be an array for values with more than 1 level of function nesting
	cache    []*estree.Node
	input    string
	options  Options
	position int
	// upstream also keeps an `unbalanced` count on the parser that nothing reads; the count that
	// matters lives on the Value and FunctionNode nodes.
	root    *estree.Node
	current *estree.Node
	tokens  []token
	spaces  string

	// byteOffsets maps a UTF-16 index into the input to its byte offset.
	byteOffsets []int
}

func newParser(input string, options Options) *parser {
	units, byteOffsets := utf16Units(input)
	parser := &parser{
		cache:       []*estree.Node{},
		input:       input,
		options:     options,
		position:    0,
		root:        newRoot(),
		byteOffsets: byteOffsets,
	}

	value := newValue()

	appendNode(parser.root, value)

	parser.current = value
	parser.tokens = tokenize(units, parser.options)
	return parser
}

func (parser *parser) parse() *estree.Node {
	return parser.loop()
}

// sourceIndex converts one of upstream's UTF-16 indexes to a byte offset. Past the end it counts one
// byte per unit, which only a malformed index reaches.
func (parser *parser) sourceIndex(index int) int {
	if index < 0 {
		return index
	}
	if index < len(parser.byteOffsets) {
		return parser.byteOffsets[index]
	}
	return parser.byteOffsets[len(parser.byteOffsets)-1] + index - (len(parser.byteOffsets) - 1)
}

// tokenSource is the { start, end } most node kinds take straight from their token.
func tokenSource(token *token) map[string]any {
	return source(position(token.startLine, token.startColumn), position(token.endLine, token.endColumnValue()))
}

func (parser *parser) colon() {
	token := parser.currToken()

	parser.newNode(leaf("colon", decode(token.value), tokenSource(token), parser.sourceIndex(token.index)))

	parser.position++
}

func (parser *parser) comma() {
	token := parser.currToken()

	parser.newNode(leaf("comma", decode(token.value), tokenSource(token), parser.sourceIndex(token.index)))

	parser.position++
}

func (parser *parser) comment() {
	inline := false
	value := commentDelimiters(decode(parser.currToken().value))

	if parser.options.Loose && strings.HasPrefix(value, "//") {
		value = value[2:]
		inline = true
	}

	node := newComment(value, inline, tokenSource(parser.currToken()), parser.sourceIndex(parser.currToken().index))

	parser.newNode(node)
	parser.position++
}

// commentDelimiters is value.replace(/\/\*|\*\//g, ""): every "/*" and "*/", left to right, not
// overlapping.
func commentDelimiters(value string) string {
	var builder strings.Builder
	for index := 0; index < len(value); index++ {
		if index+1 < len(value) && ((value[index] == '/' && value[index+1] == '*') || (value[index] == '*' && value[index+1] == '/')) {
			index++
			continue
		}
		builder.WriteByte(value[index])
	}
	return builder.String()
}

func (parser *parser) error(message string, token *token) {
	panic(&Error{Name: "ParserError", Message: message + " at line: " + itoa(token.startLine) + ", column " + itoa(token.startColumn)})
}

func (parser *parser) loop() *estree.Node {
	for parser.position < len(parser.tokens) {
		parser.parseTokens()
	}

	if last(parser.current) == nil && parser.spaces != "" {
		appendRaw(parser.current, "before", parser.spaces)
	} else if parser.spaces != "" {
		appendRaw(last(parser.current), "after", parser.spaces)
	}

	parser.spaces = ""

	return parser.root
}

func (parser *parser) operator() {
	// if a +|- operator is followed by a non-word character (. is allowed) and
	// is preceded by a non-word character. (5+5)
	char := decode(parser.currToken().value)

	if char == "+" || char == "-" {
		// only inspect if the operator is not the first token, and we're only
		// within a calc() function: the only spec-valid place for math expressions
		if !parser.options.Loose {
			if parser.position > 0 {
				if parser.current.Is("func") && parser.current.String("value") == "calc" {
					// allow operators to be proceeded by spaces and opening parens
					if kind(parser.prevToken()) != "space" && kind(parser.prevToken()) != "(" {
						parser.error("Syntax Error", parser.currToken())
					} else if kind(parser.nextToken()) != "space" && kind(parser.nextToken()) != "word" {
						// valid: calc(1 - +2)
						// invalid: calc(1 -+2)
						parser.error("Syntax Error", parser.currToken())
					} else if kind(parser.nextToken()) == "word" && property(last(parser.current), "type") != "operator" &&
						property(last(parser.current), "value") != "(" {
						// valid: calc(1 - +2)
						// valid: calc(-0.5 + 2)
						// invalid: calc(1 -2)
						parser.error("Syntax Error", parser.currToken())
					}
				} else if kind(parser.nextToken()) == "space" ||
					kind(parser.nextToken()) == "operator" ||
					kind(parser.prevToken()) == "operator" {
					// if we're not in a function and someone has doubled up on operators,
					// or they're trying to perform a calc outside of a calc
					// eg. +-4px or 5+ 5, throw an error
					parser.error("Syntax Error", parser.currToken())
				}
			}
		}

		if !parser.options.Loose {
			if kind(parser.nextToken()) == "word" {
				parser.word()
				return
			}
		} else {
			currentLast := last(parser.current)
			if (len(parser.current.List("nodes")) == 0 || (currentLast != nil && currentLast.Is("operator"))) && kind(parser.nextToken()) == "word" {
				parser.word()
				return
			}
		}
	}

	token := parser.currToken()
	// Upstream ends the operator where it starts, and takes its sourceIndex from token[4], which is
	// the end line rather than the index. Prettier reads it as an index, so it converts like one.
	node := leaf("operator", decode(token.value), source(position(token.startLine, token.startColumn), position(token.startLine, token.startColumn)), parser.sourceIndex(token.endLine))

	parser.position++

	parser.newNode(node)
}

func (parser *parser) parseTokens() {
	switch parser.currToken().kind {
	case "space":
		parser.space()
	case "colon":
		parser.colon()
	case "comma":
		parser.comma()
	case "comment":
		parser.comment()
	case "(":
		parser.parenOpen()
	case ")":
		parser.parenClose()
	case "atword", "word":
		parser.word()
	case "operator":
		parser.operator()
	case "string":
		parser.string()
	case "unicoderange":
		parser.unicodeRange()
	default:
		parser.word()
	}
}

func (parser *parser) parenOpen() {
	unbalancedCount := 1
	pos := parser.position + 1
	token := parser.currToken()

	// check for balanced parens
	for pos < len(parser.tokens) && unbalancedCount != 0 {
		tkn := &parser.tokens[pos]

		if tkn.kind == "(" {
			unbalancedCount++
		}
		if tkn.kind == ")" {
			unbalancedCount--
		}
		pos++
	}

	if unbalancedCount != 0 {
		parser.error("Expected closing parenthesis", token)
	}

	// ok, all parens are balanced. continue on

	currentLast := last(parser.current)

	if currentLast != nil && currentLast.Is("func") && unbalanced(currentLast) < 0 {
		setUnbalanced(currentLast, 0) // ok we're ready to add parens now
		parser.current = currentLast
	}

	setUnbalanced(parser.current, unbalanced(parser.current)+1)

	parser.newNode(newParen(decode(token.value), tokenSource(token), parser.sourceIndex(token.index)))

	parser.position++

	// url functions get special treatment, and anything between the function
	// parens get treated as one word, if the contents aren't not a string.
	if parser.current.Is("func") && unbalanced(parser.current) != 0 &&
		parser.current.String("value") == "url" && kind(parser.currToken()) != "string" &&
		kind(parser.currToken()) != ")" && !parser.options.Loose {
		nextToken := parser.nextToken()
		value := decode(parser.currToken().value)
		start := position(parser.currToken().startLine, parser.currToken().startColumn)

		for nextToken != nil && nextToken.kind != ")" && unbalanced(parser.current) != 0 {
			parser.position++
			value += decode(parser.currToken().value)
			nextToken = parser.nextToken()
		}

		if parser.position != len(parser.tokens)-1 {
			// skip the following word definition, or it'll be a duplicate
			parser.position++

			// Constructed directly, so unlike splitWord's words it carries no isHex or isColor.
			parser.newNode(leaf("word", value, source(start, position(parser.currToken().endLine, parser.currToken().endColumnValue())), parser.sourceIndex(parser.currToken().index)))
		}
	}
}

func (parser *parser) parenClose() {
	token := parser.currToken()

	parser.newNode(newParen(decode(token.value), tokenSource(token), parser.sourceIndex(token.index)))

	parser.position++

	if parser.position >= len(parser.tokens)-1 && unbalanced(parser.current) == 0 {
		return
	}

	setUnbalanced(parser.current, unbalanced(parser.current)-1)

	if unbalanced(parser.current) < 0 {
		parser.error("Expected opening parenthesis", token)
	}

	if unbalanced(parser.current) == 0 && len(parser.cache) > 0 {
		parser.current = parser.cache[len(parser.cache)-1]
		parser.cache = parser.cache[:len(parser.cache)-1]
	}
}

func (parser *parser) space() {
	token := parser.currToken()
	// Handle space before and after the selector
	// Upstream compares the next token's type with ',' and ')', but comma tokens are typed 'comma',
	// so only ')' and the last position ever match.
	if parser.position == len(parser.tokens)-1 || kind(parser.nextToken()) == "," || kind(parser.nextToken()) == ")" {
		currentLast := last(parser.current)
		if currentLast == nil {
			panic(typeError("raws"))
		}
		appendRaw(currentLast, "after", decode(token.value))
		parser.position++
	} else {
		parser.spaces = decode(token.value)
		parser.position++
	}
}

func (parser *parser) unicodeRange() {
	token := parser.currToken()

	parser.newNode(leaf("unicode-range", decode(token.value), tokenSource(token), parser.sourceIndex(token.index)))

	parser.position++
}

// rNumber is /^[\+\-]?((\d+(\.\d*)?)|(\.\d+))([eE][\+\-]?\d+)?/: the length of the match at the
// start of text, or -1.
func rNumber(text string) int {
	index := 0
	if index < len(text) && (text[index] == '+' || text[index] == '-') {
		index++
	}
	digits := func(from int) int {
		end := from
		for end < len(text) && text[end] >= '0' && text[end] <= '9' {
			end++
		}
		return end
	}
	if end := digits(index); end > index {
		index = end
		if index < len(text) && text[index] == '.' {
			index = digits(index + 1)
		}
	} else if index < len(text) && text[index] == '.' && digits(index+1) > index+1 {
		index = digits(index + 1)
	} else {
		return -1
	}
	if index < len(text) && (text[index] == 'e' || text[index] == 'E') {
		exponent := index + 1
		if exponent < len(text) && (text[exponent] == '+' || text[exponent] == '-') {
			exponent++
		}
		if end := digits(exponent); end > exponent {
			index = end
		}
	}
	return index
}

// rNoFollow is /^(?!\#([a-z0-9]+))[\#\{\}]/gi, tested once on a fresh regex: a word opening with
// '#', '{' or '}', unless it is '#' and an ASCII letter or digit.
func rNoFollow(word []uint16) bool {
	if len(word) == 0 || (word[0] != '#' && word[0] != '{' && word[0] != '}') {
		return false
	}
	return !(word[0] == '#' && len(word) > 1 && alphaNum(word[1]))
}

// isHex is /^#(.+)/: '#' and at least one character that is not a line terminator.
func isHex(value []uint16) bool {
	return len(value) > 1 && value[0] == '#' && value[1] != '\n' && value[1] != '\r' && value[1] != 0x2028 && value[1] != 0x2029
}

// isColor is /^#([0-9a-f]{3}|[0-9a-f]{4}|[0-9a-f]{6}|[0-9a-f]{8})$/i, ASCII only as above.
func isColor(value []uint16) bool {
	if len(value) == 0 || value[0] != '#' {
		return false
	}
	switch len(value) - 1 {
	case 3, 4, 6, 8:
	default:
		return false
	}
	for _, unit := range value[1:] {
		if !((unit >= '0' && unit <= '9') || (unit >= 'a' && unit <= 'f') || (unit >= 'A' && unit <= 'F')) {
			return false
		}
	}
	return true
}

func (parser *parser) splitWord() {
	nextToken := parser.nextToken()
	word := append([]uint16{}, parser.currToken().value...)

	// treat css-like groupings differently so they can be inspected,
	// but don't address them as anything but a word, but allow hex values
	// to pass through.
	if !rNoFollow(word) {
		for nextToken != nil && nextToken.kind == "word" {
			parser.position++

			current := parser.currToken().value
			word = append(word, current...)

			nextToken = parser.nextToken()
		}
	}

	// hasAt = indexesOf(word, '@'); indices = sortAscending(uniq(flatten([[0], hasAt]))): 0 and every
	// '@', ascending, once each.
	hasAt := []int{}
	for index, unit := range word {
		if unit == '@' {
			hasAt = append(hasAt, index)
		}
	}
	indices := []int{0}
	for _, index := range hasAt {
		if index != 0 {
			indices = append(indices, index)
		}
	}

	currToken := parser.currToken()
	for i, ind := range indices {
		index := len(word)
		if i+1 < len(indices) && indices[i+1] != 0 {
			index = indices[i+1]
		}
		value := word[ind:index]
		var node *estree.Node

		nodeSource := source(position(currToken.startLine, currToken.startColumn+ind), position(currToken.endLine, currToken.startColumn+(index-1)))
		nodeSourceIndex := parser.sourceIndex(currToken.index + indices[i])

		if containsIndex(hasAt, ind) {
			node = newAtWord(decode(value[1:]), nodeSource, nodeSourceIndex)
		} else if rNumber(decode(currToken.value)) >= 0 {
			text := decode(value)
			unit := text
			if matched := rNumber(text); matched >= 0 {
				unit = text[matched:]
			}

			// value.replace(unit, ''): the first occurrence, which is not always the suffix.
			node = newNumber(strings.Replace(text, unit, "", 1), nodeSource, nodeSourceIndex, unit)
		} else {
			if nextToken != nil && nextToken.kind == "(" {
				node = newFunc(decode(value), nodeSource, nodeSourceIndex)
				parser.cache = append(parser.cache, parser.current)
			} else {
				node = leaf("word", decode(value), nodeSource, nodeSourceIndex)
				node.Set("isHex", isHex(value))
				node.Set("isColor", isColor(value))
			}
		}

		parser.newNode(node)
	}

	parser.position++
}

func containsIndex(indexes []int, index int) bool {
	for _, each := range indexes {
		if each == index {
			return true
		}
	}
	return false
}

func (parser *parser) string() {
	token := parser.currToken()
	value := token.value
	// rQuote is /^(\"|\')/
	quoted := len(value) > 0 && (value[0] == '"' || value[0] == '\'')
	quote := ""

	if quoted {
		quote = string(rune(value[0]))
		// set value to the string within the quotes
		// quotes are stored in raws
		value = value[1 : len(value)-1]
	}

	node := newString(decode(value), tokenSource(token), parser.sourceIndex(token.index), quoted)

	raws(node)["quote"] = quote

	parser.newNode(node)
	parser.position++
}

func (parser *parser) word() {
	parser.splitWord()
}

func (parser *parser) newNode(node *estree.Node) {
	if parser.spaces != "" {
		appendRaw(node, "before", parser.spaces)
		parser.spaces = ""
	}

	appendNode(parser.current, node)
}

func (parser *parser) currToken() *token { return parser.tokenAt(parser.position) }

func (parser *parser) nextToken() *token { return parser.tokenAt(parser.position + 1) }

func (parser *parser) prevToken() *token { return parser.tokenAt(parser.position - 1) }

func (parser *parser) tokenAt(index int) *token {
	if index < 0 || index >= len(parser.tokens) {
		return nil
	}
	return &parser.tokens[index]
}

// kind is token[0], which throws when the token is undefined.
func kind(token *token) string {
	if token == nil {
		panic(typeError("0"))
	}
	return token.kind
}

// property is node.type or node.value, which throws when the node is undefined.
func property(node *estree.Node, key string) string {
	if node == nil {
		panic(typeError(key))
	}
	if key == "type" {
		return node.Type()
	}
	return node.String(key)
}

// typeError is V8's TypeError for reading a property of undefined.
func typeError(key string) *Error {
	return &Error{Name: "TypeError", Message: "Cannot read properties of undefined (reading '" + key + "')"}
}

// decode turns UTF-16 units back into a Go string.
func decode(units []uint16) string {
	return string(utf16Decode(units))
}

func utf16Decode(units []uint16) []rune {
	return utf16.Decode(units)
}
