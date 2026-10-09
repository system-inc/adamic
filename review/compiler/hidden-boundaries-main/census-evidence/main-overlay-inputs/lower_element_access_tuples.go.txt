package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A fixed tuple is already held as named slots. Numeric dispatch preserves that
// layout, and a shorter member of a tuple union has no slot past its own end.
func (l *lowering) indexedTupleRead(node *ast.Node, object ir.Expression) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
	members := []*checker.Type{receiver}
	if receiver.Flags()&checker.TypeFlagsUnion != 0 {
		members = receiver.Types()
	}
	for _, member := range members {
		if !checker.IsTupleType(member) {
			return nil, false, nil
		}
	}
	if access.QuestionDotToken != nil {
		return nil, false, nil
	}
	if checker.IsTupleType(receiver) && ast.SkipParentheses(access.ArgumentExpression).Kind == ast.KindNumericLiteral {
		return nil, false, nil
	}
	stored := []ir.Type{}
	for _, member := range members {
		if checker.TupleType_combinedFlags(member.TargetTupleType())&checker.ElementFlagsVariable != 0 {
			return nil, true, l.notYet(node, "numeric indexing of a tuple with variable elements")
		}
		elements := l.checker.GetTypeArguments(member)
		for index, element := range elements {
			of, known := l.representation(element)
			if !known || censusFieldSlotless(of) || of == ir.Weak || l.includesNull(element) {
				return nil, true, l.notYet(node, "numeric indexing of an unrepresented tuple element")
			}
			if index == len(stored) {
				stored = append(stored, of)
			} else if stored[index] != of {
				return nil, true, l.notYet(node, "a tuple union storing an index in different representations")
			}
		}
	}
	result, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	// The unmatched numeric keys (including negative, fractional and NaN) read
	// undefined. Do not manufacture a value when the result cannot hold it.
	missing := fit(ir.Undefined{}, result)
	if missing.Type() != result || (!result.IsReference() && !result.IsMaybe()) {
		return nil, true, l.notYet(node, "a dynamic tuple read without an undefined representation")
	}
	key, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.Number {
		return nil, true, l.notYet(node, "a tuple index that is not a number")
	}
	b := l.libraryArrayBuilder([]ir.Expression{object, key})
	heldObject, heldKey := b.read(b.parameters[0]), b.read(b.parameters[1])
	for index, of := range stored {
		absent := false
		for _, member := range members {
			if index >= len(l.checker.GetTypeArguments(member)) {
				absent = true
			}
		}
		if absent {
			of = ir.Maybe(of)
		}
		read := fit(ir.Property{Object: heldObject, Name: strconv.Itoa(index), Of: of, Absent: absent}, result)
		if read.Type() != result {
			return nil, true, l.notYet(node, "a dynamic tuple result with an incompatible element representation")
		}
		b.body = append(b.body, ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: heldKey, Right: ir.NumberConstant{Value: float64(index)}}, Then: []ir.Statement{ir.Return{Value: read}}})
	}
	return b.finish("element_access_tuple", missing), true, nil
}
