package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) objectIntrinsicBuiltin(node *ast.Node) (ir.Expression, bool, error) {
	if value, handled, err := l.libraryObjectStaticCall(node); handled {
		return value, handled, err
	}
	if value, handled, err := l.libraryObjectCoercionCall(node); handled {
		return value, handled, err
	}
	if value, handled, err := l.libraryObjectIntrinsicCall(node); handled {
		return value, handled, err
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind == ast.KindPropertyAccessExpression && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
		return l.objectNamesCall(node, callee.Name().Text())
	}
	return nil, false, nil
}

// These observations retain the library declaration's identity and cannot let a
// callable escape. Still visit every child through the ordinary refusal pass.
func (l *lowering) visitObjectIntrinsicObservation(node *ast.Node, visit ast.Visitor) bool {
	if !l.libraryObjectBoundMethod(node) && !(l.objectIntrinsic(node) != "" && objectTruthinessUse(node)) {
		return false
	}
	node.ForEachChild(visit)
	return true
}

func (l *lowering) objectNumericFieldName(name *ast.Node) (string, bool) {
	if name.Kind == ast.KindNumericLiteral {
		if symbol := l.checker.GetSymbolAtLocation(name); symbol != nil {
			return symbol.Name, true
		}
	}
	return "", false
}

func objectTruthinessUse(node *ast.Node) bool {
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
		switch parent.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken:
			return objectTruthinessUse(parent)
		}
	}
	return false
}
