package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func typedArrayName(kind ir.Type) string {
	switch kind {
	case ir.Uint16Array:
		return "Uint16Array"
	case ir.Uint8Array:
		return "Uint8Array"
	case ir.Int32Array:
		return "Int32Array"
	case ir.Float64Array:
		return "Float64Array"
	}
	return "typed array"
}

func (l *lowering) typedArrayKind(proven *checker.Type) ir.Type {
	for _, kind := range []ir.Type{ir.Uint8Array, ir.Uint16Array, ir.Int32Array, ir.Float64Array} {
		if l.isLibraryType(proven, typedArrayName(kind)) {
			return kind
		}
	}
	return 0
}

// Reject unsupported buffer facilities while their library identity is still visible.
func (l *lowering) typedArrayUnsupported(node *ast.Node) error {
	if node.Kind == ast.KindTypeReference {
		proven := l.checker.GetTypeAtLocation(node)
		if l.isLibraryType(proven, "SharedArrayBuffer") {
			return l.notYet(node, "sharing typed arrays across parallel tasks")
		}
		for _, name := range []string{"ArrayBuffer", "DataView", "Int8Array", "Uint8ClampedArray", "Int16Array", "Uint32Array", "Float32Array", "BigInt64Array", "BigUint64Array"} {
			if l.isLibraryType(proven, name) {
				if l.program.UsesProjectOptions() && l.numericTypedArray(proven) {
					continue
				}
				return l.notYet(node, "typed array facility "+name)
			}
		}
	}
	if node.Kind == ast.KindNewExpression {
		made := node.AsNewExpression()
		if l.isLibraryGlobal(made.Expression, "SharedArrayBuffer") {
			return l.notYet(node, "sharing typed arrays across parallel tasks")
		}
		if l.isLibraryGlobal(made.Expression, "ArrayBuffer") {
			if made.Arguments != nil && len(made.Arguments.Nodes) > 1 {
				return l.notYet(node, "resizable ArrayBuffer buffers")
			}
			return l.notYet(node, "ArrayBuffer")
		}
		if l.isLibraryGlobal(made.Expression, "DataView") {
			return l.notYet(node, "DataView")
		}
		for _, name := range []string{"Int8Array", "Uint8ClampedArray", "Int16Array", "Uint32Array", "Float32Array", "BigInt64Array", "BigUint64Array"} {
			if l.isLibraryGlobal(made.Expression, name) {
				if l.program.UsesProjectOptions() && l.numericTypedArray(l.checker.GetTypeAtLocation(node)) {
					continue
				}
				return l.notYet(node, "typed array element type "+name)
			}
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		access := node.AsPropertyAccessExpression()
		if kind := l.typedArrayKind(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))); kind != 0 && access.Name().Text() == "buffer" {
			return l.notYet(node, "constructing a view over another view's buffer")
		}
	}
	return nil
}

func (l *lowering) typedArrayExpression(node *ast.Node) (ir.Expression, bool, error) {
	if err := l.typedArrayUnsupported(node); err != nil {
		return nil, true, err
	}
	switch node.Kind {
	case ast.KindNewExpression:
		made := node.AsNewExpression()
		for _, kind := range []ir.Type{ir.Uint8Array, ir.Uint16Array, ir.Int32Array, ir.Float64Array} {
			if !l.isLibraryGlobal(made.Expression, typedArrayName(kind)) {
				continue
			}
			var arguments []*ast.Node
			if made.Arguments != nil {
				arguments = made.Arguments.Nodes
			}
			if len(arguments) > 1 {
				return nil, true, l.notYet(node, "constructing a typed array view over a buffer")
			}
			if len(arguments) == 0 {
				return ir.TypedArrayNew{Of: kind, Source: ir.NumberConstant{Value: 0}}, true, nil
			}
			if literal := ast.SkipParentheses(arguments[0]); literal.Kind == ast.KindArrayLiteralExpression && len(literal.AsArrayLiteralExpression().Elements.Nodes) == 0 {
				return ir.TypedArrayNew{Of: kind, Source: ir.ArrayLiteral{Element: ir.Number}, FromArray: true}, true, nil
			}
			source, err := l.expression(arguments[0])
			if err != nil {
				return nil, true, err
			}
			if source.Type() == ir.Number {
				return ir.TypedArrayNew{Of: kind, Source: source}, true, nil
			}
			if source.Type() != ir.Array {
				return nil, true, l.notYet(node, "constructing a typed array from other than a length or number[]")
			}
			element, err := l.elementType(arguments[0])
			if err != nil {
				return nil, true, err
			}
			if element != ir.Number {
				return nil, true, l.notYet(node, "constructing a typed array from other than number[]")
			}
			return ir.TypedArrayNew{Of: kind, Source: source, FromArray: true}, true, nil
		}
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		kind := l.typedArrayKind(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)))
		if kind == 0 {
			break
		}
		if access.QuestionDotToken == nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {
			return nil, true, l.notYet(node, "an optional chain longer than one step")
		}
		member := access.Name().Text()
		if member != "length" && member != "byteLength" {
			return nil, true, l.notYet(node, "typed array member "+access.Name().Text())
		}
		array, err := l.expression(access.Expression)
		if err != nil {
			return nil, true, err
		}
		if member == "byteLength" {
			if access.QuestionDotToken != nil {
				return nil, true, l.notYet(node, "optional typed array byteLength access")
			}
			width := 1.0
			switch kind {
			case ir.Uint16Array:
				width = 2
			case ir.Int32Array:
				width = 4
			case ir.Float64Array:
				width = 8
			}
			return ir.Binary{Operator: ir.Multiply, Left: ir.Length{Array: array}, Right: ir.NumberConstant{Value: width}}, true, nil
		}
		return ir.Length{Array: array, Optional: access.QuestionDotToken != nil}, true, nil
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		if l.typedArrayKind(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))) == 0 {
			break
		}
		if access.QuestionDotToken == nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {
			value, err := l.optionalIndexContinuation(node)
			return value, true, err
		}
		if access.QuestionDotToken != nil {
			array, err := l.expression(access.Expression)
			if err != nil {
				return nil, true, err
			}
			value, err := l.optionalIndex(node, array)
			return value, true, err
		}
		array, err := l.expression(access.Expression)
		if err != nil {
			return nil, true, err
		}
		index, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, true, err
		}
		if index.Type() != ir.Number {
			return nil, true, l.notYet(node, "a typed array index that isn't a number")
		}
		read := ir.Expression(ir.ArrayIndex{Array: array, Index: index, Element: ir.Number})
		if held, _ := l.representation(l.checker.GetTypeAtLocation(node)); held == ir.Number && !l.acceptsUndefined(node) {
			read = ir.Unwrap{Value: read}
		}
		return read, true, nil
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression {
			break
		}
		receiver := callee.AsPropertyAccessExpression().Expression
		kind := l.typedArrayKind(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)))
		if kind == 0 {
			break
		}
		if err := l.optionalCall(node); err != nil {
			return nil, true, err
		}
		method := callee.Name().Text()
		args := call.Arguments.Nodes
		switch method {
		case "fill":
			if len(args) < 1 || len(args) > 3 {
				return nil, true, l.notYet(node, "typed array fill argument count")
			}
		case "set":
			if len(args) < 1 || len(args) > 2 {
				return nil, true, l.notYet(node, "typed array set argument count")
			}
		case "subarray":
			if len(args) > 2 {
				return nil, true, l.notYet(node, "typed array subarray argument count")
			}
		default:
			return nil, true, l.notYet(node, "typed array method "+method)
		}
		array, err := l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
		lowered := []ir.Expression{}
		for index, argument := range args {
			value, err := l.expression(argument)
			if err != nil {
				return nil, true, err
			}
			if index == 0 && method == "set" {
				if value.Type() != kind {
					return nil, true, l.notYet(argument, "typed array set from a different element type or a plain array")
				}
			} else if method == "fill" && index == 0 {
				if value.Type() != ir.Number {
					return nil, true, l.notYet(argument, "typed array fill with other than a number")
				}
			} else {
				// Optional numeric arguments may explicitly be undefined, as well as omitted.
				value = fit(value, ir.MaybeNumber)
				if value.Type() != ir.MaybeNumber {
					return nil, true, l.notYet(argument, "a typed array offset that isn't number or undefined")
				}
			}
			lowered = append(lowered, value)
		}
		switch method {
		case "fill":
			return ir.TypedArrayFill{Array: array, Arguments: lowered}, true, nil
		case "set":
			return ir.TypedArraySet{Array: array, Arguments: lowered}, true, nil
		case "subarray":
			return ir.TypedArraySubarray{Array: array, Arguments: lowered}, true, nil
		}
	}
	return nil, false, nil
}

func (l *lowering) typedArrayWrite(target, valueNode *ast.Node) ([]ir.Statement, bool, error) {
	access := target.AsElementAccessExpression()
	if l.typedArrayKind(l.checker.GetTypeAtLocation(access.Expression)) == 0 {
		return nil, false, nil
	}
	array, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	index, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, true, err
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, true, err
	}
	if index.Type() != ir.Number || value.Type() != ir.Number {
		return nil, true, l.notYet(target, "a typed array write with nonnumeric index or value")
	}
	return []ir.Statement{ir.SetIndex{Array: array, Index: index, Value: value, Element: ir.Number, Site: l.writeSite(access.Expression)}}, true, nil
}

func (l *lowering) typedArrayForOf(node *ast.Node) ([]ir.Statement, bool, error) {
	statement := node.AsForInOrOfStatement()
	if l.typedArrayKind(l.checker.GetTypeAtLocation(statement.Expression)) == 0 {
		return nil, false, nil
	}
	initializer := statement.Initializer
	if statement.AwaitModifier != nil {
		return nil, true, l.notYet(node, "for await over a typed array")
	}
	if initializer.Kind != ast.KindVariableDeclarationList || initializer.Flags&ast.NodeFlagsBlockScoped == 0 {
		return nil, true, l.notYet(initializer, "a typed array loop without a const or let declaration")
	}
	declarations := initializer.AsVariableDeclarationList().Declarations.Nodes
	if len(declarations) != 1 || !ast.IsIdentifier(declarations[0].Name()) {
		return nil, true, l.notYet(initializer, "a typed array loop without a plain element name")
	}
	array, err := l.expression(statement.Expression)
	if err != nil {
		return nil, true, err
	}
	local, err := l.declareLocal(declarations[0].Name())
	if err != nil {
		return nil, true, err
	}
	body, err := l.statement(statement.Statement)
	if err != nil {
		return nil, true, err
	}
	return []ir.Statement{ir.ForOf{Iterable: array, Element: ir.Number, Local: local, Body: body}}, true, nil
}

// set's ArrayLike<number> parameter is a library protocol, not a kept object view.
// Its dedicated lowering checks same-kind storage and never passes an object shape.
func (l *lowering) typedArraySetArgument(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "set" {
		return false
	}
	return l.typedArrayKind(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))) != 0
}
