// Spec: numbers, booleans, strings, arrays, maps, plain objects
// Expect: lowers
const user = { eval(x: number): number { return x; } }; console.log(String(user.eval(1)));
