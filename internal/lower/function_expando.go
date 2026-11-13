package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// functionExpandoRefusal judges the checker's assignment-declared properties, not every
// field of a callable. A namespace's declared members and an object's fields keep their
// existing paths. Directives and casts through any keep their own refusals (#4g1avf7).
func (l *lowering) functionExpandoRefusal(node *ast.Node) error {
	if node.Kind != ast.KindBinaryExpression || !ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
		return nil
	}
	target := ast.SkipParentheses(node.AsBinaryExpression().Left)
	if target.Kind != ast.KindPropertyAccessExpression && target.Kind != ast.KindElementAccessExpression {
		return nil
	}
	receiver := ast.SkipParentheses(target.Expression())
	function := l.symbol(receiver)
	if !expandoFunction(function) {
		return nil
	}
	member := l.checker.GetSymbolAtLocation(target)
	if member == nil && target.Kind == ast.KindElementAccessExpression {
		// The checker puts no symbol on the whole bracket access. Resolve its literal key
		// against the receiver, which carries the same assignment-declared member as dot access.
		key := l.checker.GetTypeAtLocation(target.AsElementAccessExpression().ArgumentExpression)
		if key.Flags()&checker.TypeFlagsStringLiteral != 0 {
			member = l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(receiver), key.AsLiteralType().Value().(string))
		}
	}
	if member == nil || member.Flags&ast.SymbolFlagsAssignment == 0 {
		return nil
	}
	name := function.Name
	if ast.IsIdentifier(receiver) {
		name = receiver.Text()
	}
	return &Refused{
		Where: l.program.Where(node),
		What:  "an expando: adding " + member.Name + " to function " + name + " after it's made",
		Fix:   "declare " + member.Name + " on an object with a call method, or put " + name + " and " + member.Name + " in a namespace (adamic/no-expando)",
	}
}

// expandoFunction matches the function declarations and const initializers on which tsc
// declares expando properties. A callable interface alone does not prove this origin.
func expandoFunction(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration {
			return true
		}
		if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent != nil && declaration.Parent.Flags&ast.NodeFlagsConst != 0 && declaration.Initializer() != nil {
			initializer := ast.SkipParentheses(declaration.Initializer())
			if initializer.Kind == ast.KindFunctionExpression || initializer.Kind == ast.KindArrowFunction {
				return true
			}
		}
	}
	return false
}

func init() {
	// Before cast and relation rules, and before lowering's field-store NotYet: an expando
	// is permanently refused, even when its value would otherwise fit the new property.
	refusalRules.register(1450, "functionExpandoRefusal", (*lowering).functionExpandoRefusal)
}
