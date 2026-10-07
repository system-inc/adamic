// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// console lowers console.log and console.error with one string.
func (l *lowering) console(call *ast.Node) (ir.Statement, error) {
	stream, err := l.consoleStream(call.AsCallExpression().Expression)
	if err != nil {
		return nil, err
	}
	arguments := call.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		// The prelude declares one parameter, so the checker has already refused any other count.
		return nil, errors.New("lower: " + l.program.Where(call) + ": console takes one argument, and the checker let another count through")
	}
	value, err := l.expression(arguments[0])
	if err != nil {
		return nil, err
	}
	if _, null := value.(ir.Null); null {
		value = ir.StringConstant{Index: l.constant("null")}
	} else if _, undefined := value.(ir.Undefined); undefined {
		value = ir.StringConstant{Index: l.constant("undefined")}
	} else {
		value = l.spelled(arguments[0], value)
	}
	return ir.WriteLine{Stream: stream, Value: value}, nil
}

// isPanicCall reports whether an expression is a call to the prelude's panic, which never returns.
func (l *lowering) isPanicCall(expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	return expression.Kind == ast.KindCallExpression && l.isPreludeFunction(expression.AsCallExpression().Expression, "panic")
}

// isPreludeFunction reports whether a callee is the prelude's function of that name, as imported
// from 'adamic', and not one of the program's that shares the name.
func (l *lowering) isPreludeFunction(callee *ast.Node, name string) bool {
	callee = ast.SkipParentheses(callee)
	if !ast.IsIdentifier(callee) {
		return false
	}
	symbol := l.symbol(callee)
	return symbol != nil && symbol.Name == name && len(symbol.Declarations) > 0 && load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// isConsole reports whether a callee is a method of the prelude's console.
func (l *lowering) isConsole(callee *ast.Node) bool {
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(callee.AsPropertyAccessExpression().Expression)
	return symbol != nil && len(symbol.Declarations) > 0 && load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// consoleStream is the stream a callee writes to, when it is the prelude's console.log or
// console.error. A local named console is someone else's, and isn't lowered as this one.
func (l *lowering) consoleStream(callee *ast.Node) (ir.Stream, error) {
	if callee.Kind != ast.KindPropertyAccessExpression {
		return 0, l.notYet(callee, "a call to "+describe(callee))
	}
	object := callee.AsPropertyAccessExpression().Expression
	symbol := l.checker.GetSymbolAtLocation(object)
	if symbol == nil || len(symbol.Declarations) == 0 || !load.IsPrelude(ast.GetSourceFileOfNode(symbol.Declarations[0])) {
		return 0, l.notYet(callee, "a method call")
	}
	switch callee.Name().Text() {
	case "log":
		return ir.Stdout, nil
	case "error":
		return ir.Stderr, nil
	}
	return 0, l.notYet(callee, "console."+callee.Name().Text())
}
