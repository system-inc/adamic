interface H { f(x: number): number }
const h: H = { f: x => x + 1 };
const { f } = h;
console.log(String(f(2)));
