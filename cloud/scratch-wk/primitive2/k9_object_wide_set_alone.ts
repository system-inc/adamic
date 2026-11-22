const m = new WeakMap<object, number>();
function keep(key: object | string): void {
	if (typeof key === 'string') {
		console.log('string key');
		return;
	}
	m.set(key, 1);
}
keep('abc');
keep({ name: 'k' });
console.log('done');
