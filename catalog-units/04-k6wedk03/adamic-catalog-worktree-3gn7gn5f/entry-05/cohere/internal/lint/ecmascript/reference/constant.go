package reference

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ConstantStringIn is ConstantString with a scope, as eslint-utils' getStringIfConstant(node, scope)
// reads an expression: a name is followed to its binding's initializer when the binding counts as
// constant, and a regex literal is its text, `/pattern/flags`, as String() of the RegExp makes it.
//
// A binding counts as constant by eslint-utils' canBeConsideredConst: one declaration, a variable
// whose declarator names it alone (no pattern), and either `const` or never written after its
// initializer. Only this file's declarations are read, as ESLint's scope sees one file. What
// getStaticValue reaches beyond that, a built-in global's value or a safe built-in call such as
// `String.raw`, answers false here: the silent direction for every rule that asks.
//
// It takes the rule's context rather than a checker and a file so that every declaration it reads is
// visibly the rule's own file's, which the type-reach scan checks: a rule keyed on its imports' shapes
// may ask this.
func ConstantStringIn(ctx rule.Context, expression *ast.Node) (string, bool) {
	if ctx.TypeChecker == nil || ctx.SourceFile == nil {
		return ConstantString(expression)
	}
	following := map[*ast.Symbol]bool{}
	var resolve func(*ast.Node) (string, bool, bool)
	resolve = func(node *ast.Node) (string, bool, bool) {
		if node.Kind == ast.KindRegularExpressionLiteral {
			return node.Text(), true, true
		}
		initializer, symbol := constantInitializer(ctx, node)
		// A binding read inside its own initializer, `const a = a + 'x'`, has no value to give.
		if initializer == nil || following[symbol] {
			return "", false, false
		}
		following[symbol] = true
		defer delete(following, symbol)
		return constantValue(initializer, resolve)
	}
	text, _, isConstant := constantValue(expression, resolve)
	return text, isConstant
}

// constantInitializer returns the initializer a name's binding was declared with, when the binding
// counts as constant, and the symbol it read.
func constantInitializer(ctx rule.Context, name *ast.Node) (*ast.Node, *ast.Symbol) {
	symbol := ReadSymbol(ctx.TypeChecker, name)
	declarations := rule.DeclarationsIn(ctx.SourceFile, symbol)
	if symbol == nil || len(declarations) != 1 || rule.IsDeclaredOnlyInDeclarationFiles(symbol) {
		return nil, nil
	}
	declaration := declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration {
		return nil, nil
	}
	// A name a pattern declares has its binding element as its declaration, not the declarator, so a
	// declarator here always names one identifier.
	variable := declaration.AsVariableDeclaration()
	if variable.Initializer == nil {
		return nil, nil
	}
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return nil, nil
	}
	if list.Flags&ast.NodeFlagsConst == 0 && isWrittenAfterDeclaration(ctx, symbol, variable.Name()) {
		return nil, nil
	}
	return variable.Initializer, symbol
}

// isWrittenAfterDeclaration reports whether anything in the file writes the binding, the declaration's
// own name aside: eslint-utils' isEffectivelyConst, one initializing write and every other reference a
// read.
func isWrittenAfterDeclaration(ctx rule.Context, symbol *ast.Symbol, declarationName *ast.Node) bool {
	written := false
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if written {
			return true
		}
		if node.Kind == ast.KindIdentifier && node != declarationName && node.Text() == declarationName.Text() &&
			WritesToBinding(node) && ReadSymbol(ctx.TypeChecker, node) == symbol {
			written = true
			return true
		}
		return node.ForEachChild(visit)
	}
	ctx.SourceFile.AsNode().ForEachChild(visit)
	return written
}

// IsConstantRegExpIn reports whether an expression's static value is a RegExp object rather than a
// string: a regex literal, or a name whose constant binding (as ConstantStringIn counts one) was
// initialized with one, followed through any number of such names. eslint-utils' getStaticValue answers
// a RegExp there, which a rule that wants a pattern's text, not the String() of a RegExp, has to tell
// apart from a string constant.
func IsConstantRegExpIn(ctx rule.Context, expression *ast.Node) bool {
	following := map[*ast.Symbol]bool{}
	for expression != nil {
		expression = ast.SkipParentheses(expression)
		switch expression.Kind {
		case ast.KindRegularExpressionLiteral:
			return true
		case ast.KindIdentifier:
			if ctx.TypeChecker == nil || ctx.SourceFile == nil {
				return false
			}
			initializer, symbol := constantInitializer(ctx, expression)
			if initializer == nil || following[symbol] {
				return false
			}
			following[symbol] = true
			expression = initializer
		default:
			return false
		}
	}
	return false
}
