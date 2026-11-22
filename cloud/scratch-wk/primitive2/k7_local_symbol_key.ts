const m = new WeakMap<WeakKey, number>();
const local: symbol = Symbol('local');
m.set(local, 1);
console.log(`${m.has(local)} ${m.get(local)}`);
