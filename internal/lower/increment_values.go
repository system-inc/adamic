package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// incrementValue uses ordinary IR so ownership, effects and both backends see the write.
// A helper saves the old number before writing. Receiver and index arguments are evaluated
// once, in source order. Local writes use a capture cell rather than a copied argument.
func (l *lowering) incrementValue(node *ast.Node) (ir.Expression, error) {
	var operator ast.Kind
	var target *ast.Node
	postfix := node.Kind == ast.KindPostfixUnaryExpression
	if postfix {
		operator, target = node.AsPostfixUnaryExpression().Operator, node.AsPostfixUnaryExpression().Operand
	} else {
		operator, target = node.AsPrefixUnaryExpression().Operator, node.AsPrefixUnaryExpression().Operand
	}
	target = ast.SkipParentheses(target)
	of, err := l.typeOf(target)
	if err != nil {
		return nil, err
	}
	if of != ir.Number {
		return nil, l.notYet(target, "a value-used increment of a "+typeName(of))
	}
	step := ir.Add
	if operator == ast.KindMinusMinusToken {
		step = ir.Subtract
	}
	var b *libraryArrayBuilder
	var current ir.Expression
	var store func(ir.Expression) ir.Statement
	closure := false
	switch target.Kind {
	case ast.KindIdentifier:
		local, known := l.local(target)
		if !known || l.result.Locals[local].Type != ir.Number {
			return nil, l.notYet(target, "a value-used increment of a non-number slot")
		}
		b = l.libraryArrayBuilder(nil)
		closure = !l.result.Locals[local].Global
		if closure {
			// Reserve the function before touch adds its environment, including propagation through
			// surrounding closures. The cell remains owned by its declaring scope.
			l.result.Functions = append(l.result.Functions, ir.Function{Name: "increment_value", Closure: true, Returns: ir.Number})
			enclosing := l.functionIndex
			l.functionIndex = b.function
			l.closures = append(l.closures, b.function)
			l.touch(local)
			l.closures = l.closures[:len(l.closures)-1]
			l.functionIndex = enclosing
		}
		checked := l.checkedModuleRead(target, local)
		current = ir.Read{Local: local, Of: ir.Number, Checked: checked}
		store = func(value ir.Expression) ir.Statement { return ir.Assign{Local: local, Value: value, Checked: checked} }
	case ast.KindPropertyAccessExpression:
		field := l.checker.GetSymbolAtLocation(target.Name())
		if field != nil && accessorSymbol(field) {
			return nil, l.notYet(target, "a value-used increment of an accessor")
		}
		if field != nil {
			stored, known := l.representation(l.checker.GetTypeOfSymbol(field))
			if !known || stored != ir.Number {
				return nil, l.notYet(target, "a value-used increment of a non-number field slot")
			}
		}
		receiver, err := l.expression(target.AsPropertyAccessExpression().Expression)
		if err != nil {
			return nil, err
		}
		if receiver.Type() != ir.Object {
			return nil, l.notYet(target, "a value-used increment of this field receiver")
		}
		receiver = l.privateStaticReceiver(target.Name(), receiver, false)
		b = l.libraryArrayBuilder([]ir.Expression{receiver})
		object := b.read(b.parameters[0])
		name, class := l.fieldName(target.Name()), l.classOf(target)
		current = ir.Property{Object: object, Name: name, Of: ir.Number, Class: class}
		store = func(value ir.Expression) ir.Statement {
			return ir.SetProperty{Object: object, Name: name, Value: value, Class: class, Site: l.writeSite(target.AsPropertyAccessExpression().Expression)}
		}
	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		array, err := l.expression(access.Expression)
		if err != nil {
			return nil, err
		}
		element, err := l.elementType(access.Expression)
		if err != nil {
			return nil, err
		}
		if array.Type() != ir.Array || element != ir.Number {
			return nil, l.notYet(target, "a value-used increment of a non-number array slot")
		}
		index, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		if index.Type() != ir.Number {
			return nil, l.notYet(target, "a value-used increment with a non-number index")
		}
		b = l.libraryArrayBuilder([]ir.Expression{array, index})
		held, at := b.read(b.parameters[0]), b.read(b.parameters[1])
		current = ir.Unwrap{Value: ir.ArrayIndex{Array: held, Index: at, Element: ir.Number}}
		store = func(value ir.Expression) ir.Statement {
			return ir.SetIndex{Array: held, Index: at, Value: value, Element: ir.Number, Site: l.writeSite(access.Expression)}
		}
	default:
		return nil, l.notYet(target, "a value-used increment of "+describe(target))
	}
	old := b.declare("old", current)
	updated := b.declare("updated", ir.Binary{Operator: step, Left: b.read(old), Right: ir.NumberConstant{Value: 1}})
	b.body = append(b.body, store(b.read(updated)))
	result := b.read(old)
	if !postfix {
		result = b.read(updated)
	}
	if closure {
		b.body = append(b.body, ir.Return{Value: result})
		l.result.Functions[b.function].Body = b.body
		return ir.CallClosure{Closure: ir.MakeClosure{Function: b.function}, Returns: ir.Number}, nil
	}
	return b.finish("increment_value", result), nil
}
