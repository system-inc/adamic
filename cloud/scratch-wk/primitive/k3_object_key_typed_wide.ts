function keep(m: WeakMap<object, number>, key: object | string): void {
	if (typeof key === 'string') {
		console.log('string key');
		return;
	}
	m.set(key, 1);
	console.log(`${m.has(key)}`);
}
const m = new WeakMap<object, number>();
keep(m, 'abc');
keep(m, { name: 'k' });
