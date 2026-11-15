// Narrowed and still present: a plain read and a push, no ?., agreeing with Node.
const r: Record<string, number[]> = {};
r["k"] = [];
r["k"].push(1);
console.log(String(r["k"].length));
