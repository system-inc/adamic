interface H { f(x: number): number }
const h: H = { f: x => x + 1 };
const counts: number[] = [0]; const i: number = 0;
counts[i] = 1;
const { f } = h;
console.log(String(f(2) + counts[0]!));
