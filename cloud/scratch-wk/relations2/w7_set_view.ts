interface Adder {
	add(k: object): unknown;
}
const s = new WeakSet<object>();
const k = { name: 'k' };
const a: Adder = s;
a.add(k);
console.log(`${s.has(k)}`);
