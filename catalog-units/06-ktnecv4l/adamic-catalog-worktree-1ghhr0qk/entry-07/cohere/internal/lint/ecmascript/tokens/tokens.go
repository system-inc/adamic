// Package tokens is a node's tokens, comments and whitespace excluded, and the questions ESLint's
// SourceCode asks of them: getTokenBefore, getTokenAfter and getFirstTokenBetween.
//
// A fixer ported from ESLint reads its edit off tokens rather than nodes, because the comma between two
// specifiers and the braces around a list are tokens that no node owns. Two rules needed the same
// small scan (consistent-type-imports and no-unused-vars, both removing import specifiers), so it lives
// here once.
package tokens

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// Token is one token: its kind and its span, trivia excluded.
type Token struct {
	Kind  ast.Kind
	Start int
	End   int
}

// List is a node's tokens in source order.
type List []Token

// Of scans a node's tokens, from the start of its own text (leading trivia excluded, as an ESTree range
// is) to its end.
//
// For a statement-level node such as an import declaration, which holds no template or regular
// expression whose scanning depends on context. A node that can hold one would need the parser's rescans.
func Of(sourceFile *ast.SourceFile, node *ast.Node) List {
	nodeRange := rule.TokenRange(sourceFile, node)
	var list List
	tokenScanner := scanner.GetScannerForSourceFile(sourceFile, nodeRange.Pos())
	for tokenScanner.Token() != ast.KindEndOfFile && tokenScanner.TokenStart() < nodeRange.End() {
		list = append(list, Token{Kind: tokenScanner.Token(), Start: tokenScanner.TokenStart(), End: tokenScanner.TokenEnd()})
		tokenScanner.Scan()
	}
	return list
}

// Before is ESLint's getTokenBefore: the last token ending at or before position.
func (list List) Before(position int) (Token, bool) {
	for index := len(list) - 1; index >= 0; index-- {
		if list[index].End <= position {
			return list[index], true
		}
	}
	return Token{}, false
}

// After is ESLint's getTokenAfter: the first token starting at or after position.
func (list List) After(position int) (Token, bool) {
	for _, token := range list {
		if token.Start >= position {
			return token, true
		}
	}
	return Token{}, false
}

// FirstBetween is ESLint's getFirstTokenBetween for one kind: the first token of that kind lying wholly
// within [from, to).
func (list List) FirstBetween(from int, to int, kind ast.Kind) (Token, bool) {
	for _, token := range list {
		if token.Start >= from && token.End <= to && token.Kind == kind {
			return token, true
		}
	}
	return Token{}, false
}
