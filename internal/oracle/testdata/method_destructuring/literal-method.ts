interface H { f(x: number): number }
function call({ f }: H): number { return f(2); }
const a: H = { f: x => x + 1 };
const b: H = { f(x) { return x; } };
console.log(String(call(a) + call(b)));
