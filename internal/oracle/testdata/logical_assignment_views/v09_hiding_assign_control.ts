const full = { a: 1, b: "str" };
const narrow: { a: number } = full;
let opt: { a: number; b?: number } | undefined = undefined;
opt = narrow;
console.log(String((opt.b ?? 0) + 1));
