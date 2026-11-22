function kind(value: WeakSet<object> | string): string {
	return typeof value === 'string' ? value : `set ${value.has({})}`;
}
console.log(kind(new WeakSet<object>()));
console.log(kind('text'));
