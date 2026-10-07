package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func (p *Program) classHeritageNames(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "class-heritage-names" || node.Kind != ast.KindIdentifier {
		return "", fmt.Errorf("class-heritage-names requires an identifier")
	}
	var names []string
	for depth := 0; depth < 16; depth++ {
		symbol := c.GetSymbolAtLocation(node)
		if symbol == nil {
			break
		}
		symbol = checker.SkipAlias(symbol, c)
		var class *ast.Node
		for _, d := range symbol.Declarations {
			if d.Kind == ast.KindClassDeclaration || d.Kind == ast.KindClassExpression {
				class = d
				break
			}
		}
		if class == nil {
			break
		}
		base := ast.GetClassExtendsHeritageElement(class)
		if base == nil {
			break
		}
		node = base.AsExpressionWithTypeArguments().Expression
		for node != nil && (node.Kind == ast.KindParenthesizedExpression || node.Kind == ast.KindNonNullExpression || node.Kind == ast.KindAsExpression || node.Kind == ast.KindSatisfiesExpression) {
			switch node.Kind {
			case ast.KindParenthesizedExpression:
				node = node.AsParenthesizedExpression().Expression
			case ast.KindNonNullExpression:
				node = node.AsNonNullExpression().Expression
			case ast.KindAsExpression:
				node = node.AsAsExpression().Expression
			case ast.KindSatisfiesExpression:
				node = node.AsSatisfiesExpression().Expression
			}
		}
		if node == nil || node.Kind != ast.KindIdentifier {
			break
		}
		names = append(names, node.Text())
	}
	out.number(uint64(len(names)))
	for _, name := range names {
		out.text(name)
	}
	return out.String(), nil
}
