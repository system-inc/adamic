interface H { f(x: number): number }
type Loose = { f: (x: number) => number };
function call({ f }: H): number { return f(2); }
const loose: Loose = { f(x) { return x; } };
console.log(String(call(loose)));
