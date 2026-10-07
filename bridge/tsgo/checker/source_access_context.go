package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

// Compiler syntax facts only. Consumers decide which accesses matter to a rule.
func (p *Program) sourceAccessContext(out *fields, node *ast.Node, question string) (string, error) {
	if question != "source-access-context" {
		return "", fmt.Errorf("unsupported checker question: %s", question)
	}
	out.yes(ast.IsDeclarationName(node))
	out.yes(ast.IsWriteAccess(node))
	out.yes(ast.IsPartOfTypeNode(node))
	out.yes(node.Kind == ast.KindIdentifier && scanner.IsIntrinsicJsxName(node.Text()))
	return out.String(), nil
}
