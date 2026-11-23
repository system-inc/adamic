const o = { get g(): number { return 2; }, a: 1 };
const seen: { a: number } = o;
function values(): number[] { return Object.values(seen); }
console.log(values().join(','));
