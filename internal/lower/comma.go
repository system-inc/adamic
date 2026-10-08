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
		return l.commaValue(binary.Right)
	})
}

// commaValue also admits writes in the value position. The saved reference precedes the RHS,
// and the returned value is the value written, even if a setter changes its stored value.
func (l *lowering) commaValue(node *ast.Node) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		operator, compound := compoundAssignments[binary.OperatorToken.Kind]
		if binary.OperatorToken.Kind == ast.KindEqualsToken || compound {
			return l.expressionScope("comma_assignment_value", func(b *libraryArrayBuilder) (ir.Expression, error) {
				current, _, store, err := l.assignmentReference(b, binary.Left)
				if err != nil {
					return nil, err
				}
				if compound {
					current = b.read(b.declare("comma_current", current))
				}
				right, err := l.commaValue(binary.Right)
				if err != nil {
					return nil, err
				}
				if compound {
					if operator == ast.KindPlusToken {
						current, right = l.spelled(binary.Left, current), l.spelled(binary.Right, right)
					}
					right, err = l.combine(node, operator, current, right)
					if err != nil {
						return nil, err
					}
				}
				saved := b.declare("comma_value", right)
				b.body = append(b.body, store(b.read(saved)))
				return b.read(saved), nil
			})
		}
	}
	if node.Kind == ast.KindPrefixUnaryExpression || node.Kind == ast.KindPostfixUnaryExpression {
		var operator ast.Kind
		var operand *ast.Node
		if node.Kind == ast.KindPrefixUnaryExpression {
			operator, operand = node.AsPrefixUnaryExpression().Operator, node.AsPrefixUnaryExpression().Operand
		} else {
			operator, operand = node.AsPostfixUnaryExpression().Operator, node.AsPostfixUnaryExpression().Operand
		}
		if operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
			return l.expressionScope("comma_increment_value", func(b *libraryArrayBuilder) (ir.Expression, error) {
				current, _, store, err := l.assignmentReference(b, operand)
				if err != nil {
					return nil, err
				}
				if current.Type() != ir.Number {
					return nil, l.notYet(node, "a comma increment of a non-number")
				}
				old := b.declare("comma_old", current)
				step := ir.Add
				if operator == ast.KindMinusMinusToken {
					step = ir.Subtract
				}
				next := b.declare("comma_new", ir.Binary{Operator: step, Left: b.read(old), Right: ir.NumberConstant{Value: 1}})
				b.body = append(b.body, store(b.read(next)))
				if node.Kind == ast.KindPostfixUnaryExpression {
					return b.read(old), nil
				}
				return b.read(next), nil
			})
		}
	}
	return l.expression(node)
}
