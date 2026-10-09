// Package cst is eemeli/yaml 2.9.0's concrete syntax tree parser, dist/parse/, ported to Go: the lexer
// (lexer.js), the parser (parser.js), the token types and tokenType (cst.js), and the LineCounter
// (line-counter.js). The composer (internal/format/yaml/compose) builds documents from these tokens, and
// yaml-unist-parser reads them again through each node's srcToken, so the tokens have to be upstream's
// tokens, offset for offset.
//
// # Text is UTF-16
//
// A JavaScript string is UTF-16 code units, and every offset, length and index upstream counts them. So
// does this package: the source is a []uint16 (utf16.Encode([]rune(text))), a token's Source is a slice
// of it, and Offset and Indent count units. A character upstream reads as `this.buffer[i]`, possibly
// undefined, is an int here: the code unit, or -1 past either end of the string (undefined). A NUL is a
// real character in JavaScript (the string "\0" is truthy), so -1, not 0, stands for undefined.
//
// # Generators are iterators
//
// Lexer.lex and Parser.parse are generators upstream. Here Lexer.Lex and Parser.Parse return iter.Seq,
// so a consumer ranges over them as upstream's consumer runs for...of. Inside, upstream's `yield*` into a
// helper generator becomes a plain call: the helper pushes through the yield function the iterator was
// started with, and returns what the generator returned. The interleaving is upstream's exactly: the
// lexer yields a lexeme, the parser steps on it and yields whole tokens to the consumer, the consumer
// returns, the lexer resumes. When a consumer breaks out early, the iterator stops yielding and stops at
// the next state boundary; upstream would suspend instead, but nothing resumes a broken iteration.
//
// # The token shape
//
// Upstream's Token is a union of open objects told apart by their type field (cst.d.ts), and the parser
// mutates them in place as it builds: it appends to their lists, moves lists between them, and deletes
// fields. Here a Token is one struct carrying every field any member of the union has, and Type is
// upstream's string, so the composer switches on token.Type with upstream's case labels:
//
//	type                                   fields upstream gives it
//	SourceToken: "space", "comment", ...   Offset, Indent, Source
//	"error"                                Offset, Source, Message (no indent)
//	"directive"                            Offset, Source (no indent)
//	"document"                             Offset, Start, Value?, End?
//	"doc-end"                              Offset, Source, End?
//	"alias", "scalar",
//	"single-quoted-scalar",
//	"double-quoted-scalar"                 Offset, Indent, Source, End?
//	"block-scalar"                         Offset, Indent, Props, Source
//	"block-map", "block-seq"               Offset, Indent, Items
//	"flow-collection"                      Offset, Indent, FlowStart, Items, End
//
// A document's start is a list of tokens and a flow collection's start is one token, so the two are
// Start and FlowStart here. Optional fields follow JavaScript's distinctions:
//
//   - A list that may be absent (End, CollectionItem.Sep) is nil when absent and non-nil, perhaps empty,
//     when present. An empty JavaScript array is truthy, so `if (it.sep)` is `it.Sep != nil`, never
//     `len(it.Sep) > 0`. Code that builds such a list must build it non-nil.
//   - A token that may be absent (Value, CollectionItem.Key, CollectionItem.Value) is nil when absent.
//     An item's key can also be explicitly null; Key is nil then too, and HasKey tells the two apart,
//     which only the JSON shape needs.
//   - `'indent' in token` is HasIndent and `'source' in token` is HasSource.
//
// # What is not ported
//
// cst-visit.js, cst-scalar.js and cst-stringify.js (visit, createScalarToken, resolveAsScalar,
// setScalarValue, stringify): the composer never calls them, and yaml-unist-parser has its own small
// tokens() helper over token lists, which is layer 3's. cst.js's isCollection, isScalar and prettyToken:
// nothing on the path calls them (prettyToken only logs under LOG_TOKENS, which is not ported either).
package cst

// The control characters the lexer emits besides slices of the input, upstream's BOM, DOCUMENT, FLOW_END
// and SCALAR.
const (
	// BOM is the byte order mark.
	BOM uint16 = 0xFEFF
	// Document is the start of doc-mode (C0: Start of Text).
	Document uint16 = 0x02
	// FlowEnd is an unexpected end of flow-mode (C0: Cancel).
	FlowEnd uint16 = 0x18
	// Scalar says the next token is a scalar value (C0: Unit Separator).
	Scalar uint16 = 0x1F
)

// Token is any CST token, upstream's Token union. See the package comment for which fields each type has.
type Token struct {
	Type   string
	Offset int
	Indent int
	Source []uint16

	// Message is an error token's message.
	Message string

	// Start is a document's start: the tokens before its value.
	Start []*Token
	// FlowStart is a flow collection's start, its { or [ token.
	FlowStart *Token
	// Value is a document's value, nil when absent.
	Value *Token
	// End is the tokens after a document, document end marker, flow scalar or flow collection; nil when
	// absent.
	End []*Token
	// Props is a block scalar's header, and the comments and spaces on its line.
	Props []*Token
	// Items is a block map's, block sequence's or flow collection's items.
	Items []*CollectionItem

	// indentAbsent marks a token whose upstream object has no indent field: a document, directive,
	// document end or error token the parser builds. SourceTokens always have one, whatever their type.
	indentAbsent bool
}

// HasIndent is upstream's `'indent' in token`.
func (token *Token) HasIndent() bool { return !token.indentAbsent }

// HasSource is upstream's `'source' in token`: every token but a document and a collection has one.
func (token *Token) HasSource() bool {
	switch token.Type {
	case "document", "block-map", "block-seq", "flow-collection":
		return false
	}
	return true
}

// CollectionItem is an item of a block map, block sequence or flow collection, upstream's CollectionItem
// and the item types of BlockMap and BlockSequence.
type CollectionItem struct {
	Start []*Token
	// Key is nil when absent or null; HasKey tells which.
	Key *Token
	// Sep is nil when absent. Present and empty is truthy upstream.
	Sep   []*Token
	Value *Token
	// ExplicitKey is `explicitKey: true`; upstream never sets it false.
	ExplicitKey bool

	// keyPresent is `'key' in item`: true for a key that is a token or null.
	keyPresent bool
}

// HasKey is `'key' in item`: whether the item has a key field, perhaps null.
func (item *CollectionItem) HasKey() bool { return item.keyPresent }

// setKey is `item.key = key` (or Object.assign with a key), where a nil key is null.
func (item *CollectionItem) setKey(key *Token) {
	item.Key = key
	item.keyPresent = true
}

// deleteKey is `delete item.key`.
func (item *CollectionItem) deleteKey() {
	item.Key = nil
	item.keyPresent = false
}

// TokenType identifies the type of a lexer token, upstream's tokenType. It returns "" for upstream's
// null, an unknown token.
func TokenType(source []uint16) string {
	switch len(source) {
	case 0:
		// case '':
		return "newline"
	case 1:
		switch source[0] {
		case BOM:
			return "byte-order-mark"
		case Document:
			return "doc-mode"
		case FlowEnd:
			return "flow-error-end"
		case Scalar:
			return "scalar"
		case '\n':
			return "newline"
		case '-':
			return "seq-item-ind"
		case '?':
			return "explicit-key-ind"
		case ':':
			return "map-value-ind"
		case '{':
			return "flow-map-start"
		case '}':
			return "flow-map-end"
		case '[':
			return "flow-seq-start"
		case ']':
			return "flow-seq-end"
		case ',':
			return "comma"
		}
	case 2:
		if source[0] == '\r' && source[1] == '\n' {
			return "newline"
		}
	case 3:
		if source[0] == '-' && source[1] == '-' && source[2] == '-' {
			return "doc-start"
		}
		if source[0] == '.' && source[1] == '.' && source[2] == '.' {
			return "doc-end"
		}
	}
	switch source[0] {
	case ' ', '\t':
		return "space"
	case '#':
		return "comment"
	case '%':
		return "directive-line"
	case '*':
		return "alias"
	case '&':
		return "anchor"
	case '!':
		return "tag"
	case '\'':
		return "single-quoted-scalar"
	case '"':
		return "double-quoted-scalar"
	case '|', '>':
		return "block-scalar-header"
	}
	return ""
}
