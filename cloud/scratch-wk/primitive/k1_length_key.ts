interface Lengthy {
	readonly length: number;
}
const m = new WeakMap<Lengthy, number>();
const s = new WeakSet<Lengthy>();
console.log(`${m.has('abc')} ${m.get('abc')} ${m.delete('abc')} ${s.has('abc')} ${s.delete('abc')}`);
try {
	m.set('abc', 1);
	console.log('map set did not throw');
} catch (error) {
	console.log(`${(error as Error).name}: ${(error as Error).message}`);
}
try {
	s.add('abc');
	console.log('set add did not throw');
} catch (error) {
	console.log(`${(error as Error).name}: ${(error as Error).message}`);
}
