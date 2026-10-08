package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A checked compound target is read before the right operand runs. Hold its
// receiver and index independently, then check the held read and finally store.
func (l *lowering) assignmentNonNullUpdate(node, assertion *ast.Node, operator ast.Kind, rightNode *ast.Node) ([]ir.Statement, error) {
	target := ast.SkipParentheses(assertion.AsNonNullExpression().Expression)
	body := []ir.Statement{}
	hold := func(name string, value ir.Expression) ir.Read {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: value.Type(), Function: l.functionIndex})
		body = append(body, ir.Declare{Local: local, Value: value})
		return ir.Read{Local: local, Of: value.Type()}
	}
	var current ir.Expression
	var store ir.Statement
	switch target.Kind {
	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		array, err := l.expression(access.Expression)
		if err != nil {
			return nil, err
		}
		if array.Type() != ir.Array && !array.Type().IsTypedArray() {
			return nil, l.notYet(target, "a checked update element of a "+typeName(array.Type()))
		}
		element := ir.Number
		if array.Type() == ir.Array {
			element, err = l.elementType(access.Expression)
			if err != nil {
				return nil, err
			}
		}
		if element != ir.Number && element != ir.String {
			return nil, l.notYet(target, "a checked update of a "+typeName(element)+" element")
		}
		receiver := hold("update_array", array)
		index, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		if index.Type() != ir.Number {
			return nil, l.notYet(target, "a checked update index other than number")
		}
		position := hold("update_index", index)
		current = ir.ArrayIndex{Array: receiver, Index: position, Element: element}
		store = ir.SetIndex{Array: receiver, Index: position, Element: element, Site: l.writeSite(access.Expression)}
	case ast.KindPropertyAccessExpression:
		access := target.AsPropertyAccessExpression()
		if ast.SkipParentheses(access.Expression).Kind == ast.KindSuperKeyword || target.Name().Kind == ast.KindPrivateIdentifier {
			return nil, l.notYet(target, "a checked update through a private or super field")
		}
		object, err := l.expression(access.Expression)
		if err != nil {
			return nil, err
		}
		if object.Type() != ir.Object {
			return nil, l.notYet(target, "a checked update field of a "+typeName(object.Type()))
		}
		field := l.checker.GetSymbolAtLocation(target.Name())
		if field == nil {
			return nil, l.notYet(target, "a checked update field without a declared slot")
		}
		if len(field.Declarations) > 0 && ast.HasSyntacticModifier(field.Declarations[0], ast.ModifierFlagsStatic) {
			return nil, l.notYet(target, "a checked update of a static field")
		}
		of, err := l.typeOfSymbol(target, field)
		if err != nil {
			return nil, err
		}
		if of != ir.Number && of != ir.MaybeNumber && of != ir.String {
			return nil, l.notYet(target, "a checked update field of a "+typeName(of))
		}
		receiver := hold("update_object", object)
		name := l.fieldName(target.Name())
		current = ir.Property{Object: receiver, Name: name, Of: of, Class: l.classOf(target)}
		store = ir.SetProperty{Object: receiver, Name: name, Class: l.classOf(target), Site: l.writeSite(access.Expression)}
	default:
		return nil, l.notYet(target, "a checked compound assignment target other than a field or element")
	}
	checked, err := l.nonNullValue(assertion, current)
	if err != nil {
		return nil, err
	}
	if checked.Type() != ir.Number && checked.Type() != ir.String {
		return nil, l.notYet(assertion, "a checked update result of a "+typeName(checked.Type()))
	}
	old := hold("update_current", checked)
	right, err := l.expression(rightNode)
	if err != nil {
		return nil, err
	}
	if operator == ast.KindPlusToken {
		right = l.spelled(rightNode, right)
	}
	updated, err := l.combine(node, operator, old, right)
	if err != nil {
		return nil, err
	}
	switch write := store.(type) {
	case ir.SetIndex:
		write.Value = fit(updated, write.Element)
		store = write
	case ir.SetProperty:
		write.Value = fit(updated, current.Type())
		store = write
	}
	body = append(body, store)
	return []ir.Statement{ir.Block{Body: body}}, nil
}
