interface H { f(x: number): number }
const h: H = { f: x => x + 1 };
const { f: g } = h;
console.log(String(g(2)));
