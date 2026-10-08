package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// truthinessUse follows only contexts that cannot let the function escape.
// Logical operators return operands, so their result must itself be tested.
func truthinessUse(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == node
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == node
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == node
	case ast.KindForStatement:
		return parent.AsForStatement().Condition == node
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			return truthinessUse(parent)
		}
	}
	return false
}

func nullishObservation(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindNullKeyword || (ast.IsIdentifier(node) && node.Text() == "undefined")
}

func (l *lowering) methodRead(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	return symbol != nil && symbol.Flags&ast.SymbolFlagsMethod != 0
}

func (l *lowering) methodObservation(node *ast.Node) (ir.Expression, bool, error) {
	node = ast.SkipParentheses(node)
	if truthinessUse(node) {
		if value, known := l.nodeProcessMethodPresence(node); known {
			return value, true, nil
		}
	}
	if l.methodRead(node) && truthinessUse(node) {
		value, err := l.methodPresence(node)
		return value, true, err
	}
	if node.Kind != ast.KindBinaryExpression {
		return nil, false, nil
	}
	binary := node.AsBinaryExpression()
	op := binary.OperatorToken.Kind
	if (op == ast.KindAmpersandAmpersandToken || op == ast.KindBarBarToken) && truthinessUse(node) {
		// Lower both operands as conditions, preserving short-circuit evaluation.
		// Only take over logical expressions that actually contain a method observation.
		if !l.hasMethodRead(node) {
			return nil, false, nil
		}
		left, err := l.condition(binary.Left)
		if err != nil {
			return nil, true, err
		}
		right, err := l.condition(binary.Right)
		if err != nil {
			return nil, true, err
		}
		operator := ir.And
		if op == ast.KindBarBarToken {
			operator = ir.Or
		}
		return ir.Binary{Operator: operator, Left: left, Right: right}, true, nil
	}
	if op != ast.KindEqualsEqualsEqualsToken && op != ast.KindExclamationEqualsEqualsToken {
		return nil, false, nil
	}
	method, other := binary.Left, binary.Right
	if !l.methodRead(method) {
		method, other = other, method
	}
	if !l.methodRead(method) || !nullishObservation(other) {
		return nil, false, nil
	}
	// A shadowed local named undefined is not the undefined value.
	if ast.IsIdentifier(ast.SkipParentheses(other)) && !l.intrinsicUndefined(other) {
		return nil, false, nil
	}
	present, err := l.methodPresence(ast.SkipParentheses(method))
	if err != nil {
		return nil, true, err
	}
	var result ir.Expression = ir.Unary{Operator: ir.Not, Operand: present}
	if ast.SkipParentheses(other).Kind == ast.KindNullKeyword {
		// Optional methods are undefined, never null. Still evaluate the receiver.
		result = ir.IsNull{Value: present, AlwaysFalse: true}
	}
	if op == ast.KindExclamationEqualsEqualsToken {
		result = ir.Unary{Operator: ir.Not, Operand: result}
	}
	return result, true, nil
}

func (l *lowering) hasMethodRead(node *ast.Node) bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if l.methodRead(node) && truthinessUse(node) {
			found = true
			return true
		}
		node.ForEachChild(visit)
		return false
	}
	visit(node)
	return found
}

func (l *lowering) methodPresence(node *ast.Node) (ir.Expression, error) {
	if value, known := l.nodeProcessMethodPresence(node); known {
		return value, nil
	}
	access := node.AsPropertyAccessExpression()
	symbol := l.checker.GetSymbolAtLocation(node)
	if len(symbol.Declarations) == 0 || load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return nil, l.notYet(node, "presence of a host or library method without a runtime descriptor")
	}
	if l.accessorNames[node.Name().Text()] {
		return nil, l.notYet(node, "method presence through a possible accessor")
	}
	if receiver := ast.SkipParentheses(access.Expression); ast.IsIdentifier(receiver) {
		if declared := l.checker.GetSymbolAtLocation(receiver); declared != nil {
			for _, declaration := range declared.Declarations {
				for at := declaration; at != nil && at.Kind != ast.KindSourceFile; at = at.Parent {
					if ast.HasSyntacticModifier(at, ast.ModifierFlagsAmbient) {
						return nil, l.notYet(node, "presence of an ambient host method without a runtime descriptor")
					}
				}
			}
		}
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	if object.Type() != ir.Object {
		return nil, l.notYet(node, "method presence on a non-object receiver")
	}
	return ir.MethodPresence{Object: object, Name: l.fieldName(node.Name()), Optional: access.QuestionDotToken != nil}, nil
}

// methodComparisonUse admits only the actual undefined value or null.
func (l *lowering) methodComparisonUse(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	if node.Parent == nil || node.Parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := node.Parent.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken && binary.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {
		return false
	}
	other := binary.Left
	if other == node {
		other = binary.Right
	}
	if !nullishObservation(other) {
		return false
	}
	return ast.SkipParentheses(other).Kind == ast.KindNullKeyword || l.intrinsicUndefined(other)
}
