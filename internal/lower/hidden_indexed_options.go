package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// finiteObjectElement selects a known data field using an exhaustive literal-key
// domain. Hold the receiver before evaluating the key, then read only the selected
// field: evaluating a key may mutate the object whose field JavaScript reads.
func (l *lowering) finiteObjectElement(node *ast.Node, object ir.Expression) (ir.Expression, bool, error) {
	access := node.AsElementAccessExpression()
	keyType := l.checker.GetTypeAtLocation(access.ArgumentExpression)
	members := []*checker.Type{keyType}
	if keyType.Flags()&checker.TypeFlagsUnion != 0 {
		members = keyType.Types()
	}
	receiverType := l.checker.GetTypeAtLocation(access.Expression)
	names := []string{}
	fields := []*ast.Symbol{}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsStringLiteral == 0 {
			return nil, false, nil
		}
		name, ok := member.AsLiteralType().Value().(string)
		if !ok {
			return nil, false, nil
		}
		switch name {
		case "__proto__", "constructor", "toString", "toLocaleString", "valueOf", "hasOwnProperty", "propertyIsEnumerable", "isPrototypeOf", "__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__":
			return nil, false, nil
		}
		field := l.checker.GetPropertyOfType(receiverType, name)
		if field == nil || len(field.Declarations) == 0 {
			return nil, false, nil
		}
		for _, declaration := range field.Declarations {
			// Methods, accessors and class properties need their own dispatch path.
			if declaration.Kind != ast.KindPropertySignature && declaration.Kind != ast.KindPropertyAssignment && declaration.Kind != ast.KindShorthandPropertyAssignment {
				return nil, false, nil
			}
		}
		names = append(names, name)
		fields = append(fields, field)
	}
	if len(names) == 0 {
		return nil, false, nil
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	// This boundary is a boolean-option read. Keep other value families on
	// their existing lowering path until their own representation oracles exist.
	if of != ir.Boolean && of != ir.MaybeBoolean {
		return nil, false, nil
	}
	// Keep heterogeneous representation joins on their existing lowering path.
	for _, field := range fields {
		stored, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || censusFieldSlotless(stored) || (stored != of && !(stored.IsMaybe() && stored.Present() == of) && !(of.IsMaybe() && of.Present() == stored)) {
			return nil, false, nil
		}
	}
	key, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.String {
		return nil, false, nil
	}
	b := l.libraryArrayBuilder([]ir.Expression{object, key})
	readObject := b.read(b.parameters[0])
	readKey := b.read(b.parameters[1])
	var selected ir.Expression
	for i := len(names) - 1; i >= 0; i-- {
		declared := l.checker.GetTypeOfSymbol(fields[i])
		stored, _ := l.representation(declared)
		property := ir.Property{Object: readObject, Name: names[i], Of: stored,
			Absent: fields[i].Flags&ast.SymbolFlagsOptional != 0,
			View:   sourceExpression(node), Readiness: sourceExpression(node),
			ViewType: l.checker.TypeToString(declared), ViewAllowed: l.viewLiterals(declared),
		}
		value := fit(property, of)
		if selected == nil {
			selected = value
			continue
		}
		selected = ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: readKey, Right: ir.StringConstant{Index: l.constant(names[i])}}, WhenTrue: value, WhenNot: selected, Of: of}
	}
	return b.finish("indexed_data_field", selected), true, nil
}
