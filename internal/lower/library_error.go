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
	if node.Kind != ast.KindPropertyAccessExpression || node.Name().Text() != "captureStackTrace" {
		return false
	}
	receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	// Upstream Debug.fail discovers this Node member through Error as any. Only
	// this exact intrinsic access is admitted; no erased value is read or called.
	if receiver.Kind == ast.KindAsExpression && receiver.AsAsExpression().Type.Kind == ast.KindAnyKeyword {
		receiver = ast.SkipParentheses(receiver.AsAsExpression().Expression)
	}
	return l.isLibraryGlobal(receiver, "Error")
}

// The empty IR function is the definition on both backends. In particular the
// JavaScript backend never delegates to V8's stack capture implementation.
func (l *lowering) errorCaptureFunction() int {
	index := len(l.result.Functions)
	function := ir.Function{Name: "library_Error_captureStackTrace", Closure: true}
	l.result.Functions = append(l.result.Functions, function)
	return index
}

func (l *lowering) libraryErrorValue(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "stack" {
		return nil, true, l.notYet(node, "Error.stack: native frames have no JavaScript source stack; stack reads need a shared definition before lowering")
	}
	if node.Kind == ast.KindElementAccessExpression {
		key := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if key.Kind == ast.KindStringLiteral && key.Text() == "stack" {
			return nil, true, l.notYet(node, "Error.stack: native frames have no JavaScript source stack; stack reads need a shared definition before lowering")
		}
	}
	if node.Kind == ast.KindTypeOfExpression && l.isLibraryGlobal(node.AsTypeOfExpression().Expression, "Error") {
		return ir.StringConstant{Index: l.constant("function")}, true, nil
	}
	if l.isLibraryGlobal(node, "Error") {
		return nil, true, l.notYet(node, "Error as a value outside typeof and captureStackTrace discovery: constructor aliases and overloaded calls are not lowered")
	}
	if l.errorCaptureRead(node) {
		return l.errorCaptureValue(), true, nil
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
		if index == 1 {
			if _, absent := value.(ir.Undefined); absent {
				value = ir.Undefined{Of: ir.Closure}
			}
		}
		if (index == 0 && value.Type() != ir.Object) || (index == 1 && value.Type() != ir.Closure) {
			return nil, true, l.notYet(argument, "Error.captureStackTrace argument representation: target must be an object and constructorOpt a function")
		}
		arguments = append(arguments, value)
	}
	if len(arguments) == 1 {
		arguments = append(arguments, ir.Undefined{Of: ir.Closure})
	}
	return ir.CallClosure{Closure: l.errorCaptureValue(), Arguments: arguments}, true, nil
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

// Hold one closure for the static member so repeated reads preserve identity.
// It deliberately binds no arguments: a no-op must also accept an omitted
// constructorOpt when called through a detached function value.
func (l *lowering) errorCaptureValue() ir.Expression {
	if held, found := l.forwarders[-1]; found {
		return ir.Read{Local: held, Of: ir.Closure}
	}
	function := l.errorCaptureFunction()
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "Error_captureStackTrace_value", Type: ir.Closure, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: held, Value: ir.MakeClosure{Function: function}})
	if l.forwarders == nil {
		l.forwarders = map[int]int{}
	}
	// Ordinary function forwarders use nonnegative function indexes.
	l.forwarders[-1] = held
	return ir.Read{Local: held, Of: ir.Closure}
}

// Destructuring reads an own field without a property-access AST node. Keep
// that observation behind the same stack boundary, including renamed bindings.
func (l *lowering) libraryErrorBindingRefusal(node *ast.Node) error {
	if node.Kind != ast.KindBindingElement || node.Parent == nil || node.Parent.Kind != ast.KindObjectBindingPattern {
		return nil
	}
	name := node.Name()
	if declared := node.AsBindingElement(); declared.PropertyName != nil {
		name = declared.PropertyName
	}
	if name != nil && (ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral) && name.Text() == "stack" {
		return l.notYet(node, "Error.stack: native frames have no JavaScript source stack; stack reads need a shared definition before lowering")
	}
	return nil
}
