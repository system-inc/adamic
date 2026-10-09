package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// A function tag certifies kind only. Lowering must prove the signature before
// emitting a call. Methods use the backend's explicit receiver convention.
const viewCallablesRuntime = `const adamicViewCallableRead = (object, name, expression, method = false) => {
    let holder = object;
    if (method && holder !== null && holder !== undefined) {
        while (!Object.hasOwn(holder, name) && Object.getPrototypeOf(holder) !== null) holder = Object.getPrototypeOf(holder);
    }
    const value = adamicReadField(holder, name, expression, false, false, "function");
    if (adamicTypeOf(value) !== "function") panic("field read failed: " + expression + " is not a function; expected function, found " + (value === null ? "null" : adamicTypeOf(value)));
    return value;
};
const adamicViewCallableCall = (value, receiver, arguments_, method) => method ? value(receiver, ...arguments_) : adamicCall(value, arguments_);
const adamicViewCallablePreparedRead = (object, name, expression, method, optional, absent, expected) => {
    let holder = object;
    if (method && holder !== null && holder !== undefined) {
        while (!Object.hasOwn(holder, name) && Object.getPrototypeOf(holder) !== null) holder = Object.getPrototypeOf(holder);
    }
    return adamicReadField(holder, name, expression, optional, absent, expected);
};
` + viewCallableShapeRuntime + viewCallableAdapterRuntime

func emitViewCallableRead(object, member, expression string, method bool) string {
	return fmt.Sprintf("adamicViewCallableRead(%s, %s, %s, %t)", object, quote(member), quote(expression), method)
}

func emitViewCallableCall(value, receiver, arguments string, method bool) string {
	return fmt.Sprintf("adamicViewCallableCall(%s, %s, [%s], %t)", value, receiver, arguments, method)
}

func (e *emitter) emitViewCallableProperty(property ir.Property) string {
	if property.ViewEscape && !property.ViewEscapeAdamic {
		return e.emitViewCallableEscape(property)
	}
	if property.ViewContract != 0 && (e.program.ViewContracts[property.ViewContract-1].Result != 0 || e.program.ViewContracts[property.ViewContract-1].DiscardResult) {
		expected := e.program.ViewContracts[property.ViewContract-1].Name
		raw := func(object string) string {
			return fmt.Sprintf("adamicViewCallablePreparedRead(%s, %s, %s, %t, %t, %t, %s)", object, quote(property.Name), quote(property.View), property.Method, property.Optional, property.Absent, quote(expected))
		}
		if property.Method {
			value := e.emitViewCallableCertificate(property, raw("object"))
			return "((object) => { const value = " + value + "; return value instanceof AdamicClosure ? value : {code: (self, values) => value(object, ...values)}; })(" + e.value(property.Object) + ")"
		}
		return e.emitViewCallableCertificate(property, raw(e.value(property.Object)))
	}
	if property.Method {
		return "((object) => { const value = " + emitViewCallableRead("object", property.Name, property.View, true) + "; return value instanceof AdamicClosure ? value : {code: (self, values) => value(object, ...values)}; })(" + e.value(property.Object) + ")"
	}
	value := emitViewCallableRead(e.value(property.Object), property.Name, property.View, false)
	if property.ViewContract > 0 {
		contract := e.program.ViewContracts[property.ViewContract-1]
		if contract.Kind == ir.ViewUnion && contract.Of == ir.Closure {
			return e.emitViewCallableCertificate(property, value)
		}
	}
	return value
}
