package checker

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// symbolLocations exposes every declaration without requiring its name to
// have textual syntax. Destructuring declarations have binding patterns.
func (p *Program) symbolLocations(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "symbol-locations" {
		return "", fmt.Errorf("unexpected symbol-locations suffix")
	}
	symbol := c.GetSymbolAtLocation(node)
	var declarations []*ast.Node
	if symbol != nil {
		declarations = symbol.Declarations
	}
	out.number(uint64(len(declarations)))
	for _, declaration := range declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil {
			return "", fmt.Errorf("symbol declaration has no source")
		}
		out.text(file.FileName())
		out.text(strings.TrimPrefix(declaration.Kind.String(), "Kind"))
		out.number(uint64(declaration.Pos()))
		out.number(uint64(declaration.End()))
	}
	return out.String(), nil
}
