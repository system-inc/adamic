function put(m: Map<object, unknown>, k: object): void {
	m.set(k, { x: 1 });
}
const m = new Map<object, string>();
const k = { name: 'k' };
put(m, k);
console.log(typeof m.get(k));
