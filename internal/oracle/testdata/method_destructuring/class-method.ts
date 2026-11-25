interface H { f(x: number): number }
class K implements H { f(x: number): number { return x; } }
const h: H = new K();
const { f } = h;
console.log(String(f(2)));
