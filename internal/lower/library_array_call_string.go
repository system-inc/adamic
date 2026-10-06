package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Port the immutable string branches of V8 GenericArrayPush/GenericArrayPop
// (src/builtins/builtins-array.cc), GenericArrayShift (array-shift.tq), and
// GenericArrayUnshift (array-unshift.tq). The fresh string wrapper cannot escape;
// all numeric characters are present, nonconfigurable and readonly. Temporary
// writes beyond them have no observers, and the first readonly operation throws.
func (l *lowering) libraryArrayStringMutation(node *ast.Node, name string, b *libraryArrayBuilder, count int) ir.Expression {
	lengthError := b.typeError(node, "Cannot assign to read only property 'length' of object '[object String]'")
	if name == "push" || name == "unshift" && count == 0 {
		b.body = append(b.body, lengthError...)
		return b.finish("array_string_"+name, ir.Undefined{})
	}
	length := b.declare("length", ir.StringLength{Value: b.read(b.parameters[0])})
	zero, one := ir.NumberConstant{}, ir.NumberConstant{Value: 1}
	indexedError := func(index ir.Expression, deleting bool) []ir.Statement {
		prefix, suffix := "Cannot assign to read only property '", "' of object '[object String]'"
		if deleting {
			prefix, suffix = "Cannot delete property '", "' of [object String]"
		}
		return b.typeErrorMessage(node, ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant(prefix)}, ir.NumberToString{Value: index}, ir.StringConstant{Index: l.constant(suffix)}}})
	}
	var present []ir.Statement
	switch name {
	case "pop":
		present = indexedError(ir.Binary{Operator: ir.Subtract, Left: b.read(length), Right: one}, true)
	case "shift":
		present = []ir.Statement{ir.If{Condition: ir.Binary{Operator: ir.Greater, Left: b.read(length), Right: one}, Then: indexedError(zero, false), Else: indexedError(zero, true)}}
	case "unshift":
		index := ir.Conditional{Condition: ir.Binary{Operator: ir.Greater, Left: b.read(length), Right: ir.NumberConstant{Value: float64(count)}}, WhenTrue: ir.Binary{Operator: ir.Subtract, Left: b.read(length), Right: one}, WhenNot: zero}
		present = indexedError(index, false)
	}
	b.body = append(b.body, ir.If{Condition: ir.Binary{Operator: ir.Greater, Left: b.read(length), Right: zero}, Then: present, Else: lengthError})
	return b.finish("array_string_"+name, ir.Undefined{})
}
