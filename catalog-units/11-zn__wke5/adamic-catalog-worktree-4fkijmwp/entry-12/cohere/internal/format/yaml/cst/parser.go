package cst

// Ported from eemeli/yaml 2.9.0, dist/parse/parser.js.
//
// Upstream builds tokens as objects and moves arrays between them: `top.end = last.start`, `it.value.end
// = it.sep`, `sep = scalar.end; delete scalar.end`. Each move hands the array to one new owner, so a Go
// slice field assigned the same way behaves the same. Where upstream mutates an array in place through
// a reference it got from elsewhere (getPrevProps then getFirstKeyStartProps splicing it, `end.push`
// through `it.value.end`), the Go passes a pointer to the owning field or writes the field back, and
// every array spliced out is copied, so a later append to the remainder can never write into it.

import (
	"github.com/system-inc/cohere/internal/format/arena"
	"iter"
	"unicode/utf16"
)

func includesToken(list []*Token, tokenType string) bool {
	for i := 0; i < len(list); i++ {
		if list[i].Type == tokenType {
			return true
		}
	}
	return false
}

func findNonEmptyIndex(list []*Token) int {
	for i := 0; i < len(list); i++ {
		switch list[i].Type {
		case "space", "comment", "newline":
		default:
			return i
		}
	}
	return -1
}

func isFlowToken(token *Token) bool {
	if token == nil {
		return false
	}
	switch token.Type {
	case "alias", "scalar", "single-quoted-scalar", "double-quoted-scalar", "flow-collection":
		return true
	default:
		return false
	}
}

// getPrevProps returns the array upstream returns, as a pointer to the field that owns it, because
// getFirstKeyStartProps splices it in place.
func getPrevProps(parent *Token) *[]*Token {
	switch parent.Type {
	case "document":
		return &parent.Start
	case "block-map":
		it := parent.Items[len(parent.Items)-1]
		// `it.sep ?? it.start`: an empty sep is still sep.
		if it.Sep != nil {
			return &it.Sep
		}
		return &it.Start
	case "block-seq":
		return &parent.Items[len(parent.Items)-1].Start
	default:
		// should not happen
		empty := []*Token{}
		return &empty
	}
}

// getFirstKeyStartProps. Note: May modify input array.
func getFirstKeyStartProps(prevField *[]*Token) []*Token {
	prev := *prevField
	if len(prev) == 0 {
		return []*Token{}
	}
	i := len(prev)
loop:
	for {
		i--
		if i < 0 {
			break
		}
		switch prev[i].Type {
		case "doc-start", "explicit-key-ind", "map-value-ind", "seq-item-ind", "newline":
			break loop
		}
	}
	for {
		// `while (prev[++i]?.type === 'space')`
		i++
		if !(i < len(prev) && prev[i].Type == "space") {
			break
		}
	}
	// `prev.splice(i, prev.length)`
	return spliceTail(prevField, i)
}

// spliceTail is `list.splice(start)`: it removes the elements from start on and returns them, as a
// new array.
func spliceTail(field *[]*Token, start int) []*Token {
	list := *field
	start = clamp(start, 0, len(list))
	removed := append([]*Token{}, list[start:]...)
	*field = list[:start:start]
	return removed
}

func fixFlowSeqItems(fc *Token) {
	if fc.FlowStart.Type == "flow-seq-start" {
		for _, it := range fc.Items {
			if it.Sep != nil &&
				it.Value == nil &&
				!includesToken(it.Start, "explicit-key-ind") &&
				!includesToken(it.Sep, "map-value-ind") {
				if it.Key != nil {
					it.Value = it.Key
				}
				it.deleteKey()
				if isFlowToken(it.Value) {
					if it.Value.End != nil {
						it.Value.End = append(it.Value.End, it.Sep...)
					} else {
						it.Value.End = it.Sep
					}
				} else {
					it.Start = append(it.Start, it.Sep...)
				}
				it.Sep = nil
			}
		}
	}
}

// Parser is a YAML concrete syntax tree (CST) parser.
//
//	for token := range cst.NewParser(nil).Parse(source, false) {
//		// token: *Token
//	}
type Parser struct {
	// atNewLine: if true, space and sequence indicators count as indentation.
	atNewLine bool
	// atScalar: if true, next token is a scalar value.
	atScalar bool
	// indent: current indentation level.
	indent int
	// offset: current offset since the start of parsing.
	offset int
	// onKeyLine: on the same line with a block map key.
	onKeyLine bool
	// stack: top indicates the node that's currently being built.
	stack []*Token
	// source: the source of the current token, set in parse().
	source []uint16
	// tokenType: the type of the current token, set in parse(). Upstream's this.type.
	tokenType string

	lexer     *Lexer
	onNewLine func(offset int)

	// Tokens is where the parse's tokens come from, nil to allocate each (#v6ksqg3). A caller that sets it
	// owns the tokens' lifetime: unist.Parse resets it once the tree is built, since no node keeps a token.
	Tokens *arena.Arena[Token]

	// yield is the consumer of the running Parse iteration; stopped is set once it returns false.
	yield   func(*Token) bool
	stopped bool
}

// NewParser is upstream's `new Parser(onNewLine)`. If onNewLine is not nil, it is called separately with
// the start position of each new line (in Parse, including the start of input).
func NewParser(onNewLine func(offset int)) *Parser {
	return &Parser{
		atNewLine: true,
		stack:     []*Token{},
		source:    []uint16{},
		lexer:     NewLexer(),
		onNewLine: onNewLine,
	}
}

// emit is upstream's `yield token`.
func (parser *Parser) emit(token *Token) {
	if parser.stopped {
		return
	}
	if !parser.yield(token) {
		parser.stopped = true
	}
}

// Parse parses source as a YAML stream. If incomplete, a part of the last line may be left as a buffer for
// the next call.
//
// Errors are not thrown, but yielded as error tokens.
//
// It returns an iterator over the tokens representing each directive, document, and other structure.
func (parser *Parser) Parse(source []uint16, incomplete bool) iter.Seq[*Token] {
	return func(yield func(*Token) bool) {
		parser.yield = yield
		parser.stopped = false
		if parser.onNewLine != nil && parser.offset == 0 {
			parser.onNewLine(0)
		}
		for lexeme := range parser.lexer.Lex(source, incomplete) {
			parser.next(lexeme)
			if parser.stopped {
				return
			}
		}
		if !incomplete {
			parser.end()
		}
	}
}

// next advances the parser by the source of one lexical token. Upstream's LOG_TOKENS logging is not
// ported.
func (parser *Parser) next(source []uint16) {
	parser.source = source
	if parser.atScalar {
		parser.atScalar = false
		parser.step()
		parser.offset += len(source)
		return
	}
	tokenType := TokenType(source)
	if tokenType == "" {
		message := "Not a YAML token: " + string(utf16.Decode(source))
		parser.pop(parser.Tokens.New(Token{Type: "error", Offset: parser.offset, Message: message, Source: source, indentAbsent: true}))
		parser.offset += len(source)
	} else if tokenType == "scalar" {
		parser.atNewLine = false
		parser.atScalar = true
		parser.tokenType = "scalar"
	} else {
		parser.tokenType = tokenType
		parser.step()
		switch tokenType {
		case "newline":
			parser.atNewLine = true
			parser.indent = 0
			if parser.onNewLine != nil {
				parser.onNewLine(parser.offset + len(source))
			}
		case "space":
			if parser.atNewLine && source[0] == ' ' {
				parser.indent += len(source)
			}
		case "explicit-key-ind", "map-value-ind", "seq-item-ind":
			if parser.atNewLine {
				parser.indent += len(source)
			}
		case "doc-mode", "flow-error-end":
			return
		default:
			parser.atNewLine = false
		}
		parser.offset += len(source)
	}
}

// end is called at end of input to push out any remaining constructions.
func (parser *Parser) end() {
	for len(parser.stack) > 0 {
		parser.pop(nil)
	}
}

// sourceToken is upstream's sourceToken getter: a new token each time.
func (parser *Parser) sourceToken() *Token {
	return parser.Tokens.New(Token{
		Type:   parser.tokenType,
		Offset: parser.offset,
		Indent: parser.indent,
		Source: parser.source,
	})
}

func (parser *Parser) step() {
	top := parser.peek(1)
	if parser.tokenType == "doc-end" && (top == nil || top.Type != "doc-end") {
		for len(parser.stack) > 0 {
			parser.pop(nil)
		}
		parser.stack = append(parser.stack, parser.Tokens.New(Token{
			Type:         "doc-end",
			Offset:       parser.offset,
			Source:       parser.source,
			indentAbsent: true,
		}))
		return
	}
	if top == nil {
		parser.stream()
		return
	}
	switch top.Type {
	case "document":
		parser.document(top)
		return
	case "alias", "scalar", "single-quoted-scalar", "double-quoted-scalar":
		parser.scalar(top)
		return
	case "block-scalar":
		parser.blockScalar(top)
		return
	case "block-map":
		parser.blockMap(top)
		return
	case "block-seq":
		parser.blockSequence(top)
		return
	case "flow-collection":
		parser.flowCollection(top)
		return
	case "doc-end":
		parser.documentEnd(top)
		return
	}
	// should not happen
	parser.pop(nil)
}

// peek returns nil for upstream's undefined.
func (parser *Parser) peek(n int) *Token {
	index := len(parser.stack) - n
	if index < 0 || index >= len(parser.stack) {
		return nil
	}
	return parser.stack[index]
}

// pop is `pop(error)`: with a nil errorToken, it pops the top of the stack.
func (parser *Parser) pop(errorToken *Token) {
	token := errorToken
	if token == nil && len(parser.stack) > 0 {
		token = parser.stack[len(parser.stack)-1]
		parser.stack = parser.stack[:len(parser.stack)-1]
	}
	if token == nil {
		// should not happen
		message := "Tried to pop an empty stack"
		parser.emit(parser.Tokens.New(Token{Type: "error", Offset: parser.offset, Source: []uint16{}, Message: message, indentAbsent: true}))
	} else if len(parser.stack) == 0 {
		parser.emit(token)
	} else {
		top := parser.peek(1)
		if token.Type == "block-scalar" {
			// Block scalars use their parent rather than header indent. `'indent' in top ? top.indent
			// : 0`: a top without an indent field has Indent 0.
			if top.HasIndent() {
				token.Indent = top.Indent
			} else {
				token.Indent = 0
			}
		} else if token.Type == "flow-collection" && top.Type == "document" {
			// Ignore all indent for top-level flow collections
			token.Indent = 0
		}
		if token.Type == "flow-collection" {
			fixFlowSeqItems(token)
		}
		switch top.Type {
		case "document":
			top.Value = token
		case "block-scalar":
			top.Props = append(top.Props, token) // error
		case "block-map":
			it := top.Items[len(top.Items)-1]
			if it.Value != nil {
				item := &CollectionItem{Start: []*Token{}, Sep: []*Token{}}
				item.setKey(token)
				top.Items = append(top.Items, item)
				parser.onKeyLine = true
				return
			} else if it.Sep != nil {
				it.Value = token
			} else {
				it.setKey(token)
				it.Sep = []*Token{}
				parser.onKeyLine = !it.ExplicitKey
				return
			}
		case "block-seq":
			it := top.Items[len(top.Items)-1]
			if it.Value != nil {
				top.Items = append(top.Items, &CollectionItem{Start: []*Token{}, Value: token})
			} else {
				it.Value = token
			}
		case "flow-collection":
			var it *CollectionItem
			if len(top.Items) > 0 {
				it = top.Items[len(top.Items)-1]
			}
			if it == nil || it.Value != nil {
				item := &CollectionItem{Start: []*Token{}, Sep: []*Token{}}
				item.setKey(token)
				top.Items = append(top.Items, item)
			} else if it.Sep != nil {
				it.Value = token
			} else {
				it.setKey(token)
				it.Sep = []*Token{}
			}
			return
		default:
			// should not happen
			parser.pop(nil)
			parser.pop(token)
		}
		if (top.Type == "document" || top.Type == "block-map" || top.Type == "block-seq") &&
			(token.Type == "block-map" || token.Type == "block-seq") {
			if len(token.Items) > 0 {
				last := token.Items[len(token.Items)-1]
				if last.Sep == nil &&
					last.Value == nil &&
					len(last.Start) > 0 &&
					findNonEmptyIndex(last.Start) == -1 &&
					(token.Indent == 0 || everyCommentLessIndented(last.Start, token.Indent)) {
					if top.Type == "document" {
						top.End = last.Start
					} else {
						top.Items = append(top.Items, &CollectionItem{Start: last.Start})
					}
					// `token.items.splice(-1, 1)`
					token.Items = token.Items[: len(token.Items)-1 : len(token.Items)-1]
				}
			}
		}
	}
}

// everyCommentLessIndented is `start.every(st => st.type !== 'comment' || st.indent < indent)`.
func everyCommentLessIndented(start []*Token, indent int) bool {
	for _, st := range start {
		if !(st.Type != "comment" || st.Indent < indent) {
			return false
		}
	}
	return true
}

func (parser *Parser) stream() {
	switch parser.tokenType {
	case "directive-line":
		parser.emit(parser.Tokens.New(Token{Type: "directive", Offset: parser.offset, Source: parser.source, indentAbsent: true}))
		return
	case "byte-order-mark", "space", "comment", "newline":
		parser.emit(parser.sourceToken())
		return
	case "doc-mode", "doc-start":
		doc := parser.Tokens.New(Token{
			Type:         "document",
			Offset:       parser.offset,
			Start:        []*Token{},
			indentAbsent: true,
		})
		if parser.tokenType == "doc-start" {
			doc.Start = append(doc.Start, parser.sourceToken())
		}
		parser.stack = append(parser.stack, doc)
		return
	}
	parser.emit(parser.Tokens.New(Token{
		Type:         "error",
		Offset:       parser.offset,
		Message:      "Unexpected " + parser.tokenType + " token in YAML stream",
		Source:       parser.source,
		indentAbsent: true,
	}))
}

func (parser *Parser) document(doc *Token) {
	if doc.Value != nil {
		parser.lineEnd(doc)
		return
	}
	switch parser.tokenType {
	case "doc-start":
		if findNonEmptyIndex(doc.Start) != -1 {
			parser.pop(nil)
			parser.step()
		} else {
			doc.Start = append(doc.Start, parser.sourceToken())
		}
		return
	case "anchor", "tag", "space", "comment", "newline":
		doc.Start = append(doc.Start, parser.sourceToken())
		return
	}
	bv := parser.startBlockValue(doc)
	if bv != nil {
		parser.stack = append(parser.stack, bv)
	} else {
		parser.emit(parser.Tokens.New(Token{
			Type:         "error",
			Offset:       parser.offset,
			Message:      "Unexpected " + parser.tokenType + " token in YAML document",
			Source:       parser.source,
			indentAbsent: true,
		}))
	}
}

func (parser *Parser) scalar(scalar *Token) {
	if parser.tokenType == "map-value-ind" {
		prev := getPrevProps(parser.peek(2))
		start := getFirstKeyStartProps(prev)
		var sep []*Token
		if scalar.End != nil {
			sep = append(scalar.End, parser.sourceToken())
			scalar.End = nil
		} else {
			sep = []*Token{parser.sourceToken()}
		}
		item := &CollectionItem{Start: start, Sep: sep}
		item.setKey(scalar)
		blockMap := parser.Tokens.New(Token{
			Type:   "block-map",
			Offset: scalar.Offset,
			Indent: scalar.Indent,
			Items:  []*CollectionItem{item},
		})
		parser.onKeyLine = true
		parser.stack[len(parser.stack)-1] = blockMap
	} else {
		parser.lineEnd(scalar)
	}
}

func (parser *Parser) blockScalar(scalar *Token) {
	switch parser.tokenType {
	case "space", "comment", "newline":
		scalar.Props = append(scalar.Props, parser.sourceToken())
		return
	case "scalar":
		scalar.Source = parser.source
		// block-scalar source includes trailing newline
		parser.atNewLine = true
		parser.indent = 0
		if parser.onNewLine != nil {
			nl := indexOf(parser.source, '\n', 0) + 1
			for nl != 0 {
				parser.onNewLine(parser.offset + nl)
				nl = indexOf(parser.source, '\n', nl) + 1
			}
		}
		parser.pop(nil)
	default:
		// should not happen
		parser.pop(nil)
		parser.step()
	}
}

func (parser *Parser) blockMap(blockMap *Token) {
	it := blockMap.Items[len(blockMap.Items)-1]
	// it.sep is true-ish if pair already has key or : separator
	switch parser.tokenType {
	case "newline":
		parser.onKeyLine = false
		if it.Value != nil {
			// `'end' in it.value ? it.value.end : undefined`, then its last element.
			end := it.Value.End
			if len(end) > 0 && end[len(end)-1].Type == "comment" {
				it.Value.End = append(end, parser.sourceToken())
			} else {
				blockMap.Items = append(blockMap.Items, &CollectionItem{Start: []*Token{parser.sourceToken()}})
			}
		} else if it.Sep != nil {
			it.Sep = append(it.Sep, parser.sourceToken())
		} else {
			it.Start = append(it.Start, parser.sourceToken())
		}
		return
	case "space", "comment":
		if it.Value != nil {
			blockMap.Items = append(blockMap.Items, &CollectionItem{Start: []*Token{parser.sourceToken()}})
		} else if it.Sep != nil {
			it.Sep = append(it.Sep, parser.sourceToken())
		} else {
			if parser.atIndentedComment(it.Start, blockMap.Indent) {
				var prev *CollectionItem
				if len(blockMap.Items) >= 2 {
					prev = blockMap.Items[len(blockMap.Items)-2]
				}
				// `prev?.value?.end`, an array when present.
				if prev != nil && prev.Value != nil && prev.Value.End != nil {
					end := append(prev.Value.End, it.Start...)
					prev.Value.End = append(end, parser.sourceToken())
					blockMap.Items = blockMap.Items[: len(blockMap.Items)-1 : len(blockMap.Items)-1]
					return
				}
			}
			it.Start = append(it.Start, parser.sourceToken())
		}
		return
	}
	if parser.indent >= blockMap.Indent {
		atMapIndent := !parser.onKeyLine && parser.indent == blockMap.Indent
		atNextItem := atMapIndent &&
			(it.Sep != nil || it.ExplicitKey) &&
			parser.tokenType != "seq-item-ind"
		// For empty nodes, assign newline-separated not indented empty tokens to following node
		start := []*Token{}
		if atNextItem && it.Sep != nil && it.Value == nil {
			nl := []int{}
			for i := 0; i < len(it.Sep); i++ {
				st := it.Sep[i]
				switch st.Type {
				case "newline":
					nl = append(nl, i)
				case "space":
				case "comment":
					if st.Indent > blockMap.Indent {
						nl = nl[:0]
					}
				default:
					nl = nl[:0]
				}
			}
			if len(nl) >= 2 {
				start = spliceTail(&it.Sep, nl[1])
			}
		}
		switch parser.tokenType {
		case "anchor", "tag":
			if atNextItem || it.Value != nil {
				start = append(start, parser.sourceToken())
				blockMap.Items = append(blockMap.Items, &CollectionItem{Start: start})
				parser.onKeyLine = true
			} else if it.Sep != nil {
				it.Sep = append(it.Sep, parser.sourceToken())
			} else {
				it.Start = append(it.Start, parser.sourceToken())
			}
			return
		case "explicit-key-ind":
			if it.Sep == nil && !it.ExplicitKey {
				it.Start = append(it.Start, parser.sourceToken())
				it.ExplicitKey = true
			} else if atNextItem || it.Value != nil {
				start = append(start, parser.sourceToken())
				blockMap.Items = append(blockMap.Items, &CollectionItem{Start: start, ExplicitKey: true})
			} else {
				parser.stack = append(parser.stack, parser.Tokens.New(Token{
					Type:   "block-map",
					Offset: parser.offset,
					Indent: parser.indent,
					Items:  []*CollectionItem{{Start: []*Token{parser.sourceToken()}, ExplicitKey: true}},
				}))
			}
			parser.onKeyLine = true
			return
		case "map-value-ind":
			if it.ExplicitKey {
				if it.Sep == nil {
					if includesToken(it.Start, "newline") {
						it.setKey(nil)
						it.Sep = []*Token{parser.sourceToken()}
					} else {
						start := getFirstKeyStartProps(&it.Start)
						item := &CollectionItem{Start: start, Sep: []*Token{parser.sourceToken()}}
						item.setKey(nil)
						parser.stack = append(parser.stack, parser.Tokens.New(Token{
							Type:   "block-map",
							Offset: parser.offset,
							Indent: parser.indent,
							Items:  []*CollectionItem{item},
						}))
					}
				} else if it.Value != nil {
					item := &CollectionItem{Start: []*Token{}, Sep: []*Token{parser.sourceToken()}}
					item.setKey(nil)
					blockMap.Items = append(blockMap.Items, item)
				} else if includesToken(it.Sep, "map-value-ind") {
					item := &CollectionItem{Start: start, Sep: []*Token{parser.sourceToken()}}
					item.setKey(nil)
					parser.stack = append(parser.stack, parser.Tokens.New(Token{
						Type:   "block-map",
						Offset: parser.offset,
						Indent: parser.indent,
						Items:  []*CollectionItem{item},
					}))
				} else if isFlowToken(it.Key) &&
					!includesToken(it.Sep, "newline") {
					start := getFirstKeyStartProps(&it.Start)
					key := it.Key
					sep := append(it.Sep, parser.sourceToken())
					it.deleteKey()
					it.Sep = nil
					item := &CollectionItem{Start: start, Sep: sep}
					item.setKey(key)
					parser.stack = append(parser.stack, parser.Tokens.New(Token{
						Type:   "block-map",
						Offset: parser.offset,
						Indent: parser.indent,
						Items:  []*CollectionItem{item},
					}))
				} else if len(start) > 0 {
					// Not actually at next item
					// `it.sep.concat(start, this.sourceToken)`: a new array.
					sep := make([]*Token, 0, len(it.Sep)+len(start)+1)
					sep = append(sep, it.Sep...)
					sep = append(sep, start...)
					it.Sep = append(sep, parser.sourceToken())
				} else {
					it.Sep = append(it.Sep, parser.sourceToken())
				}
			} else {
				if it.Sep == nil {
					it.setKey(nil)
					it.Sep = []*Token{parser.sourceToken()}
				} else if it.Value != nil || atNextItem {
					item := &CollectionItem{Start: start, Sep: []*Token{parser.sourceToken()}}
					item.setKey(nil)
					blockMap.Items = append(blockMap.Items, item)
				} else if includesToken(it.Sep, "map-value-ind") {
					item := &CollectionItem{Start: []*Token{}, Sep: []*Token{parser.sourceToken()}}
					item.setKey(nil)
					parser.stack = append(parser.stack, parser.Tokens.New(Token{
						Type:   "block-map",
						Offset: parser.offset,
						Indent: parser.indent,
						Items:  []*CollectionItem{item},
					}))
				} else {
					it.Sep = append(it.Sep, parser.sourceToken())
				}
			}
			parser.onKeyLine = true
			return
		case "alias", "scalar", "single-quoted-scalar", "double-quoted-scalar":
			fs := parser.flowScalar(parser.tokenType)
			if atNextItem || it.Value != nil {
				item := &CollectionItem{Start: start, Sep: []*Token{}}
				item.setKey(fs)
				blockMap.Items = append(blockMap.Items, item)
				parser.onKeyLine = true
			} else if it.Sep != nil {
				parser.stack = append(parser.stack, fs)
			} else {
				it.setKey(fs)
				it.Sep = []*Token{}
				parser.onKeyLine = true
			}
			return
		default:
			bv := parser.startBlockValue(blockMap)
			if bv != nil {
				if bv.Type == "block-seq" {
					if !it.ExplicitKey &&
						it.Sep != nil &&
						!includesToken(it.Sep, "newline") {
						parser.pop(parser.Tokens.New(Token{
							Type:         "error",
							Offset:       parser.offset,
							Message:      "Unexpected block-seq-ind on same line with key",
							Source:       parser.source,
							indentAbsent: true,
						}))
						return
					}
				} else if atMapIndent {
					blockMap.Items = append(blockMap.Items, &CollectionItem{Start: start})
				}
				parser.stack = append(parser.stack, bv)
				return
			}
		}
	}
	parser.pop(nil)
	parser.step()
}

func (parser *Parser) blockSequence(seq *Token) {
	it := seq.Items[len(seq.Items)-1]
	switch parser.tokenType {
	case "newline":
		if it.Value != nil {
			end := it.Value.End
			if len(end) > 0 && end[len(end)-1].Type == "comment" {
				it.Value.End = append(end, parser.sourceToken())
			} else {
				seq.Items = append(seq.Items, &CollectionItem{Start: []*Token{parser.sourceToken()}})
			}
		} else {
			it.Start = append(it.Start, parser.sourceToken())
		}
		return
	case "space", "comment":
		if it.Value != nil {
			seq.Items = append(seq.Items, &CollectionItem{Start: []*Token{parser.sourceToken()}})
		} else {
			if parser.atIndentedComment(it.Start, seq.Indent) {
				var prev *CollectionItem
				if len(seq.Items) >= 2 {
					prev = seq.Items[len(seq.Items)-2]
				}
				if prev != nil && prev.Value != nil && prev.Value.End != nil {
					end := append(prev.Value.End, it.Start...)
					prev.Value.End = append(end, parser.sourceToken())
					seq.Items = seq.Items[: len(seq.Items)-1 : len(seq.Items)-1]
					return
				}
			}
			it.Start = append(it.Start, parser.sourceToken())
		}
		return
	case "anchor", "tag":
		if it.Value != nil || parser.indent <= seq.Indent {
			break
		}
		it.Start = append(it.Start, parser.sourceToken())
		return
	case "seq-item-ind":
		if parser.indent != seq.Indent {
			break
		}
		if it.Value != nil || includesToken(it.Start, "seq-item-ind") {
			seq.Items = append(seq.Items, &CollectionItem{Start: []*Token{parser.sourceToken()}})
		} else {
			it.Start = append(it.Start, parser.sourceToken())
		}
		return
	}
	if parser.indent > seq.Indent {
		bv := parser.startBlockValue(seq)
		if bv != nil {
			parser.stack = append(parser.stack, bv)
			return
		}
	}
	parser.pop(nil)
	parser.step()
}

func (parser *Parser) flowCollection(fc *Token) {
	var it *CollectionItem
	if len(fc.Items) > 0 {
		it = fc.Items[len(fc.Items)-1]
	}
	if parser.tokenType == "flow-error-end" {
		for {
			parser.pop(nil)
			top := parser.peek(1)
			if top == nil || top.Type != "flow-collection" {
				break
			}
		}
	} else if len(fc.End) == 0 {
		switch parser.tokenType {
		case "comma", "explicit-key-ind":
			if it == nil || it.Sep != nil {
				fc.Items = append(fc.Items, &CollectionItem{Start: []*Token{parser.sourceToken()}})
			} else {
				it.Start = append(it.Start, parser.sourceToken())
			}
			return
		case "map-value-ind":
			if it == nil || it.Value != nil {
				item := &CollectionItem{Start: []*Token{}, Sep: []*Token{parser.sourceToken()}}
				item.setKey(nil)
				fc.Items = append(fc.Items, item)
			} else if it.Sep != nil {
				it.Sep = append(it.Sep, parser.sourceToken())
			} else {
				it.setKey(nil)
				it.Sep = []*Token{parser.sourceToken()}
			}
			return
		case "space", "comment", "newline", "anchor", "tag":
			if it == nil || it.Value != nil {
				fc.Items = append(fc.Items, &CollectionItem{Start: []*Token{parser.sourceToken()}})
			} else if it.Sep != nil {
				it.Sep = append(it.Sep, parser.sourceToken())
			} else {
				it.Start = append(it.Start, parser.sourceToken())
			}
			return
		case "alias", "scalar", "single-quoted-scalar", "double-quoted-scalar":
			fs := parser.flowScalar(parser.tokenType)
			if it == nil || it.Value != nil {
				item := &CollectionItem{Start: []*Token{}, Sep: []*Token{}}
				item.setKey(fs)
				fc.Items = append(fc.Items, item)
			} else if it.Sep != nil {
				parser.stack = append(parser.stack, fs)
			} else {
				it.setKey(fs)
				it.Sep = []*Token{}
			}
			return
		case "flow-map-end", "flow-seq-end":
			fc.End = append(fc.End, parser.sourceToken())
			return
		}
		bv := parser.startBlockValue(fc)
		// should not happen
		if bv != nil {
			parser.stack = append(parser.stack, bv)
		} else {
			parser.pop(nil)
			parser.step()
		}
	} else {
		parent := parser.peek(2)
		if parent.Type == "block-map" &&
			((parser.tokenType == "map-value-ind" && parent.Indent == fc.Indent) ||
				(parser.tokenType == "newline" &&
					parent.Items[len(parent.Items)-1].Sep == nil)) {
			parser.pop(nil)
			parser.step()
		} else if parser.tokenType == "map-value-ind" &&
			parent.Type != "flow-collection" {
			prev := getPrevProps(parent)
			start := getFirstKeyStartProps(prev)
			fixFlowSeqItems(fc)
			// `fc.end.splice(1, fc.end.length)`
			sep := spliceTail(&fc.End, 1)
			sep = append(sep, parser.sourceToken())
			item := &CollectionItem{Start: start, Sep: sep}
			item.setKey(fc)
			blockMap := parser.Tokens.New(Token{
				Type:   "block-map",
				Offset: fc.Offset,
				Indent: fc.Indent,
				Items:  []*CollectionItem{item},
			})
			parser.onKeyLine = true
			parser.stack[len(parser.stack)-1] = blockMap
		} else {
			parser.lineEnd(fc)
		}
	}
}

func (parser *Parser) flowScalar(tokenType string) *Token {
	if parser.onNewLine != nil {
		nl := indexOf(parser.source, '\n', 0) + 1
		for nl != 0 {
			parser.onNewLine(parser.offset + nl)
			nl = indexOf(parser.source, '\n', nl) + 1
		}
	}
	return parser.Tokens.New(Token{
		Type:   tokenType,
		Offset: parser.offset,
		Indent: parser.indent,
		Source: parser.source,
	})
}

// startBlockValue returns nil for upstream's null.
func (parser *Parser) startBlockValue(parent *Token) *Token {
	switch parser.tokenType {
	case "alias", "scalar", "single-quoted-scalar", "double-quoted-scalar":
		return parser.flowScalar(parser.tokenType)
	case "block-scalar-header":
		return parser.Tokens.New(Token{
			Type:   "block-scalar",
			Offset: parser.offset,
			Indent: parser.indent,
			Props:  []*Token{parser.sourceToken()},
			Source: []uint16{},
		})
	case "flow-map-start", "flow-seq-start":
		return parser.Tokens.New(Token{
			Type:      "flow-collection",
			Offset:    parser.offset,
			Indent:    parser.indent,
			FlowStart: parser.sourceToken(),
			Items:     []*CollectionItem{},
			End:       []*Token{},
		})
	case "seq-item-ind":
		return parser.Tokens.New(Token{
			Type:   "block-seq",
			Offset: parser.offset,
			Indent: parser.indent,
			Items:  []*CollectionItem{{Start: []*Token{parser.sourceToken()}}},
		})
	case "explicit-key-ind":
		parser.onKeyLine = true
		prev := getPrevProps(parent)
		start := getFirstKeyStartProps(prev)
		start = append(start, parser.sourceToken())
		return parser.Tokens.New(Token{
			Type:   "block-map",
			Offset: parser.offset,
			Indent: parser.indent,
			Items:  []*CollectionItem{{Start: start, ExplicitKey: true}},
		})
	case "map-value-ind":
		parser.onKeyLine = true
		prev := getPrevProps(parent)
		start := getFirstKeyStartProps(prev)
		item := &CollectionItem{Start: start, Sep: []*Token{parser.sourceToken()}}
		item.setKey(nil)
		return parser.Tokens.New(Token{
			Type:   "block-map",
			Offset: parser.offset,
			Indent: parser.indent,
			Items:  []*CollectionItem{item},
		})
	}
	return nil
}

func (parser *Parser) atIndentedComment(start []*Token, indent int) bool {
	if parser.tokenType != "comment" {
		return false
	}
	if parser.indent <= indent {
		return false
	}
	for _, st := range start {
		if !(st.Type == "newline" || st.Type == "space") {
			return false
		}
	}
	return true
}

func (parser *Parser) documentEnd(docEnd *Token) {
	if parser.tokenType != "doc-mode" {
		if docEnd.End != nil {
			docEnd.End = append(docEnd.End, parser.sourceToken())
		} else {
			docEnd.End = []*Token{parser.sourceToken()}
		}
		if parser.tokenType == "newline" {
			parser.pop(nil)
		}
	}
}

func (parser *Parser) lineEnd(token *Token) {
	switch parser.tokenType {
	case "comma", "doc-start", "doc-end", "flow-seq-end", "flow-map-end", "map-value-ind":
		parser.pop(nil)
		parser.step()
	case "newline":
		parser.onKeyLine = false
		fallthrough
	default:
		// space, comment, and all other values (which are errors)
		if token.End != nil {
			token.End = append(token.End, parser.sourceToken())
		} else {
			token.End = []*Token{parser.sourceToken()}
		}
		if parser.tokenType == "newline" {
			parser.pop(nil)
		}
	}
}
