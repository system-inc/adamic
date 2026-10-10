package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// The adapter has the viewed ABI. Its invocation checks and converts against
// the underlying producer, using the same dispatch as an immediate view call.
func (e *emitter) emitViewCallableEscape(property ir.Property, receiver, method string) string {
	e.declarations = append(e.declarations, "#include \"view_callables.h\"")
	target := e.program.ViewContracts[property.ViewEscapeContract-1]
	call := ir.CallClosure{CallContract: property.ViewEscapeContract, CallWhere: "escaping call (read at " + property.ViewWhere + ")", Returns: e.program.ViewContracts[target.Result-1].Of}
	if e.program.ViewContracts[target.Result-1].Name == "void" {
		call.Returns = 0
	}
	for _, parameter := range target.Parameters {
		call.Arguments = append(call.Arguments, ir.Read{Of: e.program.ViewContracts[parameter-1].Of})
	}
	invoke := e.directViewCallableInvoke(call, property)
	wrapper, make := e.temporary(), e.temporary()
	methodZero := "NULL"
	if e.program.ClosureConventionNeeded() {
		methodZero = "(adamic_method_entry){0}"
	}
	count := fmt.Sprint(len(target.Parameters))
	extra := ""
	constructor := "adamic_closure_new"
	if e.program.ClosureConventionNeeded() {
		extra = ", size_t count"
		count = "count - 1"
		constructor = "adamic_counted_closure_new"
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_value %s(adamic_closure *self, adamic_value *arguments%s) { return %s(self->view->underlying, arguments[0].reference, %s, arguments + 1, %s); }", wrapper, extra, invoke, methodZero, count))
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_closure *%s(adamic_closure *underlying, const void *key) { (void)underlying; (void)key; adamic_closure *adapter = %s(%s, 0); adapter->receiver = true; return adapter; }", make, constructor, wrapper))
	key := fmt.Sprintf("adamic_view_adapter_key_%d", property.ViewTypeID)
	declaration := "static const char " + key + ";"
	found := false
	for _, existing := range e.declarations {
		found = found || existing == declaration
	}
	if !found {
		e.declarations = append(e.declarations, declaration)
	}
	// Presence/readiness checks retain their ordinary field semantics; signature
	// relations are checked only by wrapper code after arguments have evaluated.
	value := e.own(ir.Closure, fmt.Sprintf("adamic_retain(adamic_view_callable_typed(%s, %s, &%s, &%s, %s, %s, %t, %t))", receiver, cString(property.Name), e.cache(), method, cString(property.View), cString(property.ViewType), property.Absent, property.Optional))
	return e.own(ir.Closure, fmt.Sprintf("adamic_view_adapter_intern(%s, &%s, %s)", value, key, make))
}
