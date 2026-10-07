package javascript

import "fmt"

// Lane 1 includes this with the shared readiness runtime when array contracts
// reach backend dispatch. check returns the value or a transitive checked view.
// Keeping it outside the array preserves identity and observes mutable aliases.
const viewArraysRuntime = `const adamicViewArray = (value, expression, expected) => {
    if (!Array.isArray(value)) panic("field read failed: " + expression + " expected " + expected + ", found " + (value === null ? "null" : typeof value));
    return value;
};
const adamicViewArrayRead = (array, index, expression, check) => {
    const value = array[index];
    return check(value, expression + "[" + index + "]");
};
`

// Operands and contract checker are supplied by shared dispatch. The helper reads
// only the selected element; neither a field read nor this hook scans the array.
func emitViewArrayRead(array, index, expression, check string) string {
	return fmt.Sprintf("adamicViewArrayRead(%s, %s, %s, %s)", array, index, quote(expression), check)
}

// Presence and readiness share lane 1's state. The kind check deliberately does
// not call the element checker; elements are checked at their own read sites.
func emitViewArrayFieldRead(object, member, expression, expected string) string {
	return fmt.Sprintf("adamicViewArray(adamicReadField(%s, %s, %s, false, false, %s), %s, %s)", object, quote(member), quote(expression), quote(expected), quote(expression), quote(expected))
}
