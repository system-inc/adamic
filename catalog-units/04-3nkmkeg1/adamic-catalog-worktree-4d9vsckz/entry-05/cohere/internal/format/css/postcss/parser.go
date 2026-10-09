package postcss

import "strings"

// postcss 8.5.16, lib/parser.js.
//
// Upstream throws CssSyntaxError from deep inside the parser; here the parser panics with the
// *CssSyntaxError and Parse recovers it into the error it returns, so no panic leaves Parse.
//
// Token arrays are Go slices. The helpers upstream calls to pop, shift or splice an array the caller
// keeps using take a pointer to the slice; everywhere upstream copies (slice, concat) the port copies too,
// so no two slices share elements a later write could reach.

// safeCommentNeighbor is SAFE_COMMENT_NEIGHBOR.
func safeCommentNeighbor(kind string) bool {
	return kind == "empty" || kind == "space"
}

// findLastWithPosition is the last token's `token[3] || token[2]` that is truthy, -1 for undefined.
func findLastWithPosition(tokens []token) int {
	for index := len(tokens) - 1; index >= 0; index-- {
		each := tokens[index]
		pos := each.endOrStart()
		if pos > 0 {
			return pos
		}
	}
	return -1
}

func tokensToString(tokens []token, from int, to int) string {
	var result strings.Builder
	for index := from; index < to; index++ {
		result.WriteString(tokens[index].value)
	}
	return result.String()
}

// copyTokens is tokens.slice(from, to): a new array.
func copyTokens(tokens []token, from int, to int) []token {
	result := make([]token, to-from)
	copy(result, tokens[from:to])
	return result
}

// parserMethods are the methods of Parser that postcss-scss's ScssParser overrides (scss_parser.go).
// Go has no inheritance, so where upstream calls one of them on `this` the port calls it through
// parser.this, which is the parser itself for css and the scssParser around it for scss; a subclass's
// `super.method()` is a call on the embedded *parser. Every other method is called on the parser directly,
// as no subclass the port carries replaces it.
type parserMethods interface {
	atrule(atToken token)
	comment(commentToken token)
	createTokenizer()
	raw(each *node, prop string, tokens []token, customProperty bool)
	rule(tokens []token)
}

type parser struct {
	input *input

	// this is upstream's `this` for the overridable methods.
	this parserMethods

	root      *node
	current   *node
	spaces    string
	semicolon bool

	tokenizer tokenStream
}

func newParser(in *input) *parser {
	p := &parser{}
	p.this = p
	p.construct(in)
	return p
}

// construct is Parser's constructor, split from newParser so a subclass sets this first and its own
// createTokenizer is the one the constructor calls.
func (p *parser) construct(in *input) {
	p.input = in

	p.root = newRoot()
	p.current = p.root
	p.spaces = ""
	p.semicolon = false

	p.this.createTokenizer()
	p.root.source = &source{start: &position{column: 1, line: 1, offset: 0}}
	p.root.mark("source")
}

func (p *parser) atrule(atToken token) {
	node := newAtRule()
	node.name = atToken.value[1:]
	node.mark("name")
	if node.name == "" {
		p.unnamedAtrule(node, atToken)
	}
	p.init(node, atToken.start)

	var kind string
	var prev *token
	var shift int
	last := false
	open := false
	params := []token{}
	brackets := []string{}

	for !p.tokenizer.endOfFile() {
		each, _ := p.tokenizer.nextToken()
		kind = each.kind

		if kind == "(" || kind == "[" {
			if kind == "(" {
				brackets = append(brackets, ")")
			} else {
				brackets = append(brackets, "]")
			}
		} else if kind == "{" && len(brackets) > 0 {
			brackets = append(brackets, "}")
		} else if len(brackets) > 0 && kind == brackets[len(brackets)-1] {
			brackets = brackets[:len(brackets)-1]
		}

		if len(brackets) == 0 {
			if kind == ";" {
				node.source.end = p.getPosition(each.start)
				node.source.end.offset++
				p.semicolon = true
				break
			} else if kind == "{" {
				open = true
				break
			} else if kind == "}" {
				if len(params) > 0 {
					shift = len(params) - 1
					prev = &params[shift]
					for prev != nil && prev.kind == "space" {
						shift--
						if shift >= 0 {
							prev = &params[shift]
						} else {
							prev = nil
						}
					}
					if prev != nil {
						node.source.end = p.getPosition(prev.endOrStart())
						node.source.end.offset++
					}
				}
				p.end(each)
				break
			} else {
				params = append(params, each)
			}
		} else {
			params = append(params, each)
		}

		if p.tokenizer.endOfFile() {
			last = true
			break
		}
	}

	node.raws["between"] = p.spacesAndCommentsFromEnd(&params)
	if len(params) > 0 {
		node.raws["afterName"] = p.spacesAndCommentsFromStart(&params)
		p.this.raw(node, "params", params, false)
		if last {
			each := params[len(params)-1]
			node.source.end = p.getPosition(each.endOrStart())
			node.source.end.offset++
			p.spaces = node.rawString("between")
			node.raws["between"] = ""
		}
	} else {
		node.raws["afterName"] = ""
		node.params = ""
		node.mark("params")
	}

	if open {
		node.nodes = emptyNodes()
		node.mark("nodes")
		p.current = node
	}
}

func (p *parser) checkMissedSemicolon(tokens []token) {
	colon := p.colon(tokens)
	if colon < 0 {
		return
	}

	founded := 0
	var each token
	for j := colon - 1; j >= 0; j-- {
		each = tokens[j]
		if each.kind != "space" {
			founded += 1
			if founded == 2 {
				break
			}
		}
	}
	// If the token is a word, e.g. `!important`, `red` or any other valid
	// property's value. Then we need to return the colon after that word
	// token. [3] is the "end" colon of that word. And because we need it
	// after that one we do +1 to get the next one.
	offset := each.start
	if each.kind == "word" {
		offset = each.end + 1
	}
	panic(p.input.errorAt("Missed semicolon", offset))
}

// colon is the index of the first top-level ":", -1 where upstream returns false.
func (p *parser) colon(tokens []token) int {
	brackets := 0
	var prev *token
	for index := range tokens {
		each := &tokens[index]
		kind := each.kind

		if kind == "(" {
			brackets += 1
		}
		if kind == ")" {
			brackets -= 1
		}
		if brackets == 0 && kind == ":" {
			if prev == nil {
				p.doubleColon(*each)
			} else if prev.kind == "word" && prev.value == "progid" {
				continue
			} else {
				return index
			}
		}

		prev = each
	}
	return -1
}

func (p *parser) comment(commentToken token) {
	node := newComment()
	p.init(node, commentToken.start)
	node.source.end = p.getPosition(commentToken.endOrStart())
	node.source.end.offset++

	text := commentToken.value[2 : len(commentToken.value)-2]
	if strings.TrimFunc(text, isJavaScriptWhitespace) == "" {
		node.text = ""
		node.raws["left"] = text
		node.raws["right"] = ""
	} else {
		// text.match(/^(\s*)([^]*\S)(\s*)$/): the leading whitespace, the rest up to the last
		// non-whitespace character, and the trailing whitespace.
		trimmedStart := strings.TrimLeftFunc(text, isJavaScriptWhitespace)
		trimmed := strings.TrimRightFunc(trimmedStart, isJavaScriptWhitespace)
		node.text = trimmed
		node.raws["left"] = text[:len(text)-len(trimmedStart)]
		node.raws["right"] = trimmedStart[len(trimmed):]
	}
	node.mark("text")
}

func (p *parser) createTokenizer() {
	p.tokenizer = newTokenizer(p.input)
}

func (p *parser) decl(tokens []token, customProperty bool) {
	node := newDeclaration()
	p.init(node, tokens[0].start)

	last := tokens[len(tokens)-1]
	if last.kind == ";" {
		p.semicolon = true
		tokens = tokens[:len(tokens)-1]
	}

	endOffset := last.endOrStart()
	if endOffset <= 0 {
		endOffset = findLastWithPosition(tokens)
	}
	node.source.end = p.getPosition(endOffset)
	node.source.end.offset++

	start := 0
	for tokens[start].kind != "word" {
		if start == len(tokens)-1 {
			p.unknownWord([]token{tokens[start]})
		}
		start++
	}
	node.raws["before"] = node.rawString("before") + tokensToString(tokens, 0, start)
	node.source.start = p.getPosition(tokens[start].start)

	propStart := start
	for start < len(tokens) {
		kind := tokens[start].kind
		if kind == ":" || kind == "space" || kind == "comment" {
			break
		}
		start++
	}
	node.prop = tokensToString(tokens, propStart, start)
	node.mark("prop")

	betweenStart := start
	var each token
	for start < len(tokens) {
		each = tokens[start]
		start++
		if each.kind == ":" {
			break
		}
		if each.kind == "word" && hasWordCharacter(each.value) {
			p.unknownWord([]token{each})
		}
	}
	node.raws["between"] = tokensToString(tokens, betweenStart, start)

	if node.prop[0] == '_' || node.prop[0] == '*' {
		node.raws["before"] = node.rawString("before") + node.prop[:1]
		node.prop = node.prop[1:]
	}

	firstSpacesStart := start
	for start < len(tokens) {
		next := tokens[start].kind
		if next != "space" && next != "comment" {
			break
		}
		start++
	}
	firstSpaces := copyTokens(tokens, firstSpacesStart, start)

	tokens = copyTokens(tokens, start, len(tokens))

	// precheckMissedSemicolon is a hook for postcss-safe-parser; the css parser's does nothing.

	for index := len(tokens) - 1; index >= 0; index-- {
		each = tokens[index]
		lowered := asciiLowerCase(each.value)
		if lowered == "!important" {
			node.important = true
			node.mark("important")
			text := p.stringFrom(&tokens, index)
			text = p.spacesFromEnd(&tokens) + text
			if text != " !important" {
				node.raws["important"] = text
			}
			break
		} else if lowered == "important" {
			cache := copyTokens(tokens, 0, len(tokens))
			str := ""
			for j := index; j > 0; j-- {
				kind := cache[j].kind
				if strings.HasPrefix(strings.TrimFunc(str, isJavaScriptWhitespace), "!") && kind != "space" {
					break
				}
				str = cache[len(cache)-1].value + str
				cache = cache[:len(cache)-1]
			}
			if strings.HasPrefix(strings.TrimFunc(str, isJavaScriptWhitespace), "!") {
				node.important = true
				node.mark("important")
				node.raws["important"] = str
				tokens = cache
			}
		}

		if each.kind != "space" && each.kind != "comment" {
			break
		}
	}

	hasWord := false
	for _, candidate := range tokens {
		if candidate.kind != "space" && candidate.kind != "comment" {
			hasWord = true
			break
		}
	}

	if hasWord {
		node.raws["between"] = node.rawString("between") + tokensToString(firstSpaces, 0, len(firstSpaces))
		firstSpaces = []token{}
	}
	valueTokens := make([]token, 0, len(firstSpaces)+len(tokens))
	valueTokens = append(valueTokens, firstSpaces...)
	valueTokens = append(valueTokens, tokens...)
	p.this.raw(node, "value", valueTokens, customProperty)

	if strings.Contains(node.value, ":") && !customProperty {
		p.checkMissedSemicolon(tokens)
	}
}

func (p *parser) doubleColon(each token) {
	panic(p.input.errorBetween("Double colon", each.start, each.start+utf16Length(each.value)))
}

func (p *parser) emptyRule(each token) {
	node := newRule()
	p.init(node, each.start)
	node.selector = ""
	node.mark("selector")
	node.raws["between"] = ""
	p.current = node
}

func (p *parser) end(each token) {
	if p.current.hasNodes() && len(p.current.nodes) > 0 {
		p.current.raws["semicolon"] = p.semicolon
	}
	p.semicolon = false

	p.current.raws["after"] = p.current.rawString("after") + p.spaces
	p.spaces = ""

	if p.current.parent != nil {
		p.current.source.end = p.getPosition(each.start)
		p.current.source.end.offset++
		p.current = p.current.parent
	} else {
		p.unexpectedClose(each)
	}
}

func (p *parser) endFile() {
	if p.current.parent != nil {
		p.unclosedBlock()
	}
	if p.current.hasNodes() && len(p.current.nodes) > 0 {
		p.current.raws["semicolon"] = p.semicolon
	}
	p.current.raws["after"] = p.current.rawString("after") + p.spaces
	p.root.source.end = p.getPosition(p.tokenizer.position())
}

func (p *parser) freeSemicolon(each token) {
	p.spaces += each.value
	if p.current.hasNodes() {
		var prev *node
		if len(p.current.nodes) > 0 {
			prev = p.current.nodes[len(p.current.nodes)-1]
		}
		if prev != nil && prev.nodeType == "rule" && prev.rawString("ownSemicolon") == "" {
			prev.raws["ownSemicolon"] = p.spaces
			p.spaces = ""
			prev.source.end = p.getPosition(each.start)
			prev.source.end.offset += utf16Length(prev.rawString("ownSemicolon"))
		}
	}
}

// Helpers

func (p *parser) getPosition(offset int) *position {
	pos := p.input.fromOffset(offset)
	return &position{
		column: pos.column,
		line:   pos.line,
		offset: offset,
	}
}

func (p *parser) init(each *node, offset int) {
	p.current.push(each)
	each.source = &source{start: p.getPosition(offset)}
	each.mark("source")
	each.raws["before"] = p.spaces
	p.spaces = ""
	if each.nodeType != "comment" {
		p.semicolon = false
	}
}

func (p *parser) other(start token) {
	end := false
	kind := ""
	colon := false
	var bracket *token
	brackets := []string{}
	customProperty := strings.HasPrefix(start.value, "--")

	tokens := []token{}
	each := start
	present := true
	for present {
		kind = each.kind
		tokens = append(tokens, each)

		if kind == "(" || kind == "[" {
			if bracket == nil {
				opened := each
				bracket = &opened
			}
			if kind == "(" {
				brackets = append(brackets, ")")
			} else {
				brackets = append(brackets, "]")
			}
		} else if customProperty && colon && kind == "{" {
			if bracket == nil {
				opened := each
				bracket = &opened
			}
			brackets = append(brackets, "}")
		} else if len(brackets) == 0 {
			if kind == ";" {
				if colon {
					p.decl(tokens, customProperty)
					return
				} else {
					break
				}
			} else if kind == "{" {
				p.this.rule(tokens)
				return
			} else if kind == "}" {
				p.tokenizer.back(tokens[len(tokens)-1])
				tokens = tokens[:len(tokens)-1]
				end = true
				break
			} else if kind == ":" {
				colon = true
			}
		} else if kind == brackets[len(brackets)-1] {
			brackets = brackets[:len(brackets)-1]
			if len(brackets) == 0 {
				bracket = nil
			}
		}

		each, present = p.tokenizer.nextToken()
	}

	if p.tokenizer.endOfFile() {
		end = true
	}
	if len(brackets) > 0 {
		p.unclosedBracket(*bracket)
	}

	if end && colon {
		if !customProperty {
			for len(tokens) > 0 {
				kind = tokens[len(tokens)-1].kind
				if kind != "space" && kind != "comment" {
					break
				}
				p.tokenizer.back(tokens[len(tokens)-1])
				tokens = tokens[:len(tokens)-1]
			}
		}
		p.decl(tokens, customProperty)
	} else {
		p.unknownWord(tokens)
	}
}

func (p *parser) parse() {
	for !p.tokenizer.endOfFile() {
		each, _ := p.tokenizer.nextToken()

		switch each.kind {
		case "space":
			p.spaces += each.value

		case ";":
			p.freeSemicolon(each)

		case "}":
			p.end(each)

		case "comment":
			p.this.comment(each)

		case "at-word":
			p.this.atrule(each)

		case "{":
			p.emptyRule(each)

		default:
			p.other(each)
		}
	}
	p.endFile()
}

func (p *parser) raw(each *node, prop string, tokens []token, customProperty bool) {
	length := len(tokens)
	var value strings.Builder
	clean := true

	for index := 0; index < length; index++ {
		current := tokens[index]
		kind := current.kind
		if kind == "space" && index == length-1 && !customProperty {
			clean = false
		} else if kind == "comment" {
			prev := "empty"
			if index-1 >= 0 {
				prev = tokens[index-1].kind
			}
			next := "empty"
			if index+1 < length {
				next = tokens[index+1].kind
			}
			if !safeCommentNeighbor(prev) && !safeCommentNeighbor(next) {
				if strings.HasSuffix(value.String(), ",") {
					clean = false
				} else {
					value.WriteString(current.value)
				}
			} else {
				clean = false
			}
		} else {
			value.WriteString(current.value)
		}
	}
	if !clean {
		raw := tokensToString(tokens, 0, length)
		each.raws[prop] = &rawValue{raw: raw, value: value.String()}
	}
	each.setField(prop, value.String())
}

func (p *parser) rule(tokens []token) {
	tokens = tokens[:len(tokens)-1]

	node := newRule()
	p.init(node, tokens[0].start)

	node.raws["between"] = p.spacesAndCommentsFromEnd(&tokens)
	p.this.raw(node, "selector", tokens, false)
	p.current = node
}

func (p *parser) spacesAndCommentsFromEnd(tokens *[]token) string {
	spaces := ""
	for len(*tokens) > 0 {
		lastTokenType := (*tokens)[len(*tokens)-1].kind
		if lastTokenType != "space" && lastTokenType != "comment" {
			break
		}
		spaces = (*tokens)[len(*tokens)-1].value + spaces
		*tokens = (*tokens)[:len(*tokens)-1]
	}
	return spaces
}

// Errors

func (p *parser) spacesAndCommentsFromStart(tokens *[]token) string {
	spaces := ""
	for len(*tokens) > 0 {
		next := (*tokens)[0].kind
		if next != "space" && next != "comment" {
			break
		}
		spaces += (*tokens)[0].value
		*tokens = (*tokens)[1:]
	}
	return spaces
}

func (p *parser) spacesFromEnd(tokens *[]token) string {
	spaces := ""
	for len(*tokens) > 0 {
		lastTokenType := (*tokens)[len(*tokens)-1].kind
		if lastTokenType != "space" {
			break
		}
		spaces = (*tokens)[len(*tokens)-1].value + spaces
		*tokens = (*tokens)[:len(*tokens)-1]
	}
	return spaces
}

func (p *parser) stringFrom(tokens *[]token, from int) string {
	result := tokensToString(*tokens, from, len(*tokens))
	*tokens = (*tokens)[:from]
	return result
}

func (p *parser) unclosedBlock() {
	pos := p.current.source.start
	panic(p.input.errorAtLineAndColumn("Unclosed block", pos.line, pos.column))
}

func (p *parser) unclosedBracket(bracket token) {
	panic(p.input.errorBetween("Unclosed bracket", bracket.start, bracket.start+1))
}

func (p *parser) unexpectedClose(each token) {
	panic(p.input.errorBetween("Unexpected }", each.start, each.start+1))
}

func (p *parser) unknownWord(tokens []token) {
	panic(p.input.errorBetween("Unknown word "+tokens[0].value, tokens[0].start, tokens[0].start+utf16Length(tokens[0].value)))
}

func (p *parser) unnamedAtrule(_ *node, each token) {
	panic(p.input.errorBetween("At-rule without name", each.start, each.start+utf16Length(each.value)))
}

// isJavaScriptWhitespace is a character JavaScript's \s and String.prototype.trim treat as whitespace:
// WhiteSpace and LineTerminator in ECMA-262, the byte order mark among them.
func isJavaScriptWhitespace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0xA0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return character >= 0x2000 && character <= 0x200A
}

// hasWordCharacter is /\w/.test(text): an ASCII letter, digit or underscore.
func hasWordCharacter(text string) bool {
	for index := 0; index < len(text); index++ {
		character := text[index]
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '_' {
			return true
		}
	}
	return false
}

// asciiLowerCase stands in for toLowerCase where the result is only compared with "!important" and
// "important": no character outside ASCII lowercases to one of those strings' letters in JavaScript
// (U+0130 becomes "i" plus a combining dot, two units), so lowering ASCII alone decides the same.
// Go's strings.ToLower would not: it maps U+0130 to a plain "i".
func asciiLowerCase(text string) string {
	lowered := []byte(text)
	for index, character := range lowered {
		if character >= 'A' && character <= 'Z' {
			lowered[index] = character + 'a' - 'A'
		}
	}
	return string(lowered)
}

// utf16Length is a string's JavaScript length: its UTF-16 units, a byte that is not UTF-8 counting one.
func utf16Length(text string) int {
	length := 0
	for _, character := range text {
		if character > 0xFFFF {
			length += 2
		} else {
			length++
		}
	}
	return length
}
