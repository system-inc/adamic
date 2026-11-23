class C { a?: number; get g(): number { return 2; } }
const x = new C();
x.a = 1;
const seen = x as { a: number };
function values(): number[] { return Object.values(seen); }
console.log(values().join(','));
