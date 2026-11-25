interface H { f(x: number): number }
function inc(x: number): number { return x + 1; }
const h: H = { f: inc };
const { f } = h;
console.log(String(f(2)));
