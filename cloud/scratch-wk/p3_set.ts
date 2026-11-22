const s = new WeakSet<object>();
const a = { name: 'a' };
const b = { name: 'b' };
s.add(a);
console.log(s.has(a), s.has(b));
s.add(a);
console.log(s.delete(a), s.delete(a), s.has(a));
