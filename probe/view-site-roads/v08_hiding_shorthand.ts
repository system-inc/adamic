const full = { a: 1, b: "str" };
const narrow: { a: number } = full;
const kept = narrow;
const holder: { kept: { a: number; b?: number } } = { kept };
console.log(String((holder.kept.b ?? 0) + 1));
