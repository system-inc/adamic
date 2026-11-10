function three(a: number, b: number, c: number): number { return a + b + c; }
const bound = three.bind(null, 1);
const viewed: Function = bound;
console.log(String(viewed.length));
