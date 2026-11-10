function three(a: number, b: number, c: number): number { return a + b + c; }
function arity(f: Function): number { return f.length; }
console.log(String(arity(three.bind(null, 1))));
console.log(String(arity(three.bind(null))));
