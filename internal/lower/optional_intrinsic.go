package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Receiver-optional library methods keep the receiver once and evaluate the call
// (including its arguments) only when it is present. Optional function calls and
// continuations need a different guard and ABI.
func (l *lowering) optionalIntrinsic(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	if call.QuestionDotToken != nil || hasSpread(node) || optionalIntrinsicContinues(node) {
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
	if access.QuestionDotToken == nil {
		return nil, false, nil
	}
	value, known, err := func() (ir.Expression, bool, error) {
		previous := l.optionalReceiverCall
		l.optionalReceiverCall = node
		defer func() { l.optionalReceiverCall = previous }()
		return l.builtin(node)
	}()
	if !known || err != nil {
		return nil, known, err
	}
	receiver, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	if !receiver.Type().IsReference() {
		return nil, true, l.notYet(node, "an optional intrinsic without a reference receiver")
	}
	occurrences := 0
	walk(value, func(node any) bool {
		if expression, ok := node.(ir.Expression); ok && reflect.DeepEqual(expression, receiver) {
			occurrences++
			return false
		}
		return true
	})
	if occurrences != 1 {
		return nil, true, l.notYet(node, "an optional intrinsic without exactly one explicit receiver operand")
	}
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	read := ir.Read{Local: local, Of: receiver.Type()}
	replaced := false
	var replace func(reflect.Value) reflect.Value
	replace = func(v reflect.Value) reflect.Value {
		if !v.IsValid() {
			return v
		}
		if !replaced && v.CanInterface() {
			if expression, ok := v.Interface().(ir.Expression); ok && reflect.DeepEqual(expression, receiver) {
				replaced = true
				if v.Kind() == reflect.Interface {
					r := reflect.New(v.Type()).Elem()
					r.Set(reflect.ValueOf(read))
					return r
				}
				return reflect.ValueOf(read)
			}
		}
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return v
			}
			r := reflect.New(v.Type()).Elem()
			r.Set(replace(v.Elem()))
			return r
		case reflect.Struct:
			r := reflect.New(v.Type()).Elem()
			for i := 0; i < v.NumField(); i++ {
				r.Field(i).Set(replace(v.Field(i)))
			}
			return r
		case reflect.Slice:
			if v.IsNil() {
				return v
			}
			r := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				r.Index(i).Set(replace(v.Index(i)))
			}
			return r
		}
		return v
	}
	lowered := replace(reflect.ValueOf(value)).Interface().(ir.Expression)
	if !replaced {
		return nil, true, l.notYet(node, "an optional intrinsic whose receiver is not an explicit operand")
	}
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_intrinsic_receiver", Type: receiver.Type(), Function: function, ExpressionAssigned: true})
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
	return ir.Effects{Body: body, Result: ir.Conditional{
		Condition: present,
		WhenTrue:  fit(lowered, of), WhenNot: fit(ir.Undefined{}, of), Of: of,
	}}, true, nil
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
// narrowing. Other type queries retain the checker's original nullish members.
func (l *lowering) intrinsicReceiverType(node *ast.Node) *checker.Type {
	proven := l.checker.GetTypeAtLocation(node)
	if call := l.optionalReceiverCall; call != nil {
		callee := ast.SkipParentheses(call.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && callee.AsPropertyAccessExpression().Expression == node {
			return l.checker.GetNonNullableType(proven)
		}
	}
	return proven
}
