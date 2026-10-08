package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Receiver-optional, zero-argument library methods need one saved receiver and
// a missing result. An optional function call has a different guard and ABI.
func (l *lowering) optionalIntrinsic(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	if call.QuestionDotToken != nil || len(call.Arguments.Nodes) != 0 {
		return nil, false, nil
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || !l.libraryMember(callee) {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	if access.QuestionDotToken == nil {
		return nil, false, nil
	}
	value, known, err := l.builtin(node)
	if !known || err != nil {
		return nil, known, err
	}
	receiver, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	if !receiver.Type().IsReference() || value.Type() == 0 {
		return nil, true, l.notYet(node, "an optional intrinsic without a reference receiver and a value result")
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
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
	return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: receiver}}, Result: ir.Conditional{
		Condition: ir.Unary{Operator: ir.Not, Operand: ir.IsUndefined{Value: read}},
		WhenTrue:  fit(lowered, of), WhenNot: fit(ir.Undefined{}, of), Of: of,
	}}, true, nil
}
