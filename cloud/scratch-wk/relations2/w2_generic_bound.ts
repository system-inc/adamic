function wipe<M extends WeakMap<object, unknown>>(m: M, k: object): void {
	m.set(k, { x: 1 });
}
const m = new WeakMap<object, string>();
const k = { name: 'k' };
wipe(m, k);
console.log(typeof m.get(k));
