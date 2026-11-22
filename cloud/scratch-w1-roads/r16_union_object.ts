function describe(value: WeakSet<object> | { size: number }): string {
	return 'size' in value ? `object ${value.size}` : `set ${value.has({})}`;
}
console.log(describe(new WeakSet<object>()));
console.log(describe({ size: 2 }));
