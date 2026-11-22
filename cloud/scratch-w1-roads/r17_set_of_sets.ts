const outer = new WeakSet<WeakSet<object>>();
const inner = new WeakSet<object>();
outer.add(inner);
console.log(`${outer.has(inner)}`);
