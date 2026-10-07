// A port of cohere's internal/format/graphql/token.go to Adamic 0.1: graphql-js 17.0.2,
// language/tokenKind.js, and the Token class from language/ast.js.
//
// Offsets are UTF-16 indexes, as graphql-js's are; the Go's are bytes, and the test's Go side converts
// them.
//
// Not here: Token's prev and next. graphql-js links every token, ignored ones included, into a doubly
// linked list, and the Go keeps the list. A token holding the token after it and the one before is a
// cycle, which 0.1's cycle rule refuses, and the parser never walks the list backwards, so the lexer
// keeps the tokens in an array instead, in the list's order (lexer.ts).

// TokenKind is the kind of a lexed token; the values are graphql-js's, which its error messages print.
export type TokenKind =
    | '<SOF>'
    | '<EOF>'
    | '!'
    | '$'
    | '&'
    | '('
    | ')'
    | '.'
    | '...'
    | ':'
    | '='
    | '@'
    | '['
    | ']'
    | '{'
    | '|'
    | '}'
    | 'Name'
    | 'Int'
    | 'Float'
    | 'String'
    | 'BlockString'
    | 'Comment';

// Token represents a range of characters represented by a lexical token within a Source.
//
// value is in the form the test's output writes every string in (escaped, in lexer.ts): a String
// token's value is cooked as graphql-js cooks it, and then written that way, since 0.1's
// String.fromCodePoint, which cooking a \u escape needs, doesn't lower yet (gitignore's GAPS.md, gap 1;
// GAPS.md here, "Cooking a string without String.fromCodePoint"). A Name, Int or Float's value is ASCII
// letters, digits, `_`, `.`, `+` and `-`, which that form writes as themselves.
export class Token {
    // The kind of token.
    readonly kind: TokenKind;
    // The character offset at which this Node begins.
    readonly start: number;
    // The character offset at which this Node ends.
    readonly end: number;
    // The 1-indexed line number on which this Token appears.
    readonly line: number;
    // The 1-indexed column number at which this Token begins.
    readonly column: number;
    // For non-punctuation tokens, represents the interpreted value of the token.
    readonly value: string | undefined;

    constructor(kind: TokenKind, start: number, end: number, line: number, column: number, value: string | undefined) {
        this.kind = kind;
        this.start = start;
        this.end = end;
        this.line = line;
        this.column = column;
        this.value = value;
    }
}
