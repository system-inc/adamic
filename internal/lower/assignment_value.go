package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// assignmentValue keeps writes visible to analysis. Simple assignment yields the
// saved right value, even when a setter or a later operand changes the destination.
func (l *lowering) assignmentValue(node *ast.Node) (ir.Expression, error) {
	if node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		return l.simpleAssignmentValue(node)
	}
	target := ast.SkipParentheses(node.AsBinaryExpression().Left)
	if !ast.IsIdentifier(target) {
		return nil, l.notYet(target, "an assignment value to a member")
	}
	body, err := l.assignment(node)
	if err != nil {
		return nil, err
	}
	local, known := l.local(target)
	if !known {
		return nil, l.notYet(target, "an assignment value to this binding")
	}
	l.result.Locals[local].ExpressionAssigned = true
	result, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	var value ir.Expression = ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checked(local)}
	if value.Type() == ir.Union && result != ir.Union {
		value = ir.Narrow{Value: value, To: result}
	} else {
		value = fit(value, result)
	}
	if value.Type() != result {
		return nil, l.notYet(node, "an assignment result with different storage")
	}
	return ir.Effects{Body: body, Result: value}, nil
}

// assignmentRight is the origin of a simple assignment's value. It deliberately
// leaves compound and logical assignments alone: their results need other proofs.
func assignmentRight(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	for node.Kind == ast.KindBinaryExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
		node = ast.SkipParentheses(node.AsBinaryExpression().Right)
	}
	return node
}

func (l *lowering) assignmentSnapshot(body *[]ir.Statement, name string, value ir.Expression) ir.Expression {
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: value.Type(), Function: l.functionIndex})
	*body = append(*body, ir.Declare{Local: local, Value: value})
	return ir.Read{Local: local, Of: value.Type()}
}

func (l *lowering) simpleAssignmentValue(node *ast.Node) (ir.Expression, error) {
	writes, err := l.assignment(node)
	if err != nil {
		return nil, err
	}
	if len(writes) != 1 {
		return nil, l.notYet(node, "an assignment value with a destructuring target")
	}
	body := []ir.Statement{}
	var stored ir.Expression
	switch write := writes[0].(type) {
	case ir.Assign:
		l.result.Locals[write.Local].ExpressionAssigned = true
		stored = write.Value
	case ir.SetProperty:
		write.Object = l.assignmentSnapshot(&body, "assignment_object", write.Object)
		stored = write.Value
		writes[0] = write
	case ir.SetIndex:
		write.Array = l.assignmentSnapshot(&body, "assignment_array", write.Array)
		write.Index = l.assignmentSnapshot(&body, "assignment_index", write.Index)
		stored = write.Value
		writes[0] = write
	default:
		return nil, l.notYet(node, "an assignment value to this represented target")
	}
	result, err := l.typeOf(node)
	// Absent values use the destination's existing representation; their
	// checker types alone have no independent native storage type.
	if l.checker.GetTypeAtLocation(node).Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull) != 0 {
		result = stored.Type()
	} else if err != nil {
		return nil, err
	}
	// Undo only the store's representation adapter. Checks inside the RHS stay
	// attached to that single evaluation, and the saved value keeps its own type.
	value := stored
	switch adapted := stored.(type) {
	case ir.Box:
		if adapted.Value.Type() == result {
			value = adapted.Value
		}
	case ir.MaybeOf:
		if adapted.Value != nil && adapted.Value.Type() == result {
			value = adapted.Value
		}
	case ir.WeakOf:
		if adapted.Value.Type() == result {
			value = adapted.Value
		}
	}
	saved := l.assignmentSnapshot(&body, "assignment_value", value)
	switch write := writes[0].(type) {
	case ir.Assign:
		write.Value = fit(saved, stored.Type())
		writes[0] = write
	case ir.SetProperty:
		write.Value = fit(saved, stored.Type())
		writes[0] = write
	case ir.SetIndex:
		write.Value = fit(saved, stored.Type())
		writes[0] = write
	}
	body = append(body, writes[0])
	if saved.Type() == ir.Union && result != ir.Union {
		saved = ir.Narrow{Value: saved, To: result}
	} else {
		saved = fit(saved, result)
	}
	if saved.Type() != result {
		return nil, l.notYet(node, "an assignment result with different storage")
	}
	return ir.Effects{Body: body, Result: saved}, nil
}
