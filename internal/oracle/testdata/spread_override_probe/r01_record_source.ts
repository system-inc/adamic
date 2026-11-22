const s: Record<string, number> = {};
const c = { ...s, k: 1 };
console.log(String(c.k));
