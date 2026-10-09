package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// The callee has already been evaluated by callThrough. Evaluate thisArg before
// packing arguments; the adapter runs its producer checks after both are ready.
func (e *emitter) callWithReceiver(expression ir.CallClosure, closure string) string {
	declaration := "static adamic_value adamic_view_surface_call(adamic_closure *value, void *receiver, adamic_value *arguments, size_t count, size_t slots) { return adamic_closure_receiver_call(value, receiver, arguments, count, slots); }"
	found := false
	for _, existing := range e.declarations {
		found = found || existing == declaration
	}
	if !found {
		e.declarations = append(e.declarations, declaration)
	}
	receiver := e.value(expression.Receiver)
	packed, count := e.closureArguments(expression)
	call := fmt.Sprintf("adamic_view_surface_call(%s, %s, %s, %s, %d)", closure, receiver, packed, count, e.packedArgumentSize(expression))
	if expression.Returns == 0 {
		e.line("%s;", call)
		e.closureThrown()
		return "0"
	}
	result := e.temporary()
	e.line("adamic_value %s = %s;", result, call)
	e.closureThrown()
	if expression.Returns.IsReference() {
		return e.own(expression.Returns, fmt.Sprintf("(%s)%s.reference", cType(expression.Returns), result))
	}
	return e.snapshot(expression.Returns, unslotted(expression.Returns, result+"."+member(expression.Returns)))
}
