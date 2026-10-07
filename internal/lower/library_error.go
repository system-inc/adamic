package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	RegisterNodeLibraryMembers("node:globals.ErrorConstructor.captureStackTrace")
}

func (l *lowering) errorCaptureRead(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	return node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "captureStackTrace" && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "Error")
}

// The empty IR function is the definition on both backends. In particular the
// JavaScript backend never delegates to V8's stack capture implementation.
func (l *lowering) errorCaptureFunction() int {
	for index, function := range l.result.Functions {
		if function.Name == "library_Error_captureStackTrace" {
			return index
		}
	}
	index := len(l.result.Functions)
	function := ir.Function{Name: "library_Error_captureStackTrace", Closure: true}
	for _, parameter := range []struct {
		name string
		of   ir.Type
	}{{"target", ir.Object}, {"constructorOpt", ir.Closure}} {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: parameter.name, Type: parameter.of, Function: index})
		function.Parameters = append(function.Parameters, local)
	}
	l.result.Functions = append(l.result.Functions, function)
	return index
}

func (l *lowering) libraryErrorValue(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind == ast.KindTypeOfExpression && l.isLibraryGlobal(node.AsTypeOfExpression().Expression, "Error") {
		return ir.StringConstant{Index: l.constant("function")}, true, nil
	}
	if l.isLibraryGlobal(node, "Error") {
		return nil, true, l.notYet(node, "Error as a value outside typeof and captureStackTrace discovery: constructor aliases and overloaded calls are not lowered")
	}
	if l.errorCaptureRead(node) {
		return ir.MakeClosure{Function: l.errorCaptureFunction()}, true, nil
	}
	if node.Kind != ast.KindCallExpression || !l.errorCaptureRead(node.AsCallExpression().Expression) {
		return nil, false, nil
	}
	written := nodesOf(node.AsCallExpression().Arguments)
	if len(written) < 1 || len(written) > 2 || hasSpread(node) {
		return nil, true, l.notYet(node, "Error.captureStackTrace needs a target and an optional constructor function")
	}
	arguments := []ir.Expression{}
	for index, argument := range written {
		value, err := l.errorCaptureArgument(argument)
		if err != nil {
			return nil, true, err
		}
		if (index == 0 && value.Type() != ir.Object) || (index == 1 && value.Type() != ir.Closure && value.Type() != (ir.Undefined{}).Type()) {
			return nil, true, l.notYet(argument, "Error.captureStackTrace argument representation: target must be an object and constructorOpt a function")
		}
		arguments = append(arguments, value)
	}
	if len(arguments) == 1 {
		arguments = append(arguments, ir.Undefined{Of: ir.Closure})
	}
	return ir.CallClosure{Closure: ir.MakeClosure{Function: l.errorCaptureFunction()}, Arguments: arguments}, true, nil
}

func (l *lowering) libraryErrorCondition(node *ast.Node) (ir.Expression, bool, error) {
	if !l.errorCaptureRead(node) {
		return nil, false, nil
	}
	value, err := l.expression(node)
	return ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: value}}, true, err
}

// Debug.fail selects its crawl marker with ||. Functions are truthy, and the
// only absent member admitted here is undefined, so coalescing is exact.
func (l *lowering) errorCaptureArgument(node *ast.Node) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken {
		binary := node.AsBinaryExpression()
		left, err := l.expression(binary.Left)
		if err != nil {
			return nil, err
		}
		right, err := l.expression(binary.Right)
		if err != nil {
			return nil, err
		}
		if left.Type() != ir.Closure || right.Type() != ir.Closure {
			return nil, l.notYet(node, "captureStackTrace crawl marker selection needs function values or undefined")
		}
		return ir.Coalesce{Value: left, Fallback: right, Of: ir.Closure}, nil
	}
	return l.expression(node)
}
