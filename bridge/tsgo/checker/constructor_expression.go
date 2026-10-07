package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// constructorExpression returns syntax metadata, never a lint decision.
func (p *Program) constructorExpression(out *fields, node *ast.Node, question string) (string, error) {
	if question != "constructor-expression" || node.Kind != ast.KindNewExpression {
		return "", fmt.Errorf("constructor-expression requires a NewExpression")
	}
	expression := node.AsNewExpression().Expression
	out.yes(expression != nil)
	if expression != nil {
		out.text(strings.TrimPrefix(expression.Kind.String(), "Kind"))
		out.number(uint64(expression.Pos()))
		out.number(uint64(expression.End()))
	}
	return out.String(), nil
}
