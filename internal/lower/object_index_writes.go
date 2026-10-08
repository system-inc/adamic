package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A finite string key dispatches to the same fixed slots as named writes.
// Capture receiver, key and right side before dispatching, in source order.
func (l *lowering) setObjectIndex(target, valueNode *ast.Node, object ir.Expression) ([]ir.Statement, error) {
	access := target.AsElementAccessExpression()
	keyType := l.checker.GetTypeAtLocation(access.ArgumentExpression)
	members := []*checker.Type{keyType}
	if keyType.Flags()&checker.TypeFlagsUnion != 0 {
		members = keyType.Types()
	}
	names := []string{}
	types := []ir.Type{}
	receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsStringLiteral == 0 {
			return nil, l.notYet(target, "assigning an object element with a key that is not a finite string union")
		}
		name := member.AsLiteralType().Value().(string)
		field := l.checker.GetPropertyOfType(receiver, name)
		if field == nil {
			return nil, l.notYet(target, "assigning an object element outside its named fields")
		}
		if field.Flags&ast.SymbolFlagsOptional != 0 {
			return nil, l.notYet(target, "assigning an optional object field through a computed key requires growing its own shape")
		}
		if accessorSymbol(field) || name == "__proto__" {
			return nil, l.notYet(target, "assigning a computed object accessor or prototype field")
		}
		of, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || censusFieldSlotless(of) {
			return nil, l.notYet(target, "assigning a computed field with an unsupported representation")
		}
		names = append(names, name)
		types = append(types, of)
	}
	key, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	b := l.libraryArrayBuilder([]ir.Expression{object, key, value})
	heldObject, heldKey, heldValue := b.read(b.parameters[0]), b.read(b.parameters[1]), b.read(b.parameters[2])
	otherwise := []ir.Statement{ir.Panic{Message: ir.StringConstant{Index: l.constant("computed object key outside its proven finite union")}}}
	for index := len(names) - 1; index >= 0; index-- {
		converted := fit(heldValue, types[index])
		if converted.Type() != types[index] {
			return nil, l.notYet(target, "assigning a computed field with a differently held value")
		}
		otherwise = []ir.Statement{ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: heldKey, Right: ir.StringConstant{Index: l.constant(names[index])}}, Then: []ir.Statement{ir.SetProperty{Object: heldObject, Name: names[index], Value: converted, Site: l.writeSite(access.Expression)}}, Else: otherwise}}
	}
	b.body = append(b.body, otherwise...)
	return []ir.Statement{ir.Evaluate{Value: b.finish("object_index_write", heldValue)}}, nil
}
