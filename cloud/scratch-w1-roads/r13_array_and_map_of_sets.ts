const all = [new WeakSet<object>(), new WeakSet<object>()];
const key = {};
all[0].add(key);
const named = new Map<string, WeakSet<object>>();
named.set('a', all[1]);
named.get('a')?.add(key);
console.log(`${all[0].has(key)} ${all[1].has(key)} ${named.get('a')?.has(key)}`);
