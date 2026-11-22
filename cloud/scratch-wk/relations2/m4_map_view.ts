interface Store {
	set(k: object, v: unknown): unknown;
}
const m = new Map<object, string>();
const k = { name: 'k' };
const s: Store = m;
s.set(k, { x: 1 });
console.log(typeof m.get(k));
