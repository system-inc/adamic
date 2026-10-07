package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// File mode is an AST fact used to choose the native parser's await context.
func (p *Program) streamFile(out *fields, node *ast.Node, source *ast.SourceFile) (string, error) {
	if node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("stream-file requires a SourceFile")
	}
	out.yes(ast.IsExternalModule(source))
	out.yes(source.IsDeclarationFile)
	return out.String(), nil
}
