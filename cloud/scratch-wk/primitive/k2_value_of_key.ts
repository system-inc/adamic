interface Numeric {
	valueOf(): number;
}
const m = new WeakMap<Numeric, string>();
try {
	m.set(7, 'seven');
	console.log('map set did not throw');
} catch (error) {
	console.log(`${(error as Error).name}: ${(error as Error).message}`);
}
