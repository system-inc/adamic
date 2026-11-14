// Spec: No compiler at runtime in a native binary
// Expect: refused constructing a function with the global Function
const f = new Function("return 1");
