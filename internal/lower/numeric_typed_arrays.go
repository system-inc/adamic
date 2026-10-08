package lower

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

var numericTypedArrays = []string{"Int8Array", "Uint8Array", "Uint8ClampedArray", "Int16Array", "Uint16Array", "Int32Array", "Uint32Array", "Float32Array", "Float64Array"}

func (l *lowering) numericTypedArray(proven *checker.Type) bool {
	if proven == nil || l.typedArrayKind(l.checker.GetNonNullableType(proven)) != 0 {
		return false
	}
	proven = l.concrete(l.checker.GetNonNullableType(proven))
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if l.numericTypedArray(member) {
				return true
			}
		}
		return false
	}
	return l.isLibraryType(proven, numericTypedArrays...)
}

func (l *lowering) newTypedArray(node *ast.Node) (ir.Expression, bool, error) {
	made := node.AsNewExpression()
	name := ""
	for _, candidate := range numericTypedArrays {
		if l.isLibraryGlobal(made.Expression, candidate) {
			name = candidate
			break
		}
	}
	if name == "" {
		return nil, false, nil
	}
	length := ir.Expression(ir.NumberConstant{Value: 0})
	if made.Arguments != nil && len(made.Arguments.Nodes) != 0 {
		if len(made.Arguments.Nodes) != 1 {
			return nil, true, l.notYet(node, "typed array construction from buffers or multiple arguments")
		}
		var err error
		length, err = l.expression(made.Arguments.Nodes[0])
		if err != nil {
			return nil, true, err
		}
	}
	constant, known := length.(ir.NumberConstant)
	if !known && made.Arguments != nil && len(made.Arguments.Nodes) == 1 {
		constant, known = l.provenArrayLength(made.Arguments.Nodes[0], length)
	}
	// This bounded subset is valid on Node and every supported native target.
	// Dynamic lengths need constructor exceptions and ToIndex conversion.
	if !known || constant.Value < 0 || constant.Value > 1048576 || math.Trunc(constant.Value) != constant.Value || math.IsNaN(constant.Value) {
		return nil, true, l.notYet(node, "typed array length not statically proven to be an integer from 0 to 1048576")
	}
	return ir.NumericTypedArrayNew{Name: name, Length: length}, true, nil
}

func (l *lowering) typedArrayCall(node *ast.Node) error {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind == ast.KindPropertyAccessExpression {
		receiver := callee.AsPropertyAccessExpression().Expression
		if l.numericTypedArray(l.checker.GetTypeAtLocation(receiver)) {
			return l.notYet(node, "typed array methods require their buffer, coercion and view semantics")
		}
		if (l.isLibraryGlobal(receiver, "Array") && callee.Name().Text() != "isArray") || l.isLibraryGlobal(receiver, "Map") || l.isLibraryGlobal(receiver, "Set") {
			for _, argument := range call.Arguments.Nodes {
				if l.numericTypedArray(l.checker.GetTypeAtLocation(argument)) {
					return l.notYet(node, "library iteration over typed arrays requires their iterator semantics")
				}
			}
		}
		if l.isLibraryGlobal(receiver, "Object") && callee.Name().Text() != "is" {
			for _, argument := range call.Arguments.Nodes {
				if l.numericTypedArray(l.checker.GetTypeAtLocation(argument)) {
					return l.notYet(node, "Object operations on typed arrays require their own property and buffer semantics")
				}
			}
		}
	}
	return nil
}

func (l *lowering) typedArrayIndex(node *ast.Node, object ir.Expression) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	if access.QuestionDotToken != nil {
		return nil, l.notYet(node, "optional typed array indexing")
	}
	index, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	if index.Type() != ir.Number {
		return nil, l.notYet(node, "a typed array index other than a number")
	}
	return l.defined(node, ir.ArrayIndex{Array: ir.TypedArrayData{Value: object}, Index: index, Element: ir.Number}), nil
}

func (l *lowering) numericTypedArrayUnsupportedUse(node *ast.Node) error {
	if !l.numericTypedArray(l.checker.GetTypeAtLocation(node)) {
		return nil
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent != nil && (parent.Kind == ast.KindSpreadAssignment || parent.Kind == ast.KindSpreadElement) {
		return l.notYet(node, "typed array spread requires their own numeric keys or iterator semantics")
	}
	return nil
}
