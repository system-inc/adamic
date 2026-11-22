interface Keyed {
	name: string;
}
const m = new WeakMap<Keyed, number>();
const a: Keyed = { name: 'a' };
const b: Keyed = { name: 'b' };
m.set(a, 1);
console.log(`${m.has(a)} ${m.has(b)}`);
console.log(`${m.get(a)} ${m.get(b)}`);
console.log(`${m.delete(a)} ${m.has(a)}`);
