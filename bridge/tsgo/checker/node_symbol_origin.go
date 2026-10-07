package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) nodeSymbolOrigin(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "node-symbol-origin" {
		return "", fmt.Errorf("unexpected node-symbol-origin suffix")
	}
	writeSymbolOrigin(out, p, c.GetSymbolAtLocation(node))
	return out.String(), nil
}
