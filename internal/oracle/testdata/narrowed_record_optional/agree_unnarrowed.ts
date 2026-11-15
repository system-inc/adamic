// A key never narrowed: its read is number[] | undefined, a maybe the ?. tests, no Defined.
const r: Record<string, number[]> = {};
try { console.log(String(r["k"]?.includes(1))); } catch (error) { console.log("caught"); }
console.log(String(r["k"]?.length));
console.log("after");
