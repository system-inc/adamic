package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// An immediate scalar assertion consumes one owned snapshot of a readonly
// boxed index. Source object members are not asserted to have a scalar shape;
// the scalar cast helper checks the snapshot's runtime kind before conversion.
func (l *lowering) scalarCastIndexOperand(node *ast.Node) (ir.Expression, bool, error) {
	operand := ast.SkipParentheses(node.AsAsExpression().Expression)
	if operand.Kind != ast.KindElementAccessExpression {
		return nil, false, nil
	}
	access := operand.AsElementAccessExpression()
	declared := l.concrete(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression)))
	if !l.checker.IsArrayType(declared) || !l.isLibraryType(declared, "ReadonlyArray") {
		return nil, false, nil
	}
	element := l.concrete(l.checker.GetElementTypeOfArrayType(declared))
	of, known := l.representation(element)
	if !known || of != ir.Union {
		return nil, false, nil
	}
	// Complete flat readonly object alternatives and primitives only. No
	// callable, any, unknown, index signature, nominal or nested object adapter.
	for _, member := range castMembers(element) {
		if !interfaceScalar(member) && !l.interfaceScalarShape(member) {
			return nil, false, nil
		}
	}
	array, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	index, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, true, err
	}
	if array.Type() != ir.Array || index.Type() != ir.Number {
		return nil, false, nil
	}
	if _, err := l.viewContract(node, declared); err != nil {
		return nil, true, err
	}
	// Activate producer metadata needed by the physical snapshot. This is not
	// evidence about the unread object payload or about other array elements.
	l.recordViewCastOrigin(node, array)
	return ir.ArrayIndex{Array: array, Index: index, Element: ir.Union, ScalarCastSnapshot: true, View: sourceExpression(operand), ViewType: l.checker.TypeToString(element)}, true, nil
}
