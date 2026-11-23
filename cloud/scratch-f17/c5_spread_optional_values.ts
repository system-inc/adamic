function make(flag: boolean) { return { get g(): number { return 2; }, ...(flag ? { a: 1 } : {}) }; }
const seen = make(true) as { a: number };
function values(): number[] { return Object.values(seen); }
console.log(values().join(','));
