function put(m: WeakMap<object, object>, k: object): void {
	m.set(k, { x: 1 });
}
const m = new WeakMap<object, string>();
const k = { name: 'k' };
put(m, k);
console.log(typeof m.get(k));
