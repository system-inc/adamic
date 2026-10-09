package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func logicalAssignment(kind ast.Kind) bool {
	return kind == ast.KindBarBarEqualsToken || kind == ast.KindAmpersandAmpersandEqualsToken || kind == ast.KindQuestionQuestionEqualsToken
}

func (l *lowering) logicalAssignment(node *ast.Node) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	target := ast.SkipParentheses(binary.Left)
	var right ir.Expression
	result, err := l.typeOf(node)
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
		right, err = l.expression(binary.Right)
		if err != nil {
			return nil, err
		}
		local, known := l.local(target)
		if !known || l.caught[l.symbol(target)] || l.alwaysUndefined[l.symbol(target)] {
			return nil, l.notYet(target, "a logical assignment to this binding")
		}
		of := l.result.Locals[local].Type
		l.result.Locals[local].ExpressionAssigned = true
		assignment.Read = ir.Read{Local: local, Of: of, Checked: l.checked(local)}
		assignment.Write = ir.Assign{Local: local, Value: fit(right, of), Checked: l.checked(local)}
	case ast.KindPropertyAccessExpression:
		write, err := l.setProperty(target, binary.Right)
		if err != nil {
			return nil, err
		}
		store, known := write[0].(ir.SetProperty)
		if !known {
			return nil, l.notYet(target, "a logical assignment to an accessor")
		}
		field := l.checker.GetSymbolAtLocation(target.Name())
		if field != nil && accessorSymbol(field) {
			return nil, l.notYet(target, "a logical assignment to an accessor")
		}
		assignment.Write = store
		assignment.Read = ir.Property{Object: store.Object, Name: store.Name, Of: store.Value.Type(), Class: store.Class, Absent: field != nil && field.Flags&ast.SymbolFlagsOptional != 0}
	case ast.KindElementAccessExpression:
		right, err = l.expression(binary.Right)
		if err != nil {
			return nil, err
		}
		access := target.AsElementAccessExpression()
		object, err := l.expression(access.Expression)
		if err != nil {
			return nil, err
		}
		key, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		if object.Type() == ir.Array {
			element, err := l.elementType(access.Expression)
			if err != nil {
				return nil, err
			}
			if key.Type() != ir.Number {
				return nil, l.notYet(target, "an array index that isn't a number")
			}
			assignment.Read = ir.ArrayIndex{Array: object, Index: key, Element: element}
			assignment.Write = ir.SetIndex{Array: object, Index: key, Value: fit(right, element), Element: element, Site: l.writeSite(access.Expression)}
		} else if object.Type() == ir.Object && key.Type() == ir.String {
			of, err := l.typeOf(target)
			if err != nil || slotless(of) {
				return nil, l.notYet(target, "a logical assignment to this computed field")
			}
			names := l.concrete(l.checker.GetTypeAtLocation(access.ArgumentExpression))
			members := []*checker.Type{names}
			if names.Flags()&checker.TypeFlagsUnion != 0 {
				members = names.Types()
			}
			for _, name := range members {
				if name.Flags()&checker.TypeFlagsStringLiteral == 0 {
					return nil, l.notYet(target, "a computed logical assignment key without a finite set of names")
				}
				text, known := name.AsLiteralType().Value().(string)
				if !known {
					return nil, l.notYet(target, "a computed logical assignment key")
				}
				field := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(access.Expression), text)
				if field != nil && field.Flags&ast.SymbolFlagsOptional != 0 {
					return nil, l.notYet(target, "writing a possibly absent optional own field")
				}
				if field != nil && accessorSymbol(field) {
					return nil, l.notYet(target, "a logical assignment to an accessor")
				}
				if l.result.CheckedWrites[text] {
					return nil, l.notYet(target, "a checked write through a computed logical-assignment key")
				}
				assignment.KeyNames = append(assignment.KeyNames, text)
			}
			assignment.Key = key
			assignment.Read = ir.Property{Object: object, Of: of, Absent: true}
			assignment.Write = ir.SetProperty{Object: object, Value: fit(right, of), Site: l.writeSite(access.Expression)}
		} else {
			return nil, l.notYet(target, "a logical assignment to this member")
		}
	default:
		return nil, l.notYet(target, "a logical assignment to this target")
	}
	return assignment, nil
}
