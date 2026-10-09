// Package core ports ESLint's own core rules, the ones that guard JavaScript rather than any
// convention of ours.
//
// These rules carry no house style. They catch code that reads as a mistake in any codebase: a
// switch arm that can never run, a binding whose scope is wider than its initialization, a catch
// that does nothing but hand the error back. Porting them here rather than leaving them to a
// JavaScript linter is the whole point of the tool, since a rule living inside the checker costs
// a walk that already happened.
package core

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// tokenSignature renders an expression as a string that is equal for two expressions exactly when
// they are written the same way, ignoring comments and whitespace.
//
// This is our answer to ESLint's `equalTokens`, which compares two nodes token by token against the
// source text. We cannot borrow that shape directly, because it needs a token stream and
// typescript-go hands us a tree whose ForEachChild deliberately skips punctuation: `a.b` yields two
// identifiers and no dot, and a PrefixUnaryExpression yields its operand while its operator lives in
// a struct field the walk never visits. A signature built from child nodes alone therefore reads
// `-1` and `+1` as the same expression, which is a false finding on code that is correct.
//
// So the signature is built from both halves of what the parser knows. Each node contributes its
// kind, which is ESLint's node-type check and also recovers every operator the walk dropped, since
// the operator is what distinguishes one kind from another wherever a kind carries only one. Then
// the gaps between children are scanned for the punctuation that sits in them, which recovers the
// rest: the dot in `a.b`, the `?.` that makes it a different expression, the parentheses and commas
// of a call.
//
// Scanning the gaps rather than the whole node is what keeps regular expressions and strings intact.
// A scanner run over `case /[/*]/` from the start would have to decide whether a slash opens a regex
// or divides, and it decides wrong without the parser's context; a string containing `/* b */` would
// lose its middle to comment-stripping. A leaf keeps its own source text verbatim instead, so the
// literal is compared as the parser saw it, and only the punctuation between nodes, which cannot be
// a literal, is ever re-scanned.
func tokenSignature(sourceFile *ast.SourceFile, node *ast.Node) string {
	var signature strings.Builder
	appendNodeSignature(sourceFile, node, &signature)
	return signature.String()
}

func appendNodeSignature(sourceFile *ast.SourceFile, node *ast.Node, signature *strings.Builder) {
	signature.WriteString(strconv.Itoa(int(node.Kind)))
	signature.WriteByte('(')

	nodeRange := rule.TokenRange(sourceFile, node)
	cursor := nodeRange.Pos()
	hasChildren := false

	node.ForEachChild(func(child *ast.Node) bool {
		hasChildren = true
		childRange := rule.TokenRange(sourceFile, child)
		appendTokensBetween(sourceFile, cursor, childRange.Pos(), signature)
		appendNodeSignature(sourceFile, child, signature)
		signature.WriteByte(',')
		cursor = childRange.End()
		return false
	})

	if hasChildren {
		appendTokensBetween(sourceFile, cursor, nodeRange.End(), signature)
	} else {
		// A leaf is its own text: a literal must be compared exactly as written, since two strings
		// differing only in whitespace are two different values.
		signature.WriteString(sourceFile.Text()[nodeRange.Pos():nodeRange.End()])
	}

	signature.WriteByte(')')
}

// appendTokensBetween writes the tokens in a half-open source range, dropping comments and
// whitespace because the scanner never emits them.
//
// The scanner returned by GetScannerForSourceFile has already scanned the token at the start
// position, so its current token must be read before advancing. Calling Scan first, which is the
// natural-looking loop, silently drops the first token of every range and was the defect that made
// an earlier draft of this signature blind to unary operators.
func appendTokensBetween(sourceFile *ast.SourceFile, start int, end int, signature *strings.Builder) {
	if start >= end {
		return
	}
	tokenScanner := scanner.GetScannerForSourceFile(sourceFile, start)
	for tokenScanner.Token() != ast.KindEndOfFile && tokenScanner.TokenStart() < end {
		signature.WriteString(tokenScanner.TokenText())
		signature.WriteByte(' ')
		tokenScanner.Scan()
	}
}
