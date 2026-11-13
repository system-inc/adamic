// Spec: No compiler at runtime in a native binary
// Expect: refused reading the global eval
eval("1 + 1");
