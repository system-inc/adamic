interface H { f(x: number): number }
const h: H = { f: x => x + 1 };
let f: (x: number) => number = x => x - 1;
({ f } = h);
console.log(String(f(2)));
