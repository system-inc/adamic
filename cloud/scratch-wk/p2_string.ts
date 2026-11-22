const m = new WeakMap<object, string>();
const a = { name: 'a' };
const b = [1, 2];
m.set(a, 'first');
m.set(b, 'second');
console.log(`${m.get(a)} ${m.get(b)} ${m.has(a)}`);
m.set(a, 'third');
console.log(`${m.get(a)}`);
console.log(`${m.delete(b)} ${m.get(b)}`);
