package javascript

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) optionalClosureCall(call ir.CallClosure) string {
	callee := ""
	if property, ok := call.Closure.(ir.Property); ok && property.Method {
		lookup := "adamicCallee(object, " + quote(property.Name) + ", true)"
		if property.Optional {
			lookup = "object == null ? undefined : " + lookup
		}
		callee = "((object) => " + lookup + ")(" + e.value(property.Object) + ")"
	} else {
		callee = e.value(call.Closure)
	}
	return "((callee) => callee == null ? undefined : adamicCall(callee, [" + e.values(call.Arguments) + "]))(" + callee + ")"
}
