package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

func (p *Program) isolatedDeclarations(out *fields, node *ast.Node, question string) (string, error) {
	if question != "isolated-declarations" || node.Kind != ast.KindSourceFile {
		return "", fmt.Errorf("isolated-declarations requires a SourceFile")
	}
	out.yes(p.Compiler.Options().IsolatedDeclarations.IsTrue())
	return out.String(), nil
}
