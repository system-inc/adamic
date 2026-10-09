package nextjs

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// HasFileDirective reports whether a source file opens with a given directive, such as
// `use client` or `use server`.
//
// The comparison is against the directive's parsed value rather than its source text, so
// `'use client'` and `"use client"` both answer true. That is what oxc compares
// (`directive.directive.as_str()`), and it matters on our own trees: every one of the 874
// client files in `~/Projects/ahra` writes single quotes with a trailing comment, so a
// predicate matching raw source text would answer false on all of them while every imported
// fixture, which writes double quotes, passed.
//
// # Why this walks rather than asking the shim predicate
//
// `ast.IsPrologueDirective` looks like the whole answer and is not. It is purely local: it
// asks whether a statement is an expression statement whose expression is a string literal,
// with no knowledge of where the statement sits. Measured, it answers true for the middle
// statement of:
//
//	const z = 1
//	"use client"
//	export default async function MyComponent() {}
//
// and oxc is silent on that file, because a directive prologue ends at the first statement
// that is not one. So the position is supplied by this walk and the predicate only classifies.
// A rule reaching for the predicate alone reports a file upstream exempts, and no imported
// fixture writes that shape.
//
// A prologue may hold several directives, and only the run of them from the top is consulted:
//
//	"use strict"
//	'use client'          answers true, pinned against the release oxlint binary
//
// A template literal is not a directive per the specification and the parser agrees, so a
// backtick-quoted `use client` answers false without needing a check here.
func HasFileDirective(sourceFile *ast.SourceFile, directive string) bool {
	if sourceFile == nil || sourceFile.Statements == nil {
		return false
	}
	for _, statement := range sourceFile.Statements.Nodes {
		// The prologue ends at the first statement that is not a directive. Everything after it
		// is ordinary code, and a string sitting there is an expression rather than a directive.
		if !ast.IsPrologueDirective(statement) {
			return false
		}
		expression := statement.AsExpressionStatement().Expression
		if expression.Kind == ast.KindStringLiteral && expression.Text() == directive {
			return true
		}
	}
	return false
}
