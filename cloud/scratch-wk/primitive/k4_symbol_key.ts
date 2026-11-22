const m = new WeakMap<WeakKey, number>();
const local = Symbol('local');
m.set(local, 1);
console.log(`${m.has(local)}`);
try {
	m.set(Symbol.for('registered'), 2);
	console.log('registered symbol did not throw');
} catch (error) {
	console.log(`${(error as Error).name}: ${(error as Error).message}`);
}
