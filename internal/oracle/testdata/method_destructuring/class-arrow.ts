interface H { f(x: number): number }
class K { f = (x: number): number => x + 1; }
const h: H = new K();
const { f } = h;
console.log(String(f(2)));
