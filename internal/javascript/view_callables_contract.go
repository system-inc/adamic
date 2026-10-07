package javascript

import "fmt"

// Producer metadata is an explicit argument until the closure owner wires its
// immutable recorded signature. Never infer it from Function.length or a view.
// Presence and initialization are checked by the shared field read first.
const viewCallableShapeRuntime = `const adamicViewCallableShape = (value, recorded, expected, expression, optional = false) => {
    if (optional && value === undefined) return undefined;
    let found = value === null ? "null" : adamicTypeOf(value);
    if (found === "function") {
        found = "function with unknown signature";
        if (recorded !== undefined && recorded !== null && expected !== undefined && expected !== null) {
            if (recorded.parameters.length !== expected.parameters.length) found = "function with arity " + recorded.parameters.length;
            else if (recorded.result === 0 || expected.result === 0) found = "function with unknown signature";
            else if (recorded.result !== expected.result) found = "function with incompatible result representation";
            else if (recorded.parameters.every((representation, index) => representation !== 0 && expected.parameters[index] !== 0 && representation === expected.parameters[index])) return value;
            else found = "function with incompatible parameter representations";
        }
    }
    const wanted = expected === undefined || expected === null ? "function with known signature" : expected.name;
    panic("field read failed: " + expression + " expected " + wanted + ", found " + found);
};
`

// Operands are already evaluated by shared field dispatch. This helper inserts
// no wrapper and changes neither callable identity nor receiver binding.
func emitViewCallableShape(value, recorded, expected, expression string, optional bool) string {
	return fmt.Sprintf("adamicViewCallableShape(%s, %s, %s, %s, %t)", value, recorded, expected, quote(expression), optional)
}
