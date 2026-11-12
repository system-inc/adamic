const full: { a: number; b?: number } = { a: 1, b: 2 };
let opt: { a: number; b?: number } | undefined = undefined;
opt ??= full;
console.log(String((opt.b ?? 0) + 1));
