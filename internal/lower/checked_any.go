package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Any travels with unknown's tags. A contextual type is a use's requirement,
// never evidence about the value. Checks use ordinary IR in both backends.
func (l *lowering) checkedAnyContext(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	// Intrinsics and already checked arithmetic may have a precise IR result
	// despite an erased any in the library declaration. Their own lowering proves
	// that representation; only a tagged dynamic value needs this boundary.
	if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsAny == 0 {
		return value, nil
	}
	target := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	if target == nil || target.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return value, nil
	}
	if value.Type() == ir.Object || value.Type() == ir.Array || value.Type() == ir.Map || value.Type() == ir.Closure {
		if to, known := l.representation(target); known && to == value.Type() {
			return value, nil
		}
	}
	return l.checkedAnyType(node, value, target)
}

func (l *lowering) checkedAnyType(node *ast.Node, value ir.Expression, target *checker.Type) (ir.Expression, error) {
	target = l.concrete(target)
	// A tag alone cannot establish literal values, brands, callable signatures,
	// object fields or the invariants of mutable aliases.
	var to ir.Type
	switch {
	case target.Flags() == checker.TypeFlagsNumber:
		to = ir.Number
	case target.Flags() == checker.TypeFlagsString:
		to = ir.String
	case target.Flags()&checker.TypeFlagsBoolean != 0 && target.Flags()&checker.TypeFlagsBooleanLiteral == 0:
		to = ir.Boolean
	default:
		return nil, l.notYet(node, "checked any used as "+l.checker.TypeToString(target)+" (a scalar tag does not establish this contract)")
	}
	if value.Type() == to {
		return value, nil
	}
	b := l.libraryArrayBuilder([]ir.Expression{fit(value, ir.Union)})
	held := b.read(b.parameters[0])
	tag := ir.TypeOf{Value: held}
	matches := ir.Binary{Operator: ir.Equal, Left: tag, Right: ir.StringConstant{Index: l.constant(typeName(to))}}
	prefix := "checked any: " + sourceExpression(node) + " needs " + l.checker.TypeToString(target) + ", found "
	found := ir.Conditional{Condition: ir.IsNull{Value: held}, WhenTrue: ir.StringConstant{Index: l.constant("null")}, WhenNot: tag}
	message := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant(prefix)}, found}}
	b.body = append(b.body, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: matches}, Then: []ir.Statement{ir.Panic{Message: message}}})
	return b.finish("checked_any_use", ir.Narrow{Value: held, To: to}), nil
}

// Arithmetic uses a scalar contract, not JavaScript's implicit coercions. For
// overloaded + and ordering, a known scalar operand selects the contract;
// two dynamic operands still need an operation-specific dynamic dispatch.
func (l *lowering) checkedAnyOperands(node *ast.Node, operator ast.Kind, left, right ir.Expression) (ir.Expression, ir.Expression, error) {
	if node.Kind != ast.KindBinaryExpression {
		return left, right, nil
	}
	binary := node.AsBinaryExpression()
	leftAny := l.checker.GetTypeAtLocation(binary.Left).Flags()&checker.TypeFlagsAny != 0
	rightAny := l.checker.GetTypeAtLocation(binary.Right).Flags()&checker.TypeFlagsAny != 0
	if !leftAny && !rightAny {
		return left, right, nil
	}
	_, arithmeticUse := arithmetic[operator]
	_, bitwiseUse := bitwise[operator]
	_, orderingUse := comparisons[operator]
	if !arithmeticUse && !bitwiseUse && !orderingUse {
		return left, right, nil
	}
	target := l.checker.GetNumberType()
	if operator == ast.KindPlusToken || orderingUse {
		known := left
		if leftAny {
			known = right
		}
		if leftAny && rightAny {
			return nil, nil, l.notYet(node, "checked any with two dynamic operands for + or ordering")
		}
		switch known.Type() {
		case ir.String:
			target = l.checker.GetStringType()
		case ir.Number:
		default:
			return nil, nil, l.notYet(node, "checked any arithmetic with a nonscalar operand")
		}
	}
	var err error
	if leftAny {
		left, err = l.checkedAnyType(binary.Left, left, target)
		if err != nil {
			return nil, nil, err
		}
	}
	if rightAny {
		right, err = l.checkedAnyType(binary.Right, right, target)
		if err != nil {
			return nil, nil, err
		}
	}
	return left, right, nil
}

func (l *lowering) checkedAnyPropertyReceiver(node *ast.Node, value ir.Expression) ir.Expression {
	b := l.libraryArrayBuilder([]ir.Expression{value})
	held := b.read(b.parameters[0])
	nullish := ir.Binary{Operator: ir.Or, Left: ir.IsNull{Value: held}, Right: ir.IsUndefined{Value: held}}
	found := ir.Conditional{Condition: ir.IsNull{Value: held}, WhenTrue: ir.StringConstant{Index: l.constant("null")}, WhenNot: ir.StringConstant{Index: l.constant("undefined")}}
	message := ir.Concat{Parts: []ir.Expression{ir.StringConstant{Index: l.constant("checked any: " + sourceExpression(node) + " needs non-null property receiver, found ")}, found}}
	b.body = append(b.body, ir.If{Condition: nullish, Then: []ir.Statement{ir.Panic{Message: message}}})
	return b.finish("checked_any_property_receiver", held)
}
