package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func logicalAssignment(kind ast.Kind) bool {
	return kind == ast.KindBarBarEqualsToken ||
		kind == ast.KindAmpersandAmpersandEqualsToken ||
		kind == ast.KindQuestionQuestionEqualsToken
}

func (lowering *lowering) logicalAssignment(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	target := ast.SkipParentheses(binary.Left)
	var right ir.Expression
	result, err := lowering.typeOf(node)
	if err != nil {
		return nil, err
	}
	assignment := ir.LogicalAssignment{Of: result, Operator: "||="}
	if binary.OperatorToken.Kind == ast.KindAmpersandAmpersandEqualsToken {
		assignment.Operator = "&&="
	}
	if binary.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken {
		assignment.Operator = "??="
	}
	switch target.Kind {
	case ast.KindIdentifier:
		right, err = lowering.expression(binary.Right)
		if err != nil {
			return nil, err
		}
		local, known := lowering.local(target)
		if !known || lowering.caught[lowering.symbol(target)] || lowering.alwaysUndefined[lowering.symbol(target)] {
			return nil, lowering.notYet(target, "a logical assignment to this binding")
		}
		of := lowering.result.Locals[local].Type
		lowering.result.Locals[local].ExpressionAssigned = true
		assignment.Read = ir.Read{Local: local, Of: of, Checked: lowering.checked(local)}
		assignment.Write = ir.Assign{Local: local, Value: fit(right, of), Checked: lowering.checked(local)}
	case ast.KindPropertyAccessExpression:
		write, err := lowering.setProperty(target, binary.Right)
		if err != nil {
			return nil, err
		}
		store, known := write[0].(ir.SetProperty)
		if !known {
			return nil, lowering.notYet(target, "a logical assignment to an accessor")
		}
		// A node array's fields live on its metadata object, and native's logical assignment takes the
		// receiver as an object, not through the array to its metadata.
		if store.Object.Type() == ir.Array {
			return nil, lowering.notYet(target, "a logical assignment to a node array's field")
		}
		field := lowering.checker.GetSymbolAtLocation(target.Name())
		if field != nil && accessorSymbol(field) {
			return nil, lowering.notYet(target, "a logical assignment to an accessor")
		}
		assignment.Write = store
		// Through a checked view, the field is read as a plain read through it is.
		assignment.Read = lowering.viewedField(target, ir.Property{
			Object: store.Object,
			Name:   store.Name,
			Of:     store.Value.Type(),
			Class:  store.Class,
			Absent: field != nil && field.Flags&ast.SymbolFlagsOptional != 0,
		})
	case ast.KindElementAccessExpression:
		right, err = lowering.expression(binary.Right)
		if err != nil {
			return nil, err
		}
		access := target.AsElementAccessExpression()
		object, err := lowering.expression(access.Expression)
		if err != nil {
			return nil, err
		}
		key, err := lowering.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		if object.Type() == ir.Array {
			element, err := lowering.elementType(access.Expression)
			if err != nil {
				return nil, err
			}
			if key.Type() != ir.Number {
				return nil, lowering.notYet(target, "an array index that isn't a number")
			}
			assignment.Read = ir.ArrayIndex{Array: object, Index: key, Element: element}
			assignment.Write = ir.SetIndex{
				Array:   object,
				Index:   key,
				Value:   fit(right, element),
				Element: element,
				Site:    lowering.writeSite(access.Expression),
			}
		} else if object.Type() == ir.Object && key.Type() == ir.String {
			of, err := lowering.typeOf(target)
			if err != nil || slotless(of) {
				return nil, lowering.notYet(target, "a logical assignment to this computed field")
			}
			names := lowering.concrete(lowering.checker.GetTypeAtLocation(access.ArgumentExpression))
			members := []*checker.Type{names}
			if names.Flags()&checker.TypeFlagsUnion != 0 {
				members = names.Types()
			}
			for _, name := range members {
				if name.Flags()&checker.TypeFlagsStringLiteral == 0 {
					return nil, lowering.notYet(
						target,
						"a computed logical assignment key without a finite set of names",
					)
				}
				text, known := name.AsLiteralType().Value().(string)
				if !known {
					return nil, lowering.notYet(target, "a computed logical assignment key")
				}
				field := lowering.checker.GetPropertyOfType(lowering.checker.GetTypeAtLocation(access.Expression), text)
				if field != nil && field.Flags&ast.SymbolFlagsOptional != 0 {
					return nil, lowering.notYet(target, "writing a possibly absent optional own field")
				}
				if field != nil && accessorSymbol(field) {
					return nil, lowering.notYet(target, "a logical assignment to an accessor")
				}
				assignment.KeyNames = append(assignment.KeyNames, text)
			}
			assignment.Key = key
			assignment.Read = ir.Property{Object: object, Of: of, Absent: true}
			assignment.Write = ir.SetProperty{
				Object:             object,
				Value:              fit(right, of),
				Site:               lowering.writeSite(access.Expression),
				ViewReceiverTypeID: lowering.viewWriteReceiverTypeID(target, access.Expression, assignment.KeyNames...),
			}
		} else {
			return nil, lowering.notYet(target, "a logical assignment to this member")
		}
	default:
		return nil, lowering.notYet(target, "a logical assignment to this target")
	}
	return assignment, nil
}
