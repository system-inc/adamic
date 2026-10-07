package javascript

import "fmt"

// A function tag certifies kind only. Lowering must prove the signature before
// emitting a call. Methods use the backend's explicit receiver convention.
const viewCallablesRuntime = `const adamicViewCallableRead = (object, name, expression, method = false) => {
    let holder = object;
    if (method && holder !== null && holder !== undefined) {
        while (!Object.hasOwn(holder, name) && Object.getPrototypeOf(holder) !== null) holder = Object.getPrototypeOf(holder);
    }
    const value = adamicReadField(holder, name, expression, false, false, "function");
    if (adamicTypeOf(value) !== "function") panic("field read failed: " + expression + " expected function, found " + (value === null ? "null" : adamicTypeOf(value)));
    return value;
};
const adamicViewCallableCall = (value, receiver, arguments_, method) => method ? value(receiver, ...arguments_) : adamicCall(value, arguments_);
`

func emitViewCallableRead(object, member, expression string, method bool) string {
	return fmt.Sprintf("adamicViewCallableRead(%s, %s, %s, %t)", object, quote(member), quote(expression), method)
}

func emitViewCallableCall(value, receiver, arguments string, method bool) string {
	return fmt.Sprintf("adamicViewCallableCall(%s, %s, [%s], %t)", value, receiver, arguments, method)
}
