package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Binding identities and type-node membership preserve the compiler's AST facts.
func (p *Program) nodeBinding(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "node-binding" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("node-binding requires an Identifier")
	}
	out.number(p.symbolID(c.GetSymbolAtLocation(node)))
	out.number(p.symbolID(c.GetExportSpecifierLocalTargetSymbol(node)))
	out.yes(ast.IsPartOfTypeNode(node))
	return out.String(), nil
}
