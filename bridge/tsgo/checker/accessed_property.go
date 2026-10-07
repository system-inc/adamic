package checker

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// accessedProperty exposes the checker-resolved key, including enum and literal keys.
func (p *Program) accessedProperty(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "accessed-property" || !ast.IsAccessExpression(node) {
		return "", fmt.Errorf("accessed-property requires an access expression")
	}
	name, _ := checker.Checker_getAccessedPropertyName(c, node)
	out.text(name)
	return out.String(), nil
}
