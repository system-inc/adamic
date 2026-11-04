package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// A function declaration with a `this` parameter is called as f(1) while its parameter list reads
// (this: void, n: number), so lowering it as written gives n nothing and the JavaScript backend prints
// NaN where Node prints 2 (#j8mt4pr). A class method or an object literal's method has the same shift:
// signature() (functions.go) declares every parameter, this included, after the receiver it already
// takes as its first. Its siblings already stop: a function with a this parameter used as a value
// (expression.go), a function expression with one (library_function_expressions.go), a nested
// declaration (nested_functions.go) and a namespace function (namespaces.go). An arrow, a constructor
// and an accessor can't have one (tsc's TS2730, TS2681 and TS2784). A declaration or a method stops the
// same way until the this parameter is dropped from the lowered parameters, which is the real fix.
func (l *lowering) thisParameterDeclarationGap(node *ast.Node) error {
	if node.Kind != ast.KindFunctionDeclaration && node.Kind != ast.KindMethodDeclaration {
		return nil
	}
	for _, parameter := range node.Parameters() {
		if ast.IsIdentifier(parameter.Name()) && parameter.Name().Text() == "this" {
			return l.notYet(parameter, "a function declaration or method with a this parameter (dynamic receivers are not implemented)")
		}
	}
	return nil
}

func init() {
	refusalRules.register(1450, "thisParameterDeclarationGap", (*lowering).thisParameterDeclarationGap)
}
