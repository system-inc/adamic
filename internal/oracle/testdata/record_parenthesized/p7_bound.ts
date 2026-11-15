// A parenthesized record read bound to a const, then read (#grc95d6).
const r: Record<string, number[]> = { k: [1, 2] };
const n: number[] = (r["k"]);
console.log(String(n.length));
