package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	RegisterNodeLibraryMembers("node:globals.ErrorConstructor.captureStackTrace", "node:globals.ErrorConstructor.stackTraceLimit")
}

func (l *lowering) errorMember(node *ast.Node, name string) bool {
	node = ast.SkipParentheses(node)
	var receiver *ast.Node
	if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == name {
		receiver = node.AsPropertyAccessExpression().Expression
	} else if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind == ast.KindStringLiteral && key.Text() == name {
			receiver = access.Expression
		}
	}
	if receiver == nil {
		return false
	}
	receiver = ast.SkipParentheses(receiver)
	if receiver.Kind == ast.KindAsExpression && receiver.AsAsExpression().Type.Kind == ast.KindAnyKeyword {
		receiver = ast.SkipParentheses(receiver.AsAsExpression().Expression)
	}
	return l.isLibraryGlobal(receiver, "Error")
}

func (l *lowering) errorCaptureRead(node *ast.Node) bool {
	return l.errorMember(node, "captureStackTrace")
}

// Follow only immutable aliases of the intrinsic, never an arbitrary function
// whose type happens to have the same signature.
func (l *lowering) errorCaptureCallee(node *ast.Node) bool {
	for depth := 0; depth < 16; depth++ {
		node = ast.SkipParentheses(node)
		if l.errorCaptureRead(node) {
			return true
		}
		if !ast.IsIdentifier(node) {
			return false
		}
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		node = declaration.AsVariableDeclaration().Initializer
		if node == nil {
			return false
		}
	}
	return false
}

// This global is ordinary IR, so the same initialization, reads and writes run
// on both backends. Negative keys do not collide with named function forwarders.
func (l *lowering) errorLimitLocal() int {
	if held, found := l.forwarders[-2]; found {
		return held
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "Error_stackTraceLimit", Type: ir.Number, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: held, Value: ir.NumberConstant{Value: 10}})
	if l.forwarders == nil {
		l.forwarders = map[int]int{}
	}
	l.forwarders[-2] = held
	return held
}

func (l *lowering) errorCaptureValue() ir.Expression {
	if held, found := l.forwarders[-1]; found {
		return ir.Read{Local: held, Of: ir.Closure}
	}
	index := len(l.result.Functions)
	target := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "target", Type: ir.Object, Function: index})
	// constructorOpt affects only omitted frame text. Its expression is still
	// evaluated by the caller, including when this intrinsic is detached.
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "library_Error_captureStackTrace", Closure: true, Parameters: []int{target}, Body: []ir.Statement{ir.Evaluate{Value: ir.ObjectCall{Method: "errorCaptureStack", Arguments: []ir.Expression{ir.Read{Local: target, Of: ir.Object}}}}}})
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "Error_captureStackTrace", Type: ir.Closure, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}})
	if l.forwarders == nil {
		l.forwarders = map[int]int{}
	}
	l.forwarders[-1] = held
	return ir.Read{Local: held, Of: ir.Closure}
}

func (l *lowering) libraryErrorValue(node *ast.Node) (ir.Expression, bool, error) {
	if l.errorMember(node, "stackTraceLimit") {
		return ir.Read{Local: l.errorLimitLocal(), Of: ir.Number}, true, nil
	}
	if node.Kind == ast.KindBinaryExpression && l.errorMember(node.AsBinaryExpression().Left, "stackTraceLimit") {
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindEqualsToken {
			if _, assignment := compoundAssignments[binary.OperatorToken.Kind]; assignment {
				return nil, true, l.notYet(node, "Error.stackTraceLimit compound updates need an evaluation-order proof")
			}
			return nil, false, nil
		}
		value, err := l.expression(binary.Right)
		if err != nil {
			return nil, true, err
		}
		local := l.errorLimitLocal()
		if value.Type() != ir.Number {
			return nil, true, l.notYet(node, "Error.stackTraceLimit requires a number")
		}
		b := l.libraryArrayBuilder([]ir.Expression{value})
		b.body = append(b.body, ir.Assign{Local: local, Value: b.read(b.parameters[0])})
		return b.finish("Error_setStackTraceLimit", b.read(b.parameters[0])), true, nil
	}
	if node.Kind == ast.KindTypeOfExpression && l.isLibraryGlobal(node.AsTypeOfExpression().Expression, "Error") {
		return ir.StringConstant{Index: l.constant("function")}, true, nil
	}
	if l.errorCaptureRead(node) {
		return l.errorCaptureValue(), true, nil
	}
	if l.isLibraryGlobal(node, "Error") {
		return nil, true, l.notYet(node, "Error constructor aliases outside typeof and captureStackTrace are not lowered")
	}
	var receiver *ast.Node
	if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "stack" {
		receiver = node.AsPropertyAccessExpression().Expression
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		key := ast.SkipParentheses(access.ArgumentExpression)
		if key.Kind == ast.KindStringLiteral && key.Text() == "stack" {
			receiver = access.Expression
		}
	}
	if receiver != nil {
		if l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "Error") && !l.errorCapturedBefore(node, receiver) {
			return nil, true, l.notYet(node, "Error.stack without a preceding captureStackTrace: native frames have no JavaScript source stack")
		}
		if !l.errorPlainTarget(receiver) && !l.errorCapturedBefore(node, receiver) {
			return nil, true, l.notYet(node, "stack receiver must be a known plain object or previously captured Error")
		}
		of, known := l.representation(l.checker.GetTypeAtLocation(node))
		if !known || of != ir.String {
			return nil, false, nil
		}
		object, err := l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
		if object.Type() != ir.Object {
			return nil, true, l.notYet(node, "captured stack on other than a present object")
		}
		return ir.ObjectCall{Method: "errorReadStack", Arguments: []ir.Expression{object}, Returns: ir.String}, true, nil
	}
	if value, known, err := l.errorStackHasOwn(node); known {
		return value, true, err
	}
	if node.Kind != ast.KindCallExpression {
		return nil, false, nil
	}
	call := node.AsCallExpression()
	if !l.errorCaptureCallee(call.Expression) {
		return nil, false, nil
	}
	if l.errorCaptureFrozenProgram() {
		return nil, true, l.notYet(node, "captureStackTrace in a program that freezes objects: defining the stack property needs catchable descriptor failures")
	}
	written := nodesOf(call.Arguments)
	if len(written) < 1 || len(written) > 2 || hasSpread(node) {
		return nil, true, l.notYet(node, "Error.captureStackTrace needs a target and optional constructor")
	}
	arguments := []ir.Expression{}
	for index, argument := range written {
		value, err := l.errorCaptureArgument(argument)
		if err != nil {
			return nil, true, err
		}
		if index == 0 {
			if value.Type() != ir.Object || l.includesUndefined(l.checker.GetTypeAtLocation(argument)) {
				return nil, true, l.notYet(argument, "Error.captureStackTrace target must be a present native object")
			}
			if stack := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(argument), "stack"); stack != nil {
				of, known := l.representation(l.checker.GetTypeOfSymbol(stack))
				if !known || of != ir.String {
					return nil, true, l.notYet(argument, "captureStackTrace would overwrite a non-string stack field")
				}
			}
			if !l.errorPlainTarget(argument) && !l.errorNativeTarget(argument) {
				return nil, true, l.notYet(argument, "captureStackTrace target shape must be a known plain object or Error")
			}
			if stack := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(argument), "stack"); stack != nil {
				representation, known := l.representation(l.checker.GetTypeOfSymbol(stack))
				if !known || representation != ir.String {
					return nil, true, l.notYet(argument, "captureStackTrace would overwrite a non-string stack field")
				}
			}
		} else if value.Type() != ir.Closure {
			if _, absent := value.(ir.Undefined); !absent {
				return nil, true, l.notYet(argument, "captureStackTrace constructorOpt must be a function or undefined")
			}
		}
		if index == 1 {
			if _, absent := value.(ir.Undefined); absent {
				value = ir.Undefined{Of: ir.Closure}
			}
		}
		arguments = append(arguments, value)
	}
	closure, err := l.expression(call.Expression)
	if err != nil {
		return nil, true, err
	}
	return ir.CallClosure{Closure: closure, Arguments: arguments}, true, nil
}

func (l *lowering) libraryErrorCondition(node *ast.Node) (ir.Expression, bool, error) {
	if !l.errorCaptureRead(node) {
		return nil, false, nil
	}
	value, err := l.expression(node)
	return ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: value}}, true, err
}

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

// The ordinary Error stack remains refused. A direct capture on the same const
// binding in an earlier sibling statement proves this specific read is captured.
func (l *lowering) errorCapturedBefore(node, receiver *ast.Node) bool {
	receiver = ast.SkipParentheses(receiver)
	if !ast.IsIdentifier(receiver) {
		return false
	}
	symbol := l.symbol(receiver)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	for statement := node; statement.Parent != nil; statement = statement.Parent {
		parent := statement.Parent
		if ast.IsFunctionLike(parent) {
			return false
		}
		var siblings []*ast.Node
		if parent.Kind == ast.KindBlock {
			siblings = parent.AsBlock().Statements.Nodes
		} else if parent.Kind == ast.KindSourceFile {
			siblings = parent.AsSourceFile().Statements.Nodes
		} else {
			continue
		}
		for _, earlier := range siblings {
			if earlier == statement {
				break
			}
			if earlier.Kind != ast.KindExpressionStatement {
				continue
			}
			expression := ast.SkipParentheses(earlier.AsExpressionStatement().Expression)
			if expression.Kind != ast.KindCallExpression || !l.errorCaptureCallee(expression.AsCallExpression().Expression) {
				continue
			}
			arguments := nodesOf(expression.AsCallExpression().Arguments)
			if len(arguments) > 0 && l.symbol(ast.SkipParentheses(arguments[0])) == symbol {
				return true
			}
		}
	}
	return false
}

// Following immutable initializers keeps structural views from hiding an Error
// or a non-string field that capture would replace with a string.
func (l *lowering) errorPlainTarget(node *ast.Node) bool {
	for depth := 0; depth < 16; depth++ {
		node = ast.SkipParentheses(node)
		if node.Kind == ast.KindObjectLiteralExpression {
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				if (property.Kind == ast.KindGetAccessor || property.Kind == ast.KindSetAccessor) && property.Name() != nil && property.Name().Text() == "stack" {
					return false
				}
				if property.Kind == ast.KindSpreadAssignment {
					return false
				}
			}
			// Existing shaped stack fields need per-field enumerability after
			// replacement and copying; an added stack has dedicated storage.
			if l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(node), "stack") != nil {
				return false
			}
			return true
		}
		if !ast.IsIdentifier(node) {
			return false
		}
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		node = declaration.AsVariableDeclaration().Initializer
		if node == nil {
			return false
		}
	}
	return false
}

// Stack capture adds exactly this one known own property. Its presence test does
// not need the fixed-shape proof required by arbitrary Object reflection.
func (l *lowering) errorStackHasOwn(node *ast.Node) (ir.Expression, bool, error) {
	if node.Kind != ast.KindCallExpression {
		return nil, false, nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	arguments := nodesOf(call.Arguments)
	var target, key *ast.Node
	access := callee.AsPropertyAccessExpression()
	if access.Name().Text() == "hasOwn" && l.isLibraryGlobal(access.Expression, "Object") && len(arguments) == 2 {
		target, key = arguments[0], arguments[1]
	} else if access.Name().Text() == "hasOwnProperty" && l.libraryMember(callee) && len(arguments) == 1 {
		target, key = access.Expression, arguments[0]
	} else {
		return nil, false, nil
	}
	key = ast.SkipParentheses(key)
	if key.Kind != ast.KindStringLiteral || key.Text() != "stack" {
		return nil, false, nil
	}
	if l.isLibraryType(l.checker.GetTypeAtLocation(target), "Error") && !l.errorCapturedBefore(node, target) {
		return nil, true, l.notYet(node, "Error.stack own-property observation without a preceding captureStackTrace")
	}
	if !l.errorPlainTarget(target) && !l.errorCapturedBefore(node, target) {
		return nil, true, l.notYet(node, "stack own-property receiver must be a known plain object or previously captured Error")
	}
	value, err := l.expression(target)
	if err != nil {
		return nil, true, err
	}
	if value.Type() != ir.Object || l.includesUndefined(l.checker.GetTypeAtLocation(target)) {
		return nil, true, l.notYet(node, "captured stack own-property check needs a present object")
	}
	return ir.ObjectCall{Method: "hasOwn", Arguments: []ir.Expression{value, ir.StringConstant{Index: l.constant("stack")}}, Returns: ir.Boolean}, true, nil
}

// Freezing prevents defining stack. Keep those descriptor failures out of the
// native path until this intrinsic participates in exception unwinding.
func (l *lowering) errorCaptureFrozenProgram() bool {
	found := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "freeze" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
				found = true
			}
		}
		if !found {
			node.ForEachChild(visit)
		}
		return found
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if !file.IsDeclarationFile {
			file.AsNode().ForEachChild(visit)
		}
	}
	return found
}

// Destructuring bypasses ordinary property lowering; keep it refused until its
// source and capture dominance are proven by that path as well.
func (l *lowering) errorStackSyntaxRefusal(node *ast.Node) error {
	if node.Kind == ast.KindBindingElement {
		binding := node.AsBindingElement()
		key := binding.PropertyName
		if key == nil {
			key = binding.Name()
		}
		if key != nil && (ast.IsIdentifier(key) || key.Kind == ast.KindStringLiteral) && key.Text() == "stack" {
			return l.notYet(node, "stack destructuring needs a captured-property proof; use a direct stack read")
		}
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		key := ast.SkipParentheses(binary.Left)
		if binary.OperatorToken.Kind == ast.KindInKeyword && key.Kind == ast.KindStringLiteral && key.Text() == "stack" && !l.errorPlainTarget(binary.Right) && !l.errorCapturedBefore(node, binary.Right) {
			return l.notYet(node, "stack presence receiver must be a known plain object or previously captured Error (host descriptor metadata)")
		}
	}
	return nil
}

func (l *lowering) errorNativeTarget(node *ast.Node) bool {
	for depth := 0; depth < 16; depth++ {
		node = ast.SkipParentheses(node)
		if node.Kind == ast.KindNewExpression {
			return l.isLibraryGlobal(node.AsNewExpression().Expression, "Error")
		}
		if !ast.IsIdentifier(node) {
			return false
		}
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		node = declaration.AsVariableDeclaration().Initializer
		if node == nil {
			return false
		}
	}
	return false
}
