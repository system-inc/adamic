interface H { f(x: number): number }
function plus(x: number): number { return x + 10; }
const h: H = { f: x => x + 1 };
Object.assign(h, { f: plus });
const { f } = h;
console.log(String(f(2)));
