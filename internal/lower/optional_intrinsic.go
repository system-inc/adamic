package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Receiver-optional library methods keep the receiver once and evaluate the call
// (including its arguments) only when it is present. Optional function calls need a
// different guard and ABI, and a chain that continues past the call is
// optionalChain's (optional_chain.go).
func (l *lowering) optionalIntrinsic(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	if l.questionDot(node) || hasSpread(node) || optionalIntrinsicContinues(node) {
		return nil, false, nil
	}
	// join already owns its receiver guard and effect-local lifetime.
	if l.isOptionalJoin(node) {
		return nil, false, nil
	}
	if value, handled, err := l.optionalMapGet(node); handled {
		return value, true, err
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || (!l.libraryMember(callee) && !l.nodeFSEntryMethod(callee)) {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	if !l.questionDot(callee) {
		return nil, false, nil
	}
	receiver, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	read := ir.Read{Local: local, Of: receiver.Type()}
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_intrinsic_receiver", Type: receiver.Type(), Function: function, ExpressionAssigned: true})
	lowered, known, uses, err := func() (ir.Expression, bool, int, error) {
		previousCall, previousBinding := l.optionalReceiverCall, l.optionalReceiverBinding
		l.optionalReceiverCall = node
		l.optionalReceiverBinding = optionalReceiverBinding{call: node, receiver: access.Expression, read: read}
		defer func() {
			l.optionalReceiverCall, l.optionalReceiverBinding = previousCall, previousBinding
		}()
		value, known, err := l.builtin(node)
		return value, known, l.optionalReceiverBinding.uses, err
	}()
	if !known || err != nil {
		return nil, known, err
	}
	if !receiver.Type().IsReference() {
		return nil, true, l.notYet(node, "an optional intrinsic without a reference receiver")
	}
	if uses == 0 {
		return nil, true, l.notYet(node, "an optional intrinsic whose receiver is not an explicit operand")
	}
	if uses > 1 {
		return nil, true, l.notYet(node, "an optional intrinsic that reads its receiver more than once")
	}
	body := []ir.Statement{ir.Declare{Local: local, Value: receiver}}
	present := ir.Unary{Operator: ir.Not, Operand: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}}}
	if lowered.Type() == 0 {
		body = append(body, ir.If{Condition: present, Then: []ir.Statement{ir.Evaluate{Value: lowered}}})
		return ir.Effects{Body: body, Result: ir.Undefined{}}, true, nil
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	// The receiver is tested whatever its type says, since a record's value read can be missing
	// where the type holds none (a.r["k"]?.push(a)). So a number or a boolean result is held as a
	// maybe even where TypeScript types the call without undefined, and reads undefined there as
	// Node's does; held as the scalar, the absent arm would be undefined in a double (#qgthc6g).
	of = ir.Maybe(of)
	return ir.Effects{Body: body, Result: ir.Conditional{
		Condition: present,
		WhenTrue:  fit(lowered, of), WhenNot: fit(ir.Undefined{}, of), Of: of,
	}}, true, nil
}

// This binding belongs to one call. Snapshots copy its counter by value, and a
// nested optional call restores the enclosing binding when its builtin returns.
type optionalReceiverBinding struct {
	call, receiver *ast.Node
	read           ir.Read
	uses           int
}

// A non-null assertion stays in the chain, while parentheses end it. A new
// optional access has its own guard and can read this call's undefined result.
func optionalIntrinsicContinues(node *ast.Node) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindNonNullExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		access := parent.AsPropertyAccessExpression()
		return access.Expression == node && access.QuestionDotToken == nil
	case ast.KindElementAccessExpression:
		access := parent.AsElementAccessExpression()
		return access.Expression == node && access.QuestionDotToken == nil
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		return call.Expression == node && call.QuestionDotToken == nil
	}
	return false
}

// A guarded builtin sees the same generic receiver type as its plain call after
// narrowing, a guarded optional chain's steps included (optionalChainOperandType).
// Other type queries retain the checker's original nullish members.
func (l *lowering) intrinsicReceiverType(node *ast.Node) *checker.Type {
	proven := l.optionalChainOperandType(node)
	if call := l.optionalReceiverCall; call != nil {
		callee := ast.SkipParentheses(call.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && callee.AsPropertyAccessExpression().Expression == node {
			return l.checker.GetNonNullableType(proven)
		}
	}
	return proven
}
