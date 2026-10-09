// Package parser re-exports the typescript-go parser entry points cohere needs.
//
// Hand-written rather than generated: internal/types/tools/generate_shims covers the eleven packages tsgolint needed,
// and the parser was not among them because a headless linter is handed a program rather than
// parsing files itself. Rule fixtures do parse single files, so this exists for them.
package parser

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
)

// ParseSourceFile parses one file into an AST, with no program and no type information.
func ParseSourceFile(opts ast.SourceFileParseOptions, sourceText string, scriptKind core.ScriptKind) *ast.SourceFile {
	return parser.ParseSourceFile(opts, sourceText, scriptKind)
}
