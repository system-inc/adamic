interface H { f(x: number): number }
let reads = 0;
const h: H = { get f() { reads += 1; return (x: number) => x * 2; } };
const { f } = h;
console.log(String(f(3)) + ":" + String(reads));
