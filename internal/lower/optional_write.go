package lower

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// absentLiteralField distinguishes a missing own slot from a present optional
// reference slot. Native objects currently keep the shape of their initializer.
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
	return initializer.Kind == ast.KindObjectLiteralExpression && l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(initializer), target.Name().Text()) == nil
}

// presentLiteralField proves an optional slot exists in this variable's initializer.
func (l *lowering) presentLiteralField(target *ast.Node) bool {
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
	return initializer.Kind == ast.KindObjectLiteralExpression && l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(initializer), target.Name().Text()) != nil
}
