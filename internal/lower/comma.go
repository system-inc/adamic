package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// discardExpression keeps the effects of the left operand, including void calls and writes.
func (l *lowering) discardExpression(node *ast.Node) ([]ir.Statement, error) {
	node = ast.SkipParentheses(node)
	statement := node.Kind == ast.KindCallExpression
	if node.Kind == ast.KindBinaryExpression {
		kind := node.AsBinaryExpression().OperatorToken.Kind
		_, compound := compoundAssignments[kind]
		statement = kind == ast.KindCommaToken || kind == ast.KindEqualsToken || compound || logicalAssignment(kind)
	}
	if node.Kind == ast.KindPrefixUnaryExpression || node.Kind == ast.KindPostfixUnaryExpression {
		var kind ast.Kind
		if node.Kind == ast.KindPrefixUnaryExpression {
			kind = node.AsPrefixUnaryExpression().Operator
		} else {
			kind = node.AsPostfixUnaryExpression().Operator
		}
		statement = kind == ast.KindPlusPlusToken || kind == ast.KindMinusMinusToken
	}
	if statement {
		return l.expressionStatement(node)
	}
	value, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Evaluate{Value: value}}, nil
}

func (l *lowering) commaExpression(node *ast.Node) (ir.Expression, error) {
	return l.expressionScope("comma_expression", func(b *libraryArrayBuilder) (ir.Expression, error) {
		binary := node.AsBinaryExpression()
		left, err := l.discardExpression(binary.Left)
		if err != nil {
			return nil, err
		}
		b.body = append(b.body, left...)
		if l.checker.GetTypeAtLocation(binary.Right).Flags()&checker.TypeFlagsVoid != 0 {
			right, err := l.discardExpression(binary.Right)
			b.body = append(b.body, right...)
			return ir.Undefined{}, err
		}
		return l.expression(binary.Right)
	})
}
