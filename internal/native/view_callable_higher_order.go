package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// A forward declaration breaks generation cycles when an unrelated producer in
// the registry itself takes callbacks. The runtime adapter still has one layer.
func (e *emitter) viewCallableCallbackAdapter(id ir.ViewContractID, value string, origin ir.Property, where string) string {
	name := fmt.Sprintf("adamic_view_callback_%d", id)
	prototype := "static adamic_closure *" + name + "(adamic_closure *value);"
	for _, existing := range e.declarations {
		if existing == prototype {
			return name + "(" + value + ")"
		}
	}
	e.declarations = append(e.declarations, prototype)
	target := e.program.ViewContracts[id-1]
	property := ir.Property{ViewEscape: true, View: origin.View + " callback", ViewWhere: where, ViewType: target.Name}
	call := ir.CallClosure{CallContract: id, CallWhere: "callback adapted at " + where, Returns: e.program.ViewContracts[target.Result-1].Of}
	if e.program.ViewContracts[target.Result-1].Kind == ir.ViewUndefined {
		call.Returns = 0
	}
	for _, parameter := range target.Parameters {
		call.Arguments = append(call.Arguments, ir.Read{Of: e.program.ViewContracts[parameter-1].Of})
	}
	invoke := e.directViewCallableInvoke(call, property)
	wrapper, make := e.temporary(), e.temporary()
	zero, extra, count, constructor := "NULL", "", fmt.Sprint(len(target.Parameters)), "adamic_closure_new"
	if e.program.ClosureConventionNeeded() {
		zero = "(adamic_method_entry){0}"
		extra = ", size_t count"
		count = "count - 1"
		constructor = "adamic_counted_closure_new"
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_value %s(adamic_closure *self, adamic_value *arguments%s) { return %s(self->view->underlying, arguments[0].reference, %s, arguments + 1, %s); }", wrapper, extra, invoke, zero, count))
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_closure *%s(adamic_closure *underlying, const void *key) { (void)underlying; (void)key; adamic_closure *adapter=%s(%s, 0); adapter->receiver=true; return adapter; }", make, constructor, wrapper))
	key := fmt.Sprintf("adamic_view_adapter_key_%d", target.CallableTypeID)
	declaration := "static const char " + key + ";"
	found := false
	for _, existing := range e.declarations {
		found = found || existing == declaration
	}
	if !found {
		e.declarations = append(e.declarations, declaration)
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_closure *%s(adamic_closure *value) { return adamic_view_adapter_intern(value, &%s, %s); }", name, key, make))
	return name + "(" + value + ")"
}
