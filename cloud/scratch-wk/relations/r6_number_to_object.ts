function put(m: Map<object, object>, k: object): void {
	m.set(k, { x: 1 });
}
const m = new Map<object, number>();
const k = { name: 'k' };
put(m, k);
console.log(typeof m.get(k));
