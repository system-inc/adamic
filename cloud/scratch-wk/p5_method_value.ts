const m = new WeakMap<object, number>();
const k = { name: 'k' };
m.set(k, 3);
const g = m.get;
console.log(typeof g);
console.log(`${g.call(m, k)}`);
