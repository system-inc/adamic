package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// absentLiteralField distinguishes a missing own slot from a present optional
// reference slot that was not reserved by its contextual literal type.
func (l *lowering) absentLiteralField(target *ast.Node) bool {
	receiver := ast.SkipParentheses(target.AsPropertyAccessExpression().Expression)
	if !ast.IsIdentifier(receiver) {
		return false
	}
	symbol := l.symbol(receiver)
	if symbol == nil || symbol.ValueDeclaration == nil || symbol.ValueDeclaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	initializer := symbol.ValueDeclaration.Initializer()
	if initializer == nil {
		return false
	}
	initializer = ast.SkipParentheses(initializer)
	if initializer.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	reserved, _ := l.optionalLiteralSlots(initializer, nil)
	for _, field := range reserved {
		if field.Name == target.Name().Text() {
			return false
		}
	}
	return l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(initializer), target.Name().Text()) == nil
}
