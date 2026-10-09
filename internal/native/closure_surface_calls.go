package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// The callee has already been evaluated by callThrough. Evaluate thisArg before
// packing arguments; the adapter runs its producer checks after both are ready.
func (e *emitter) callWithReceiver(expression ir.CallClosure, closure string) string {
	e.viewCallableBlameRuntime()
	declaration := "static adamic_value adamic_view_surface_call(adamic_closure *value, void *receiver, adamic_value *arguments, size_t count, size_t slots, const char *site) { const char *previous = adamic_view_call_site; if (site != NULL) adamic_view_call_site = site; adamic_value result = adamic_closure_receiver_call(value, receiver, arguments, count, slots); adamic_view_call_site = previous; return result; }"
	found := false
	for _, existing := range e.declarations {
		found = found || existing == declaration
	}
	if !found {
		e.declarations = append(e.declarations, declaration)
	}
	receiver := e.value(expression.Receiver)
	packed, count := e.closureArguments(expression)
	if expression.ArgumentCount != nil {
		count = e.value(expression.ArgumentCount)
	}
	if expression.CheckBound {
		property := ir.Property{View: expression.BoundView, ViewType: e.program.ViewContracts[expression.CallContract-1].Name}
		invoke := e.directViewCallableInvoke(expression, property)
		method := "NULL"
		if e.program.ClosureConventionNeeded() {
			method = "(adamic_method_entry){0}"
		}
		e.line("(void)%s(%s, %s, %s, %s, %s);", invoke, closure, receiver, method, packed, count)
		return "0"
	}
	call := fmt.Sprintf("adamic_view_surface_call(%s, %s, %s, %s, %d, %s)", closure, receiver, packed, count, e.packedArgumentSize(expression), viewCallableSourceSite(expression))
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
