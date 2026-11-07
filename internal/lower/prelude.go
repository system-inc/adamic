// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// console lowers console.log and console.error with one string.
func (lowering *lowering) console(call *ast.Node) (ir.Statement, error) {
	stream, consoleStreamError := lowering.consoleStream(call.AsCallExpression().Expression)
	if consoleStreamError != nil {
		return nil, consoleStreamError
	}
	arguments := call.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		// The prelude declares one parameter, so the checker has already refused any other count.
		return nil, lowering.notYet(
			call,
			"console."+call.AsCallExpression().Expression.Name().Text()+" with other than one string argument",
		)
	}
	value, expressionError := lowering.expression(arguments[0])
	if expressionError != nil {
		return nil, expressionError
	}
	if value.Type() != ir.String {
		return nil, lowering.notYet(
			call,
			"console."+call.AsCallExpression().Expression.Name().Text()+" with a non-string argument",
		)
	}
	return ir.WriteLine{Stream: stream, Value: lowering.spelled(arguments[0], value)}, nil
}

// isPanicCall reports whether an expression is a call to the prelude's panic, which never returns.
func (lowering *lowering) isPanicCall(expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	return expression.Kind == ast.KindCallExpression &&
		lowering.isPreludeFunction(expression.AsCallExpression().Expression, "panic")
}

// isPreludeFunction reports whether a callee is the prelude's function of that name, as imported
// from 'adamic', and not one of the program's that shares the name.
func (lowering *lowering) isPreludeFunction(callee *ast.Node, name string) bool {
	callee = ast.SkipParentheses(callee)
	if !ast.IsIdentifier(callee) {
		return false
	}
	symbol := lowering.symbol(callee)
	return symbol != nil &&
		symbol.Name == name &&
		len(symbol.Declarations) > 0 &&
		load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// isConsole reports whether a callee is a method of the prelude's console.
func (lowering *lowering) isConsole(callee *ast.Node) bool {
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	symbol := lowering.checker.GetSymbolAtLocation(callee.AsPropertyAccessExpression().Expression)
	return isPreludeConsole(symbol)
}

// consoleStream is the stream a callee writes to, when it is the prelude's console.log or
// console.error. A local named console is someone else's, and isn't lowered as this one.
func (lowering *lowering) consoleStream(callee *ast.Node) (ir.Stream, error) {
	if callee.Kind != ast.KindPropertyAccessExpression {
		return 0, lowering.notYet(callee, "a call to "+describe(callee))
	}
	object := callee.AsPropertyAccessExpression().Expression
	symbol := lowering.checker.GetSymbolAtLocation(object)
	if !isPreludeConsole(symbol) {
		return 0, lowering.notYet(callee, "a method call")
	}
	switch callee.Name().Text() {
	case "log":
		return ir.Stdout, nil
	case "error":
		return ir.Stderr, nil
	}
	return 0, lowering.notYet(callee, "console."+callee.Name().Text())
}

// Node's global console declaration merges with the prelude variable. Inspect
// all declarations so source ordering cannot turn the real console into a user one.
func isPreludeConsole(symbol *ast.Symbol) bool {
	if symbol == nil || symbol.Name != "console" {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if load.IsPrelude(ast.GetSourceFileOfNode(declaration)) {
			return true
		}
	}
	return false
}
