package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func logicalAssignment(operator ast.Kind) bool {
	return operator == ast.KindBarBarEqualsToken || operator == ast.KindAmpersandAmpersandEqualsToken || operator == ast.KindQuestionQuestionEqualsToken
}

// expressionScope puts an expression's statements in an immediately called closure. Captures
// share the original bindings, while ordinary IR exposes every conditional write to the analyses.
func (l *lowering) expressionScope(name string, build func(*libraryArrayBuilder) (ir.Expression, error)) (ir.Expression, error) {
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: name, Closure: true})
	outerFunction, outerIndex := l.function, l.functionIndex
	function := l.result.Functions[index]
	l.function, l.functionIndex = &function, index
	l.closures = append(l.closures, index)
	defer func() {
		l.function, l.functionIndex = outerFunction, outerIndex
		l.closures = l.closures[:len(l.closures)-1]
	}()
	b := &libraryArrayBuilder{l: l, function: index}
	value, err := build(b)
	if err != nil {
		return nil, err
	}
	l.result.Functions[index].Body = append(b.body, ir.Return{Value: value})
	l.result.Functions[index].Returns = value.Type()
	return ir.CallClosure{Closure: ir.MakeClosure{Function: index}, Returns: value.Type()}, nil
}

// assignmentReference snapshots the base and index before reading, even when either is a
// variable a right-hand call can replace. Its store always writes that original reference.
func (l *lowering) assignmentReference(b *libraryArrayBuilder, target *ast.Node) (ir.Expression, ir.Type, func(ir.Expression) ir.Statement, error) {
	target = ast.SkipParentheses(target)
	if ast.IsIdentifier(target) {
		local, known := l.local(target)
		if !known || l.caught[l.symbol(target)] || l.alwaysUndefined[l.symbol(target)] {
			return nil, 0, nil, l.notYet(target, "assigning to "+describe(target))
		}
		of := l.result.Locals[local].Type
		read := ir.Read{Local: local, Of: of, Checked: l.checkedModuleRead(target, local)}
		return read, of, func(value ir.Expression) ir.Statement {
			return ir.Assign{Local: local, Value: fit(value, of), Checked: l.checked(local)}
		}, nil
	}
	if target.Kind == ast.KindPropertyAccessExpression {
		access := target.AsPropertyAccessExpression()
		object, err := l.expression(access.Expression)
		if err != nil {
			return nil, 0, nil, err
		}
		if object.Type() != ir.Object {
			return nil, 0, nil, l.notYet(target, "assigning a field of a "+typeName(object.Type()))
		}
		object = l.privateStaticReceiver(access.Name(), object, false)
		held := b.declare("assignment_object", object)
		l.noteLocal(held, l.checker.GetTypeAtLocation(access.Expression), target)
		object = b.read(held)
		of, err := l.typeOf(target)
		if symbol := l.checker.GetSymbolAtLocation(access.Name()); symbol != nil {
			of, err = l.typeOfSymbol(target, symbol)
		}
		if err != nil {
			return nil, 0, nil, err
		}
		if slotless(of) {
			return nil, 0, nil, l.notYet(target, "a field of "+typeName(of))
		}
		name, class := l.fieldName(access.Name()), l.classOf(target)
		read := ir.Property{Object: object, Name: name, Of: of, Class: class, Absent: access.Name() != nil && l.checker.GetSymbolAtLocation(access.Name()) != nil && l.checker.GetSymbolAtLocation(access.Name()).Flags&ast.SymbolFlagsOptional != 0}
		return read, of, func(value ir.Expression) ir.Statement {
			return ir.SetProperty{Object: object, Name: name, Value: fit(value, of), Class: class, Site: l.writeSite(access.Expression)}
		}, nil
	}
	if target.Kind == ast.KindElementAccessExpression {
		access := target.AsElementAccessExpression()
		array, err := l.expression(access.Expression)
		if err != nil {
			return nil, 0, nil, err
		}
		if array.Type() != ir.Array {
			return nil, 0, nil, l.notYet(target, "assigning an element of a "+typeName(array.Type()))
		}
		element, err := l.elementType(access.Expression)
		if err != nil {
			return nil, 0, nil, err
		}
		held := b.declare("assignment_array", array)
		l.noteLocal(held, l.checker.GetTypeAtLocation(access.Expression), target)
		index, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, 0, nil, err
		}
		if index.Type() != ir.Number {
			return nil, 0, nil, l.notYet(target, "an array index that isn't a number")
		}
		at := b.declare("assignment_index", index)
		array, index = b.read(held), b.read(at)
		read := ir.ArrayIndex{Array: array, Index: index, Element: element}
		return read, element, func(value ir.Expression) ir.Statement {
			return ir.SetIndex{Array: array, Index: index, Value: fit(value, element), Element: element, Site: l.writeSite(access.Expression)}
		}, nil
	}
	return nil, 0, nil, l.notYet(target, "assigning to "+describe(target))
}

func (l *lowering) logicalAssignment(node *ast.Node) (ir.Expression, error) {
	name := "logical_or_assignment"
	if node.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandEqualsToken {
		name = "logical_and_assignment"
	}
	if node.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken {
		name = "logical_nullish_assignment"
	}
	return l.expressionScope(name, func(b *libraryArrayBuilder) (ir.Expression, error) {
		binary := node.AsBinaryExpression()
		current, _, store, err := l.assignmentReference(b, binary.Left)
		if err != nil {
			return nil, err
		}
		if current.Type() == ir.Weak {
			declared := l.assignmentDeclaredType(binary.Left)
			members := declared.Types()
			if declared.Flags()&checker.TypeFlagsUnion == 0 {
				members = []*checker.Type{declared}
			}
			var target ir.Type
			for _, member := range members {
				if weak := l.weakTarget(member); weak != nil {
					target, _ = l.representation(weak)
				}
			}
			if !target.IsReference() {
				return nil, l.notYet(binary.Left, "a logical assignment of an unknown Weak target")
			}
			current = ir.WeakTarget{Value: current, To: target, Present: false}
		}
		held := b.declare("assignment_current", current)
		current = b.read(held)
		right, err := l.expression(binary.Right)
		if err != nil {
			return nil, err
		}
		resultType, err := l.typeOf(node)
		if err != nil {
			return nil, err
		}
		if resultType == ir.Weak {
			resultType = current.Type()
		}
		result := b.local("assignment_right", right.Type())
		taken := []ir.Statement{ir.Declare{Local: result, Value: right}, store(b.read(result)), ir.Return{Value: assignmentResult(b.read(result), resultType)}}
		var take ir.Expression
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarEqualsToken:
			take = ir.Unary{Operator: ir.Not, Operand: l.assignmentTruthy(current)}
		case ast.KindAmpersandAmpersandEqualsToken:
			take = l.assignmentTruthy(current)
		case ast.KindQuestionQuestionEqualsToken:
			take = ir.BooleanConstant{Value: false}
			if current.Type().IsMaybe() || current.Type().IsReference() {
				take = ir.IsUndefined{Value: current}
				declared := l.assignmentDeclaredType(binary.Left)
				if declared != nil && declared.Flags()&checker.TypeFlagsNever == 0 && l.includesNull(l.concrete(declared)) {
					take = ir.Binary{Operator: ir.Or, Left: take, Right: ir.IsNull{Value: current}}
				}
			}
		}
		b.body = append(b.body, ir.If{Condition: take, Then: taken})
		return assignmentResult(current, resultType), nil
	})
}

func assignmentResult(value ir.Expression, to ir.Type) ir.Expression {
	if value.Type() == ir.Union && to != ir.Union {
		return ir.Narrow{Value: value, To: to}
	}
	return fit(value, to)
}

// ECMAScript ToBoolean, restricted only by the representations the checker can give a target.
// The current value is a saved local, so these tests never repeat a source read or a getter.
func (l *lowering) assignmentTruthy(value ir.Expression) ir.Expression {
	not := func(v ir.Expression) ir.Expression { return ir.Unary{Operator: ir.Not, Operand: v} }
	and := func(a, b ir.Expression) ir.Expression { return ir.Binary{Operator: ir.And, Left: a, Right: b} }
	switch value.Type() {
	case ir.Boolean:
		return value
	case ir.Number:
		return and(ir.Binary{Operator: ir.NotEqual, Left: value, Right: ir.NumberConstant{Value: 0}}, ir.Binary{Operator: ir.Equal, Left: value, Right: value})
	case ir.MaybeNumber, ir.MaybeBoolean:
		return ir.Conditional{Condition: ir.IsUndefined{Value: value}, WhenTrue: ir.BooleanConstant{Value: false}, WhenNot: l.assignmentTruthy(ir.Unwrap{Value: value})}
	case ir.String:
		return and(not(ir.IsUndefined{Value: value}), ir.Binary{Operator: ir.NotEqual, Left: ir.StringLength{Value: value}, Right: ir.NumberConstant{Value: 0}})
	case ir.Union:
		answer := ir.Expression(and(not(ir.IsUndefined{Value: value}), not(ir.IsNull{Value: value})))
		for _, member := range []struct {
			name string
			of   ir.Type
		}{{"number", ir.Number}, {"boolean", ir.Boolean}, {"string", ir.String}} {
			answer = ir.Conditional{Condition: ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: value}, Right: ir.StringConstant{Index: l.constant(member.name)}}, WhenTrue: l.assignmentTruthy(ir.Narrow{Value: value, To: member.of}), WhenNot: answer}
		}
		return answer
	default:
		return and(not(ir.IsUndefined{Value: value}), not(ir.IsNull{Value: value}))
	}
}

func (l *lowering) assignmentDeclaredType(node *ast.Node) *checker.Type {
	target := ast.SkipParentheses(node)
	declared := l.checker.GetTypeAtLocation(target)
	if symbol := l.symbol(target); symbol != nil {
		declared = l.checker.GetTypeOfSymbol(symbol)
	}
	if target.Kind == ast.KindElementAccessExpression {
		holder := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(target.AsElementAccessExpression().Expression))
		if l.checker.IsArrayType(holder) {
			declared = l.checker.GetElementTypeOfArrayType(holder)
		}
	}
	return l.concrete(declared)
}
