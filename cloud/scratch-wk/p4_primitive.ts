const m = new WeakMap<object, number>();
const s = new WeakSet<object>();
const number = 1 as unknown as object;
const text = 'x' as unknown as object;
console.log(m.has(number), m.get(number), m.delete(number), s.has(text), s.delete(text));
try {
	m.set(number, 1);
	console.log('map set did not throw');
} catch (error) {
	console.log((error as Error).name, (error as Error).message);
}
try {
	s.add(text);
	console.log('set add did not throw');
} catch (error) {
	console.log((error as Error).name, (error as Error).message);
}
