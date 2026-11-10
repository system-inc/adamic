function three(a: number, b: number, c: number): number { return a + b + c; }
function make(): Function { return three.bind(null, 1, 2); }
const f = make();
console.log(String(f.length));
