package graphql

// graphql-js 17.0.2, language/tokenKind.js, and the Token class from language/ast.js.

// tokenKind is the kind of a lexed token; the values are graphql-js's, which its error messages print.
type tokenKind string

const (
	tokenKindSOF         tokenKind = "<SOF>"
	tokenKindEOF         tokenKind = "<EOF>"
	tokenKindBang        tokenKind = "!"
	tokenKindDollar      tokenKind = "$"
	tokenKindAmp         tokenKind = "&"
	tokenKindParenL      tokenKind = "("
	tokenKindParenR      tokenKind = ")"
	tokenKindDot         tokenKind = "."
	tokenKindSpread      tokenKind = "..."
	tokenKindColon       tokenKind = ":"
	tokenKindEquals      tokenKind = "="
	tokenKindAt          tokenKind = "@"
	tokenKindBracketL    tokenKind = "["
	tokenKindBracketR    tokenKind = "]"
	tokenKindBraceL      tokenKind = "{"
	tokenKindPipe        tokenKind = "|"
	tokenKindBraceR      tokenKind = "}"
	tokenKindName        tokenKind = "Name"
	tokenKindInt         tokenKind = "Int"
	tokenKindFloat       tokenKind = "Float"
	tokenKindString      tokenKind = "String"
	tokenKindBlockString tokenKind = "BlockString"
	tokenKindComment     tokenKind = "Comment"
)

// token represents a range of characters represented by a lexical token within a Source.
//
// Start and end are byte offsets into the UTF-8 text, where graphql-js's are UTF-16 indexes, and so is
// column: 1 + start - lineStart, counted in bytes. Nothing reads a token's column but an error, and an
// error computes its own (see syntaxError).
type token struct {
	// The kind of token.
	kind tokenKind

	// The character offset at which this Node begins.
	start int

	// The character offset at which this Node ends.
	end int

	// The 1-indexed line number on which this Token appears.
	line int

	// The 1-indexed column number at which this Token begins.
	column int

	// For non-punctuation tokens, represents the interpreted value of the token. hasValue is whether
	// graphql-js's value is defined at all: an empty string is a value, a punctuator has none.
	value    string
	hasValue bool

	// Tokens exist as nodes in a double-linked-list amongst all tokens including ignored tokens.
	// <SOF> is always the first node and <EOF> the last.
	prev *token
	next *token
}

func newToken(kind tokenKind, start int, end int, line int, column int, value ...string) *token {
	created := &token{kind: kind, start: start, end: end, line: line, column: column}
	if len(value) > 0 {
		created.value, created.hasValue = value[0], true
	}
	return created
}
