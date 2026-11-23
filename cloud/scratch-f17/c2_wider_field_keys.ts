function pick(n: number): number | string { return n > 0 ? n : 'none'; }
const o = { get g(): number { return 2; }, a: pick(1) };
const seen = o as { a: number };
console.log(Object.keys(seen).join(','));
