function pick(n: number): number | string { return n > 0 ? n : 'none'; }
const o = { get g(): number { return 2; }, a: pick(1) };
const seen = o as { a: number };
function values(): number[] { return Object.values(seen); }
console.log(values().join(','));
