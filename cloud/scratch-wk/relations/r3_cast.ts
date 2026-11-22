const m = new Map<object, string>();
const k = { name: 'k' };
const w = m as unknown as Map<object, object>;
w.set(k, { x: 1 });
console.log(typeof m.get(k));
