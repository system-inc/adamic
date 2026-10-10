package javascript

import "fmt"

// Producer metadata is an explicit argument until the closure owner wires its
// immutable recorded signature. Never infer it from Function.length or a view.
// Presence and initialization are checked by the shared field read first.
const unionCallableShapeRuntime = `const adamicCallableRepresentationCompatible = (from, to, fromMembers = 0, toMembers = 0) => {
    if (from === 0 || to === 0) return false;
    if (from !== 10 && to !== 10 && from !== to) return false;
    if (from !== 10 && to !== 10 && (fromMembers === 0 || toMembers === 0)) return from === to;
    const given = fromMembers !== 0 || from === 10 ? fromMembers : (1 << from);
    const wanted = toMembers !== 0 || to === 10 ? toMembers : (1 << to);
    return given !== 0 && wanted !== 0 && (given & wanted) === given;
};
const adamicViewCallableSignaturesMatch = (recorded, expected) => recorded!==undefined && recorded!==null && expected!==undefined && expected!==null && recorded.result!==0 && expected.result!==0 && (expected.result===255 || adamicCallableRepresentationCompatible(recorded.result, expected.result, recorded.resultMask, expected.resultMask)) && recorded.parameters.length===expected.parameters.length && recorded.parameters.every((value,index)=>adamicCallableRepresentationCompatible(expected.parameters[index], value, expected.parameterMasks?.[index], recorded.parameterMasks?.[index]));
const adamicViewCallableShape = (value, recorded, expected, expression, optional = false) => {
    if (optional && value === undefined) return undefined;
    let found = value === null ? "null" : adamicTypeOf(value);
    if (found === "function") {
        found = "function with unknown signature";
        if (recorded !== undefined && recorded !== null && expected !== undefined && expected !== null) {
            if (recorded.parameters.length !== expected.parameters.length) found = "function with arity " + recorded.parameters.length;
            else if (recorded.result === 0 || expected.result === 0) found = "function with unknown signature";
            else if (expected.result !== 255 && !adamicCallableRepresentationCompatible(recorded.result, expected.result, recorded.resultMask, expected.resultMask)) found = "function with incompatible result representation";
            else if (adamicViewCallableSignaturesMatch(recorded, expected)) return value;
            else found = "function with incompatible parameter representations";
        }
    }
    const wanted = expected === undefined || expected === null ? "function with known signature" : expected.name;
    panic("field read failed: " + expression + " expected " + wanted + ", found " + found);
};
`

// Operands are already evaluated by shared field dispatch. This helper inserts
// no wrapper and changes neither callable identity nor receiver binding.
func emitUnionCallableShape(value, recorded, expected, expression string, optional bool) string {
	return fmt.Sprintf("adamicViewCallableShape(%s, %s, %s, %s, %t)", value, recorded, expected, quote(expression), optional)
}
