package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) assignmentValue(node *ast.Node) (ir.Expression, error) {
	l.result.AssignmentValues = true
	stores, err := l.assignment(node)
	if err != nil {
		return nil, err
	}
	if len(stores) != 1 {
		return nil, l.notYet(node, "an assignment value without a single represented store")
	}
	switch store := stores[0].(type) {
	case ir.Assign:
		l.result.Locals[store.Local].ExpressionAssigned = true
	case ir.SetProperty, ir.SetIndex:
	default:
		return nil, l.notYet(node, "an assignment value through an accessor or destructuring target")
	}
	assignment := ir.AssignmentValue{Store: stores[0]}
	var stored ir.Expression
	switch store := stores[0].(type) {
	case ir.Assign:
		stored = store.Value
	case ir.SetProperty:
		stored = store.Value
	case ir.SetIndex:
		stored = store.Value
	}
	of, err := l.typeOf(node)
	if err != nil {
		if assignmentStoresUndefined(stored) {
			// This is the existing Undefined constant, whose IR Type is Object
			// and whose native value is NULL, not an allocated object.
			of = (ir.Undefined{}).Type()
		} else {
			return nil, err
		}
	}
	for stored.Type() != of {
		switch fitted := stored.(type) {
		case ir.Box:
			stored = fitted.Value
		case ir.MaybeOf:
			if fitted.Value == nil {
				stored = ir.Undefined{Of: of}
				goto projected
			}
			stored = fitted.Value
		case ir.Undefined:
			stored = ir.Undefined{Of: of}
		case ir.Null:
			stored = ir.Null{Of: of}
		case ir.WeakOf:
			stored = fitted.Value
		default:
			goto projected
		}
		if stored == nil {
			break
		}
	}
projected:
	if stored != nil {
		assignment.Value = stored
	}
	value := ir.Expression(assignment)
	if value.Type() == of {
		return value, nil
	}
	if value.Type() == ir.Union {
		return ir.Narrow{Value: value, To: of}, nil
	}
	value = fit(value, of)
	if value.Type() != of {
		return nil, l.notYet(node, "an assignment result requiring another representation")
	}
	return value, nil
}

// Admit only an existing constant-undefined store. A checker type alone does not
// justify manufacturing an object representation for an arbitrary expression.
func assignmentStoresUndefined(value ir.Expression) bool {
	switch value := value.(type) {
	case ir.Undefined:
		return true
	case ir.MaybeOf:
		return value.Value == nil
	case ir.Box:
		return assignmentStoresUndefined(value.Value)
	case ir.WeakOf:
		return assignmentStoresUndefined(value.Value)
	}
	return false
}
