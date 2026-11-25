interface H { f(x: number): number }
type Loose = { f: (x: number) => number };
function twelve(x: number) { return x * 12; }
const h: H = { f: x => x + 1 };
const loose: Loose = h;
loose["f"] = twelve;
const { f } = h;
console.log(String(f(1)));
