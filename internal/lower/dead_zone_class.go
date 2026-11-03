package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// A class binding is in its temporal dead zone until its declaration runs, as a let's is, and Node
// throws "Cannot access 'Late' before initialization" on a new Late() that runs before then (#aw65psm).
// Stage 0 gives a class no readiness: its constructor is static code, so new Late() inside a function
// works whenever the function runs, and a call that runs before class Late { } is reached prints what
// Node refuses. tsc rejects a use before the declaration at the module's own top level (TS2449), so
// the road is a function, a method or an arrow that holds the reference and is called before the
// declaration runs. That needs a statement before the declaration that can run code at all, so a
// reference inside any function in the module is not yet when such a statement comes first, and
// otherwise lowers as before. A reference inside the class's own body can't run before its
// declaration, since nothing can make an instance or reach a static before it. A reference from
// another module is judged where the modules are ordered (import cycles' own checks). Readiness for
// classes lifts this.
func (l *lowering) classDeadZoneGap(node *ast.Node) error {
	if node.Kind != ast.KindIdentifier || ast.IsPartOfTypeNode(node) {
		return nil
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsClass == 0 {
		return nil
	}
	declaration := symbol.ValueDeclaration
	if declaration == nil || declaration.Kind != ast.KindClassDeclaration || declaration.Parent == nil || declaration.Parent.Kind != ast.KindSourceFile {
		return nil
	}
	if node == declaration.Name() || ast.GetSourceFileOfNode(node) != ast.GetSourceFileOfNode(declaration) {
		return nil
	}
	inFunction := false
	for at := node.Parent; at != nil; at = at.Parent {
		if at == declaration {
			return nil
		}
		if ast.IsFunctionLike(at) {
			inFunction = true
		}
	}
	if !inFunction || !codeRunsBefore(declaration) {
		return nil
	}
	return l.notYet(node, "a class used from a function that may run before the class's declaration (its temporal dead zone)")
}

func init() { refusalRules.register(2050, "classDeadZoneGap", (*lowering).classDeadZoneGap) }

// codeRunsBefore is whether a top-level statement before declaration can run code, and so call a
// function that reaches it: anything but a declaration that only names things, or a variable whose
// initializers are literals and functions.
func codeRunsBefore(declaration *ast.Node) bool {
	for _, statement := range declaration.Parent.AsSourceFile().Statements.Nodes {
		if statement == declaration {
			return false
		}
		if !inertStatement(statement) {
			return true
		}
	}
	return false
}

// inertStatement is a top-level statement that runs no code of the program's when it's reached.
func inertStatement(statement *ast.Node) bool {
	switch statement.Kind {
	case ast.KindFunctionDeclaration, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindImportDeclaration, ast.KindImportEqualsDeclaration, ast.KindEmptyStatement:
		return true
	case ast.KindExportDeclaration:
		return statement.ModuleSpecifier() == nil
	case ast.KindVariableStatement:
		for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			if initializer := declaration.AsVariableDeclaration().Initializer; initializer != nil && !inertExpression(initializer) {
				return false
			}
		}
		return true
	case ast.KindClassDeclaration:
		// A class's heritage, static initializers, static blocks and computed names run at its declaration.
		for _, clause := range nodesOf(statement.AsClassDeclaration().HeritageClauses) {
			if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
				return false
			}
		}
		for _, member := range statement.Members() {
			if member.Kind == ast.KindClassStaticBlockDeclaration || member.Name() != nil && member.Name().Kind == ast.KindComputedPropertyName {
				return false
			}
			if member.Kind == ast.KindPropertyDeclaration && ast.HasStaticModifier(member) && member.AsPropertyDeclaration().Initializer != nil {
				return false
			}
		}
		return true
	}
	return false
}

// inertExpression is an initializer that runs nothing: a literal, a function, a name read, or an array
// or object literal of those. An object literal's methods and accessors are made, not run; a computed
// key or a spread would run code, so either makes it not inert.
func inertExpression(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindArrowFunction, ast.KindFunctionExpression:
		return true
	case ast.KindIdentifier:
		return true
	case ast.KindPrefixUnaryExpression:
		return node.AsPrefixUnaryExpression().Operand.Kind == ast.KindNumericLiteral
	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindPropertyAssignment:
				if property.Name().Kind == ast.KindComputedPropertyName || !inertExpression(property.AsPropertyAssignment().Initializer) {
					return false
				}
			case ast.KindShorthandPropertyAssignment:
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				if property.Name().Kind == ast.KindComputedPropertyName {
					return false
				}
			default:
				return false
			}
		}
		return true
	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			if !inertExpression(element) {
				return false
			}
		}
		return true
	}
	return false
}
