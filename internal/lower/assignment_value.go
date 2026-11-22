package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// assignmentValue keeps the assignment visible to write analysis and reads its
// stored value immediately. A local has no setter that could replace that value.
func (l *lowering) assignmentValue(node *ast.Node) (ir.Expression, error) {
	if l.uninitializedInitializer(node.AsBinaryExpression().Right) {
		return nil, l.notYet(node, "using the result of a placeholder reset")
	}
	target := ast.SkipParentheses(node.AsBinaryExpression().Left)
	if target.Kind == ast.KindPropertyAccessExpression && node.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken && !l.namespaceMember(target) {
		return l.memberAssignmentValue(node, target)
	}
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
	if origin := l.result.Locals[local].Placeholder; origin != "" && !l.placeholderAllowsUnset(node) {
		value = ir.PlaceholderUse{Value: value, Origin: origin, Use: "assignment result", Path: sourceExpression(node), Where: l.program.Where(node), Of: result}
	}
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

// memberAssignmentValue lowers o.f = v used as a value, as tsc's sources write it (la = la.next ||
// (la.next = new LinkedArray(la)), lnIdx = la.pos = 1; revcomp, #4gxxhvf). The write is the statement
// form's, unchanged, so every rule a plain o.f = v meets (Rule C's holds, checked views, readonly and
// presence slots, a host member's write) judges it here too, and a target that form refuses stays refused.
// Its receiver and its value are each held in a local, in Node's order: the receiver, then the value, then
// the store, so a value that rebinds the receiver's name still stores into the object named first. The
// result reads the value's local back: the value is evaluated once, and a store is never read back, since
// a getter may not give it.
func (l *lowering) memberAssignmentValue(node, target *ast.Node) (ir.Expression, error) {
	body, err := l.assignment(node)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, l.notYet(target, "an assignment value to a member with no field store")
	}
	store, isStore := body[len(body)-1].(ir.SetProperty)
	if !isStore || store.Uninitialized || store.Unset || store.Define {
		// A setter, a private static, a record or foreign write: each stores through a call or a shape this
		// can't hold the value of.
		return nil, l.notYet(target, "an assignment value to a member that isn't a plain field store")
	}
	result, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	receiver := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "assigned_receiver", Type: store.Object.Type(), Function: l.functionIndex, ExpressionAssigned: true})
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "assigned", Type: store.Value.Type(), Function: l.functionIndex, ExpressionAssigned: true})
	statements := append(append([]ir.Statement{}, body[:len(body)-1]...), ir.Declare{Local: receiver, Value: store.Object}, ir.Declare{Local: held, Value: store.Value})
	store.Object = ir.Read{Local: receiver, Of: l.result.Locals[receiver].Type}
	store.Value = ir.Read{Local: held, Of: l.result.Locals[held].Type}
	statements = append(statements, store)
	var value ir.Expression = ir.Read{Local: held, Of: l.result.Locals[held].Type}
	if value.Type() == ir.Union && result != ir.Union {
		value = ir.Narrow{Value: value, To: result}
	} else {
		value = fit(value, result)
	}
	if value.Type() != result {
		return nil, l.notYet(node, "an assignment result with different storage")
	}
	return ir.Effects{Body: statements, Result: value}, nil
}
