package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

func nonNullAssignmentTarget(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNonNullExpression {
		return node
	}
	return nil
}

// Hold the receiver and index before reading, and hold the checked read before the
// RHS runs. The final store uses those same held values even when a call mutates them.
func (l *lowering) nonNullUpdate(node, assertion, target *ast.Node, operator ast.Kind) ([]ir.Statement, error) {
	body := []ir.Statement{}
	hold := func(name string, value ir.Expression, where *ast.Node) ir.Read {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: value.Type(), Function: l.functionIndex})
		if where != nil {
			l.noteLocal(local, l.concrete(l.checker.GetTypeAtLocation(where)), where)
		}
		body = append(body, ir.Declare{Local: local, Value: value})
		return ir.Read{Local: local, Of: value.Type()}
	}
	var current ir.Expression
	var store func(ir.Expression) ir.Statement
	switch target.Kind {
	case ast.KindIdentifier:
		local, known := l.local(target)
		if !known || l.caught[l.symbol(target)] || l.alwaysUndefined[l.symbol(target)] {
			return nil, l.notYet(target, "updating this asserted variable")
		}
		var err error
		current, err = l.expression(target)
		if err != nil {
			return nil, err
		}
		store = func(value ir.Expression) ir.Statement {
			return ir.Assign{Local: local, Value: fit(value, l.result.Locals[local].Type), Checked: l.checked(local)}
		}
	case ast.KindPropertyAccessExpression:
		receiver := target.AsPropertyAccessExpression().Expression
		if ast.SkipParentheses(receiver).Kind == ast.KindSuperKeyword {
			return nil, l.notYet(target, "updating an asserted super property")
		}
		object, err := l.expression(receiver)
		if err != nil {
			return nil, err
		}
		if object.Type() != ir.Object {
			return nil, l.notYet(target, "updating an asserted field of a "+typeName(object.Type()))
		}
		object = l.privateStaticReceiver(target.Name(), object, false)
		held := hold("assertion_object", object, receiver)
		of, err := l.typeOf(target)
		if field := l.checker.GetSymbolAtLocation(target.Name()); field != nil {
			of, err = l.typeOfSymbol(target, field)
		}
		if err != nil || slotless(of) {
			return nil, l.notYet(target, "updating an asserted field without a stored representation")
		}
		name := l.fieldName(target.Name())
		current = l.readObjectField(target, ir.Property{Object: held, Name: name, Of: of, Class: l.classOf(target)})
		store = func(value ir.Expression) ir.Statement {
			return ir.SetProperty{Object: held, Name: name, Value: fit(value, of), Class: l.classOf(target), Site: l.writeSite(receiver)}
		}
	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		array, err := l.expression(access.Expression)
		if err != nil {
			return nil, err
		}
		if array.Type() != ir.Array && !array.Type().IsTypedArray() {
			return nil, l.notYet(target, "updating an asserted index of a "+typeName(array.Type()))
		}
		element := ir.Number
		if array.Type() == ir.Array {
			element, err = l.elementType(access.Expression)
			if err != nil {
				return nil, err
			}
		}
		heldArray := hold("assertion_array", array, access.Expression)
		index, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		if index.Type() != ir.Number {
			return nil, l.notYet(target, "an asserted array index that isn't a number")
		}
		heldIndex := hold("assertion_index", index, nil)
		current = ir.ArrayIndex{Array: heldArray, Index: heldIndex, Element: element}
		store = func(value ir.Expression) ir.Statement {
			return ir.SetIndex{Array: heldArray, Index: heldIndex, Value: fit(value, element), Element: element, Site: l.writeSite(access.Expression)}
		}
	default:
		return nil, l.notYet(target, "updating an asserted "+describe(target))
	}
	checked, err := l.nonNullValue(assertion, current)
	if err != nil {
		return nil, err
	}
	old := hold("assertion_value", checked, nil)
	right, err := l.expression(node.AsBinaryExpression().Right)
	if err != nil {
		return nil, err
	}
	if logicalAssignment(node.AsBinaryExpression().OperatorToken.Kind) {
		condition := ir.Expression(ir.BooleanConstant{Value: false})
		if node.AsBinaryExpression().OperatorToken.Kind != ast.KindQuestionQuestionEqualsToken {
			condition = l.assignmentTruthy(old)
			if node.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarEqualsToken {
				condition = ir.Unary{Operator: ir.Not, Operand: condition}
			}
		}
		body = append(body, ir.If{Condition: condition, Then: []ir.Statement{store(right)}})
	} else {
		left := ir.Expression(old)
		if operator == ast.KindPlusToken {
			left, right = l.spelled(assertion, left), l.spelled(node.AsBinaryExpression().Right, right)
		}
		updated, err := l.combine(node, operator, left, right)
		if err != nil {
			return nil, err
		}
		body = append(body, store(updated))
	}
	return []ir.Statement{ir.Block{Body: body}}, nil
}

// Logical assignment has JavaScript's short-circuit value semantics, including NaN.
func (l *lowering) assignmentTruthy(value ir.Expression) ir.Expression {
	switch value.Type() {
	case ir.Boolean:
		return value
	case ir.Number:
		return ir.Binary{Operator: ir.And, Left: ir.Binary{Operator: ir.NotEqual, Left: value, Right: ir.NumberConstant{}}, Right: ir.Binary{Operator: ir.Equal, Left: value, Right: value}}
	case ir.String:
		return ir.Binary{Operator: ir.NotEqual, Left: ir.StringLength{Value: value}, Right: ir.NumberConstant{}}
	case ir.Union:
		result := ir.Expression(ir.BooleanConstant{Value: true})
		for _, member := range []struct {
			name string
			of   ir.Type
		}{{"string", ir.String}, {"boolean", ir.Boolean}, {"number", ir.Number}} {
			condition := ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: value}, Right: ir.StringConstant{Index: l.constant(member.name)}}
			result = ir.Conditional{Condition: condition, WhenTrue: l.assignmentTruthy(ir.Narrow{Value: value, To: member.of}), WhenNot: result}
		}
		return result
	}
	return ir.BooleanConstant{Value: true}
}
