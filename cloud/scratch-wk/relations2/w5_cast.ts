const m = new WeakMap<object, string>();
const k = { name: 'k' };
const w = m as unknown as WeakMap<object, object>;
w.set(k, { x: 1 });
console.log(typeof m.get(k));
