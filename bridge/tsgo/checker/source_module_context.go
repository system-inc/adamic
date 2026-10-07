package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Parsing context only. No lint decision is made on the Go side.
func (p *Program) sourceModuleContext(out *fields, node *ast.Node, question string) (string, error) {
	if question != "source-module-context" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("source-module-context requires a SourceFile")
	}
	out.yes(ast.IsExternalModule(node.AsSourceFile()))
	return out.String(), nil
}
