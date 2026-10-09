package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Only constructor origins and immutable aliases prove these host shapes. A structural
// Date/RegExp/Map interface alone says nothing about extra own properties.
func (l *lowering) objectIntegrityShape(node *ast.Node, depth int) int {
	if depth > 16 {
		return 0
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		expression := ast.SkipParentheses(call.Expression)
		if expression.Kind == ast.KindPropertyAccessExpression && call.Arguments != nil && len(call.Arguments.Nodes) == 1 {
			property := expression.AsPropertyAccessExpression()
			name := property.Name().Text()
			if l.isLibraryGlobal(property.Expression, "Object") && (name == "freeze" || name == "seal" || name == "preventExtensions") {
				return l.objectIntegrityShape(call.Arguments.Nodes[0], depth+1)
			}
		}
	}
	if node.Kind == ast.KindNewExpression {
		constructor := node.AsNewExpression().Expression
		for _, name := range []string{"Date", "Map", "Set"} {
			if l.isLibraryGlobal(constructor, name) {
				return 1
			}
		}
		created := node.AsNewExpression()
		if l.isLibraryGlobal(constructor, "Object") && (created.Arguments == nil || len(created.Arguments.Nodes) == 0) {
			return 1
		}
		if l.isLibraryGlobal(constructor, "RegExp") {
			return 2
		}
	}
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol != nil && len(symbol.Declarations) == 1 {
			declaration := symbol.Declarations[0]
			if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
				variable := declaration.AsVariableDeclaration()
				if variable.Initializer != nil {
					return l.objectIntegrityShape(variable.Initializer, depth+1)
				}
			}
		}
	}
	return 0
}

func (l *lowering) objectIntegrityValue(argument, call *ast.Node) (ir.Expression, error) {
	argument = ast.SkipParentheses(argument)
	if argument.Kind == ast.KindNewExpression {
		created := argument.AsNewExpression()
		empty := created.Arguments == nil || len(created.Arguments.Nodes) == 0
		if empty && l.isLibraryGlobal(created.Expression, "Object") {
			return ir.ObjectLiteral{}, nil
		}
		// Vacuous key/value representation is valid only if this fresh empty collection is
		// discarded. Never let an any/unknown collection escape into a binding or operation.
		outer := call
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		if empty && outer.Parent != nil && outer.Parent.Kind == ast.KindExpressionStatement {
			if l.isLibraryGlobal(created.Expression, "Map") {
				return ir.MapNew{Key: ir.Number, Value: ir.Number}, nil
			}
			if l.isLibraryGlobal(created.Expression, "Set") {
				return ir.SetNew{Element: ir.Number}, nil
			}
		}
		if empty && (l.isLibraryGlobal(created.Expression, "Map") || l.isLibraryGlobal(created.Expression, "Set")) {
			for _, element := range l.typeArguments(l.checker.GetTypeAtLocation(argument)) {
				if element.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
					return nil, l.notYet(argument, "an untyped collection escaping Object integrity; provide explicit element types")
				}
			}
		}
	}
	return l.expression(argument)
}
