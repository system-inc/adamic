interface H { f(x: number): number }
const f = (x: number): number => x + 1;
const first: H = { f };
const second: H = { f };
const { f: a } = first;
const { f: b } = second;
console.log(String(a(2) + b(2)));
